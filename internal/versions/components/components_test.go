/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package components

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	crictl = &Component{
		Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.28.0/crictl-v1.28.0-linux-amd64.tar.gz", // renovate:crictl-release
		Hash:        "sha256:8dc78774f7cbeaf787994d386eec663f0a3cf24de1ea4893598096cb39ef2508",
		InstallPath: "/bin",
		Extract:     true,
	}
	kubeadm = &Component{
		Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.11/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
		Hash:        "sha256:58f886e39e517ba1a92493f136e80f1b6ea9362966ad9d2accdf2133004161f2",
		InstallPath: "/bin/kubeadm",
	}
	kubeapi = &Component{
		Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI2LjExQHNoYTI1NjozOTUzNWQwZWZlODk1YWU5MWI1NTExZmRhZGI1MmVjOTMyOWYzODk4NzYxMTYzYThjMGRlMjAzZTIzZTMzODUzIn1d",
		InstallPath: "/etc/kubernetes/patches/kube-apiserver+json.json",
	}
)

func TestGetKubeadmComponent(t *testing.T) {
	assert := assert.New(t)
	got, err := Components{crictl, kubeadm, kubeapi}.GetKubeadmComponent()
	assert.NoError(err)
	assert.Same(kubeadm, got)
}

func TestGetKubeadmComponentWithoutKubeadm(t *testing.T) {
	assert := assert.New(t)
	got, err := Components{crictl, kubeapi}.GetKubeadmComponent()
	assert.Error(err)
	assert.Nil(got)
}

func TestGetUpgradableComponents(t *testing.T) {
	assert := assert.New(t)
	got := Components{crictl, kubeadm, kubeapi}.GetUpgradableComponents()
	assert.Contains(got, kubeadm)
	assert.Contains(got, kubeapi)
	assert.NotContains(got, crictl)
}

func TestUnmarshalComponents(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	legacyFormat := `[{"URL":"https://example.com/foo.tar.gz","Hash":"1234567890","InstallPath":"/foo","Extract":true}]`
	newFormat := `[{"url":"https://example.com/foo.tar.gz","hash":"1234567890","install_path":"/foo","extract":true}]`

	var fromLegacy Components
	require.NoError(json.Unmarshal([]byte(legacyFormat), &fromLegacy))

	var fromNew Components
	require.NoError(json.Unmarshal([]byte(newFormat), &fromNew))

	assert.Equal(fromLegacy, fromNew)
}
