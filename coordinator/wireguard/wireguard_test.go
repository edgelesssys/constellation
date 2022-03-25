package wireguard

import (
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestUpdatePeer(t *testing.T) {
	requirePre := require.New(t)

	firstKey, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peer1 := peer.Peer{PublicEndpoint: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: firstKey[:]}
	firstKeyUpd, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peer1KeyUpd := peer.Peer{PublicEndpoint: "192.0.2.11:2000", VPNIP: "192.0.2.21", VPNPubKey: firstKeyUpd[:]}
	peer1EndpUpd := peer.Peer{PublicEndpoint: "192.0.2.110:2000", VPNIP: "192.0.2.21", VPNPubKey: firstKey[:]}
	secondKey, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peer2 := peer.Peer{PublicEndpoint: "192.0.2.12:2000", VPNIP: "192.0.2.22", VPNPubKey: secondKey[:]}
	thirdKey, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peer3 := peer.Peer{PublicEndpoint: "192.0.2.13:2000", VPNIP: "192.0.2.23", VPNPubKey: thirdKey[:]}
	fourthKey, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peerSelf := peer.Peer{PublicEndpoint: "192.0.2.10:2000", VPNIP: "192.0.2.20", VPNPubKey: fourthKey[:]}

	checkError := func(peers []wgtypes.Peer, err error) []wgtypes.Peer {
		requirePre.NoError(err)
		return peers
	}

	testCases := map[string]struct {
		storePeers       []peer.Peer
		vpnPeers         []wgtypes.Peer
		expectErr        bool
		expectedVPNPeers []wgtypes.Peer
	}{
		"basic": {
			storePeers:       []peer.Peer{peer1, peer3},
			vpnPeers:         checkError(transformToWgpeer([]peer.Peer{peer1, peer2}, "")),
			expectedVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer1, peer3}, "")),
		},
		"previously empty": {
			storePeers:       []peer.Peer{peer1, peer2},
			expectedVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer1, peer2}, "")),
		},
		"no changes": {
			storePeers:       []peer.Peer{peer1, peer2},
			vpnPeers:         checkError(transformToWgpeer([]peer.Peer{peer1, peer2}, "")),
			expectedVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer1, peer2}, "")),
		},
		"key update": {
			storePeers:       []peer.Peer{peer1KeyUpd, peer3},
			vpnPeers:         checkError(transformToWgpeer([]peer.Peer{peer1, peer2}, "")),
			expectedVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer1KeyUpd, peer3}, "")),
		},
		"public endpoint update": {
			storePeers:       []peer.Peer{peer1EndpUpd, peer3},
			vpnPeers:         checkError(transformToWgpeer([]peer.Peer{peer1, peer2}, "")),
			expectedVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer1EndpUpd, peer3}, "")),
		},
		"dont add self": {
			storePeers:       []peer.Peer{peerSelf, peer3},
			vpnPeers:         checkError(transformToWgpeer([]peer.Peer{peer2, peer3}, "")),
			expectedVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer3}, "")),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fakewg := fakewgClient{}
			fakewg.devices = make(map[string]*wgtypes.Device)
			wg := Wireguard{client: &fakewg, getInterfaceIP: func(s string) (string, error) {
				return "192.0.2.20", nil
			}}

			fakewg.devices[netInterface] = &wgtypes.Device{Peers: tc.vpnPeers}

			updateErr := wg.UpdatePeers(tc.storePeers)

			if tc.expectErr {
				assert.Error(updateErr)
				return
			}
			require.NoError(updateErr)

			assert.ElementsMatch(tc.expectedVPNPeers, fakewg.devices[netInterface].Peers)
		})
	}
}

type fakewgClient struct {
	devices map[string]*wgtypes.Device
}

func (w *fakewgClient) Device(name string) (*wgtypes.Device, error) {
	if val, ok := w.devices[name]; ok {
		return val, nil
	}
	return nil, errors.New("device does not exist")
}

func (w *fakewgClient) ConfigureDevice(name string, cfg wgtypes.Config) error {
	var newPeerList []wgtypes.Peer
	var operation bool
	vpnPeers := make(map[wgtypes.Key]wgtypes.Peer)

	for _, peer := range w.devices[netInterface].Peers {
		vpnPeers[peer.PublicKey] = peer
	}

	for _, configPeer := range cfg.Peers {
		operation = false
		for _, vpnPeer := range w.devices[netInterface].Peers {
			// wireguard matches internally via pubkey
			if vpnPeer.PublicKey == configPeer.PublicKey {
				operation = true
				if configPeer.Remove {
					delete(vpnPeers, vpnPeer.PublicKey)
					continue
				}
				if configPeer.UpdateOnly {
					vpnPeers[vpnPeer.PublicKey] = wgtypes.Peer{
						PublicKey:  vpnPeer.PublicKey,
						AllowedIPs: vpnPeer.AllowedIPs,
						Endpoint:   configPeer.Endpoint,
					}
				}
			}
		}
		if !operation {
			vpnPeers[configPeer.PublicKey] = wgtypes.Peer{
				PublicKey:  configPeer.PublicKey,
				AllowedIPs: configPeer.AllowedIPs,
				Endpoint:   configPeer.Endpoint,
			}
		}
	}
	for _, peer := range vpnPeers {
		newPeerList = append(newPeerList, peer)
	}
	w.devices[netInterface].Peers = newPeerList

	return nil
}

func (w *fakewgClient) Close() error {
	return nil
}
