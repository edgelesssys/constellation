package core

import (
	"errors"
	"net"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGetPeers(t *testing.T) {
	peer1 := peer.Peer{PublicEndpoint: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: []byte{1, 2, 3}}
	peer2 := peer.Peer{PublicEndpoint: "192.0.2.12:2000", VPNIP: "192.0.2.22", VPNPubKey: []byte{2, 3, 4}}

	testCases := map[string]struct {
		storePeers      []peer.Peer
		resourceVersion int
		expectedPeers   []peer.Peer
	}{
		"request version 0": { // store has version 2
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 0,
			expectedPeers:   []peer.Peer{peer1, peer2},
		},
		"request version 1": {
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 1,
			expectedPeers:   []peer.Peer{peer1, peer2},
		},
		"request version 2": {
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 2,
			expectedPeers:   nil,
		},
		"request version 3": {
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 3,
			expectedPeers:   []peer.Peer{peer1, peer2},
		},
		"request version 4": {
			storePeers:      []peer.Peer{peer1, peer2},
			resourceVersion: 4,
			expectedPeers:   []peer.Peer{peer1, peer2},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			core, err := NewCore(&stubVPN{}, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil)
			require.NoError(err)

			// prepare store
			for _, p := range tc.storePeers {
				require.NoError(core.data().PutPeer(p))
			}
			require.NoError(core.data().IncrementPeersResourceVersion())

			resourceVersion, peers, err := core.GetPeers(tc.resourceVersion)
			require.NoError(err)

			assert.Equal(2, resourceVersion)
			assert.ElementsMatch(tc.expectedPeers, peers)
		})
	}
}

func TestAddPeer(t *testing.T) {
	someErr := errors.New("failed")
	testPeer := peer.Peer{
		PublicEndpoint: "192.0.2.11:2000",
		VPNIP:          "192.0.2.21",
		VPNPubKey:      []byte{2, 3, 4},
	}
	expectedVPNPeers := []stubVPNPeer{{
		pubKey:   testPeer.VPNPubKey,
		publicIP: "192.0.2.11",
		vpnIP:    testPeer.VPNIP,
	}}

	testCases := map[string]struct {
		peer               peer.Peer
		vpn                stubVPN
		expectErr          bool
		expectedVPNPeers   []stubVPNPeer
		expectedStorePeers []peer.Peer
	}{
		"add peer": {
			peer:               testPeer,
			expectedVPNPeers:   expectedVPNPeers,
			expectedStorePeers: []peer.Peer{testPeer},
		},
		"don't add self to vpn": {
			peer:               testPeer,
			vpn:                stubVPN{interfaceIP: testPeer.VPNIP},
			expectedStorePeers: []peer.Peer{testPeer},
		},
		"public endpoint without port": {
			peer: peer.Peer{
				PublicEndpoint: "192.0.2.11",
				VPNIP:          "192.0.2.21",
				VPNPubKey:      []byte{2, 3, 4},
			},
			expectErr: true,
		},
		"vpn add peer error": {
			peer:      testPeer,
			vpn:       stubVPN{addPeerErr: someErr},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			core, err := NewCore(&tc.vpn, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil)
			require.NoError(err)

			err = core.AddPeer(tc.peer)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(tc.expectedVPNPeers, tc.vpn.peers)

			actualStorePeers, err := core.data().GetPeers()
			require.NoError(err)
			assert.Equal(tc.expectedStorePeers, actualStorePeers)
		})
	}
}

