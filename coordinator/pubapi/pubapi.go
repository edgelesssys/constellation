// Package pubapi implements the API that a peer exposes publicly.
package pubapi

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/state/setup"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

const (
	deadlineDuration = time.Minute
	endpointAVPNPort = "9000"
	vpnAPIPort       = "9027"
	updateInterval   = 10 * time.Second
)

// API is the API.
type API struct {
	mut             sync.Mutex
	logger          *zap.Logger
	core            Core
	dialer          Dialer
	vpnAPIServer    VPNAPIServer
	validator       atls.Validator
	getPublicIPAddr GetIPAddrFunc
	stopUpdate      chan struct{}
	wgClose         sync.WaitGroup
	resourceVersion int
	peerFromContext PeerFromContextFunc
	pubproto.UnimplementedAPIServer
}

// New creates a new API.
func New(logger *zap.Logger, core Core, dialer Dialer, vpnAPIServer VPNAPIServer, validator atls.Validator, getPublicIPAddr GetIPAddrFunc, peerFromContext PeerFromContextFunc) *API {
	return &API{
		logger:          logger,
		core:            core,
		dialer:          dialer,
		vpnAPIServer:    vpnAPIServer,
		validator:       validator,
		getPublicIPAddr: getPublicIPAddr,
		stopUpdate:      make(chan struct{}, 1),
		peerFromContext: peerFromContext,
	}
}

// GetState is the RPC call to get the peer's state.
func (a *API) GetState(ctx context.Context, in *pubproto.GetStateRequest) (*pubproto.GetStateResponse, error) {
	return &pubproto.GetStateResponse{State: uint32(a.core.GetState())}, nil
}

// Close closes the API.
func (a *API) Close() {
	a.stopUpdate <- struct{}{}
	if a.vpnAPIServer != nil {
		a.vpnAPIServer.Close()
	}
	a.wgClose.Wait()
}

func (a *API) dial(ctx context.Context, target string) (*grpc.ClientConn, error) {
	tlsConfig, err := atls.CreateAttestationClientTLSConfig([]atls.Validator{a.validator})
	if err != nil {
		return nil, err
	}

	return grpc.DialContext(ctx, target,
		a.grpcWithDialer(),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
}

func (a *API) dialInsecure(ctx context.Context, target string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target,
		a.grpcWithDialer(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

func (a *API) dialNoVerify(ctx context.Context, target string) (*grpc.ClientConn, error) {
	tlsConfig, err := atls.CreateUnverifiedClientTLSConfig()
	if err != nil {
		return nil, err
	}

	return grpc.DialContext(ctx, target,
		a.grpcWithDialer(),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
}

func (a *API) grpcWithDialer() grpc.DialOption {
	return grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return a.dialer.DialContext(ctx, "tcp", addr)
	})
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
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
