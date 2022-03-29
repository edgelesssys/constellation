package vpn

import (
	"fmt"
	"net"
	"testing"

	wgquick "github.com/nmiculinic/wg-quick-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type stubNetworkLink struct {
	link netlink.Link
	addr string
	up   bool
}

func newStubNetworkLink() *stubNetworkLink {
	return &stubNetworkLink{}
}

func (s *stubNetworkLink) LinkAdd(link netlink.Link) error {
	s.link = link
	return nil
}

func (s *stubNetworkLink) LinkByName(name string) (netlink.Link, error) {
	if name != s.link.Attrs().Name {
		return nil, fmt.Errorf("could not find interface with name %v", name)
	}
	return s.link, nil
}

func (s *stubNetworkLink) ParseAddr(addr string) (*netlink.Addr, error) {
	return netlink.ParseAddr(addr)
}

func (s *stubNetworkLink) AddrAdd(link netlink.Link, addr *netlink.Addr) error {
	if link.Attrs().Name != s.link.Attrs().Name {
		return fmt.Errorf("could not find interface with name %v", link.Attrs().Name)
	}
	s.addr = addr.IP.String()
	return nil
}

func (s *stubNetworkLink) LinkSetUp(link netlink.Link) error {
	if link.Attrs().Name != s.link.Attrs().Name {
		return fmt.Errorf("could not find interface with name %v", link.Attrs().Name)
	}
	s.up = true
	return nil
}

type stubVPN struct {
	name   string
	config wgtypes.Config
}

func newStubVPN() *stubVPN {
	return &stubVPN{}
}

func (s *stubVPN) ConfigureDevice(name string, cfg wgtypes.Config) error {
	s.name = name
	s.config = cfg
	return nil
}

func TestConfigurer(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	link := newStubNetworkLink()
	vpn := newStubVPN()
	client, err := NewConfigurer(link, vpn)
	require.NoError(err)
	coordinatorPubKey, err := wgtypes.GenerateKey()
	require.NoError(err)
	clientPrivKey, err := wgtypes.GenerateKey()
	require.NoError(err)
	clientVpnIp := "192.0.2.1"
	coordinatorPubIp := "192.0.2.2"
	assert.NoError(client.Configure(clientVpnIp, coordinatorPubKey.String(), coordinatorPubIp, clientPrivKey.String()))

	// assert expected interface
	assert.Equal(interfaceName, link.link.Attrs().Name)
	assert.NotNil(link.addr)
	assert.True(link.up)

	// assert vpn config
	config := client.vpn.(*stubVPN).config
	assert.Equal(wireguardPort, *config.ListenPort)
	assert.Equal(clientPrivKey, *config.PrivateKey)
	assert.Less(0, len(config.Peers))
	assert.Equal(coordinatorPubKey, config.Peers[0].PublicKey)
	assert.Equal(net.JoinHostPort(coordinatorPubIp, "51820"), config.Peers[0].Endpoint.String())
	assert.Equal("10.118.0.1/32", config.Peers[0].AllowedIPs[0].String())
}

func TestNewConfig(t *testing.T) {
	require := require.New(t)

	testKey, err := wgtypes.GeneratePrivateKey()
	require.NoError(err)

	testCases := map[string]struct {
		coordinatorPubKey wgtypes.Key
		coordinatorPubIP  string
		clientPrivKey     wgtypes.Key
		wantErr           bool
	}{
		"valid": {
			coordinatorPubKey: testKey.PublicKey(),
			coordinatorPubIP:  "192.0.2.1",
			clientPrivKey:     testKey,
		},
		"empty coordinator pub ip": {
			coordinatorPubKey: testKey.PublicKey(),
			clientPrivKey:     testKey,
		},
		"empty coordinator public key": {
			coordinatorPubKey: wgtypes.Key{},
			coordinatorPubIP:  "192.0.2.1",
			clientPrivKey:     testKey,
			wantErr:           true,
		},
		"empty client private key": {
			coordinatorPubKey: testKey.PublicKey(),
			coordinatorPubIP:  "192.0.2.1",
			clientPrivKey:     wgtypes.Key{},
			wantErr:           true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			var coordinatorPubKeyStr, clientPrivKeyStr string
			if tc.coordinatorPubKey != (wgtypes.Key{}) {
				coordinatorPubKeyStr = tc.coordinatorPubKey.String()
			}
			if tc.clientPrivKey != (wgtypes.Key{}) {
				clientPrivKeyStr = tc.clientPrivKey.String()
			}
			config, err := NewConfig(coordinatorPubKeyStr, tc.coordinatorPubIP, clientPrivKeyStr)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.coordinatorPubKey, config.Peers[0].PublicKey)
				assert.Equal(tc.clientPrivKey, *config.PrivateKey)
			}
		})
	}
}

func TestNewWGQuickConfig(t *testing.T) {
	require := require.New(t)

	testKey, err := wgtypes.GeneratePrivateKey()
	require.NoError(err)
	testConfig := wgtypes.Config{
		PrivateKey: &testKey,
	}

	testCases := map[string]struct {
		config      wgtypes.Config
		clientVPNIP string
		wantErr     bool
	}{
		"valid config": {
			clientVPNIP: "192.0.2.1",
			config:      testConfig,
		},
		"empty client vpn ip": {
			config:  testConfig,
			wantErr: true,
		},
		"config without private key": {
			clientVPNIP: "192.0.2.1",
			config:      wgtypes.Config{},
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			quickFile, err := NewWGQuickConfig(tc.config, tc.clientVPNIP)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				var quickConfig wgquick.Config
				assert.NoError(quickConfig.UnmarshalText(quickFile))
				assert.Equal(tc.config.PrivateKey, quickConfig.PrivateKey)
				assert.Equal(tc.clientVPNIP, quickConfig.Address[0].IP.String())
			}
		})
	}
}
