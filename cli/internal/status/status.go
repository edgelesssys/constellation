package status

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	grpcstatus "google.golang.org/grpc/status"
)

// TODO(katexochen): Use protoClient for waiter?

// Waiter waits for PeerStatusServer to reach a specific state. The waiter needs
// to be initialized before usage.
type Waiter struct {
	initialized bool
	interval    time.Duration
	newConn     func(ctx context.Context, target string, opts ...grpc.DialOption) (ClientConn, error)
	newClient   func(cc grpc.ClientConnInterface) pubproto.APIClient
}

// NewWaiter returns a default Waiter with probing inteval of 10 seconds,
// attested gRPC connection and PeerStatusClient.
func NewWaiter() *Waiter {
	return &Waiter{
		interval:  10 * time.Second,
		newClient: pubproto.NewAPIClient,
	}
}

// InitializeValidators initializes the validators for the attestation.
func (w *Waiter) InitializeValidators(validators []atls.Validator) error {
	if len(validators) == 0 {
		return errors.New("no validators provided to initialize status waiter")
	}
	w.newConn = newAttestedConnGenerator(validators)
	w.initialized = true
	return nil
}

// WaitFor waits for a PeerStatusServer, which is reachable under the given endpoint
// to reach the specified state.
func (w *Waiter) WaitFor(ctx context.Context, endpoint string, status ...state.State) error {
	if !w.initialized {
		return errors.New("waiter not initialized")
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Check once before waiting
	resp, err := w.probe(ctx, endpoint)
	if err != nil && grpcstatus.Code(err) != grpccodes.Unavailable {
		return err
	}
	if resp != nil && containsState(state.State(resp.State), status...) {
		return nil
	}

	// Periodically check status again
	for {
		select {
		case <-ticker.C:
			resp, err := w.probe(ctx, endpoint)
			if grpcstatus.Code(err) == grpccodes.Unavailable {
				// The server isn't reachable yet.
				continue
			}
			if err != nil {
				return err
			}
			if containsState(state.State(resp.State), status...) {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// probe sends a PeerStatusCheck request to a PeerStatusServer and returns the response.
func (w *Waiter) probe(ctx context.Context, endpoint string) (*pubproto.GetStateResponse, error) {
	conn, err := w.newConn(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := w.newClient(conn)
	return client.GetState(ctx, &pubproto.GetStateRequest{})
}

// WaitForAll waits for a list of PeerStatusServers, which listen on the handed
// endpoints, to reach the specified state.
func (w *Waiter) WaitForAll(ctx context.Context, endpoints []string, status ...state.State) error {
	if !w.initialized {
		return errors.New("waiter not initialized")
	}

	for _, endpoint := range endpoints {
		if err := w.WaitFor(ctx, endpoint, status...); err != nil {
			return err
		}
	}
	return nil
}

// newAttestedConnGenerator creates a function returning a default attested grpc connection.
func newAttestedConnGenerator(validators []atls.Validator) func(ctx context.Context, target string, opts ...grpc.DialOption) (ClientConn, error) {
	return func(ctx context.Context, target string, opts ...grpc.DialOption) (ClientConn, error) {
		tlsConfig, err := atls.CreateAttestationClientTLSConfig(nil, validators)
		if err != nil {
			return nil, err
		}

		return grpc.DialContext(
			ctx, target, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		)
	}
}

// ClientConn is the gRPC connection a PeerStatusClient uses to connect to a server.
type ClientConn interface {
	grpc.ClientConnInterface
	io.Closer
}

// containsState checks if current state is one of the given states.
func containsState(s state.State, states ...state.State) bool {
	for _, state := range states {
		if state == s {
			return true
		}
	}
	return false
}
