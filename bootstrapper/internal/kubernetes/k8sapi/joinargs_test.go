/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package k8sapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

func TestParseJoinCommand(t *testing.T) {
	testCases := map[string]struct {
		joinCommand  string
		wantJoinArgs kubeadm.BootstrapTokenDiscovery
		wantErr      bool
	}{
		"join command can be parsed": {
			joinCommand: "kubeadm join 192.0.2.0:8443 --token dummy-token --discovery-token-ca-cert-hash sha512:dummy-hash --control-plane",
			wantJoinArgs: kubeadm.BootstrapTokenDiscovery{
				APIServerEndpoint: "192.0.2.0:8443",
				Token:             "dummy-token",
				CACertHashes:      []string{"sha512:dummy-hash"},
			},
			wantErr: false,
		},
		"incorrect join command returns error": {
			joinCommand: "some string",
			wantErr:     true,
		},
		"missing api server endpoint is checked": {
			joinCommand: "kubeadm join --token dummy-token --discovery-token-ca-cert-hash sha512:dummy-hash --control-plane",
			wantErr:     true,
		},
		"missing token is checked": {
			joinCommand: "kubeadm join 192.0.2.0:8443 --discovery-token-ca-cert-hash sha512:dummy-hash --control-plane",
			wantErr:     true,
		},
		"missing discovery-token-ca-cert-hash is checked": {
			joinCommand: "kubeadm join 192.0.2.0:8443 --token dummy-token --control-plane",
			wantErr:     true,
		},
		"missing control-plane": {
			joinCommand: "kubeadm join 192.0.2.0:8443 --token dummy-token --discovery-token-ca-cert-hash sha512:dummy-hash",
			wantJoinArgs: kubeadm.BootstrapTokenDiscovery{
				APIServerEndpoint: "192.0.2.0:8443",
				Token:             "dummy-token",
				CACertHashes:      []string{"sha512:dummy-hash"},
			},
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			joinArgs, err := ParseJoinCommand(tc.joinCommand)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(&tc.wantJoinArgs, joinArgs)
		})
	}
}
