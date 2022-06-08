package statuswaiter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/coordinator/atls"
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
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
	assert.NoError(waiter.InitializeValidators([]atls.Validator{core.NewMockValidator()}))
	assert.NoError(waiter.WaitFor(context.Background(), "someIP", state.IsNode))
}

func TestWaitForAndWaitForAll(t *testing.T) {
	var noErr error
	someErr := errors.New("failed")

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
