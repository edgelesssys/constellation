package pubapi

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/coordinator/util/testdialer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestActivateAsAdditionalCoordinator(t *testing.T) {
	coordinatorPubKey := []byte{6, 7, 8}
	testCoord1 := stubPeer{peer: peer.Peer{PublicIP: "192.0.2.11", VPNPubKey: []byte{1, 2, 3}, VPNIP: "10.118.0.1", Role: role.Coordinator}}
	stubVPN := stubVPNAPI{joinArgs: kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: "endp",
		Token:             "token",
		CACertHashes:      []string{"dis"},
	}}

	someErr := errors.New("some error")
	testCases := map[string]struct {
		coordinators               stubPeer
		state                      state.State
		expectedState              state.State
		vpnapi                     stubVPNAPI
		expectErr                  bool
		switchToPersistentStoreErr error
		k8sJoinargsErr             error
		k8sCertKeyErr              error
	}{
		"basic": {
			coordinators:  testCoord1,
			state:         state.AcceptingInit,
			expectedState: state.ActivatingNodes,
			vpnapi:        stubVPN,
		},
		"already activated": {
			state:         state.ActivatingNodes,
			expectErr:     true,
			expectedState: state.ActivatingNodes,
			vpnapi:        stubVPN,
		},
		"SwitchToPersistentStore error": {
			coordinators:               testCoord1,
			state:                      state.AcceptingInit,
			switchToPersistentStoreErr: someErr,
			expectErr:                  true,
			expectedState:              state.Failed,
			vpnapi:                     stubVPN,
		},
		"GetK8SJoinArgs error": {
			coordinators:               testCoord1,
			state:                      state.AcceptingInit,
			switchToPersistentStoreErr: someErr,
			expectErr:                  true,
			expectedState:              state.Failed,
			vpnapi:                     stubVPN,
			k8sJoinargsErr:             someErr,
		},
		"GetK8SCertificateKeyErr error": {
			coordinators:               testCoord1,
			state:                      state.AcceptingInit,
			switchToPersistentStoreErr: someErr,
			expectErr:                  true,
			expectedState:              state.Failed,
			vpnapi:                     stubVPN,
			k8sCertKeyErr:              someErr,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			tc.vpnapi.getJoinArgsErr = tc.k8sJoinargsErr
			tc.vpnapi.getK8SCertKeyErr = tc.k8sCertKeyErr
			core := &fakeCore{
				state:                      tc.state,
				vpnPubKey:                  coordinatorPubKey,
				switchToPersistentStoreErr: tc.switchToPersistentStoreErr,
				kubeconfig:                 []byte("kubeconfig"),
				ownerID:                    []byte("ownerID"),
				clusterID:                  []byte("clusterID"),
			}
			dialer := testdialer.NewBufconnDialer()

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), core, dialer, stubVPNAPIServer{}, fakeValidator{}, getPublicIPAddr, nil)
			defer api.Close()

			// spawn vpnServer
			vpnapiServer := tc.vpnapi.newServer()
			go vpnapiServer.Serve(dialer.GetListener(net.JoinHostPort(tc.coordinators.peer.VPNIP, vpnAPIPort)))
			defer vpnapiServer.GracefulStop()

			_, err := api.ActivateAsAdditionalCoordinator(context.Background(), &pubproto.ActivateAsAdditionalCoordinatorRequest{
				AssignedVpnIp:             "10.118.0.2",
				ActivatingCoordinatorData: peer.ToPubProto([]peer.Peer{tc.coordinators.peer})[0],
				OwnerId:                   core.ownerID,
				ClusterId:                 core.clusterID,
			})

			assert.Equal(tc.expectedState, core.state)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestTriggerCoordinatorUpdate(t *testing.T) {
	// someErr := errors.New("failed")
	peers := []peer.Peer{
		{PublicIP: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: []byte{1, 2, 3}},
		{PublicIP: "192.0.2.12:2000", VPNIP: "192.0.2.22", VPNPubKey: []byte{2, 3, 4}},
	}

	testCases := map[string]struct {
		peers        []peer.Peer
		state        state.State
		getUpdateErr error
		expectErr    bool
	}{
		"basic": {
			peers: peers,
			state: state.ActivatingNodes,
		},
		"not activated": {
			peers:     peers,
			state:     state.AcceptingInit,
			expectErr: true,
		},
		"wrong peer kind": {
			peers:     peers,
			state:     state.IsNode,
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			logger := zaptest.NewLogger(t)
			core := &fakeCore{
				state: tc.state,
				peers: tc.peers,
			}
			dialer := testdialer.NewBufconnDialer()

			api := New(logger, core, dialer, nil, nil, nil, nil)

			_, err := api.TriggerCoordinatorUpdate(context.Background(), &pubproto.TriggerCoordinatorUpdateRequest{})
			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			// second update should be a noop
			_, err = api.TriggerCoordinatorUpdate(context.Background(), &pubproto.TriggerCoordinatorUpdateRequest{})
			require.NoError(err)

			require.Len(core.updatedPeers, 1)
			assert.Equal(tc.peers, core.updatedPeers[0])
		})
	}
}

