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
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/certificate"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/diskencryption"
	"github.com/edgelesssys/constellation/v2/internal/attestation"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/nodestate"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
	"github.com/spf13/afero"
	"go.uber.org/zap"
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
	cleaner     cleaner
	metadataAPI MetadataAPI

	log *logger.Logger

	mux      sync.Mutex
	stopC    chan struct{}
	stopDone chan struct{}
}

// New creates a new JoinClient.
func New(lock locker, dial grpcDialer, joiner ClusterJoiner, meta MetadataAPI, log *logger.Logger) *JoinClient {
	return &JoinClient{
		nodeLock:    lock,
		disk:        diskencryption.New(),
		fileHandler: file.NewHandler(afero.NewOsFs()),
		timeout:     timeout,
		joinTimeout: joinTimeout,
		interval:    interval,
		clock:       clock.RealClock{},
		dialer:      dial,
		joiner:      joiner,
		metadataAPI: meta,
		log:         log.Named("join-client"),
	}
}

// Start starts the client routine. The client will make the needed API calls to join
// the cluster with the role it receives from the metadata API.
// After receiving the needed information, the node will join the cluster.
// Multiple calls of start on the same client won't start a second routine if there is
// already a routine running.
func (c *JoinClient) Start(cleaner cleaner) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.stopC != nil { // daemon already running
		return
	}

	c.log.Infof("Starting")
	c.stopC = make(chan struct{}, 1)
	c.stopDone = make(chan struct{}, 1)
	c.cleaner = cleaner

	ticker := c.clock.NewTicker(c.interval)
	go func() {
		defer ticker.Stop()
		defer func() { c.stopDone <- struct{}{} }()
		defer c.log.Infof("Client stopped")

		diskUUID, err := c.getDiskUUID()
		if err != nil {
			c.log.With(zap.Error(err)).Errorf("Failed to get disk UUID")
			return
		}
		c.diskUUID = diskUUID

		for {
			err := c.getNodeMetadata()
			if err == nil {
				c.log.With(zap.String("role", c.role.String()), zap.String("name", c.nodeName)).Infof("Received own instance metadata")
				break
			}
			c.log.With(zap.Error(err)).Errorf("Failed to retrieve instance metadata")

			c.log.With(zap.Duration("interval", c.interval)).Infof("Sleeping")
			select {
			case <-c.stopC:
				return
			case <-ticker.C():
			}
		}

		for {
			err := c.tryJoinWithAvailableServices()
			if err == nil {
				c.log.Infof("Joined successfully. Client is shutting down")
				return
			} else if isUnrecoverable(err) {
				c.log.With(zap.Error(err)).Errorf("Unrecoverable error occurred")
				return
			}
			c.log.With(zap.Error(err)).Warnf("Join failed for all available endpoints")

			c.log.With(zap.Duration("interval", c.interval)).Infof("Sleeping")
			select {
			case <-c.stopC:
				return
			case <-ticker.C():
			}
		}
	}()
}

// Stop stops the client and blocks until the client's routine is stopped.
func (c *JoinClient) Stop() {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.stopC == nil { // daemon not running
		return
	}

	c.log.Infof("Stopping")

	c.stopC <- struct{}{}
	<-c.stopDone

	c.stopC = nil
	c.stopDone = nil

	c.log.Infof("Stopped")
}

func (c *JoinClient) tryJoinWithAvailableServices() error {
	ips, err := c.getControlPlaneIPs()
	if err != nil {
		return fmt.Errorf("failed to get control plane IPs: %w", err)
	}

	ip, _, err := c.metadataAPI.GetLoadBalancerEndpoint(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to get load balancer endpoint: %w", err)
	}
	ips = append(ips, ip)

	if len(ips) == 0 {
		return errors.New("no control plane IPs found")
	}

	for _, ip := range ips {
		err = c.join(net.JoinHostPort(ip, strconv.Itoa(constants.JoinServiceNodePort)))
		if err == nil {
			return nil
		}
		if isUnrecoverable(err) {
			return err
		}
	}

	return err
}

