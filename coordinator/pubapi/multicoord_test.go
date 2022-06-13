package pubapi

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/logging"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/coordinator/state"
	"github.com/edgelesssys/constellation/internal/grpc/dialer"
	"github.com/edgelesssys/constellation/internal/grpc/testdialer"
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
		wantState                  state.State
		vpnapi                     stubVPNAPI
		wantErr                    bool
		switchToPersistentStoreErr error
		k8sJoinargsErr             error
		k8sCertKeyErr              error
	}{
		"basic": {
			coordinators: testCoord1,
			state:        state.AcceptingInit,
			wantState:    state.ActivatingNodes,
			vpnapi:       stubVPN,
		},
		"already activated": {
			state:     state.ActivatingNodes,
			wantErr:   true,
			wantState: state.ActivatingNodes,
			vpnapi:    stubVPN,
		},
		"SwitchToPersistentStore error": {
			coordinators:               testCoord1,
			state:                      state.AcceptingInit,
			switchToPersistentStoreErr: someErr,
			wantErr:                    true,
			wantState:                  state.Failed,
			vpnapi:                     stubVPN,
		},
		"GetK8SJoinArgs error": {
			coordinators:               testCoord1,
			state:                      state.AcceptingInit,
			switchToPersistentStoreErr: someErr,
			wantErr:                    true,
			wantState:                  state.Failed,
			vpnapi:                     stubVPN,
			k8sJoinargsErr:             someErr,
		},
		"GetK8SCertificateKeyErr error": {
			coordinators:               testCoord1,
			state:                      state.AcceptingInit,
			switchToPersistentStoreErr: someErr,
			wantErr:                    true,
			wantState:                  state.Failed,
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
			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, fakeValidator{}, netDialer)

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), &logging.NopLogger{}, core, dialer, stubVPNAPIServer{}, getPublicIPAddr, nil)
			defer api.Close()

			// spawn vpnServer
			vpnapiServer := tc.vpnapi.newServer()
			go vpnapiServer.Serve(netDialer.GetListener(net.JoinHostPort(tc.coordinators.peer.VPNIP, vpnAPIPort)))
			defer vpnapiServer.GracefulStop()

			_, err := api.ActivateAsAdditionalCoordinator(context.Background(), &pubproto.ActivateAsAdditionalCoordinatorRequest{
				AssignedVpnIp:             "10.118.0.2",
				ActivatingCoordinatorData: peer.ToPubProto([]peer.Peer{tc.coordinators.peer})[0],
				OwnerId:                   core.ownerID,
				ClusterId:                 core.clusterID,
			})

			assert.Equal(tc.wantState, core.state)

			if tc.wantErr {
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
		wantErr      bool
	}{
		"basic": {
			peers: peers,
			state: state.ActivatingNodes,
		},
		"not activated": {
			peers:   peers,
			state:   state.AcceptingInit,
			wantErr: true,
		},
		"wrong peer kind": {
			peers:   peers,
			state:   state.IsNode,
			wantErr: true,
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
			dialer := dialer.New(nil, fakeValidator{}, nil)

			api := New(logger, &logging.NopLogger{}, core, dialer, nil, nil, nil)

			_, err := api.TriggerCoordinatorUpdate(context.Background(), &pubproto.TriggerCoordinatorUpdateRequest{})
			if tc.wantErr {
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
		wantErr         bool
		wantState       state.State
	}{
		"basic": {
			coordinators: testCoord1,
			state:        state.ActivatingNodes,
			wantState:    state.ActivatingNodes,
		},
		"Activation Err": {
			coordinators: testCoord1,
			state:        state.ActivatingNodes,
			wantState:    state.ActivatingNodes,
			activateErr:  someErr,
			wantErr:      true,
		},
		"Not in exprected state": {
			coordinators: testCoord1,
			state:        state.AcceptingInit,
			wantState:    state.AcceptingInit,
			wantErr:      true,
		},
		"getPeerPublicKey error": {
			coordinators:    testCoord1,
			state:           state.ActivatingNodes,
			wantState:       state.ActivatingNodes,
			getPublicKeyErr: someErr,
			wantErr:         true,
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
			netDialer := testdialer.NewBufconnDialer()
			dialer := dialer.New(nil, fakeValidator{}, netDialer)

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), &logging.NopLogger{}, core, dialer, stubVPNAPIServer{}, getPublicIPAddr, nil)
			defer api.Close()

			// spawn coordinator
			tc.coordinators.activateErr = tc.activateErr
			tc.coordinators.getPubKeyErr = tc.getPublicKeyErr
			server := tc.coordinators.newServer()
			go server.Serve(netDialer.GetListener(net.JoinHostPort(tc.coordinators.peer.PublicIP, endpointAVPNPort)))
			defer server.GracefulStop()

			_, err := api.ActivateAdditionalCoordinator(context.Background(), &pubproto.ActivateAdditionalCoordinatorRequest{CoordinatorPublicIp: tc.coordinators.peer.PublicIP})

			assert.Equal(tc.wantState, core.state)

			if tc.wantErr {
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
		wantErr         bool
	}{
		"basic": {
			coordinator: testCoord,
		},
		"Activation Err": {
			coordinator:     testCoord,
			getVPNPubKeyErr: someErr,
			wantErr:         true,
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
			dialer := dialer.New(nil, fakeValidator{}, testdialer.NewBufconnDialer())

			getPublicIPAddr := func() (string, error) {
				return "192.0.2.1", nil
			}

			api := New(zaptest.NewLogger(t), &logging.NopLogger{}, core, dialer, stubVPNAPIServer{}, getPublicIPAddr, nil)
			defer api.Close()

			resp, err := api.GetPeerVPNPublicKey(context.Background(), &pubproto.GetPeerVPNPublicKeyRequest{})

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.coordinator.peer.VPNPubKey, resp.CoordinatorPubKey)
		})
	}
}
