/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
