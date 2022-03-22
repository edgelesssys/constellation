package wireguard

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/edgelesssys/constellation/coordinator/util"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	netInterface = "wg0"
	port         = 51820
)

type Wireguard struct{}

func New() *Wireguard {
	return &Wireguard{}
}

func (w *Wireguard) Setup(privKey []byte) ([]byte, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to open wgctrl: %w", err)
	}
	defer client.Close()

	var key wgtypes.Key
	if len(privKey) == 0 {
		key, err = wgtypes.GeneratePrivateKey()
	} else {
		key, err = wgtypes.NewKey(privKey)
	}
	if err != nil {
		return nil, err
	}

	listenPort := port
	if err := client.ConfigureDevice(netInterface, wgtypes.Config{PrivateKey: &key, ListenPort: &listenPort}); err != nil {
		return nil, prettyWgError(err)
	}

	return key[:], nil
}

func (w *Wireguard) GetPublicKey(privKey []byte) ([]byte, error) {
	key, err := wgtypes.NewKey(privKey)
	if err != nil {
		return nil, err
	}
	pubkey := key.PublicKey()
	return pubkey[:], nil
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
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to open wgctrl: %w", err)
	}
	defer client.Close()

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

	return prettyWgError(client.ConfigureDevice(netInterface, cfg))
}

// RemovePeer removes a peer from the wireguard interface.
func (w *Wireguard) RemovePeer(pubKey []byte) error {
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to open wgctrl: %w", err)
	}
	defer client.Close()

	key, err := wgtypes.NewKey(pubKey)
	if err != nil {
		return err
	}

	cfg := wgtypes.Config{Peers: []wgtypes.PeerConfig{{PublicKey: key, Remove: true}}}

	return prettyWgError(client.ConfigureDevice(netInterface, cfg))
}

func prettyWgError(err error) error {
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("interface not found or is not a WireGuard interface")
	}
	return err
}
