package core

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/deploy/user"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/grpc/atlscredentials"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
	kms "github.com/edgelesssys/constellation/kms/server/setup"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func TestReinitializeAsNode(t *testing.T) {
	testPeers := []peer.Peer{
		{
			PublicIP:  "192.0.2.1",
			VPNIP:     "198.51.100.1",
			VPNPubKey: []byte{0x1, 0x2, 0x3},
			Role:      role.Coordinator,
		},
	}
	wantedVPNPeers := []stubVPNPeer{
		{
			publicIP: "192.0.2.1",
			vpnIP:    "198.51.100.1",
			pubKey:   []byte{0x1, 0x2, 0x3},
		},
	}
	vpnIP := "198.51.100.2"

	testCases := map[string]struct {
		getInitialVPNPeersResponses []struct {
			peers []peer.Peer
			err   error
		}
		wantErr bool
	}{
		"reinitialize as node works": {
			getInitialVPNPeersResponses: []struct {
				peers []peer.Peer
				err   error
			}{{peers: testPeers}},
		},
		"reinitialize as node will retry until vpn peers are retrieved": {
			getInitialVPNPeersResponses: []struct {
				peers []peer.Peer
				err   error
			}{
				{err: errors.New("retrieving vpn peers failed")},
				{peers: testPeers},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			coordinators := []cloudtypes.Instance{{PrivateIPs: []string{"192.0.2.1"}, Role: role.Coordinator}}
			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, &MockValidator{}, netDialer)
			server := newPubAPIServer()
			api := &pubAPIServerStub{responses: tc.getInitialVPNPeersResponses}
			pubproto.RegisterAPIServer(server, api)
			go server.Serve(netDialer.GetListener("192.0.2.1:9000"))
			defer server.Stop()
			vpn := &stubVPN{}
			fs := afero.NewMemMapFs()
			core, err := NewCore(vpn, nil, &stubMetadata{listRes: coordinators, supportedRes: true}, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(fs), user.NewLinuxUserManagerFake(fs))
			require.NoError(err)
			err = core.ReinitializeAsNode(context.Background(), dialer, vpnIP, &stubPubAPI{}, 0)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(vpnIP, vpn.interfaceIP)
			assert.Equal(wantedVPNPeers, vpn.peers)
		})
	}
}

func TestReinitializeAsCoordinator(t *testing.T) {
	testPeers := []peer.Peer{
		{
			PublicIP:  "192.0.2.1",
			VPNIP:     "198.51.100.1",
			VPNPubKey: []byte{0x1, 0x2, 0x3},
			Role:      role.Coordinator,
		},
	}
	wantedVPNPeers := []stubVPNPeer{
		{
			publicIP: "192.0.2.1",
			vpnIP:    "198.51.100.1",
			pubKey:   []byte{0x1, 0x2, 0x3},
		},
	}
	vpnIP := "198.51.100.2"

	testCases := map[string]struct {
		getInitialVPNPeersResponses []struct {
			peers []peer.Peer
			err   error
		}
		wantErr bool
	}{
		"reinitialize as coordinator works": {
			getInitialVPNPeersResponses: []struct {
				peers []peer.Peer
				err   error
			}{{peers: testPeers}},
		},
		"reinitialize as coordinator will retry until vpn peers are retrieved": {
			getInitialVPNPeersResponses: []struct {
				peers []peer.Peer
				err   error
			}{
				{err: errors.New("retrieving vpn peers failed")},
				{peers: testPeers},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			coordinators := []cloudtypes.Instance{{PrivateIPs: []string{"192.0.2.1"}, Role: role.Coordinator}}
			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, &MockValidator{}, netDialer)
			server := newPubAPIServer()
			api := &pubAPIServerStub{responses: tc.getInitialVPNPeersResponses}
			pubproto.RegisterAPIServer(server, api)
			go server.Serve(netDialer.GetListener("192.0.2.1:9000"))
			defer server.Stop()
			vpn := &stubVPN{}
			fs := afero.NewMemMapFs()
			core, err := NewCore(vpn, nil, &stubMetadata{listRes: coordinators, supportedRes: true}, nil, zaptest.NewLogger(t), nil, &fakeStoreFactory{}, file.NewHandler(fs), user.NewLinuxUserManagerFake(fs))
			require.NoError(err)
			// prepare store to emulate initialized KMS
			require.NoError(core.data().PutKMSData(kms.KMSInformation{StorageUri: kms.NoStoreURI, KmsUri: kms.ClusterKMSURI}))
			require.NoError(core.data().PutMasterSecret([]byte("master-secret")))
			err = core.ReinitializeAsCoordinator(context.Background(), dialer, vpnIP, &stubPubAPI{}, 0)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(vpnIP, vpn.interfaceIP)
			assert.Equal(wantedVPNPeers, vpn.peers)
		})
	}
}