func TestUpdatePeer(t *testing.T) {
	someErr := errors.New("failed")
	peer1 := peer.Peer{PublicEndpoint: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: []byte{1}}
	peer1KeyUpd := peer.Peer{PublicEndpoint: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: []byte{1, 1}}
	peer1EndpUpd := peer.Peer{PublicEndpoint: "192.0.2.110:2000", VPNIP: "192.0.2.21", VPNPubKey: []byte{1}}
	peer2 := peer.Peer{PublicEndpoint: "192.0.2.12:2000", VPNIP: "192.0.2.22", VPNPubKey: []byte{2}}
	peer3 := peer.Peer{PublicEndpoint: "192.0.2.13:2000", VPNIP: "192.0.2.23", VPNPubKey: []byte{3}}

	makeVPNPeers := func(peers ...peer.Peer) []stubVPNPeer {
		var result []stubVPNPeer
		for _, p := range peers {
			publicIP, _, err := net.SplitHostPort(p.PublicEndpoint)
			if err != nil {
				panic(err)
			}
			result = append(result, stubVPNPeer{pubKey: p.VPNPubKey, publicIP: publicIP, vpnIP: p.VPNIP})
		}
		return result
	}

	testCases := map[string]struct {
		peers              []peer.Peer
		storePeers         []peer.Peer
		vpn                stubVPN
		expectErr          bool
		expectedVPNPeers   []stubVPNPeer
		expectedStorePeers []peer.Peer
	}{
		"basic": {
			peers:              []peer.Peer{peer1, peer3},
			storePeers:         []peer.Peer{peer1, peer2},
			vpn:                stubVPN{peers: makeVPNPeers(peer1, peer2)},
			expectedVPNPeers:   makeVPNPeers(peer1, peer3),
			expectedStorePeers: []peer.Peer{peer1, peer3},
		},
		"previously empty": {
			peers:              []peer.Peer{peer1, peer2},
			expectedVPNPeers:   makeVPNPeers(peer1, peer2),
			expectedStorePeers: []peer.Peer{peer1, peer2},
		},
		"no changes": {
			peers:              []peer.Peer{peer1, peer2},
			storePeers:         []peer.Peer{peer1, peer2},
			vpn:                stubVPN{peers: makeVPNPeers(peer1, peer2)},
			expectedVPNPeers:   makeVPNPeers(peer1, peer2),
			expectedStorePeers: []peer.Peer{peer1, peer2},
		},
		"key update": {
			peers:              []peer.Peer{peer1KeyUpd, peer3},
			storePeers:         []peer.Peer{peer1, peer2},
			vpn:                stubVPN{peers: makeVPNPeers(peer1, peer2)},
			expectedVPNPeers:   makeVPNPeers(peer1KeyUpd, peer3),
			expectedStorePeers: []peer.Peer{peer1KeyUpd, peer3},
		},
		"public endpoint update": {
			peers:              []peer.Peer{peer1EndpUpd, peer3},
			storePeers:         []peer.Peer{peer1, peer2},
			vpn:                stubVPN{peers: makeVPNPeers(peer1, peer2)},
			expectedVPNPeers:   makeVPNPeers(peer1EndpUpd, peer3),
			expectedStorePeers: []peer.Peer{peer1EndpUpd, peer3},
		},
		"don't add self": {
			peers:              []peer.Peer{peer1, peer3},
			storePeers:         []peer.Peer{peer1, peer2},
			vpn:                stubVPN{peers: makeVPNPeers(peer1, peer2), interfaceIP: peer3.VPNIP},
			expectedVPNPeers:   makeVPNPeers(peer1),
			expectedStorePeers: []peer.Peer{peer1},
		},
		"public endpoint without port": {
			peers: []peer.Peer{
				peer1,
				{
					PublicEndpoint: "192.0.2.13",
					VPNIP:          "192.0.2.23",
					VPNPubKey:      []byte{3},
				},
			},
			storePeers: []peer.Peer{peer1, peer2},
			vpn:        stubVPN{peers: makeVPNPeers(peer1, peer2)},
			expectErr:  true,
		},
		"vpn add peer error": {
			peers:      []peer.Peer{peer1, peer3},
			storePeers: []peer.Peer{peer1, peer2},
			vpn:        stubVPN{peers: makeVPNPeers(peer1, peer2), addPeerErr: someErr},
			expectErr:  true,
		},
		"vpn remove peer error": {
			peers:      []peer.Peer{peer1, peer3},
			storePeers: []peer.Peer{peer1, peer2},
			vpn:        stubVPN{peers: makeVPNPeers(peer1, peer2), removePeerErr: someErr},
			expectErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			core, err := NewCore(&tc.vpn, nil, nil, nil, nil, zaptest.NewLogger(t), nil, nil)
			require.NoError(err)

			// prepare store
			for _, p := range tc.storePeers {
				require.NoError(core.data().PutPeer(p))
			}

			updateErr := core.UpdatePeers(tc.peers)

			actualStorePeers, err := core.data().GetPeers()
			require.NoError(err)

			if tc.expectErr {
				assert.Error(updateErr)
				assert.ElementsMatch(tc.storePeers, actualStorePeers, "store has been changed despite failure")
				return
			}
			require.NoError(updateErr)

			assert.ElementsMatch(tc.expectedVPNPeers, tc.vpn.peers)
			assert.ElementsMatch(tc.expectedStorePeers, actualStorePeers)
		})
	}
}
