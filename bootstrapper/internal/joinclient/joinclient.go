/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
# JoinClient

The JoinClient is one of the two main components of the bootstrapper.
It is responsible for for the initial setup of a node, and joining an existing Kubernetes cluster.

The JoinClient is started on each node, it then continuously checks for an existing cluster to join,
or for the InitServer to bootstrap a new cluster.

If the JoinClient finds an existing cluster, it will attempt to join it as either a control-plane or a worker node.
*/
package joinclient

import (
	"context"
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/addresses"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/certificate"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/nodestate"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	kubeconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/utils/clock"
)

const (
	interval    = 30 * time.Second
	timeout     = 30 * time.Second
	joinTimeout = 5 * time.Minute
)

// JoinClient is a client for requesting the needed information and
// joining an existing Kubernetes cluster.
type JoinClient struct {
	nodeLock    locker
	diskUUID    string
	nodeName    string
	role        role.Role
	validIPs    []net.IP
	disk        encryptedDisk
	fileHandler file.Handler

	timeout     time.Duration
	joinTimeout time.Duration
	interval    time.Duration
	clock       clock.WithTicker

	dialer      grpcDialer
	joiner      ClusterJoiner
	metadataAPI MetadataAPI

	log *slog.Logger

	stopC    chan struct{}
	stopDone chan struct{}
}

// New creates a new JoinClient.
func New(lock locker, dial grpcDialer, joiner ClusterJoiner, meta MetadataAPI, disk encryptedDisk, log *slog.Logger) *JoinClient {
	return &JoinClient{
		nodeLock:    lock,
		disk:        disk,
		fileHandler: file.NewHandler(afero.NewOsFs()),
		timeout:     timeout,
		joinTimeout: joinTimeout,
		interval:    interval,
		clock:       clock.RealClock{},
		dialer:      dial,
		joiner:      joiner,
		metadataAPI: meta,
		log:         log.WithGroup("join-client"),

		stopC:    make(chan struct{}, 1),
		stopDone: make(chan struct{}, 1),
	}
}

// Start starts the client routine. The client will make the needed API calls to join
// the cluster with the role it receives from the metadata API.
// After receiving the needed information, the node will join the cluster.
func (c *JoinClient) Start(cleaner cleaner) error {
	c.log.Info("Starting")
	ticker := c.clock.NewTicker(c.interval)
	defer ticker.Stop()
	defer func() { c.stopDone <- struct{}{} }()
	defer c.log.Info("Client stopped")

	diskUUID, err := c.getDiskUUID()
	if err != nil {
		c.log.With(slog.Any("error", err)).Error("Failed to get disk UUID")
		return err
	}
	c.diskUUID = diskUUID

	for {
		err := c.getNodeMetadata()
		if err == nil {
			c.log.With(slog.String("role", c.role.String()), slog.String("name", c.nodeName)).Info("Received own instance metadata")
			break
		}
		c.log.With(slog.Any("error", err)).Error("Failed to retrieve instance metadata")

		c.log.With(slog.Duration("interval", c.interval)).Info("Sleeping")
		select {
		case <-c.stopC:
			return nil
		case <-ticker.C():
		}
	}

	var ticket *joinproto.IssueJoinTicketResponse
	var kubeletKey []byte

	for {
		ticket, kubeletKey, err = c.tryJoinWithAvailableServices()
		if err == nil {
			c.log.Info("Successfully retrieved join ticket, starting Kubernetes node")
			break
		}
		c.log.With(slog.Any("error", err)).Warn("Join failed for all available endpoints")

		c.log.With(slog.Duration("interval", c.interval)).Info("Sleeping")
		select {
		case <-c.stopC:
			return nil
		case <-ticker.C():
		}
	}

	if err := c.startNodeAndJoin(ticket, kubeletKey, cleaner); err != nil {
		c.log.With(slog.Any("error", err)).Error("Failed to start node and join cluster")
		return err
	}

	return nil
}

