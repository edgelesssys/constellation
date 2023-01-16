/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package idkeydigest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMarshal(t *testing.T) {
	testCases := map[string]struct {
		dgst     IDKeyDigests
		wantYAML string
		wantJSON string
	}{
		"digest": {
			dgst:     IDKeyDigests{{0x01, 0x02, 0x03, 0x04}, {0xff, 0xff, 0xff, 0xff}},
			wantJSON: `["01020304","ffffffff"]`,
			wantYAML: `
- "01020304"
- "ffffffff"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			{
				// YAML
				yaml, err := yaml.Marshal(tc.dgst)
				require.NoError(err)

				assert.YAMLEq(tc.wantYAML, string(yaml))
			}

			{
				// JSON
				json, err := json.Marshal(tc.dgst)
				require.NoError(err)

				assert.JSONEq(tc.wantJSON, string(json))
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	testCases := map[string]struct {
		yaml     string
		json     string
		wantDgst IDKeyDigests
	}{
		"digest struct": {
			json: `["01020304","ffffffff"]`,
			yaml: `
- "01020304"
- "ffffffff"`,
			wantDgst: IDKeyDigests{{0x01, 0x02, 0x03, 0x04}, {0xff, 0xff, 0xff, 0xff}},
		},
		"legacy digest as string": {
			json:     `"01020304"`,
			yaml:     `"01020304"`,
			wantDgst: IDKeyDigests{{0x01, 0x02, 0x03, 0x04}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			{
				// YAML
				var dgst IDKeyDigests
				err := yaml.Unmarshal([]byte(tc.yaml), &dgst)
				require.NoError(err)

				assert.Equal(tc.wantDgst, dgst)
			}

			{
				// JSON
				var dgst IDKeyDigests
				err := json.Unmarshal([]byte(tc.json), &dgst)
				require.NoError(err)

				assert.Equal(tc.wantDgst, dgst)
			}
		})
	}
}
