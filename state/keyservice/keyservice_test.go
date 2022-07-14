package keyservice

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/internal/atls"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/oid"
	"github.com/edgelesssys/constellation/kms/kmsproto"
	"github.com/edgelesssys/constellation/state/keyservice/keyproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

func TestRequestKeyLoop(t *testing.T) {
	defaultInstance := metadata.InstanceMetadata{
		Name:       "test-instance",
		ProviderID: "/test/provider",
		Role:       role.ControlPlane,
		PrivateIPs: []string{"192.0.2.1"},
	}

	testCases := map[string]struct {
		server          *stubAPIServer
		wantCalls       int
		listResponse    []metadata.InstanceMetadata
		dontStartServer bool
	}{
		"success": {
			server:       &stubAPIServer{requestStateDiskKeyResp: &kmsproto.GetDataKeyResponse{}},
			listResponse: []metadata.InstanceMetadata{defaultInstance},
		},
		"no error if server throws an error": {
			server: &stubAPIServer{
				requestStateDiskKeyResp: &kmsproto.GetDataKeyResponse{},
				requestStateDiskKeyErr:  errors.New("error"),
			},
			listResponse: []metadata.InstanceMetadata{defaultInstance},
		},
		"no error if the server can not be reached": {
			server:          &stubAPIServer{requestStateDiskKeyResp: &kmsproto.GetDataKeyResponse{}},
			listResponse:    []metadata.InstanceMetadata{defaultInstance},
			dontStartServer: true,
		},
		"no error if no endpoint is available": {
			server: &stubAPIServer{requestStateDiskKeyResp: &kmsproto.GetDataKeyResponse{}},
		},
		"works for multiple endpoints": {
			server: &stubAPIServer{requestStateDiskKeyResp: &kmsproto.GetDataKeyResponse{}},
			listResponse: []metadata.InstanceMetadata{
				defaultInstance,
				{
					Name:       "test-instance-2",
					ProviderID: "/test/provider",
					Role:       role.ControlPlane,
					PrivateIPs: []string{"192.0.2.2"},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			keyReceived := make(chan struct{}, 1)
			listener := bufconn.Listen(1)
			defer listener.Close()

			creds := atlscredentials.New(atls.NewFakeIssuer(oid.Dummy{}), nil)
			s := grpc.NewServer(grpc.Creds(creds))
			kmsproto.RegisterAPIServer(s, tc.server)

			if !tc.dontStartServer {
				go func() { require.NoError(s.Serve(listener)) }()
			}

			keyWaiter := &KeyAPI{
				log:         logger.NewTest(t),
				metadata:    stubMetadata{listResponse: tc.listResponse},
				keyReceived: keyReceived,
				timeout:     500 * time.Millisecond,
			}

			// notify the API a key was received after 1 second
			go func() {
				time.Sleep(1 * time.Second)
				keyReceived <- struct{}{}
			}()

			keyWaiter.requestKeyLoop(
				"1234",
				grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
					return listener.DialContext(ctx)
				}),
			)

			s.Stop()
		})
	}
}

func TestPushStateDiskKey(t *testing.T) {
	testCases := map[string]struct {
		testAPI *KeyAPI
		request *keyproto.PushStateDiskKeyRequest
		wantErr bool
	}{
		"success": {
			testAPI: &KeyAPI{keyReceived: make(chan struct{}, 1)},
			request: &keyproto.PushStateDiskKeyRequest{StateDiskKey: []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")},
		},
		"key already set": {
			testAPI: &KeyAPI{
				keyReceived: make(chan struct{}, 1),
				key:         []byte("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
			},
			request: &keyproto.PushStateDiskKeyRequest{StateDiskKey: []byte("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")},
			wantErr: true,
		},
		"incorrect size of pushed key": {
			testAPI: &KeyAPI{keyReceived: make(chan struct{}, 1)},
			request: &keyproto.PushStateDiskKeyRequest{StateDiskKey: []byte("AAAAAAAAAAAAAAAA")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			tc.testAPI.log = logger.NewTest(t)
			_, err := tc.testAPI.PushStateDiskKey(context.Background(), tc.request)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.request.StateDiskKey, tc.testAPI.key)
			}
		})
	}
}

func TestResetKey(t *testing.T) {
	api := New(logger.NewTest(t), nil, nil, time.Second)

	api.key = []byte{0x1, 0x2, 0x3}
	api.ResetKey()
	assert.Nil(t, api.key)
}

type stubAPIServer struct {
	requestStateDiskKeyResp *kmsproto.GetDataKeyResponse
	requestStateDiskKeyErr  error
	kmsproto.UnimplementedAPIServer
}

func (s *stubAPIServer) GetDataKey(ctx context.Context, req *kmsproto.GetDataKeyRequest) (*kmsproto.GetDataKeyResponse, error) {
	return s.requestStateDiskKeyResp, s.requestStateDiskKeyErr
}

type stubMetadata struct {
	listResponse []metadata.InstanceMetadata
}

func (s stubMetadata) List(ctx context.Context) ([]metadata.InstanceMetadata, error) {
	return s.listResponse, nil
}
