package status

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestWaitForAndWaitForAll(t *testing.T) {
	var noErr error
	someErr := errors.New("failed")

	testCases := map[string]struct {
		waiter       Waiter
		waitForState state.State
		wantErr      bool
	}{
		"successful wait": {
			waiter: Waiter{
				interval:  time.Millisecond,
				newConn:   stubNewConnFunc(noErr),
				newClient: stubNewClientFunc(&stubPeerStatusClient{state: state.IsNode}),
			},
			waitForState: state.IsNode,
		},
		"expect timeout": {
			waiter: Waiter{
				interval:  time.Millisecond,
				newConn:   stubNewConnFunc(noErr),
				newClient: stubNewClientFunc(&stubPeerStatusClient{state: state.AcceptingInit}),
			},
			waitForState: state.IsNode,
			wantErr:      true,
		},
		"fail to check call": {
			waiter: Waiter{
				interval:  time.Millisecond,
				newConn:   stubNewConnFunc(noErr),
				newClient: stubNewClientFunc(&stubPeerStatusClient{checkErr: someErr}),
			},
			waitForState: state.IsNode,
			wantErr:      true,
		},
		"fail to create conn": {
			waiter: Waiter{
				interval:  time.Millisecond,
				newConn:   stubNewConnFunc(someErr),
				newClient: stubNewClientFunc(&stubPeerStatusClient{}),
			},
			waitForState: state.IsNode,
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

				err := tc.waiter.WaitFor(ctx, tc.waitForState, "someIP")

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
				err := tc.waiter.WaitForAll(ctx, tc.waitForState, endpoints)

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

func (c *stubClientConn) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
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
