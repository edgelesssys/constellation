/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package k8sapi

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	kubeadmUtil "k8s.io/kubernetes/cmd/kubeadm/app/util"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestInitConfiguration(t *testing.T) {
	kubeadmConfig := KubdeadmConfiguration{}

	testCases := map[string]struct {
		config KubeadmInitYAML
	}{
		"kubeadm init config can be created": {
			config: kubeadmConfig.InitConfiguration(true, versions.VersionConfigs[versions.Default].ClusterVersion),
		},
		"kubeadm init config with all fields can be created": {
			config: func() KubeadmInitYAML {
				c := kubeadmConfig.InitConfiguration(true, versions.VersionConfigs[versions.Default].ClusterVersion)
				c.SetNodeIP("192.0.2.0")
				c.SetNodeName("node")
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
			var tmp KubeadmInitYAML
			require.NoError(kubernetes.UnmarshalK8SResources(config, &tmp))
			// test on correct mashalling and unmarshalling
			assert.Equal(tc.config.ClusterConfiguration, tmp.ClusterConfiguration)
			assert.Equal(tc.config.InitConfiguration, tmp.InitConfiguration)
		})
	}
}

func TestInitConfigurationKubeadmCompatibility(t *testing.T) {
	kubeadmConfig := KubdeadmConfiguration{}

	testCases := map[string]struct {
		config          KubeadmInitYAML
		expectedVersion string
		wantErr         bool
	}{
		"Kubeadm accepts version 'Latest'": {
			config:          kubeadmConfig.InitConfiguration(true, versions.VersionConfigs[versions.Default].ClusterVersion),
			expectedVersion: versions.VersionConfigs[versions.Default].ClusterVersion,
		},
		"Kubeadm receives incompatible version": {
			config:  kubeadmConfig.InitConfiguration(true, "1.19"),
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			version, err := kubeadmUtil.KubernetesReleaseVersion(tc.config.ClusterConfiguration.KubernetesVersion)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.Equal(tc.expectedVersion, version)
			assert.NoError(err)
		})
	}
}

func TestJoinConfiguration(t *testing.T) {
	kubdeadmConfig := KubdeadmConfiguration{}

	testCases := map[string]struct {
		config KubeadmJoinYAML
	}{
		"kubeadm join config can be created": {
			config: kubdeadmConfig.JoinConfiguration(true),
		},
		"kubeadm join config with all fields can be created": {
			config: func() KubeadmJoinYAML {
				c := kubdeadmConfig.JoinConfiguration(true)
				c.SetAPIServerEndpoint("192.0.2.0:6443")
				c.SetNodeIP("192.0.2.0")
				c.SetNodeName("node")
				c.SetToken("token")
				c.AppendDiscoveryTokenCaCertHash("discovery-token-ca-cert-hash")
				c.SetProviderID("somecloudprovider://instance-id")
				c.SetControlPlane("192.0.2.0")
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
			var tmp KubeadmJoinYAML
			require.NoError(kubernetes.UnmarshalK8SResources(config, &tmp))
			// test on correct mashalling and unmarshalling
			assert.Equal(tc.config.JoinConfiguration, tmp.JoinConfiguration)
		})
	}
}
