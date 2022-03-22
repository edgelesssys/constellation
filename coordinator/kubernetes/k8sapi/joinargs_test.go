package k8sapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestParseJoinCommand(t *testing.T) {
	testCases := map[string]struct {
		joinCommand      string
		expectedJoinArgs kubeadm.BootstrapTokenDiscovery
		expectErr        bool
	}{
		"join command can be parsed": {
			joinCommand: "kubeadm join 192.0.2.0:8443 --token dummy-token --discovery-token-ca-cert-hash sha512:dummy-hash --control-plane",
			expectedJoinArgs: kubeadm.BootstrapTokenDiscovery{
				APIServerEndpoint: "192.0.2.0:8443",
				Token:             "dummy-token",
				CACertHashes:      []string{"sha512:dummy-hash"},
			},
			expectErr: false,
		},
		"incorrect join command returns error": {
			joinCommand: "some string",
			expectErr:   true,
		},
		"missing api server endpoint is checked": {
			joinCommand: "kubeadm join --token dummy-token --discovery-token-ca-cert-hash sha512:dummy-hash --control-plane",
			expectErr:   true,
		},
		"missing token is checked": {
			joinCommand: "kubeadm join 192.0.2.0:8443 --discovery-token-ca-cert-hash sha512:dummy-hash --control-plane",
			expectErr:   true,
		},
		"missing discovery-token-ca-cert-hash is checked": {
			joinCommand: "kubeadm join 192.0.2.0:8443 --token dummy-token --control-plane",
			expectErr:   true,
		},
		"missing control-plane": {
			joinCommand: "kubeadm join 192.0.2.0:8443 --token dummy-token --discovery-token-ca-cert-hash sha512:dummy-hash",
			expectedJoinArgs: kubeadm.BootstrapTokenDiscovery{
				APIServerEndpoint: "192.0.2.0:8443",
				Token:             "dummy-token",
				CACertHashes:      []string{"sha512:dummy-hash"},
			},
			expectErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			joinArgs, err := ParseJoinCommand(tc.joinCommand)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(&tc.expectedJoinArgs, joinArgs)
		})
	}
}
