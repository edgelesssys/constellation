// Package pubapi implements the API that a peer exposes publicly.
package pubapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/coordinator/logging"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/state/setup"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const (
	deadlineDuration = 5 * time.Minute
	endpointAVPNPort = "9000"
	vpnAPIPort       = "9027"
	updateInterval   = 10 * time.Second
)

// API is the API.
type API struct {
	mut             sync.Mutex
	logger          *zap.Logger
	cloudLogger     logging.CloudLogger
	core            Core
	dialer          Dialer
	vpnAPIServer    VPNAPIServer
	getPublicIPAddr GetIPAddrFunc
	stopUpdate      chan struct{}
	wgClose         sync.WaitGroup
	resourceVersion int
	peerFromContext PeerFromContextFunc
	pubproto.UnimplementedAPIServer
}

// New creates a new API.
func New(logger *zap.Logger, cloudLogger logging.CloudLogger, core Core, dialer Dialer, vpnAPIServer VPNAPIServer, getPublicIPAddr GetIPAddrFunc, peerFromContext PeerFromContextFunc) *API {
	return &API{
		logger:          logger,
		cloudLogger:     cloudLogger,
		core:            core,
		dialer:          dialer,
		vpnAPIServer:    vpnAPIServer,
		getPublicIPAddr: getPublicIPAddr,
		stopUpdate:      make(chan struct{}, 1),
		peerFromContext: peerFromContext,
	}
}

// GetState is the RPC call to get the peer's state.
func (a *API) GetState(ctx context.Context, in *pubproto.GetStateRequest) (*pubproto.GetStateResponse, error) {
	return &pubproto.GetStateResponse{State: uint32(a.core.GetState())}, nil
}

// StartVPNAPIServer starts the VPN-API server.
func (a *API) StartVPNAPIServer(vpnIP string) error {
	if err := a.vpnAPIServer.Listen(net.JoinHostPort(vpnIP, vpnAPIPort)); err != nil {
		return fmt.Errorf("start vpnAPIServer: %v", err)
	}
	a.wgClose.Add(1)
	go func() {
		defer a.wgClose.Done()
		if err := a.vpnAPIServer.Serve(); err != nil {
			panic(err)
		}
	}()
	return nil
}

// Close closes the API.
func (a *API) Close() {
	a.stopUpdate <- struct{}{}
	if a.vpnAPIServer != nil {
		a.vpnAPIServer.Close()
	}
	a.wgClose.Wait()
}

type VPNAPIServer interface {
	Listen(endpoint string) error
	Serve() error
	Close()
}

type GetIPAddrFunc func() (string, error)

// PeerFromContextFunc returns a peer endpoint (IP:port) from a given context.
type PeerFromContextFunc func(context.Context) (string, error)

// GetRecoveryPeerFromContext returns the context's IP joined with the Coordinator's default port.
func GetRecoveryPeerFromContext(ctx context.Context) (string, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return "", errors.New("unable to get peer from context")
	}

	peerIP, _, err := net.SplitHostPort(peer.Addr.String())
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(peerIP, setup.RecoveryPort), nil
}

// Dialer can open grpc client connections with different levels of ATLS encryption / verification.
type Dialer interface {
	Dial(ctx context.Context, target string) (*grpc.ClientConn, error)
	DialInsecure(ctx context.Context, target string) (*grpc.ClientConn, error)
	DialNoVerify(ctx context.Context, target string) (*grpc.ClientConn, error)
}
