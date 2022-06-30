package vpn

import (
	"errors"
	"testing"

	wgquick "github.com/nmiculinic/wg-quick-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestCreate(t *testing.T) {
	require := require.New(t)

	testKey, err := wgtypes.GeneratePrivateKey()
	require.NoError(err)

	testCases := map[string]struct {
		coordinatorPubKey string
		coordinatorPubIP  string
		clientPrivKey     string
		clientVPNIP       string
		wantErr           bool
	}{
		"valid config": {
			clientPrivKey:     testKey.String(),
			clientVPNIP:       "192.0.2.1",
			coordinatorPubKey: testKey.PublicKey().String(),
			coordinatorPubIP:  "192.0.2.1",
		},
		"valid missing endpoint": {
			clientPrivKey:     testKey.String(),
			clientVPNIP:       "192.0.2.1",
			coordinatorPubKey: testKey.PublicKey().String(),
		},
		"invalid coordinator pub key": {
			clientPrivKey:    testKey.String(),
			clientVPNIP:      "192.0.2.1",
			coordinatorPubIP: "192.0.2.1",
			wantErr:          true,
		},
		"invalid client priv key": {
			clientVPNIP:       "192.0.2.1",
			coordinatorPubKey: testKey.PublicKey().String(),
			coordinatorPubIP:  "192.0.2.1",
			wantErr:           true,
		},
		"invalid client ip": {
			clientPrivKey:     testKey.String(),
			coordinatorPubKey: testKey.PublicKey().String(),
			coordinatorPubIP:  "192.0.2.1",
			wantErr:           true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			handler := &ConfigHandler{}
			const mtu = 2

			quickConfig, err := handler.Create(tc.coordinatorPubKey, tc.coordinatorPubIP, tc.clientPrivKey, tc.clientVPNIP, mtu)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.clientPrivKey, quickConfig.PrivateKey.String())
				assert.Equal(tc.clientVPNIP, quickConfig.Address[0].IP.String())

				if tc.coordinatorPubIP != "" {
					assert.Equal(tc.coordinatorPubIP, quickConfig.Peers[0].Endpoint.IP.String())
				}
				assert.Equal(mtu, quickConfig.MTU)
			}
		})
	}
}

func TestApply(t *testing.T) {
	testKey, err := wgtypes.GeneratePrivateKey()
	require.NoError(t, err)

	testCases := map[string]struct {
		quickConfig *wgquick.Config
		upErr       error
		wantErr     bool
	}{
		"valid": {
			quickConfig: &wgquick.Config{Config: wgtypes.Config{PrivateKey: &testKey}},
		},
		"invalid apply": {
			quickConfig: &wgquick.Config{Config: wgtypes.Config{PrivateKey: &testKey}},
			upErr:       errors.New("some err"),
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			var ifaceSpy string
			var cfgSpy *wgquick.Config
			upSpy := func(cfg *wgquick.Config, iface string) error {
				ifaceSpy = iface
				cfgSpy = cfg
				return tc.upErr
			}

			handler := &ConfigHandler{up: upSpy}

			err := handler.Apply(tc.quickConfig)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(interfaceName, ifaceSpy)
				assert.Equal(tc.quickConfig, cfgSpy)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	require := require.New(t)

	testKey, err := wgtypes.GeneratePrivateKey()
	require.NoError(err)

	testCases := map[string]struct {
		quickConfig *wgquick.Config
		wantErr     bool
	}{
		"valid": {
			quickConfig: &wgquick.Config{Config: wgtypes.Config{PrivateKey: &testKey}},
		},
		"invalid config": {
			quickConfig: &wgquick.Config{Config: wgtypes.Config{}},
			wantErr:     true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			handler := &ConfigHandler{}

			data, err := handler.Marshal(tc.quickConfig)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Greater(len(data), 0)
			}
		})
	}
}