// Stop stops the client and blocks until the client's routine is stopped.
func (c *JoinClient) Stop() {
	c.log.Info("Stopping")

	c.stopC <- struct{}{}
	<-c.stopDone

	c.log.Info("Stopped")
}

func (c *JoinClient) tryJoinWithAvailableServices() (ticket *joinproto.IssueJoinTicketResponse, kubeletKey []byte, err error) {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	var endpoints []string

	endpoint, _, err := c.metadataAPI.GetLoadBalancerEndpoint(ctx)
	if err != nil {
		c.log.Warn("Failed to get load balancer endpoint", "err", err)
	}
	endpoints = append(endpoints, endpoint)

	ips, err := c.getControlPlaneIPs(ctx)
	if err != nil {
		c.log.Warn("Failed to get control plane IPs", "err", err)
	}
	endpoints = append(endpoints, ips...)

	if len(endpoints) == 0 {
		return nil, nil, errors.New("no control plane IPs found")
	}

	var joinErrs error
	for _, endpoint := range endpoints {
		ticket, kubeletKey, err := c.requestJoinTicket(net.JoinHostPort(endpoint, strconv.Itoa(constants.JoinServiceNodePort)))
		if err == nil {
			return ticket, kubeletKey, nil
		}

		joinErrs = errors.Join(joinErrs, err)
	}

	return nil, nil, fmt.Errorf("trying to join on all endpoints %v: %w", endpoints, joinErrs)
}

func (c *JoinClient) requestJoinTicket(serviceEndpoint string) (ticket *joinproto.IssueJoinTicketResponse, kubeletKey []byte, err error) {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	certificateRequest, kubeletKey, err := certificate.GetKubeletCertificateRequest(c.nodeName, c.validIPs)
	if err != nil {
		return nil, nil, err
	}

	principalList, err := addresses.GetMachineNetworkAddresses()
	if err != nil {
		c.log.With(slog.Any("error", err)).Error("Failed to get network interfaces")
		return nil, nil, err
	}
	hostname, err := os.Hostname()
	if err != nil {
		c.log.With(slog.Any("error", err)).Error("Failed to get hostname")
		return nil, nil, err
	}
	principalList = append(principalList, hostname)

	var hostKeyPubSSH ssh.PublicKey

	if _, err := c.fileHandler.Stat(constants.SSHHostKeyPath); errors.Is(err, os.ErrNotExist) {
		hostKeyPub, hostKey, err := ed25519.GenerateKey(nil)
		if err != nil {
			c.log.With(slog.Any("error", err)).Error("Failed to generate SSH host key")
			return nil, nil, err
		}

		hostKeyPubSSH, err = ssh.NewPublicKey(hostKeyPub)
		if err != nil {
			c.log.With(slog.Any("error", err)).Error("Failed to convert ed25519 pubkey to ssh pubkey")
			return nil, nil, err
		}

		pemHostKey, err := ssh.MarshalPrivateKey(hostKey, "")
		if err != nil {
			c.log.With(slog.Any("error", err)).Error("Failed to format SSH host key")
			return nil, nil, err
		}
		if err := c.fileHandler.Write(constants.SSHHostKeyPath, pem.EncodeToMemory(pemHostKey), file.OptMkdirAll); err != nil {
			c.log.With(slog.Any("error", err)).Error("Failed to write SSH host key")
			return nil, nil, err
		}
	} else {
		hostKeyData, err := c.fileHandler.Read(constants.SSHHostKeyPath)
		if err != nil {
			c.log.With(slog.Any("error", err)).Error("Failed to read SSH host key file")
			return nil, nil, err
		}

		hostKey, err := ssh.ParsePrivateKey(hostKeyData)
		if err != nil {
			c.log.With(slog.Any("error", err)).Error("Failed to parse SSH host key file")
			return nil, nil, err
		}
		hostKeyPubSSH = hostKey.PublicKey()
	}

	conn, err := c.dialer.Dial(serviceEndpoint)
	if err != nil {
		c.log.With(slog.String("endpoint", serviceEndpoint), slog.Any("error", err)).Error("Join service unreachable")
		return nil, nil, fmt.Errorf("dialing join service endpoint: %w", err)
	}
	defer conn.Close()

	protoClient := joinproto.NewAPIClient(conn)
	req := &joinproto.IssueJoinTicketRequest{
		DiskUuid:                  c.diskUUID,
		CertificateRequest:        certificateRequest,
		IsControlPlane:            c.role == role.ControlPlane,
		HostPublicKey:             hostKeyPubSSH.Marshal(),
		HostCertificatePrincipals: principalList,
	}
	ticket, err = protoClient.IssueJoinTicket(ctx, req)
	if err != nil {
		c.log.With(slog.String("endpoint", serviceEndpoint), slog.Any("error", err)).Error("Issuing join ticket failed")
		return nil, nil, fmt.Errorf("issuing join ticket: %w", err)
	}

	return ticket, kubeletKey, err
}

