package wireguard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/edgelesssys/constellation/coordinator/peer"
	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	netInterface = "wg0"
	port         = 51820
)

type Wireguard struct {
	client wgClient
}

func New() (*Wireguard, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	return &Wireguard{client: client}, nil
}

func (w *Wireguard) Setup(privKey []byte) error {
	var key wgtypes.Key
	var err error
	if len(privKey) == 0 {
		key, err = wgtypes.GeneratePrivateKey()
	} else {
		key, err = wgtypes.NewKey(privKey)
	}
	if err != nil {
		return err
	}
	listenPort := port
	return w.client.ConfigureDevice(netInterface, wgtypes.Config{PrivateKey: &key, ListenPort: &listenPort})
}

// GetPrivateKey returns the private key of the wireguard interface.
func (w *Wireguard) GetPrivateKey() ([]byte, error) {
	device, err := w.client.Device(netInterface)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve wireguard private key from device %v: %w", netInterface, err)
	}
	return device.PrivateKey[:], nil
}

func (w *Wireguard) DerivePublicKey(privKey []byte) ([]byte, error) {
	key, err := wgtypes.NewKey(privKey)
	if err != nil {
		return nil, err
	}
	pubkey := key.PublicKey()
	return pubkey[:], nil
}

func (w *Wireguard) GetPublicKey() ([]byte, error) {
	deviceData, err := w.client.Device(netInterface)
	if err != nil {
		return nil, err
	}
	return deviceData.PublicKey[:], nil
}

func (w *Wireguard) GetInterfaceIP() (string, error) {
	return util.GetInterfaceIP(netInterface)
}

// SetInterfaceIP sets the ip interface ip.
func (w *Wireguard) SetInterfaceIP(ip string) error {
	addr, err := netlink.ParseAddr(ip + "/16")
	if err != nil {
		return err
	}
	link, err := netlink.LinkByName(netInterface)
	if err != nil {
		return err
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return err
	}
	return netlink.LinkSetUp(link)
}

// AddPeer adds a new peer to a wireguard interface.
func (w *Wireguard) AddPeer(pubKey []byte, publicIP string, vpnIP string) error {
	_, allowedIPs, err := net.ParseCIDR(vpnIP + "/32")
	if err != nil {
		return err
	}

	key, err := wgtypes.NewKey(pubKey)
	if err != nil {
		return err
	}

	var endpoint *net.UDPAddr
	if ip := net.ParseIP(publicIP); ip != nil {
		endpoint = &net.UDPAddr{IP: ip, Port: port}
	}

	keepAlive := 10 * time.Second
	cfg := wgtypes.Config{
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:                   key,
				UpdateOnly:                  false,
				Endpoint:                    endpoint,
				AllowedIPs:                  []net.IPNet{*allowedIPs},
				PersistentKeepaliveInterval: &keepAlive,
			},
		},
	}

	return prettyWgError(w.client.ConfigureDevice(netInterface, cfg))
}

// RemovePeer removes a peer from the wireguard interface.
func (w *Wireguard) RemovePeer(pubKey []byte) error {
	key, err := wgtypes.NewKey(pubKey)
	if err != nil {
		return err
	}

	cfg := wgtypes.Config{Peers: []wgtypes.PeerConfig{{PublicKey: key, Remove: true}}}

	return prettyWgError(w.client.ConfigureDevice(netInterface, cfg))
}

func prettyWgError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("interface not found or is not a WireGuard interface")
	}
	return err
}

func (w *Wireguard) UpdatePeers(peers []peer.Peer) error {
	wgPeers, err := transformToWgpeer(peers)
	if err != nil {
		return fmt.Errorf("failed to transform peers to wireguard-peers: %w", err)
	}

	deviceData, err := w.client.Device(netInterface)
	if err != nil {
		return fmt.Errorf("failed to obtain device data: %w", err)
	}
	// convert to map for easier lookup
	storePeers := make(map[string]wgtypes.Peer)
	for _, p := range wgPeers {
		storePeers[p.AllowedIPs[0].String()] = p
	}
	var added []wgtypes.Peer
	var removed []wgtypes.Peer

	for _, interfacePeer := range deviceData.Peers {
		if updPeer, ok := storePeers[interfacePeer.AllowedIPs[0].String()]; ok {
			if !bytes.Equal(updPeer.PublicKey[:], interfacePeer.PublicKey[:]) {
				added = append(added, updPeer)
				removed = append(removed, interfacePeer)
			}
			delete(storePeers, updPeer.AllowedIPs[0].String())
		} else {
			removed = append(removed, interfacePeer)
		}
	}
	// remaining store peers are new ones
	for _, peer := range storePeers {
		added = append(added, peer)
	}

	keepAlive := 10 * time.Second
	var newPeerConfig []wgtypes.PeerConfig
	for _, peer := range removed {
		newPeerConfig = append(newPeerConfig, wgtypes.PeerConfig{
			// pub Key for remove matching is enought
			PublicKey: peer.PublicKey,
			Remove:    true,
		})
	}
	for _, peer := range added {
		newPeerConfig = append(newPeerConfig, wgtypes.PeerConfig{
			PublicKey:  peer.PublicKey,
			Remove:     false,
			UpdateOnly: false,
			Endpoint:   peer.Endpoint,
			AllowedIPs: peer.AllowedIPs,
			// needed, otherwise gRPC has problems establishing the initial connection.
			PersistentKeepaliveInterval: &keepAlive,
		})
	}
	if len(newPeerConfig) == 0 {
		return nil
	}
	cfg := wgtypes.Config{
		ReplacePeers: false,
		Peers:        newPeerConfig,
	}
	return prettyWgError(w.client.ConfigureDevice(netInterface, cfg))
}

func (w *Wireguard) Close() error {
	return w.client.Close()
}

// A wgClient is a type which can control a WireGuard device.
type wgClient interface {
	io.Closer
	Device(name string) (*wgtypes.Device, error)
	ConfigureDevice(name string, cfg wgtypes.Config) error
}

func transformToWgpeer(corePeers []peer.Peer) ([]wgtypes.Peer, error) {
	var wgPeers []wgtypes.Peer
	for _, peer := range corePeers {
		key, err := wgtypes.NewKey(peer.VPNPubKey)
		if err != nil {
			return nil, err
		}
		_, allowedIPs, err := net.ParseCIDR(peer.VPNIP + "/32")
		if err != nil {
			return nil, err
		}
		var endpoint *net.UDPAddr
		if ip := net.ParseIP(peer.PublicIP); ip != nil {
			endpoint = &net.UDPAddr{IP: ip, Port: port}
		}
		wgPeers = append(wgPeers, wgtypes.Peer{
			PublicKey:  key,
			Endpoint:   endpoint,
			AllowedIPs: []net.IPNet{*allowedIPs},
		})
	}
	return wgPeers, nil
}
