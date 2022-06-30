package proto

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestClose(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	client := Client{}

	// Create a connection.
	listener := bufconn.Listen(4)
	defer listener.Close()
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(err)
	defer conn.Close()

	// Wait for connection to reach 'connecting' state.
	// Connection is not yet usable in this state, but we just need
	// any stable non 'shutdown' state to validate that the state
	// previous to calling close isn't already 'shutdown'.
	err = func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if connectivity.Connecting == conn.GetState() {
				return nil
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()
	require.NoError(err)

	client.conn = conn

	// Close connection.
	assert.NoError(client.Close())
	assert.Empty(client.conn)
	assert.Equal(connectivity.Shutdown, conn.GetState())

	// Close closed connection.
	assert.NoError(client.Close())
	assert.Empty(client.conn)
	assert.Equal(connectivity.Shutdown, conn.GetState())
}

func TestGetState(t *testing.T) {
	someErr := errors.New("some error")

	testCases := map[string]struct {
		pubAPIClient pubproto.APIClient
		wantErr      bool
		wantState    state.State
	}{
		"success": {
			pubAPIClient: &stubPubAPIClient{getStateState: state.IsNode},
			wantState:    state.IsNode,
		},
		"getState error": {
			pubAPIClient: &stubPubAPIClient{getStateErr: someErr},
			wantErr:      true,
		},
		"uninitialized": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := Client{}
			if tc.pubAPIClient != nil {
				client.pubapi = tc.pubAPIClient
			}

			state, err := client.GetState(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantState, state)
			}
		})
	}
}

func TestActivate(t *testing.T) {
	testKey := base64.StdEncoding.EncodeToString([]byte("32bytesWireGuardKeyForTheTesting"))
	someErr := errors.New("failed")

	testCases := map[string]struct {
		pubAPIClient  *stubPubAPIClient
		userPublicKey string
		ips           []string
		wantErr       bool
	}{
		"normal activation": {
			pubAPIClient:  &stubPubAPIClient{},
			userPublicKey: testKey,
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			wantErr:       false,
		},
		"client without pubAPIClient": {
			userPublicKey: testKey,
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			wantErr:       true,
		},
		"empty public key parameter": {
			pubAPIClient:  &stubPubAPIClient{},
			userPublicKey: "",
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			wantErr:       true,
		},
		"invalid public key parameter": {
			pubAPIClient:  &stubPubAPIClient{},
			userPublicKey: "invalid Key",
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			wantErr:       true,
		},
		"empty ips parameter": {
			pubAPIClient:  &stubPubAPIClient{},
			userPublicKey: testKey,
			ips:           []string{},
			wantErr:       true,
		},
		"fail ActivateAsCoordinator": {
			pubAPIClient:  &stubPubAPIClient{activateAsCoordinatorErr: someErr},
			userPublicKey: testKey,
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := Client{}
			if tc.pubAPIClient != nil {
				client.pubapi = tc.pubAPIClient
			}
			_, err := client.Activate(context.Background(), []byte(tc.userPublicKey), []byte("Constellation"), tc.ips, nil, nil, "serviceaccount://test", nil)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal("32bytesWireGuardKeyForTheTesting", string(tc.pubAPIClient.activateAsCoordinatorReqKey))
				assert.Equal(tc.ips, tc.pubAPIClient.activateAsCoordinatorReqIPs)
				assert.Equal("Constellation", string(tc.pubAPIClient.activateAsCoordinatorMasterSecret))
				assert.Equal("serviceaccount://test", tc.pubAPIClient.activateCloudServiceAccountURI)
			}
		})
	}
}

type stubPubAPIClient struct {
	getStateState                     state.State
	getStateErr                       error
	activateAsCoordinatorErr          error
	activateAdditionalNodesErr        error
	activateAsCoordinatorReqKey       []byte
	activateAsCoordinatorReqIPs       []string
	activateAsCoordinatorMasterSecret []byte
	activateAdditionalNodesReqIPs     []string
	activateCloudServiceAccountURI    string
	pubproto.APIClient
}

func (s *stubPubAPIClient) GetState(ctx context.Context, in *pubproto.GetStateRequest, opts ...grpc.CallOption) (*pubproto.GetStateResponse, error) {
	return &pubproto.GetStateResponse{State: uint32(s.getStateState)}, s.getStateErr
}

func (s *stubPubAPIClient) ActivateAsCoordinator(ctx context.Context, in *pubproto.ActivateAsCoordinatorRequest,
	opts ...grpc.CallOption,
) (pubproto.API_ActivateAsCoordinatorClient, error) {
	s.activateAsCoordinatorReqKey = in.AdminVpnPubKey
	s.activateAsCoordinatorReqIPs = in.NodePublicIps
	s.activateAsCoordinatorMasterSecret = in.MasterSecret
	s.activateCloudServiceAccountURI = in.CloudServiceAccountUri
	return dummyActivateAsCoordinatorClient{}, s.activateAsCoordinatorErr
}

func (s *stubPubAPIClient) ActivateAdditionalNodes(ctx context.Context, in *pubproto.ActivateAdditionalNodesRequest,
	opts ...grpc.CallOption,
) (pubproto.API_ActivateAdditionalNodesClient, error) {
	s.activateAdditionalNodesReqIPs = in.NodePublicIps
	return dummyActivateAdditionalNodesClient{}, s.activateAdditionalNodesErr
}