func (c *JoinClient) startNodeAndJoin(ticket *joinproto.IssueJoinTicketResponse, kubeletKey []byte, cleaner cleaner) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.joinTimeout)
	defer cancel()

	clusterID, err := attestation.DeriveClusterID(ticket.MeasurementSecret, ticket.MeasurementSalt)
	if err != nil {
		return err
	}

	nodeLockAcquired, err := c.nodeLock.TryLockOnce(clusterID)
	if err != nil {
		c.log.With(slog.Any("error", err)).Error("Acquiring node lock failed")
		return fmt.Errorf("acquiring node lock: %w", err)
	}
	if !nodeLockAcquired {
		// There is already a cluster initialization in progress on
		// this node, so there is no need to also join the cluster,
		// as the initializing node is automatically part of the cluster.
		c.log.Info("Node is already being initialized. Aborting join process.")
		return nil
	}

	cleaner.Clean()

	if err := c.updateDiskPassphrase(string(ticket.StateDiskKey)); err != nil {
		return fmt.Errorf("updating disk passphrase: %w", err)
	}

	if c.role == role.ControlPlane {
		if err := c.writeControlPlaneFiles(ticket.ControlPlaneFiles); err != nil {
			return fmt.Errorf("writing control plane files: %w", err)
		}
	}
	if err := c.fileHandler.Write(certificate.CertificateFilename, ticket.KubeletCert, file.OptMkdirAll); err != nil {
		return fmt.Errorf("writing kubelet certificate: %w", err)
	}
	if err := c.fileHandler.Write(certificate.KeyFilename, kubeletKey, file.OptMkdirAll); err != nil {
		return fmt.Errorf("writing kubelet key: %w", err)
	}

	if err := c.fileHandler.Write(constants.SSHCAKeyPath, ticket.AuthorizedCaPublicKey, file.OptMkdirAll); err != nil {
		return fmt.Errorf("writing ssh ca key: %w", err)
	}

	if err := c.fileHandler.Write(constants.SSHHostCertificatePath, ticket.HostCertificate, file.OptMkdirAll); err != nil {
		return fmt.Errorf("writing ssh host certificate: %w", err)
	}

	state := nodestate.NodeState{
		Role:            c.role,
		MeasurementSalt: ticket.MeasurementSalt,
	}
	if err := state.ToFile(c.fileHandler); err != nil {
		return fmt.Errorf("persisting node state: %w", err)
	}

	btd := &kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: ticket.ApiServerEndpoint,
		Token:             ticket.Token,
		CACertHashes:      []string{ticket.DiscoveryTokenCaCertHash},
	}

	// We currently cannot recover from any failure in this function. Joining the k8s cluster
	// sometimes fails transiently, and we don't want to brick the node because of that.
	for i := range 3 {
		err = c.joiner.JoinCluster(ctx, btd, c.role, ticket.KubernetesComponents)
		if err == nil {
			break
		}
		c.log.Error("failed to join k8s cluster", "role", c.role, "attempt", i, "error", err)
	}
	if err != nil {
		return fmt.Errorf("joining Kubernetes cluster: %w", err)
	}

	return nil
}

