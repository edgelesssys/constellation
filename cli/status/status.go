package status

import (
	"context"
	"io"
	"time"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/attestation/aws"
	"github.com/edgelesssys/constellation/coordinator/attestation/azure"
	"github.com/edgelesssys/constellation/coordinator/attestation/gcp"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	grpcstatus "google.golang.org/grpc/status"
)

// Waiter waits for PeerStatusServer to reach a specific state.
type Waiter struct {
	interval  time.Duration
	newConn   func(ctx context.Context, target string, opts ...grpc.DialOption) (ClientConn, error)
	newClient func(cc grpc.ClientConnInterface) pubproto.APIClient
}

// NewWaiter returns a default Waiter with probing inteval of 10 seconds,
// attested gRPC connection and PeerStatusClient.
func NewWaiter(gcpPCRs map[uint32][]byte) Waiter {
	return Waiter{
		interval:  10 * time.Second,
		newConn:   newAttestedConnGenerator(gcpPCRs),
		newClient: pubproto.NewAPIClient,
	}
}

// WaitFor waits for a PeerStatusServer, which is reachable under the given endpoint
// to reach the specified state.
func (w Waiter) WaitFor(ctx context.Context, status state.State, endpoint string) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Check once before waiting
	resp, err := w.probe(ctx, endpoint)
	if err != nil && grpcstatus.Code(err) != grpccodes.Unavailable {
		return err
	}
	if resp != nil && resp.State == uint32(status) {
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
			if resp.State == uint32(status) {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// probe sends a PeerStatusCheck request to a PeerStatusServer and returns the response.
func (w Waiter) probe(ctx context.Context, endpoint string) (*pubproto.GetStateResponse, error) {
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
func (w Waiter) WaitForAll(ctx context.Context, status state.State, endpoints []string) error {
	for _, endpoint := range endpoints {
		if err := w.WaitFor(ctx, status, endpoint); err != nil {
			return err
		}
	}
	return nil
}

// newAttestedConnGenerator creates a function returning a default attested grpc connection.
func newAttestedConnGenerator(gcpPCRs map[uint32][]byte) func(ctx context.Context, target string, opts ...grpc.DialOption) (ClientConn, error) {
	return func(ctx context.Context, target string, opts ...grpc.DialOption) (ClientConn, error) {
		validators := []atls.Validator{
			aws.NewValidator(aws.NaAdGetVerifiedPayloadAsJson),
			gcp.NewValidator(gcpPCRs),
			gcp.NewNonCVMValidator(map[uint32][]byte{}), // TODO: Remove once we no longer use non cvms
			azure.NewValidator(map[uint32][]byte{}),
		}

		tlsConfig, err := atls.CreateAttestationClientTLSConfig(validators)
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
