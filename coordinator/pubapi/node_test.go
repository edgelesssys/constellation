package pubapi

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/util/testdialer"
	"github.com/edgelesssys/constellation/coordinator/vpnapi/vpnproto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestActivateAsNode(t *testing.T) {
	someErr := errors.New("failed")
	peer1 := peer.Peer{PublicEndpoint: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: []byte{1, 2, 3}}
	peer2 := peer.Peer{PublicEndpoint: "192.0.2.12:2000", VPNIP: "192.0.2.22", VPNPubKey: []byte{2, 3, 4}}

	testCases := map[string]struct {
		initialPeers  []peer.Peer
		updatedPeers  []peer.Peer
		state         state.State
		getUpdateErr  error
		setVPNIPErr   error
		expectErr     bool
		expectedState state.State
	}{
		"basic": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.AcceptingInit,
			expectedState: state.NodeWaitingForClusterJoin,
		},
		"already activated": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.IsNode,
			expectErr:     true,
			expectedState: state.IsNode,
		},
		"wrong peer kind": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.ActivatingNodes,
			expectErr:     true,
			expectedState: state.ActivatingNodes,
		},
		"GetUpdate error": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.AcceptingInit,
			getUpdateErr:  someErr,
			expectedState: state.NodeWaitingForClusterJoin,
		},
		"SetVPNIP error": {
			initialPeers:  []peer.Peer{peer1},
			updatedPeers:  []peer.Peer{peer2},
			state:         state.AcceptingInit,
			setVPNIPErr:   someErr,
			expectErr:     true,
			expectedState: state.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			const nodeVPNIP = "192.0.2.2"
			vpnPubKey := []byte{7, 8, 9}
			ownerID := []byte("ownerID")
			clusterID := []byte("clusterID")

			logger := zaptest.NewLogger(t)
			core := &fakeCore{state: tc.state, vpnPubKey: vpnPubKey, setVPNIPErr: tc.setVPNIPErr}
			dialer := testdialer.NewBufconnDialer()

			api := New(logger, core, dialer, nil, nil, nil)
			defer api.Close()

			vserver := grpc.NewServer()
			vapi := &stubVPNAPI{peers: tc.updatedPeers, getUpdateErr: tc.getUpdateErr}
			vpnproto.RegisterAPIServer(vserver, vapi)
			go vserver.Serve(dialer.GetListener(net.JoinHostPort(core.GetCoordinatorVPNIP(), vpnAPIPort)))
			defer vserver.GracefulStop()

			resp, err := api.ActivateAsNode(context.Background(), &pubproto.ActivateAsNodeRequest{
				NodeVpnIp: nodeVPNIP,
				Peers:     peer.ToPubProto(tc.initialPeers),
				OwnerId:   ownerID,
				ClusterId: clusterID,
			})

			assert.Equal(tc.expectedState, core.state)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(vpnPubKey, resp.NodeVpnPubKey)
			assert.Equal(nodeVPNIP, core.vpnIP)
			assert.Equal(ownerID, core.ownerID)
			assert.Equal(clusterID, core.clusterID)

			api.Close() // blocks until update loop finished

			if tc.getUpdateErr == nil {
				require.Len(core.updatedPeers, 2)
				assert.Equal(tc.updatedPeers, core.updatedPeers[1])
			} else {
				require.Len(core.updatedPeers, 1)
			}
			assert.Equal(tc.initialPeers, core.updatedPeers[0])
		})
	}
}

