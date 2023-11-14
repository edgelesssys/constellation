/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package attestationconfigapi

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSEVSNPVersionListMarshalUnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		input    SEVSNPVersionList
		output   SEVSNPVersionList
		wantDiff bool
	}{
		"success": {
			input:  SEVSNPVersionList{list: []string{"v1", "v2"}},
			output: SEVSNPVersionList{list: []string{"v1", "v2"}},
		},
		"variant is lost": {
			input:  SEVSNPVersionList{list: []string{"v1", "v2"}, variant: variant.AzureSEVSNP{}},
			output: SEVSNPVersionList{list: []string{"v1", "v2"}},
		},
		"wrong order": {
			input:    SEVSNPVersionList{list: []string{"v1", "v2"}},
			output:   SEVSNPVersionList{list: []string{"v2", "v1"}},
			wantDiff: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			inputRaw, err := tc.input.MarshalJSON()
			require.NoError(t, err)

			var actual SEVSNPVersionList
			err = actual.UnmarshalJSON(inputRaw)
			require.NoError(t, err)

			if tc.wantDiff {
				assert.NotEqual(t, tc.output, actual, "Objects are equal, expected unequal")
			} else {
				assert.Equal(t, tc.output, actual, "Objects are not equal, expected equal")
			}
		})
	}
}

func TestSEVSNPVersionListAddVersion(t *testing.T) {
	tests := map[string]struct {
		versions []string
		new      string
		expected []string
	}{
		"success": {
			versions: []string{"v1", "v2"},
			new:      "v3",
			expected: []string{"v3", "v2", "v1"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			v := SEVSNPVersionList{list: tc.versions}
			v.addVersion(tc.new)

			assert.Equal(t, tc.expected, v.list)
		})
	}
}
