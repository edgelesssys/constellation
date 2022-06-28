package joinclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/activation/activationproto"
	"github.com/edgelesssys/constellation/coordinator/internal/diskencryption"
	"github.com/edgelesssys/constellation/coordinator/nodestate"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"k8s.io/utils/clock"
)

const (
	interval = 30 * time.Second
	timeout  = 30 * time.Second
)

// JoinClient is a client for self-activation of node.
type JoinClient struct {
	nodeLock    *sync.Mutex
	diskUUID    string
	nodeName    string
	role        role.Role
	disk        EncryptedDisk
	fileHandler file.Handler

	timeout  time.Duration
	interval time.Duration
	clock    clock.WithTicker

	dialer      grpcDialer
	joiner      ClusterJoiner
	metadataAPI metadataAPI

	log *zap.Logger

	mux      sync.Mutex
	stopC    chan struct{}
	stopDone chan struct{}
}

// New creates a new SelfActivationClient.
func New(nodeLock *sync.Mutex, dial grpcDialer, joiner ClusterJoiner, meta metadataAPI, log *zap.Logger) *JoinClient {
	return &JoinClient{
		disk:        diskencryption.New(),
		fileHandler: file.NewHandler(afero.NewOsFs()),
		timeout:     timeout,
		interval:    interval,
		clock:       clock.RealClock{},
		dialer:      dial,
		joiner:      joiner,
		metadataAPI: meta,
		log:         log.Named("selfactivation-client"),
	}
}

