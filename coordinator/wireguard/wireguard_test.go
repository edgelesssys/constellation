package wireguard

import (
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestUpdatePeer(t *testing.T) {
	requirePre := require.New(t)

	firstKey, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peer1 := peer.Peer{PublicIP: "192.0.2.11", VPNIP: "192.0.2.21", VPNPubKey: firstKey[:]}
	firstKeyUpd, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peer1KeyUpd := peer.Peer{PublicIP: "192.0.2.11", VPNIP: "192.0.2.21", VPNPubKey: firstKeyUpd[:]}
	secondKey, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peer2 := peer.Peer{PublicIP: "192.0.2.12", VPNIP: "192.0.2.22", VPNPubKey: secondKey[:]}
	thirdKey, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peer3 := peer.Peer{PublicIP: "192.0.2.13", VPNIP: "192.0.2.23", VPNPubKey: thirdKey[:]}
	fourthKey, err := wgtypes.GenerateKey()
	requirePre.NoError(err)
	peerAdmin := peer.Peer{PublicIP: "192.0.2.10", VPNIP: "192.0.2.25", VPNPubKey: fourthKey[:]}
	peerAdminNoEndp := peer.Peer{VPNIP: "192.0.2.25", VPNPubKey: fourthKey[:]}

	checkError := func(peers []wgtypes.Peer, err error) []wgtypes.Peer {
		requirePre.NoError(err)
		return peers
	}

	testCases := map[string]struct {
		storePeers   []peer.Peer
		vpnPeers     []wgtypes.Peer
		excludedIP   map[string]struct{}
		wantErr      bool
		wantVPNPeers []wgtypes.Peer
	}{
		"basic": {
			storePeers:   []peer.Peer{peer1, peer3},
			vpnPeers:     checkError(transformToWgpeer([]peer.Peer{peer1, peer2})),
			wantVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer1, peer3})),
		},
		"previously empty": {
			storePeers:   []peer.Peer{peer1, peer2},
			wantVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer1, peer2})),
		},
		"no changes": {
			storePeers:   []peer.Peer{peer1, peer2},
			vpnPeers:     checkError(transformToWgpeer([]peer.Peer{peer1, peer2})),
			wantVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer1, peer2})),
		},
		"key update": {
			storePeers:   []peer.Peer{peer1KeyUpd, peer3},
			vpnPeers:     checkError(transformToWgpeer([]peer.Peer{peer1, peer2})),
			wantVPNPeers: checkError(transformToWgpeer([]peer.Peer{peer1KeyUpd, peer3})),
		},
		"not update Endpoint changes": {
			storePeers:   []peer.Peer{peerAdminNoEndp, peer3},
			vpnPeers:     checkError(transformToWgpeer([]peer.Peer{peerAdmin, peer3})),
			wantVPNPeers: checkError(transformToWgpeer([]peer.Peer{peerAdmin, peer3})),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fakewg := fakewgClient{}
			fakewg.devices = make(map[string]*wgtypes.Device)
			wg := Wireguard{client: &fakewg}

			fakewg.devices[netInterface] = &wgtypes.Device{Peers: tc.vpnPeers}

			updateErr := wg.UpdatePeers(tc.storePeers)

			if tc.wantErr {
				assert.Error(updateErr)
				return
			}
			require.NoError(updateErr)

			assert.ElementsMatch(tc.wantVPNPeers, fakewg.devices[netInterface].Peers)
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