func TestGetInitialVPNPeers(t *testing.T) {
	testPeers := []peer.Peer{
		{
			PublicIP:  "192.0.2.1",
			VPNIP:     "198.51.100.1",
			VPNPubKey: []byte{0x1, 0x2, 0x3},
			Role:      role.Coordinator,
		},
	}

	testCases := map[string]struct {
		ownCoordinatorEndpoint *string
		coordinatorIPs         []string
		metadataErr            error
		peers                  []peer.Peer
		getVPNPeersErr         error
		wantErr                bool
	}{
		"getInitialVPNPeers works from worker node": {
			coordinatorIPs: []string{"192.0.2.1"},
			peers:          testPeers,
		},
		"getInitialVPNPeers works from coordinator": {
			ownCoordinatorEndpoint: proto.String("192.0.2.2:9000"),
			coordinatorIPs:         []string{"192.0.2.1", "192.0.2.2"},
			peers:                  testPeers,
		},
		"getInitialVPNPeers filters itself": {
			ownCoordinatorEndpoint: proto.String("192.0.2.2:9000"),
			coordinatorIPs:         []string{"192.0.2.2"},
			wantErr:                true,
		},
		"getInitialVPNPeers fails if no coordinators are found": {
			wantErr: true,
		},
		"getInitialVPNPeers fails if metadata API fails to retrieve coordinators": {
			metadataErr: errors.New("metadata error"),
			wantErr:     true,
		},
		"getInitialVPNPeers fails if rpc call fails": {
			coordinatorIPs: []string{"192.0.2.1"},
			getVPNPeersErr: errors.New("rpc error"),
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			coordinators := func(ips []string) []cloudtypes.Instance {
				instances := []cloudtypes.Instance{}
				for _, ip := range ips {
					instances = append(instances, cloudtypes.Instance{PrivateIPs: []string{ip}, Role: role.Coordinator})
				}
				return instances
			}(tc.coordinatorIPs)
			zapLogger, err := zap.NewDevelopment()
			require.NoError(err)
			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, &MockValidator{}, netDialer)
			server := newPubAPIServer()
			api := &pubAPIServerStub{
				responses: []struct {
					peers []peer.Peer
					err   error
				}{{peers: tc.peers, err: tc.getVPNPeersErr}},
			}
			pubproto.RegisterAPIServer(server, api)
			go server.Serve(netDialer.GetListener("192.0.2.1:9000"))
			defer server.Stop()
			peers, err := getInitialVPNPeers(context.Background(), dialer, zapLogger, &stubMetadata{listRes: coordinators, listErr: tc.metadataErr, supportedRes: true}, tc.ownCoordinatorEndpoint)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.peers, peers)
		})
	}
}

func newPubAPIServer() *grpc.Server {
	creds := atlscredentials.New(&MockIssuer{}, nil)

	return grpc.NewServer(grpc.Creds(creds))
}

type pubAPIServerStub struct {
	responses []struct {
		peers []peer.Peer
		err   error
	}
	i int
	pubproto.UnimplementedAPIServer
}

func (s *pubAPIServerStub) GetVPNPeers(ctx context.Context, in *pubproto.GetVPNPeersRequest) (*pubproto.GetVPNPeersResponse, error) {
	if len(s.responses) == 0 {
		return nil, nil
	}
	resp := s.responses[s.i]
	s.i = (s.i + 1) % len(s.responses)
	return &pubproto.GetVPNPeersResponse{
		Peers: peer.ToPubProto(resp.peers),
	}, resp.err
}
