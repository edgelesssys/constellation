/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vincent-petithory/dataurl"
)

func TestVersionFromDockerImage(t *testing.T) {
	testCases := map[string]struct {
		imageName   string
		wantVersion string
		wantPanic   bool
	}{
		"valid image name": {
			imageName:   "registry.test.foo/kube-apiserver:v1.18.0",
			wantVersion: "v1.18.0",
		},
		"valid image name with sha": {
			imageName:   "registry.test.foo/kube-apiserver:v1.18.0@sha256:1234567890abcdef",
			wantVersion: "v1.18.0",
		},
		"invalid image name": {
			imageName: "registry.test.foo/kube-apiserver",
			wantPanic: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			if tc.wantPanic {
				assert.Panics(func() { versionFromDockerImage(tc.imageName) })
			} else {
				assert.Equal(tc.wantVersion, versionFromDockerImage(tc.imageName))
			}
		})
	}
}

func TestKubernetesImagePatchCompatibility(t *testing.T) {
	// This test ensures that pinned Kubernetes images correspond to the
	// supported Kubernetes versions.
	for v, clusterConfig := range VersionConfigs {
		t.Run(string(v), func(t *testing.T) {
			for i, component := range clusterConfig.KubernetesComponents.GetUpgradableComponents() {
				if !strings.HasPrefix(component.Url, "data:") {
					// This test only applies to kubeadm patches.
					continue
				}
				if strings.Contains(component.InstallPath, "/etcd") {
					// The etcd version is not derived from the Kubernetes version
					continue
				}
				t.Run(fmt.Sprintf("%d-%s", i, path.Base(component.InstallPath)), func(t *testing.T) {
					require := require.New(t)
					dataURL, err := dataurl.DecodeString(component.Url)
					require.NoError(err)
					require.Contains(string(dataURL.Data), clusterConfig.ClusterVersion)
				})
			}
		})
	}
}