// Start starts the client routine. The client will make the needed API calls to activate
// the node as the role it receives from the metadata API.
// Multiple calls of start on the same client won't start a second routine if there is
// already a routine running.
func (c *JoinClient) Start() {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.stopC != nil { // daemon already running
		return
	}

	c.log.Info("Starting")
	c.stopC = make(chan struct{}, 1)
	c.stopDone = make(chan struct{}, 1)

	ticker := c.clock.NewTicker(c.interval)
	go func() {
		defer ticker.Stop()
		defer func() { c.stopDone <- struct{}{} }()

		diskUUID, err := c.GetDiskUUID()
		if err != nil {
			c.log.Error("Failed to get disk UUID", zap.Error(err))
			c.log.Error("Stopping self-activation client")
			return
		}
		c.diskUUID = diskUUID

		for {
			err := c.getNodeMetadata()
			if err == nil {
				c.log.Info("Received own instance metadata", zap.String("role", c.role.String()), zap.String("name", c.nodeName))
				break
			}
			c.log.Info("Failed to retrieve instance metadata", zap.Error(err))

			c.log.Info("Sleeping", zap.Duration("interval", c.interval))
			select {
			case <-c.stopC:
				return
			case <-ticker.C():
			}
		}

		for {
			err := c.tryJoinAtAvailableServices()
			if err == nil {
				c.log.Info("Activated successfully. SelfActivationClient shut down.")
				return
			}
			c.log.Info("Activation failed for all available endpoints", zap.Error(err))

			c.log.Info("Sleeping", zap.Duration("interval", c.interval))
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

	c.log.Info("Stopping")

	c.stopC <- struct{}{}
	<-c.stopDone

	c.stopC = nil
	c.stopDone = nil

	c.log.Info("Stopped")
}

func (c *JoinClient) tryJoinAtAvailableServices() error {
	ips, err := c.getCoordinatorIPs()
	if err != nil {
		return err
	}

	if len(ips) == 0 {
		return errors.New("no coordinator IPs found")
	}

	for _, ip := range ips {
		err = c.join(net.JoinHostPort(ip, strconv.Itoa(constants.ActivationServiceNodePort)))
		if err == nil {
			return nil
		}
	}

	return err
}

func (c *JoinClient) join(serviceEndpoint string) error {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	conn, err := c.dialer.Dial(ctx, serviceEndpoint)
	if err != nil {
		c.log.Info("join service unreachable", zap.String("endpoint", serviceEndpoint), zap.Error(err))
		return fmt.Errorf("dialing join service endpoint: %v", err)
	}
	defer conn.Close()

	protoClient := activationproto.NewAPIClient(conn)

	switch c.role {
	case role.Node:
		return c.joinAsWorkerNode(ctx, protoClient)
	case role.Coordinator:
		return c.joinAsControlPlaneNode(ctx, protoClient)
	default:
		return fmt.Errorf("cannot activate as %s", role.Unknown)
	}
}

func (c *JoinClient) joinAsWorkerNode(ctx context.Context, client activationproto.APIClient) error {
	req := &activationproto.ActivateWorkerNodeRequest{
		DiskUuid: c.diskUUID,
		NodeName: c.nodeName,
	}
	resp, err := client.ActivateWorkerNode(ctx, req)
	if err != nil {
		c.log.Info("Failed to activate as worker node", zap.Error(err))
		return fmt.Errorf("activating worker node: %w", err)
	}
	c.log.Info("Activation at AaaS succeeded")

	return c.startNodeAndJoin(
		ctx,
		resp.StateDiskKey,
		resp.OwnerId,
		resp.ClusterId,
		resp.KubeletKey,
		resp.KubeletCert,
		resp.ApiServerEndpoint,
		resp.Token,
		resp.DiscoveryTokenCaCertHash,
		"",
	)
}

func (c *JoinClient) joinAsControlPlaneNode(ctx context.Context, client activationproto.APIClient) error {
	req := &activationproto.ActivateControlPlaneNodeRequest{
		DiskUuid: c.diskUUID,
		NodeName: c.nodeName,
	}
	resp, err := client.ActivateControlPlaneNode(ctx, req)
	if err != nil {
		c.log.Info("Failed to activate as control plane node", zap.Error(err))
		return fmt.Errorf("activating control plane node: %w", err)
	}
	c.log.Info("Activation at AaaS succeeded")

	return c.startNodeAndJoin(
		ctx,
		resp.StateDiskKey,
		resp.OwnerId,
		resp.ClusterId,
		resp.KubeletKey,
		resp.KubeletCert,
		resp.ApiServerEndpoint,
		resp.Token,
		resp.DiscoveryTokenCaCertHash,
		resp.CertificateKey,
	)
}

func (c *JoinClient) startNodeAndJoin(ctx context.Context, diskKey, ownerID, clusterID, kubeletKey, kubeletCert []byte, endpoint, token,
	discoveryCACertHash, certKey string,
) error {
	if ok := c.nodeLock.TryLock(); !ok {
		return errors.New("node is already being initialized")
	}

	if err := c.updateDiskPassphrase(string(diskKey)); err != nil {
		return fmt.Errorf("updating disk passphrase: %w", err)
	}

	state := nodestate.NodeState{
		Role:      c.role,
		OwnerID:   ownerID,
		ClusterID: clusterID,
	}
	if err := state.ToFile(c.fileHandler); err != nil {
		return fmt.Errorf("persisting node state: %w", err)
	}

	btd := &kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: endpoint,
		Token:             token,
		CACertHashes:      []string{discoveryCACertHash},
	}
	if err := c.joiner.JoinCluster(ctx, btd, certKey, c.role); err != nil {
		return fmt.Errorf("joining Kubernetes cluster: %w", err)
	}

	return nil
}

func (c *JoinClient) getNodeMetadata() error {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	c.log.Info("Requesting node metadata from metadata API")
	inst, err := c.metadataAPI.Self(ctx)
	if err != nil {
		return err
	}
	c.log.Info("Received node metadata", zap.Any("instance", inst))

	if inst.Name == "" {
		return errors.New("got instance metadata with empty name")
	}

	if inst.Role == role.Unknown {
		return errors.New("got instance metadata with unknown role")
	}

	c.nodeName = inst.Name
	c.role = inst.Role

	return nil
}

func (c *JoinClient) updateDiskPassphrase(passphrase string) error {
	if err := c.disk.Open(); err != nil {
		return fmt.Errorf("opening disk: %w", err)
	}
	defer c.disk.Close()
	return c.disk.UpdatePassphrase(passphrase)
}

func (c *JoinClient) GetDiskUUID() (string, error) {
	if err := c.disk.Open(); err != nil {
		return "", fmt.Errorf("opening disk: %w", err)
	}
	defer c.disk.Close()
	return c.disk.UUID()
}

func (c *JoinClient) getCoordinatorIPs() ([]string, error) {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	instances, err := c.metadataAPI.List(ctx)
	if err != nil {
		c.log.Error("Failed to list instances from metadata API", zap.Error(err))
		return nil, fmt.Errorf("listing instances from metadata API: %w", err)
	}

	ips := []string{}
	for _, instance := range instances {
		if instance.Role == role.Coordinator {
			ips = append(ips, instance.PrivateIPs...)
		}
	}

	c.log.Info("Received Coordinator endpoints", zap.Strings("IPs", ips))
	return ips, nil
}

func (c *JoinClient) timeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

type grpcDialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
}

type ClusterJoiner interface {
	JoinCluster(
		ctx context.Context,
		args *kubeadm.BootstrapTokenDiscovery,
		certKey string,
		peerRole role.Role,
	) error
}

type metadataAPI interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]metadata.InstanceMetadata, error)
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
}

// EncryptedDisk manages the encrypted state disk.
type EncryptedDisk interface {
	// Open prepares the underlying device for disk operations.
	Open() error
	// Close closes the underlying device.
	Close() error
	// UUID gets the device's UUID.
	UUID() (string, error)
	// UpdatePassphrase switches the initial random passphrase of the encrypted disk to a permanent passphrase.
	UpdatePassphrase(passphrase string) error
}
