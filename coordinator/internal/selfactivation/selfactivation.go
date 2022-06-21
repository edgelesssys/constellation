package selfactivation

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/activation/activationproto"
	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/constants"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"k8s.io/utils/clock"
)

const (
	interval = 30 * time.Second
	timeout  = 30 * time.Second
)

// SelfActivationClient is a client for self-activation of node.
type SelfActivationClient struct {
	diskUUID string
	role     role.Role

	timeout  time.Duration
	interval time.Duration
	clock    clock.WithTicker

	dialer      grpcDialer
	setterAPI   activeSetter
	metadataAPI metadataAPI

	log *zap.Logger

	mux      sync.Mutex
	stopC    chan struct{}
	stopDone chan struct{}
}

// NewClient creates a new SelfActivationClient.
func NewClient(diskUUID string, dial grpcDialer, setter activeSetter, meta metadataAPI, log *zap.Logger) *SelfActivationClient {
	return &SelfActivationClient{
		diskUUID:    diskUUID,
		timeout:     timeout,
		interval:    interval,
		clock:       clock.RealClock{},
		dialer:      dial,
		setterAPI:   setter,
		metadataAPI: meta,
		log:         log.Named("selfactivation-client"),
	}
}

// Start starts the client routine. The client will make the needed API calls to activate
// the node as the role it receives from the metadata API.
// Multiple calls of start on the same client won't start a second routine if there is
// already a routine running.
func (c *SelfActivationClient) Start() {
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

		for {
			c.role = c.getRole()
			if c.role != role.Unknown {
				break
			}

			c.log.Info("Sleeping", zap.Duration("interval", c.interval))
			select {
			case <-c.stopC:
				return
			case <-ticker.C():
			}
		}

		// TODO(katexochen): Delete when Coordinator self-activation is implemented.
		if c.role == role.Coordinator {
			c.log.Info("Role is Coordinator, terminating")
			return
		}

		for {
			err := c.tryActivationAtAvailableServices()
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
func (c *SelfActivationClient) Stop() {
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

func (c *SelfActivationClient) tryActivationAtAvailableServices() error {
	ips, err := c.getCoordinatorIPs()
	if err != nil {
		return err
	}

	if len(ips) == 0 {
		return errors.New("no coordinator IPs found")
	}

	for _, ip := range ips {
		err = c.activate(net.JoinHostPort(ip, strconv.Itoa(constants.ActivationServicePort)))
		if err == nil {
			return nil
		}
	}

	return err
}

func (c *SelfActivationClient) activate(aaasEndpoint string) error {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	conn, err := c.dialer.Dial(ctx, aaasEndpoint)
	if err != nil {
		c.log.Info("AaaS unreachable", zap.String("endpoint", aaasEndpoint), zap.Error(err))
		return fmt.Errorf("dialing AaaS endpoint: %v", err)
	}
	defer conn.Close()

	protoClient := activationproto.NewAPIClient(conn)

	switch c.role {
	case role.Node:
		return c.activateAsWorkerNode(ctx, protoClient)
	case role.Coordinator:
		return c.activateAsControlePlaneNode(ctx, protoClient)
	default:
		return fmt.Errorf("cannot activate as %s", role.Unknown)
	}
}

func (c *SelfActivationClient) activateAsWorkerNode(ctx context.Context, client activationproto.APIClient) error {
	req := &activationproto.ActivateWorkerNodeRequest{DiskUuid: c.diskUUID}
	resp, err := client.ActivateWorkerNode(ctx, req)
	if err != nil {
		c.log.Info("Failed to activate as node", zap.Error(err))
		return fmt.Errorf("activating node: %w", err)
	}
	c.log.Info("Activation at AaaS succeeded")

	return c.setterAPI.SetNodeActive(
		resp.StateDiskKey,
		resp.OwnerId,
		resp.ClusterId,
		resp.ApiServerEndpoint,
		resp.Token,
		resp.DiscoveryTokenCaCertHash,
	)
}

func (c *SelfActivationClient) activateAsControlePlaneNode(ctx context.Context, client activationproto.APIClient) error {
	panic("not implemented")
}

func (c *SelfActivationClient) getRole() role.Role {
	ctx, cancel := c.timeoutCtx()
	defer cancel()

	c.log.Info("Requesting role from metadata API")
	inst, err := c.metadataAPI.Self(ctx)
	if err != nil {
		c.log.Error("Failed to get self instance from metadata API", zap.Error(err))
		return role.Unknown
	}

	c.log.Info("Received new role", zap.String("role", inst.Role.String()))
	return inst.Role
}

func (c *SelfActivationClient) getCoordinatorIPs() ([]string, error) {
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

func (c *SelfActivationClient) timeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.timeout)
}

type grpcDialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
}

type activeSetter interface {
	SetNodeActive(diskKey, ownerID, clusterID []byte, endpoint, token, discoveryCACertHash string) error
	SetCoordinatorActive() error
}

type metadataAPI interface {
	Self(ctx context.Context) (cloudtypes.Instance, error)
	List(ctx context.Context) ([]cloudtypes.Instance, error)
}
