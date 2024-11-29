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

func TestVersionListMarshalUnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		input    List
		output   List
		wantDiff bool
	}{
		"success": {
			input:  List{List: []string{"v1", "v2"}},
			output: List{List: []string{"v1", "v2"}},
		},
		"variant is lost": {
			input:  List{List: []string{"v1", "v2"}, Variant: variant.AzureSEVSNP{}},
			output: List{List: []string{"v1", "v2"}},
		},
		"wrong order": {
			input:    List{List: []string{"v1", "v2"}},
			output:   List{List: []string{"v2", "v1"}},
			wantDiff: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			inputRaw, err := tc.input.MarshalJSON()
			require.NoError(t, err)

			var actual List
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

func TestVersionListAddVersion(t *testing.T) {
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
			v := List{List: tc.versions}
			v.AddVersion(tc.new)

			assert.Equal(t, tc.expected, v.List)
		})
	}
}
