package core

import (
	"bytes"
	"errors"

	"github.com/edgelesssys/constellation/coordinator/peer"
)

type VPN interface {
	Setup(privKey []byte) ([]byte, error)
	GetPrivateKey() ([]byte, error)
	GetPublicKey(privKey []byte) ([]byte, error)
	GetInterfaceIP() (string, error)
	SetInterfaceIP(ip string) error
	AddPeer(pubKey []byte, publicIP string, vpnIP string) error
	RemovePeer(pubKey []byte) error
	UpdatePeers(peers []peer.Peer) error
}

type stubVPN struct {
	peers             []stubVPNPeer
	interfaceIP       string
	privateKey        []byte
	addPeerErr        error
	removePeerErr     error
	getInterfaceIPErr error
	getPrivateKeyErr  error
}

func (*stubVPN) Setup(privKey []byte) ([]byte, error) {
	return []byte{2, 3, 4}, nil
}

func (v *stubVPN) GetPrivateKey() ([]byte, error) {
	return v.privateKey, v.getPrivateKeyErr
}

func (*stubVPN) GetPublicKey(privKey []byte) ([]byte, error) {
	if bytes.Equal(privKey, []byte{2, 3, 4}) {
		return []byte{3, 4, 5}, nil
	}
	return nil, errors.New("unexpected privKey")
}

func (v *stubVPN) GetInterfaceIP() (string, error) {
	return v.interfaceIP, v.getInterfaceIPErr
}

func (v *stubVPN) SetInterfaceIP(ip string) error {
	v.interfaceIP = ip
	return nil
}

func (v *stubVPN) AddPeer(pubKey []byte, publicIP string, vpnIP string) error {
	v.peers = append(v.peers, stubVPNPeer{pubKey, publicIP, vpnIP})
	return v.addPeerErr
}

func (v *stubVPN) RemovePeer(pubKey []byte) error {
	newPeerList := make([]stubVPNPeer, 0, len(v.peers))
	for _, v := range v.peers {
		if !bytes.Equal(v.pubKey, pubKey) {
			newPeerList = append(newPeerList, v)
		}
	}
	v.peers = newPeerList
	return v.removePeerErr
}

func (v *stubVPN) UpdatePeers(peers []peer.Peer) error {
	for _, peer := range peers {
		if err := v.AddPeer(peer.VPNPubKey, peer.PublicIP, peer.VPNIP); err != nil {
			return err
		}
	}
	return nil
}

type stubVPNPeer struct {
	pubKey   []byte
	publicIP string
	vpnIP    string
}
