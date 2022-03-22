// Package pubapi implements the API that a peer exposes publicly.
package pubapi

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
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
	pubproto.UnimplementedAPIServer
}

// New creates a new API.
func New(logger *zap.Logger, core Core, dialer Dialer, vpnAPIServer VPNAPIServer, validator atls.Validator, getPublicIPAddr GetIPAddrFunc) *API {
	return &API{
		logger:          logger,
		core:            core,
		dialer:          dialer,
		vpnAPIServer:    vpnAPIServer,
		validator:       validator,
		getPublicIPAddr: getPublicIPAddr,
		stopUpdate:      make(chan struct{}, 1),
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