func TestActivateAdditionalCoordinators(t *testing.T) {
	someErr := errors.New("failed")
	coordinatorPubKey := []byte{6, 7, 8}
	testCoord1 := stubPeer{peer: peer.Peer{PublicIP: "192.0.2.11", VPNPubKey: []byte{1, 2, 3}, VPNIP: "10.118.0.1", Role: role.Coordinator}}

	testCases := map[string]struct {
		coordinators    stubPeer
		state           state.State
		activateErr     error
		getPublicKeyErr error
		expectErr       bool
		expectedState   state.State
	}{
		"basic": {
			coordinators:  testCoord1,
			state:         state.ActivatingNodes,
			expectedState: state.ActivatingNodes,
		},
		"Activation Err": {
			coordinators:  testCoord1,
			state:         state.ActivatingNodes,
			expectedState: state.ActivatingNodes,
			activateErr:   someErr,
			expectErr:     true,
		},
		"Not in exprected state": {
			coordinators:  testCoord1,
			state:         state.AcceptingInit,
			expectedState: state.AcceptingInit,
			expectErr:     true,
		},
		"getPeerPublicKey error": {
			coordinators:    testCoord1,
			state:           state.ActivatingNodes,
			expectedState:   state.ActivatingNodes,
			getPublicKeyErr: someErr,
			expectErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			core := &fakeCore{
				state:      tc.state,
				vpnPubKey:  coordinatorPubKey,
				kubeconfig: []byte("kubeconfig"),
				ownerID:    []byte("ownerID"),
				clusterID:  []byte("clusterID"),
			}
			dialer := testdialer.NewBufconnDialer()

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), core, dialer, stubVPNAPIServer{}, fakeValidator{}, getPublicIPAddr, nil)
			defer api.Close()

			// spawn coordinator
			tc.coordinators.activateErr = tc.activateErr
			tc.coordinators.getPubKeyErr = tc.getPublicKeyErr
			server := tc.coordinators.newServer()
			go server.Serve(dialer.GetListener(net.JoinHostPort(tc.coordinators.peer.PublicIP, endpointAVPNPort)))
			defer server.GracefulStop()

			_, err := api.ActivateAdditionalCoordinator(context.Background(), &pubproto.ActivateAdditionalCoordinatorRequest{CoordinatorPublicIp: tc.coordinators.peer.PublicIP})

			assert.Equal(tc.expectedState, core.state)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestGetPeerVPNPublicKey(t *testing.T) {
	someErr := errors.New("failed")
	testCoord := stubPeer{peer: peer.Peer{PublicIP: "192.0.2.11", VPNPubKey: []byte{1, 2, 3}, VPNIP: "10.118.0.1", Role: role.Coordinator}}

	testCases := map[string]struct {
		coordinator     stubPeer
		getVPNPubKeyErr error
		expectErr       bool
	}{
		"basic": {
			coordinator: testCoord,
		},
		"Activation Err": {
			coordinator:     testCoord,
			getVPNPubKeyErr: someErr,
			expectErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			core := &fakeCore{
				vpnPubKey:       tc.coordinator.peer.VPNPubKey,
				getvpnPubKeyErr: tc.getVPNPubKeyErr,
			}
			dialer := testdialer.NewBufconnDialer()

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), core, dialer, stubVPNAPIServer{}, fakeValidator{}, getPublicIPAddr, nil)
			defer api.Close()

			resp, err := api.GetPeerVPNPublicKey(context.Background(), &pubproto.GetPeerVPNPublicKeyRequest{})

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.coordinator.peer.VPNPubKey, resp.CoordinatorPubKey)
		})
	}
}