func TestTriggerNodeUpdate(t *testing.T) {
	someErr := errors.New("failed")
	peers := []peer.Peer{
		{PublicEndpoint: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: []byte{1, 2, 3}},
		{PublicEndpoint: "192.0.2.12:2000", VPNIP: "192.0.2.22", VPNPubKey: []byte{2, 3, 4}},
	}

	testCases := map[string]struct {
		peers        []peer.Peer
		state        state.State
		getUpdateErr error
		expectErr    bool
	}{
		"basic": {
			peers: peers,
			state: state.IsNode,
		},
		"not activated": {
			peers:     peers,
			state:     state.AcceptingInit,
			expectErr: true,
		},
		"wrong peer kind": {
			peers:     peers,
			state:     state.ActivatingNodes,
			expectErr: true,
		},
		"GetUpdate error": {
			peers:        peers,
			state:        state.IsNode,
			getUpdateErr: someErr,
			expectErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			logger := zaptest.NewLogger(t)
			core := &fakeCore{state: tc.state}
			dialer := testdialer.NewBufconnDialer()

			api := New(logger, core, dialer, nil, nil, nil)

			vserver := grpc.NewServer()
			vapi := &stubVPNAPI{
				peers:        tc.peers,
				getUpdateErr: tc.getUpdateErr,
			}
			vpnproto.RegisterAPIServer(vserver, vapi)
			go vserver.Serve(dialer.GetListener(net.JoinHostPort(core.GetCoordinatorVPNIP(), vpnAPIPort)))
			defer vserver.GracefulStop()

			_, err := api.TriggerNodeUpdate(context.Background(), &pubproto.TriggerNodeUpdateRequest{})
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			// second update should be a noop
			_, err = api.TriggerNodeUpdate(context.Background(), &pubproto.TriggerNodeUpdateRequest{})
			require.NoError(err)

			require.Len(core.updatedPeers, 1)
			assert.Equal(tc.peers, core.updatedPeers[0])
		})
	}
}

func TestJoinCluster(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		state          state.State
		getJoinArgsErr error
		joinClusterErr error
		expectErr      bool
		expectedState  state.State
	}{
		"basic": {
			state:         state.NodeWaitingForClusterJoin,
			expectedState: state.IsNode,
		},
		"not activated": {
			state:         state.AcceptingInit,
			expectErr:     true,
			expectedState: state.AcceptingInit,
		},
		"wrong peer kind": {
			state:         state.ActivatingNodes,
			expectErr:     true,
			expectedState: state.ActivatingNodes,
		},
		"GetK8sJoinArgs error": {
			state:          state.NodeWaitingForClusterJoin,
			getJoinArgsErr: someErr,
			expectErr:      true,
			expectedState:  state.NodeWaitingForClusterJoin,
		},
		"JoinCluster error": {
			state:          state.NodeWaitingForClusterJoin,
			joinClusterErr: someErr,
			expectErr:      true,
			expectedState:  state.Failed,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			logger := zaptest.NewLogger(t)
			core := &fakeCore{state: tc.state, joinClusterErr: tc.joinClusterErr}
			dialer := testdialer.NewBufconnDialer()

			api := New(logger, core, dialer, nil, nil, nil)

			vserver := grpc.NewServer()
			vapi := &stubVPNAPI{
				joinArgs: kubeadm.BootstrapTokenDiscovery{
					APIServerEndpoint: "endp",
					Token:             "token",
					CACertHashes:      []string{"dis"},
				},
				getJoinArgsErr: tc.getJoinArgsErr,
			}
			vpnproto.RegisterAPIServer(vserver, vapi)
			go vserver.Serve(dialer.GetListener(net.JoinHostPort(core.GetCoordinatorVPNIP(), vpnAPIPort)))
			defer vserver.GracefulStop()

			_, err := api.JoinCluster(context.Background(), &pubproto.JoinClusterRequest{})

			assert.Equal(tc.expectedState, core.state)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal([]kubeadm.BootstrapTokenDiscovery{vapi.joinArgs}, core.joinArgs)
		})
	}
}

type stubVPNAPI struct {
	peers          []peer.Peer
	getUpdateErr   error
	joinArgs       kubeadm.BootstrapTokenDiscovery
	getJoinArgsErr error
	vpnproto.UnimplementedAPIServer
}

func (a *stubVPNAPI) GetUpdate(ctx context.Context, in *vpnproto.GetUpdateRequest) (*vpnproto.GetUpdateResponse, error) {
	return &vpnproto.GetUpdateResponse{ResourceVersion: 1, Peers: peer.ToVPNProto(a.peers)}, a.getUpdateErr
}

func (a *stubVPNAPI) GetK8SJoinArgs(ctx context.Context, in *vpnproto.GetK8SJoinArgsRequest) (*vpnproto.GetK8SJoinArgsResponse, error) {
	return &vpnproto.GetK8SJoinArgsResponse{
		ApiServerEndpoint:        a.joinArgs.APIServerEndpoint,
		Token:                    a.joinArgs.Token,
		DiscoveryTokenCaCertHash: a.joinArgs.CACertHashes[0],
	}, a.getJoinArgsErr
}
