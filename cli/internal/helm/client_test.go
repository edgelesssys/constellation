/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCRDs(t *testing.T) {
	testCases := map[string]struct {
		data    string
		wantErr bool
	}{
		"success": {
			data:    "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: nodeimages.update.edgeless.systems\nspec:\n  group: update.edgeless.systems\n  names:\n    kind: NodeImage\n",
			wantErr: false,
		},
		"wrong kind": {
			data:    "apiVersion: v1\nkind: Secret\ntype: Opaque\nmetadata:\n  name: supersecret\n  namespace: testNamespace\ndata:\n  data: YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE=\n",
			wantErr: true,
		},
		"decoding error": {
			data:    "asdf",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err := parseCRD([]byte(tc.data))
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestIsUpgrade(t *testing.T) {
	testCases := map[string]struct {
		currentVersion string
		newVersion     string
		wantUpgrade    bool
	}{
		"upgrade": {
			currentVersion: "0.1.0",
			newVersion:     "0.2.0",
			wantUpgrade:    true,
		},
		"downgrade": {
			currentVersion: "0.2.0",
			newVersion:     "0.1.0",
			wantUpgrade:    false,
		},
		"equal": {
			currentVersion: "0.1.0",
			newVersion:     "0.1.0",
			wantUpgrade:    false,
		},
		"invalid current version": {
			currentVersion: "asdf",
			newVersion:     "0.1.0",
			wantUpgrade:    false,
		},
		"invalid new version": {
			currentVersion: "0.1.0",
			newVersion:     "asdf",
			wantUpgrade:    false,
		},
		"patch version": {
			currentVersion: "0.1.0",
			newVersion:     "0.1.1",
			wantUpgrade:    true,
		},
		"pre-release version": {
			currentVersion: "0.1.0",
			newVersion:     "0.1.1-rc1",
			wantUpgrade:    true,
		},
		"pre-release version downgrade": {
			currentVersion: "0.1.1-rc1",
			newVersion:     "0.1.0",
			wantUpgrade:    false,
		},
		"pre-release of same version": {
			currentVersion: "0.1.0",
			newVersion:     "0.1.0-rc1",
			wantUpgrade:    false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			upgrade := isUpgrade(tc.currentVersion, tc.newVersion)
			assert.Equal(tc.wantUpgrade, upgrade)

			upgrade = isUpgrade("v"+tc.currentVersion, "v"+tc.newVersion)
			assert.Equal(tc.wantUpgrade, upgrade)
		})
	}
}
