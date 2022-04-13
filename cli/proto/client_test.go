package proto

import (
	"context"
	"encoding/base64"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

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

func TestActivate(t *testing.T) {
	testKey := base64.StdEncoding.EncodeToString([]byte("32bytesWireGuardKeyForTheTesting"))
	someErr := errors.New("failed")

	testCases := map[string]struct {
		avpn          *stubAVPNClient
		userPublicKey string
		ips           []string
		errExpected   bool
	}{
		"normal activation": {
			avpn:          &stubAVPNClient{},
			userPublicKey: testKey,
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			errExpected:   false,
		},
		"client without avpn": {
			userPublicKey: testKey,
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			errExpected:   true,
		},
		"empty public key parameter": {
			avpn:          &stubAVPNClient{},
			userPublicKey: "",
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			errExpected:   true,
		},
		"invalid public key parameter": {
			avpn:          &stubAVPNClient{},
			userPublicKey: "invalid Key",
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			errExpected:   true,
		},
		"empty ips parameter": {
			avpn:          &stubAVPNClient{},
			userPublicKey: testKey,
			ips:           []string{},
			errExpected:   true,
		},
		"fail ActivateAsCoordinator": {
			avpn:          &stubAVPNClient{activateAsCoordinatorErr: someErr},
			userPublicKey: testKey,
			ips:           []string{"192.0.2.1", "192.0.2.1", "192.0.2.1"},
			errExpected:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := Client{}
			if tc.avpn != nil {
				client.avpn = tc.avpn
			}
			_, err := client.Activate(context.Background(), []byte(tc.userPublicKey), []byte("Constellation"), tc.ips, nil, "serviceaccount://test")
			if tc.errExpected {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal("32bytesWireGuardKeyForTheTesting", string(tc.avpn.activateAsCoordinatorReqKey))
				assert.Equal(tc.ips, tc.avpn.activateAsCoordinatorReqIPs)
				assert.Equal("Constellation", string(tc.avpn.activateAsCoordinatorMasterSecret))
				assert.Equal("serviceaccount://test", tc.avpn.activateCloudServiceAccountURI)
			}
		})
	}
}

type stubAVPNClient struct {
	activateAsCoordinatorErr          error
	activateAdditionalNodesErr        error
	activateAsCoordinatorReqKey       []byte
	activateAsCoordinatorReqIPs       []string
	activateAsCoordinatorMasterSecret []byte
	activateAdditionalNodesReqIPs     []string
	activateCloudServiceAccountURI    string
	pubproto.APIClient
}

func (s *stubAVPNClient) ActivateAsCoordinator(ctx context.Context, in *pubproto.ActivateAsCoordinatorRequest,
	opts ...grpc.CallOption,
) (pubproto.API_ActivateAsCoordinatorClient, error) {
	s.activateAsCoordinatorReqKey = in.AdminVpnPubKey
	s.activateAsCoordinatorReqIPs = in.NodePublicIps
	s.activateAsCoordinatorMasterSecret = in.MasterSecret
	s.activateCloudServiceAccountURI = in.CloudServiceAccountUri
	return dummyAVPNActivateAsCoordinatorClient{}, s.activateAsCoordinatorErr
}

func (s *stubAVPNClient) ActivateAdditionalNodes(ctx context.Context, in *pubproto.ActivateAdditionalNodesRequest,
	opts ...grpc.CallOption,
) (pubproto.API_ActivateAdditionalNodesClient, error) {
	s.activateAdditionalNodesReqIPs = in.NodePublicIps
	return dummyAVPNActivateAdditionalNodesClient{}, s.activateAdditionalNodesErr
}
