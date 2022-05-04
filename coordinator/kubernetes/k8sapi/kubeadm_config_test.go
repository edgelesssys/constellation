package k8sapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestInitConfiguration(t *testing.T) {
	awsConfig := AWSConfiguration{}
	coreOSConfig := CoreOSConfiguration{}

	testCases := map[string]struct {
		config KubeadmInitYAML
	}{
		"AWS init config can be created": {
			config: awsConfig.InitConfiguration(),
		},
		"AWS init config with all fields can be created": {
			config: func() KubeadmInitYAML {
				c := awsConfig.InitConfiguration()
				c.SetApiServerAdvertiseAddress("192.0.2.0")
				c.SetNodeIP("192.0.2.0")
				c.SetNodeName("node")
				c.SetPodNetworkCIDR("10.244.0.0/16")
				c.SetServiceCIDR("10.245.0.0/24")
				c.SetProviderID("somecloudprovider://instance-id")
				return c
			}(),
		},
		"CoreOS init config can be created": {
			config: coreOSConfig.InitConfiguration(true),
		},
		"CoreOS init config with all fields can be created": {
			config: func() KubeadmInitYAML {
				c := coreOSConfig.InitConfiguration(true)
				c.SetApiServerAdvertiseAddress("192.0.2.0")
				c.SetNodeIP("192.0.2.0")
				c.SetNodeName("node")
				c.SetPodNetworkCIDR("10.244.0.0/16")
				c.SetServiceCIDR("10.245.0.0/24")
				c.SetProviderID("somecloudprovider://instance-id")
				return c
			}(),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			config, err := tc.config.Marshal()
			require.NoError(err)
			tmp, err := tc.config.Unmarshal(config)
			require.NoError(err)
			// test on correct mashalling and unmarshalling
			assert.Equal(tc.config.ClusterConfiguration, tmp.ClusterConfiguration)
			assert.Equal(tc.config.InitConfiguration, tmp.InitConfiguration)
		})
	}
}

func TestJoinConfiguration(t *testing.T) {
	awsConfig := AWSConfiguration{}
	coreOSConfig := CoreOSConfiguration{}

	testCases := map[string]struct {
		config KubeadmJoinYAML
	}{
		"AWS join config can be created": {
			config: awsConfig.JoinConfiguration(),
		},
		"AWS join config with all fields can be created": {
			config: func() KubeadmJoinYAML {
				c := awsConfig.JoinConfiguration()
				c.SetApiServerEndpoint("192.0.2.0:6443")
				c.SetNodeIP("192.0.2.0")
				c.SetNodeName("node")
				c.SetToken("token")
				c.AppendDiscoveryTokenCaCertHash("discovery-token-ca-cert-hash")
				c.SetProviderID("somecloudprovider://instance-id")
				return c
			}(),
		},
		"CoreOS join config can be created": {
			config: coreOSConfig.JoinConfiguration(true),
		},
		"CoreOS join config with all fields can be created": {
			config: func() KubeadmJoinYAML {
				c := coreOSConfig.JoinConfiguration(true)
				c.SetApiServerEndpoint("192.0.2.0:6443")
				c.SetNodeIP("192.0.2.0")
				c.SetNodeName("node")
				c.SetToken("token")
				c.AppendDiscoveryTokenCaCertHash("discovery-token-ca-cert-hash")
				c.SetProviderID("somecloudprovider://instance-id")
				c.SetControlPlane("192.0.2.0", "11111111111111111111111111111111111")
				return c
			}(),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			config, err := tc.config.Marshal()
			require.NoError(err)
			tmp, err := tc.config.Unmarshal(config)
			require.NoError(err)
			// test on correct mashalling and unmarshalling
			assert.Equal(tc.config.JoinConfiguration, tmp.JoinConfiguration)
		})
	}
}
