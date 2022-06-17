package statuswaiter

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestInitializeValidators(t *testing.T) {
	assert := assert.New(t)

	waiter := Waiter{
		interval:  time.Millisecond,
		newClient: stubNewClientFunc(&stubPeerStatusClient{state: state.IsNode}),
	}

	// Uninitialized waiter fails.
	assert.Error(waiter.WaitFor(context.Background(), "someIP", state.IsNode))

	// Initializing waiter with no validators fails
	assert.Error(waiter.InitializeValidators(nil))

	// Initialized waiter succeeds
	assert.NoError(waiter.InitializeValidators(atls.NewFakeValidators(oid.Dummy{})))
	assert.NoError(waiter.WaitFor(context.Background(), "someIP", state.IsNode))
}

func TestWaitForAndWaitForAll(t *testing.T) {
	var noErr error
	someErr := errors.New("failed")
	handshakeErr := status.Error(codes.Unavailable, `connection error: desc = "transport: authentication handshake failed"`)

	testCases := map[string]struct {
		waiter       Waiter
		waitForState []state.State
		wantErr      bool
	}{
		"successful wait": {
			waiter: Waiter{
				initialized: true,
				interval:    time.Millisecond,
				newConn:     stubNewConnFunc(noErr),
				newClient:   stubNewClientFunc(&stubPeerStatusClient{state: state.IsNode}),
			},
			waitForState: []state.State{state.IsNode},
		},
		"successful wait multi states": {
			waiter: Waiter{
				initialized: true,
				interval:    time.Millisecond,
				newConn:     stubNewConnFunc(noErr),
				newClient:   stubNewClientFunc(&stubPeerStatusClient{state: state.IsNode}),
			},
			waitForState: []state.State{state.IsNode, state.ActivatingNodes},
		},
		"expect timeout": {
			waiter: Waiter{
				initialized: true,
				interval:    time.Millisecond,
				newConn:     stubNewConnFunc(noErr),
				newClient:   stubNewClientFunc(&stubPeerStatusClient{state: state.AcceptingInit}),
			},
			waitForState: []state.State{state.IsNode},
			wantErr:      true,
		},
		"fail to check call": {
			waiter: Waiter{
				initialized: true,
				interval:    time.Millisecond,
				newConn:     stubNewConnFunc(noErr),
				newClient:   stubNewClientFunc(&stubPeerStatusClient{checkErr: someErr}),
			},
			waitForState: []state.State{state.IsNode},
			wantErr:      true,
		},
		"fail to create conn": {
			waiter: Waiter{
				initialized: true,
				interval:    time.Millisecond,
				newConn:     stubNewConnFunc(someErr),
				newClient:   stubNewClientFunc(&stubPeerStatusClient{}),
			},
			waitForState: []state.State{state.IsNode},
			wantErr:      true,
		},
		"fail TLS handshake": {
			waiter: Waiter{
				initialized: true,
				interval:    time.Millisecond,
				newConn:     stubNewConnFunc(handshakeErr),
				newClient:   stubNewClientFunc(&stubPeerStatusClient{state: state.IsNode}),
			},
			waitForState: []state.State{state.IsNode},
			wantErr:      true,
		},
	}

	t.Run("WaitFor", func(t *testing.T) {
		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				ctx := context.Background()
				ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()

				err := tc.waiter.WaitFor(ctx, "someIP", tc.waitForState...)

				if tc.wantErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
				}
			})
		}
	})

	t.Run("WaitForAll", func(t *testing.T) {
		for name, tc := range testCases {
			t.Run(name, func(t *testing.T) {
				assert := assert.New(t)

				ctx := context.Background()
				ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()

				endpoints := []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"}
				err := tc.waiter.WaitForAll(ctx, endpoints, tc.waitForState...)

				if tc.wantErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
				}
			})
		}
	})
}

func stubNewConnFunc(errStub error) func(ctx context.Context, target string, opts ...grpc.DialOption) (ClientConn, error) {
	return func(ctx context.Context, target string, opts ...grpc.DialOption) (ClientConn, error) {
		return &stubClientConn{}, errStub
	}
}

type stubClientConn struct{}

func (c *stubClientConn) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return nil
}

func (c *stubClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func (c *stubClientConn) Close() error {
	return nil
}

func stubNewClientFunc(stubClient pubproto.APIClient) func(cc grpc.ClientConnInterface) pubproto.APIClient {
	return func(cc grpc.ClientConnInterface) pubproto.APIClient {
		return stubClient
	}
}

type stubPeerStatusClient struct {
	state    state.State
	checkErr error
	pubproto.APIClient
}

func (c *stubPeerStatusClient) GetState(ctx context.Context, in *pubproto.GetStateRequest, opts ...grpc.CallOption) (*pubproto.GetStateResponse, error) {
	resp := &pubproto.GetStateResponse{State: uint32(c.state)}
	return resp, c.checkErr
}

func TestContainsState(t *testing.T) {
	testCases := map[string]struct {
		s       state.State
		states  []state.State
		success bool
	}{
		"is state": {
			s: state.IsNode,
			states: []state.State{
				state.IsNode,
			},
			success: true,
		},
		"is state multi": {
			s: state.AcceptingInit,
			states: []state.State{
				state.AcceptingInit,
				state.ActivatingNodes,
			},
			success: true,
		},
		"is not state": {
			s: state.NodeWaitingForClusterJoin,
			states: []state.State{
				state.AcceptingInit,
			},
		},
		"is not state multi": {
			s: state.NodeWaitingForClusterJoin,
			states: []state.State{
				state.AcceptingInit,
				state.ActivatingNodes,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			res := containsState(tc.s, tc.states...)
			assert.Equal(tc.success, res)
		})
	}
}

func TestIsHandshakeError(t *testing.T) {
	testCases := map[string]struct {
		err          error
		wantedResult bool
	}{
		"TLS handshake error": {
			err:          getGRPCHandshakeError(),
			wantedResult: true,
		},
		"Unavailable error": {
			err:          status.Error(codes.Unavailable, "connection error"),
			wantedResult: false,
		},
		"TLS handshake error with wrong code": {
			err:          status.Error(codes.Aborted, `connection error: desc = "transport: authentication handshake failed`),
			wantedResult: false,
		},
		"Non gRPC error": {
			err:          errors.New("error"),
			wantedResult: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			res := isGRPCHandshakeError(tc.err)
			assert.Equal(tc.wantedResult, res)
		})
	}
}

func getGRPCHandshakeError() error {
	serverCreds := atlscredentials.New(atls.NewFakeIssuer(oid.Dummy{}), nil)
	api := &fakeAPI{}
	server := grpc.NewServer(grpc.Creds(serverCreds))
	pubproto.RegisterAPIServer(server, api)

	listener := bufconn.Listen(1024)
	defer server.GracefulStop()
	go server.Serve(listener)

	clientCreds := atlscredentials.New(nil, []atls.Validator{failingValidator{oid.Dummy{}}})
	conn, err := grpc.DialContext(context.Background(), "", grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithTransportCredentials(clientCreds))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := pubproto.NewAPIClient(conn)
	_, err = client.GetState(context.Background(), &pubproto.GetStateRequest{})
	return err
}

type failingValidator struct {
	oid.Getter
}

func (v failingValidator) Validate(attDoc []byte, nonce []byte) ([]byte, error) {
	return nil, errors.New("error")
}

type fakeAPI struct {
	pubproto.UnimplementedAPIServer
}

func (f *fakeAPI) GetState(ctx context.Context, in *pubproto.GetStateRequest) (*pubproto.GetStateResponse, error) {
	return &pubproto.GetStateResponse{State: 1}, nil
}