func (c *JoinClient) getNodeMetadata() error {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	c.log.Debug("Requesting node metadata from metadata API")
	inst, err := c.metadataAPI.Self(ctx)
	if err != nil {
		return err
	}
	c.log.With(slog.Any("instance", inst)).Debug("Received node metadata")

	if inst.Name == "" {
		return errors.New("got instance metadata with empty name")
	}

	if inst.Role == role.Unknown {
		return errors.New("got instance metadata with unknown role")
	}

	var ips []net.IP

	if inst.VPCIP != "" {
		ips = append(ips, net.ParseIP(inst.VPCIP))
	}

	c.nodeName = inst.Name
	c.role = inst.Role
	c.validIPs = ips

	return nil
}

func (c *JoinClient) updateDiskPassphrase(passphrase string) error {
	free, err := c.disk.Open()
	if err != nil {
		return fmt.Errorf("opening disk: %w", err)
	}
	defer free()
	return c.disk.UpdatePassphrase(passphrase)
}

func (c *JoinClient) getDiskUUID() (string, error) {
	free, err := c.disk.Open()
	if err != nil {
		return "", fmt.Errorf("opening disk: %w", err)
	}
	defer free()
	return c.disk.UUID()
}

func (c *JoinClient) getControlPlaneIPs(ctx context.Context) ([]string, error) {
	instances, err := c.metadataAPI.List(ctx)
	if err != nil {
		c.log.With(slog.Any("error", err)).Error("Failed to list instances from metadata API")
		return nil, fmt.Errorf("listing instances from metadata API: %w", err)
	}

	ips := []string{}
	for _, instance := range instances {
		if instance.Role == role.ControlPlane && instance.VPCIP != "" {
			ips = append(ips, instance.VPCIP)
		}
	}

	c.log.With(slog.Any("IPs", ips)).Info("Received control plane endpoints")
	return ips, nil
}

func (c *JoinClient) writeControlPlaneFiles(files []*joinproto.ControlPlaneCertOrKey) error {
	for _, cert := range files {
		if err := c.fileHandler.Write(
			filepath.Join(kubeconstants.KubernetesDir, kubeconstants.DefaultCertificateDir, cert.Name),
			cert.Data,
			file.OptMkdirAll,
		); err != nil {
			return fmt.Errorf("writing control plane files: %w", err)
		}
	}

	return nil
}

func (c *JoinClient) timeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

type grpcDialer interface {
	Dial(target string) (*grpc.ClientConn, error)
}

// ClusterJoiner has the ability to join a new node to an existing cluster.
type ClusterJoiner interface {
	// JoinCluster joins a new node to an existing cluster.
	JoinCluster(
		ctx context.Context,
		args *kubeadm.BootstrapTokenDiscovery,
		peerRole role.Role,
		k8sComponents components.Components,
	) error
}

// MetadataAPI provides information about the instances.
type MetadataAPI interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]metadata.InstanceMetadata, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
	// GetLoadBalancerEndpoint retrieves the load balancer endpoint.
	GetLoadBalancerEndpoint(ctx context.Context) (host, port string, err error)
}

type encryptedDisk interface {
	// Open prepares the underlying device for disk operations.
	Open() (func(), error)
	// UUID gets the device's UUID.
	UUID() (string, error)
	// UpdatePassphrase switches the initial random passphrase of the encrypted disk to a permanent passphrase.
	UpdatePassphrase(passphrase string) error
}

type cleaner interface {
	Clean()
}

type locker interface {
	// TryLockOnce tries to lock the node. If the node is already locked, it
	// returns false. If the node is unlocked, it locks it and returns true.
	TryLockOnce(clusterID []byte) (bool, error)
}
