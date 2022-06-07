package vpn

import (
	"fmt"
	"net"
	"time"

	wgquick "github.com/nmiculinic/wg-quick-go"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	interfaceName = "wg0"
	wireguardPort = 51820
)

type ConfigHandler struct {
	up func(cfg *wgquick.Config, iface string) error
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{up: wgquick.Up}
}

func (h *ConfigHandler) Create(coordinatorPubKey, coordinatorPubIP, clientPrivKey, clientVPNIP string, mtu int) (*wgquick.Config, error) {
	return NewWGQuickConfig(coordinatorPubKey, coordinatorPubIP, clientPrivKey, clientVPNIP, mtu)
}

// Apply applies the generated WireGuard quick config.
func (h *ConfigHandler) Apply(conf *wgquick.Config) error {
	return h.up(conf, interfaceName)
}

// GetBytes returns the the bytes of the config.
func (h *ConfigHandler) Marshal(conf *wgquick.Config) ([]byte, error) {
	data, err := conf.MarshalText()
	if err != nil {
		return nil, fmt.Errorf("marshal wg-quick config: %w", err)
	}
	return data, nil
}

// newConfig creates a new WireGuard configuration.
func newConfig(coordinatorPubKey, coordinatorPubIP, clientPrivKey string) (wgtypes.Config, error) {
	_, allowedIPs, err := net.ParseCIDR("10.118.0.1/32")
	if err != nil {
		return wgtypes.Config{}, fmt.Errorf("parsing CIDR: %w", err)
	}

	coordinatorPubKeyParsed, err := wgtypes.ParseKey(coordinatorPubKey)
	if err != nil {
		return wgtypes.Config{}, fmt.Errorf("parsing coordinator public key: %w", err)
	}

	var endpoint *net.UDPAddr
	if ip := net.ParseIP(coordinatorPubIP); ip != nil {
		endpoint = &net.UDPAddr{IP: ip, Port: wireguardPort}
	} else {
		endpoint = nil
	}
	clientPrivKeyParsed, err := wgtypes.ParseKey(clientPrivKey)
	if err != nil {
		return wgtypes.Config{}, fmt.Errorf("parsing client private key: %w", err)
	}
	listenPort := wireguardPort

	keepAlive := 10 * time.Second
	return wgtypes.Config{
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
	}, nil
}

// NewWGQuickConfig create a new WireGuard wg-quick configuration file and mashals it to bytes.
func NewWGQuickConfig(coordinatorPubKey, coordinatorPubIP, clientPrivKey, clientVPNIP string, mtu int) (*wgquick.Config, error) {
	config, err := newConfig(coordinatorPubKey, coordinatorPubIP, clientPrivKey)
	if err != nil {
		return nil, err
	}

	clientIP := net.ParseIP(clientVPNIP)
	if clientIP == nil {
		return nil, fmt.Errorf("invalid client vpn ip '%s'", clientVPNIP)
	}
	quickfile := wgquick.Config{
		Config:  config,
		Address: []net.IPNet{{IP: clientIP, Mask: []byte{255, 255, 0, 0}}},
		MTU:     mtu,
	}
	return &quickfile, nil
}
