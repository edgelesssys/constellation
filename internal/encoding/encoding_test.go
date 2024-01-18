/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package encoding

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMarshalHexBytes(t *testing.T) {
	testCases := map[string]struct {
		in           HexBytes
		expectedJSON string
		expectedYAML string
		wantErr      bool
	}{
		"success": {
			in:           []byte{0xab, 0xcd, 0xef},
			expectedJSON: "\"abcdef\"",
			expectedYAML: "abcdef\n",
		},
		"empty": {
			in:           []byte{},
			expectedJSON: "\"\"",
			expectedYAML: "\"\"\n",
		},
		"nil": {
			in:           nil,
			expectedJSON: "\"\"",
			expectedYAML: "\"\"\n",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			actualYAML, errYAML := yaml.Marshal(tc.in)
			actualJSON, errJSON := json.Marshal(tc.in)

			if tc.wantErr {
				assert.Error(errYAML)
				assert.Error(errJSON)
				return
			}
			assert.NoError(errYAML)
			assert.NoError(errJSON)
			assert.Equal(tc.expectedYAML, string(actualYAML), "yaml")
			assert.Equal(tc.expectedJSON, string(actualJSON), "json")
		})
	}
}

func TestUnmarshalHexBytes(t *testing.T) {
	testCases := map[string]struct {
		yamlString string
		jsonString string
		expected   HexBytes
		wantErr    bool
	}{
		"success": {
			yamlString: "abcdef",
			jsonString: "\"abcdef\"",
			expected:   []byte{0xab, 0xcd, 0xef},
		},
		"empty": {
			yamlString: "",
			jsonString: "\"\"",
			expected:   nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			var actualYAML HexBytes
			errYAML := yaml.Unmarshal([]byte(tc.yamlString), &actualYAML)
			var actualJSON HexBytes
			errJSON := json.Unmarshal([]byte(tc.jsonString), &actualJSON)

			if tc.wantErr {
				assert.Error(errYAML)
				assert.Error(errJSON)
				return
			}
			assert.NoError(errYAML)
			assert.NoError(errJSON)
			assert.Equal(tc.expected, actualYAML, "yaml")
			assert.Equal(tc.expected, actualJSON, "json")
		})
	}
}

func TestMarshalUnmarshalHexBytes(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	in := HexBytes{0xab, 0xcd, 0xef}
	expectedJSON := "\"abcdef\""
	expectedYAML := "abcdef\n"

	actualJSON, err := json.Marshal(in)
	require.NoError(err)
	assert.Equal(expectedJSON, string(actualJSON))
	actualYAML, err := yaml.Marshal(in)
	require.NoError(err)
	assert.Equal(expectedYAML, string(actualYAML))

	var actualJSON2 HexBytes
	err = json.Unmarshal(actualJSON, &actualJSON2)
	require.NoError(err)
	assert.Equal(in, actualJSON2)
	var actualYAML2 HexBytes
	err = yaml.Unmarshal(actualYAML, &actualYAML2)
	require.NoError(err)
	assert.Equal(in, actualYAML2)
}
