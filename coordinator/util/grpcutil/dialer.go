package grpcutil

import (
	"context"
	"net"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Dialer can open grpc client connections with different levels of ATLS encryption / verification.
type Dialer struct {
	validator atls.Validator
	netDialer NetDialer
}

// NewDialer creates a new Dialer.
func NewDialer(validator atls.Validator, netDialer NetDialer) *Dialer {
	return &Dialer{
		validator: validator,
		netDialer: netDialer,
	}
}

// Dial creates a new grpc client connection to the given target using the atls validator.
func (d *Dialer) Dial(ctx context.Context, target string) (*grpc.ClientConn, error) {
	tlsConfig, err := atls.CreateAttestationClientTLSConfig(nil, []atls.Validator{d.validator})
	if err != nil {
		return nil, err
	}

	return grpc.DialContext(ctx, target,
		d.grpcWithDialer(),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
}

// DialInsecure creates a new grpc client connection to the given target without using encryption or verification.
// Only use this method when using another kind of encryption / verification (VPN, etc).
func (d *Dialer) DialInsecure(ctx context.Context, target string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, target,
		d.grpcWithDialer(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

// DialNoVerify creates a new grpc client connection to the given target without verifying the server's attestation.
func (d *Dialer) DialNoVerify(ctx context.Context, target string) (*grpc.ClientConn, error) {
	tlsConfig, err := atls.CreateAttestationClientTLSConfig(nil, nil)
	if err != nil {
		return nil, err
	}

	return grpc.DialContext(ctx, target,
		d.grpcWithDialer(),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
}

func (d *Dialer) grpcWithDialer() grpc.DialOption {
	return grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return d.netDialer.DialContext(ctx, "tcp", addr)
	})
}

// NetDialer implements the net Dialer interface.
type NetDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}