func (c *JoinClient) join(serviceEndpoint string) error {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	certificateRequest, kubeletKey, err := certificate.GetKubeletCertificateRequest(c.nodeName, c.validIPs)
	if err != nil {
		return err
	}

	conn, err := c.dialer.Dial(ctx, serviceEndpoint)
	if err != nil {
		c.log.With(zap.String("endpoint", serviceEndpoint), zap.Error(err)).Errorf("Join service unreachable")
		return fmt.Errorf("dialing join service endpoint: %w", err)
	}
	defer conn.Close()

	protoClient := joinproto.NewAPIClient(conn)
	req := &joinproto.IssueJoinTicketRequest{
		DiskUuid:           c.diskUUID,
		CertificateRequest: certificateRequest,
		IsControlPlane:     c.role == role.ControlPlane,
	}
	ticket, err := protoClient.IssueJoinTicket(ctx, req)
	if err != nil {
		c.log.With(zap.String("endpoint", serviceEndpoint), zap.Error(err)).Errorf("Issuing join ticket failed")
		return fmt.Errorf("issuing join ticket: %w", err)
	}

	return c.startNodeAndJoin(ticket, kubeletKey)
}

func (c *JoinClient) startNodeAndJoin(ticket *joinproto.IssueJoinTicketResponse, kubeletKey []byte) (retErr error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.joinTimeout)
	defer cancel()

	// If an error occurs in this func, the client cannot continue.
	defer func() {
		if retErr != nil {
			retErr = unrecoverableError{retErr}
		}
	}()

	clusterID, err := attestation.DeriveClusterID(ticket.MeasurementSecret, ticket.MeasurementSalt)
	if err != nil {
		return err
	}

	nodeLockAcquired, err := c.nodeLock.TryLockOnce(clusterID)
	if err != nil {
		c.log.With(zap.Error(err)).Errorf("Acquiring node lock failed")
		return fmt.Errorf("acquiring node lock: %w", err)
	}
	if !nodeLockAcquired {
		// There is already a cluster initialization in progress on
		// this node, so there is no need to also join the cluster,
		// as the initializing node is automatically part of the cluster.
		return errors.New("node is already being initialized")
	}

	c.cleaner.Clean()

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
	k8sComponents := components.NewComponentsFromJoinProto(ticket.KubernetesComponents)

	if err := c.joiner.JoinCluster(ctx, btd, c.role, k8sComponents, c.log); err != nil {
		return fmt.Errorf("joining Kubernetes cluster: %w", err)
	}

	return nil
}

func (c *JoinClient) getNodeMetadata() error {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	c.log.Debugf("Requesting node metadata from metadata API")
	inst, err := c.metadataAPI.Self(ctx)
	if err != nil {
		return err
	}
	c.log.With(zap.Any("instance", inst)).Debugf("Received node metadata")

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

func (c *JoinClient) getControlPlaneIPs() ([]string, error) {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	instances, err := c.metadataAPI.List(ctx)
	if err != nil {
		c.log.With(zap.Error(err)).Errorf("Failed to list instances from metadata API")
		return nil, fmt.Errorf("listing instances from metadata API: %w", err)
	}

	ips := []string{}
	for _, instance := range instances {
		if instance.Role == role.ControlPlane && instance.VPCIP != "" {
			ips = append(ips, instance.VPCIP)
		}
	}

	c.log.With(zap.Strings("IPs", ips)).Infof("Received control plane endpoints")
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

type unrecoverableError struct{ error }

func isUnrecoverable(err error) bool {
	_, ok := err.(unrecoverableError)
	return ok
}

type grpcDialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
}

// ClusterJoiner has the ability to join a new node to an existing cluster.
type ClusterJoiner interface {
	// JoinCluster joins a new node to an existing cluster.
	JoinCluster(
		ctx context.Context,
		args *kubeadm.BootstrapTokenDiscovery,
		peerRole role.Role,
		k8sComponents components.Components,
		log *logger.Logger,
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
