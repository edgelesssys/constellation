package keyservice

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/oid"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"
)

func TestRequestLoop(t *testing.T) {
	defaultInstance := core.Instance{
		Name:       "test-instance",
		ProviderID: "/test/provider",
		Role:       role.Coordinator,
		IPs:        []string{"192.0.2.1"},
	}

	testCases := map[string]struct {
		server          *stubAPIServer
		expectedCalls   int
		listResponse    []core.Instance
		dontStartServer bool
	}{
		"success": {
			server:       &stubAPIServer{requestStateDiskKeyResp: &pubproto.RequestStateDiskKeyResponse{}},
			listResponse: []core.Instance{defaultInstance},
		},
		"no error if server throws an error": {
			server: &stubAPIServer{
				requestStateDiskKeyResp: &pubproto.RequestStateDiskKeyResponse{},
				requestStateDiskKeyErr:  errors.New("error"),
			},
			listResponse: []core.Instance{defaultInstance},
		},
		"no error if the server can not be reached": {
			server:          &stubAPIServer{requestStateDiskKeyResp: &pubproto.RequestStateDiskKeyResponse{}},
			listResponse:    []core.Instance{defaultInstance},
			dontStartServer: true,
		},
		"no error if no endpoint is available": {
			server: &stubAPIServer{requestStateDiskKeyResp: &pubproto.RequestStateDiskKeyResponse{}},
		},
		"works for multiple endpoints": {
			server: &stubAPIServer{requestStateDiskKeyResp: &pubproto.RequestStateDiskKeyResponse{}},
			listResponse: []core.Instance{
				defaultInstance,
				{
					Name:       "test-instance-2",
					ProviderID: "/test/provider",
					Role:       role.Coordinator,
					IPs:        []string{"192.0.2.2"},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			keyReceived := make(chan bool, 1)
			listener := bufconn.Listen(1)
			defer listener.Close()

			tlsConfig, err := stubTLSConfig()
			require.NoError(err)
			s := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
			pubproto.RegisterAPIServer(s, tc.server)

			if !tc.dontStartServer {
				go func() { require.NoError(s.Serve(listener)) }()
			}

			keyWaiter := &keyAPI{
				metadata:    stubMetadata{listResponse: tc.listResponse},
				keyReceived: keyReceived,
				timeout:     500 * time.Millisecond,
			}

			// notify the API a key was received after 1 second
			go func() {
				time.Sleep(1 * time.Second)
				keyReceived <- true
			}()

			err = keyWaiter.requestKeyFromCoordinator(
				"1234",
				grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
					return listener.DialContext(ctx)
				}),
			)
			assert.NoError(err)

			s.Stop()
		})
	}
}

type stubAPIServer struct {
	requestStateDiskKeyResp *pubproto.RequestStateDiskKeyResponse
	requestStateDiskKeyErr  error
	pubproto.UnimplementedAPIServer
}

func (s *stubAPIServer) GetState(ctx context.Context, in *pubproto.GetStateRequest) (*pubproto.GetStateResponse, error) {
	return nil, nil
}

func (s *stubAPIServer) ActivateAsCoordinator(in *pubproto.ActivateAsCoordinatorRequest, srv pubproto.API_ActivateAsCoordinatorServer) error {
	return nil
}

func (s *stubAPIServer) ActivateAsNode(pubproto.API_ActivateAsNodeServer) error {
	return nil
}

func (s *stubAPIServer) ActivateAdditionalNodes(in *pubproto.ActivateAdditionalNodesRequest, srv pubproto.API_ActivateAdditionalNodesServer) error {
	return nil
}

func (s *stubAPIServer) JoinCluster(ctx context.Context, in *pubproto.JoinClusterRequest) (*pubproto.JoinClusterResponse, error) {
	return nil, nil
}

func (s *stubAPIServer) TriggerNodeUpdate(ctx context.Context, in *pubproto.TriggerNodeUpdateRequest) (*pubproto.TriggerNodeUpdateResponse, error) {
	return nil, nil
}

func (s *stubAPIServer) RequestStateDiskKey(ctx context.Context, in *pubproto.RequestStateDiskKeyRequest) (*pubproto.RequestStateDiskKeyResponse, error) {
	return s.requestStateDiskKeyResp, s.requestStateDiskKeyErr
}

func stubTLSConfig() (*tls.Config, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	getCertificate := func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
		serialNumber, err := util.GenerateCertificateSerialNumber()
		if err != nil {
			return nil, err
		}
		now := time.Now()
		template := &x509.Certificate{
			SerialNumber:    serialNumber,
			Subject:         pkix.Name{CommonName: "Constellation"},
			NotBefore:       now.Add(-2 * time.Hour),
			NotAfter:        now.Add(2 * time.Hour),
			ExtraExtensions: []pkix.Extension{{Id: oid.Dummy{}.OID(), Value: []byte{0x1, 0x2, 0x3}}},
		}
		cert, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
		if err != nil {
			return nil, err
		}

		return &tls.Certificate{Certificate: [][]byte{cert}, PrivateKey: priv}, nil
	}
	return &tls.Config{GetCertificate: getCertificate, MinVersion: tls.VersionTLS12}, nil
}
