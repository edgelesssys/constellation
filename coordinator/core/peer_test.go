package core

import (
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGetPeers(t *testing.T) {
	peer1 := peer.Peer{PublicIP: "192.0.2.11", VPNIP: "192.0.2.21", VPNPubKey: []byte{1, 2, 3}}
	peer2 := peer.Peer{PublicIP: "192.0.2.12", VPNIP: "192.0.2.22", VPNPubKey: []byte{2, 3, 4}}

	testCases := map[string]struct {
		storePeers      []peer.Peer
		resourceVersion int
		wantPeers       []peer.Peer
	}{
		"request version 0": { // store has version 2
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 0,
			wantPeers:       []peer.Peer{peer1, peer2},
		},
		"request version 1": {
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 1,
			wantPeers:       []peer.Peer{peer1, peer2},
		},
		"request version 2": {
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 2,
			wantPeers:       nil,
		},
		"request version 3": {
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 3,
			wantPeers:       []peer.Peer{peer1, peer2},
		},
		"request version 4": {
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 4,
			wantPeers:       []peer.Peer{peer1, peer2},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			core, err := NewCore(&stubVPN{}, nil, nil, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(afero.NewMemMapFs()))
			require.NoError(err)

			// prepare store
			for _, p := range tc.storePeers {
				require.NoError(core.data().PutPeer(p))
			}
			require.NoError(core.data().IncrementPeersResourceVersion())

			resourceVersion, peers, err := core.GetPeers(tc.resourceVersion)
			require.NoError(err)

			assert.Equal(2, resourceVersion)
			assert.ElementsMatch(tc.wantPeers, peers)
		})
	}
}

func TestAddPeer(t *testing.T) {
	someErr := errors.New("failed")
	testPeer := peer.Peer{
		PublicIP:  "192.0.2.11",
		VPNIP:     "192.0.2.21",
		VPNPubKey: []byte{2, 3, 4},
	}
	wantVPNPeers := []stubVPNPeer{{
		pubKey:   testPeer.VPNPubKey,
		publicIP: "192.0.2.11",
		vpnIP:    testPeer.VPNIP,
	}}

	testCases := map[string]struct {
		peer           peer.Peer
		vpn            stubVPN
		wantErr        bool
		wantVPNPeers   []stubVPNPeer
		wantStorePeers []peer.Peer
	}{
		"add peer": {
			peer:           testPeer,
			wantVPNPeers:   wantVPNPeers,
			wantStorePeers: []peer.Peer{testPeer},
		},
		"don't add self to vpn": {
			peer:           testPeer,
			vpn:            stubVPN{interfaceIP: testPeer.VPNIP},
			wantStorePeers: []peer.Peer{testPeer},
		},
		"vpn add peer error": {
			peer:    testPeer,
			vpn:     stubVPN{addPeerErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			core, err := NewCore(&tc.vpn, nil, nil, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil, file.NewHandler(afero.NewMemMapFs()))
			require.NoError(err)

			err = core.AddPeer(tc.peer)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(tc.wantVPNPeers, tc.vpn.peers)

			actualStorePeers, err := core.data().GetPeers()
			require.NoError(err)
			assert.Equal(tc.wantStorePeers, actualStorePeers)
		})
	}
}
