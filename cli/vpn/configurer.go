package vpn

import (
	"net"
	"time"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	interfaceName = "wg0"
	wireguardPort = 51820
)

type vpn interface {
	ConfigureDevice(name string, cfg wgtypes.Config) error
}

type networkLink interface {
	LinkAdd(link netlink.Link) error
	LinkByName(name string) (netlink.Link, error)
	ParseAddr(s string) (*netlink.Addr, error)
	AddrAdd(link netlink.Link, addr *netlink.Addr) error
	LinkSetUp(link netlink.Link) error
}

type netLink struct{}

func newNetLink() *netLink {
	return &netLink{}
}

func (n *netLink) LinkAdd(link netlink.Link) error {
	return netlink.LinkAdd(link)
}

func (n *netLink) LinkByName(name string) (netlink.Link, error) {
	return netlink.LinkByName(name)
}

func (n *netLink) ParseAddr(s string) (*netlink.Addr, error) {
	return netlink.ParseAddr(s)
}

func (n *netLink) AddrAdd(link netlink.Link, addr *netlink.Addr) error {
	return netlink.AddrAdd(link, addr)
}

func (n *netLink) LinkSetUp(link netlink.Link) error {
	return netlink.LinkSetUp(link)
}

type Configurer struct {
	netLink networkLink
	vpn     vpn
}

// NewWithDefaults creates a new vpn client.
func NewWithDefaults() (*Configurer, error) {
	vpn, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	return &Configurer{netLink: newNetLink(), vpn: vpn}, nil
}

// New creates a new vpn client with the provided
// network link and vpn interface.
func New(netLink networkLink, vpn vpn) (*Configurer, error) {
	return &Configurer{netLink: netLink, vpn: vpn}, nil
}

// Configure configures a WireGuard interface
// WireGuard will listen on its default port.
// The peer must have the IP 10.118.0.1 in the vpn.
func (c *Configurer) Configure(clientVpnIp, coordinatorPubKey, coordinatorPubIP, clientPrivKey string) error {
	if err := c.netLink.LinkAdd(&netlink.Wireguard{LinkAttrs: netlink.LinkAttrs{Name: interfaceName}}); err != nil {
		return err
	}

	link, err := c.netLink.LinkByName(interfaceName)
	if err != nil {
		return err
	}
	addr, err := c.netLink.ParseAddr(clientVpnIp + "/16")
	if err != nil {
		return err
	}
	if err := c.netLink.AddrAdd(link, addr); err != nil {
		return err
	}
	if err := c.netLink.LinkSetUp(link); err != nil {
		return err
	}

	_, allowedIPs, err := net.ParseCIDR("10.118.0.1/32")
	if err != nil {
		return err
	}

	coordinatorPubKeyParsed, err := wgtypes.ParseKey(coordinatorPubKey)
	if err != nil {
		return err
	}

	var endpoint *net.UDPAddr
	if ip := net.ParseIP(coordinatorPubIP); ip != nil {
		endpoint = &net.UDPAddr{IP: ip, Port: wireguardPort}
	} else {
		endpoint = nil
	}
	clientPrivKeyParsed, err := wgtypes.ParseKey(clientPrivKey)
	if err != nil {
		return err
	}
	listenPort := wireguardPort

	keepAlive := 10 * time.Second
	err = c.vpn.ConfigureDevice(interfaceName, wgtypes.Config{
		PrivateKey:   &clientPrivKeyParsed,
		ListenPort:   &listenPort,
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:                   coordinatorPubKeyParsed,
				UpdateOnly:                  false,
				Endpoint:                    endpoint,
				AllowedIPs:                  []net.IPNet{*allowedIPs},
				PersistentKeepaliveInterval: &keepAlive,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
