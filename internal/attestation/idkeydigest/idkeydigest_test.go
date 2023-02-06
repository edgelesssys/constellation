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
		wantErr  bool
	}{
		"digest struct": {
			json: `["57486a447ec0f1958002a22a06b7673b9fd27d11e1c6527498056054c5fa92d23c50f9de44072760fe2b6fb89740b696","0356215882a825279a85b300b0b742931d113bf7e32dde2e50ffde7ec743ca491ecdd7f336dc28a6e0b2bb57af7a44a3"]`,
			yaml: `
- "57486a447ec0f1958002a22a06b7673b9fd27d11e1c6527498056054c5fa92d23c50f9de44072760fe2b6fb89740b696"
- "0356215882a825279a85b300b0b742931d113bf7e32dde2e50ffde7ec743ca491ecdd7f336dc28a6e0b2bb57af7a44a3"`,
			wantDgst: IDKeyDigests{
				{0x57, 0x48, 0x6a, 0x44, 0x7e, 0xc0, 0xf1, 0x95, 0x80, 0x02, 0xa2, 0x2a, 0x06, 0xb7, 0x67, 0x3b, 0x9f, 0xd2, 0x7d, 0x11, 0xe1, 0xc6, 0x52, 0x74, 0x98, 0x05, 0x60, 0x54, 0xc5, 0xfa, 0x92, 0xd2, 0x3c, 0x50, 0xf9, 0xde, 0x44, 0x07, 0x27, 0x60, 0xfe, 0x2b, 0x6f, 0xb8, 0x97, 0x40, 0xb6, 0x96},
				{0x03, 0x56, 0x21, 0x58, 0x82, 0xa8, 0x25, 0x27, 0x9a, 0x85, 0xb3, 0x00, 0xb0, 0xb7, 0x42, 0x93, 0x1d, 0x11, 0x3b, 0xf7, 0xe3, 0x2d, 0xde, 0x2e, 0x50, 0xff, 0xde, 0x7e, 0xc7, 0x43, 0xca, 0x49, 0x1e, 0xcd, 0xd7, 0xf3, 0x36, 0xdc, 0x28, 0xa6, 0xe0, 0xb2, 0xbb, 0x57, 0xaf, 0x7a, 0x44, 0xa3},
			},
		},
		"legacy digest as string": {
			json:     `"57486a447ec0f1958002a22a06b7673b9fd27d11e1c6527498056054c5fa92d23c50f9de44072760fe2b6fb89740b696"`,
			yaml:     `"57486a447ec0f1958002a22a06b7673b9fd27d11e1c6527498056054c5fa92d23c50f9de44072760fe2b6fb89740b696"`,
			wantDgst: IDKeyDigests{{0x57, 0x48, 0x6a, 0x44, 0x7e, 0xc0, 0xf1, 0x95, 0x80, 0x02, 0xa2, 0x2a, 0x06, 0xb7, 0x67, 0x3b, 0x9f, 0xd2, 0x7d, 0x11, 0xe1, 0xc6, 0x52, 0x74, 0x98, 0x05, 0x60, 0x54, 0xc5, 0xfa, 0x92, 0xd2, 0x3c, 0x50, 0xf9, 0xde, 0x44, 0x07, 0x27, 0x60, 0xfe, 0x2b, 0x6f, 0xb8, 0x97, 0x40, 0xb6, 0x96}},
		},
		"invalid length": {
			json:     `"010203"`,
			yaml:     `"010203"`,
			wantDgst: IDKeyDigests{{}},
			wantErr:  true,
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
				if tc.wantErr {
					require.Error(err)
				} else {
					require.NoError(err)

					assert.Equal(tc.wantDgst, dgst)
				}
			}

			{
				// JSON
				var dgst IDKeyDigests
				err := json.Unmarshal([]byte(tc.json), &dgst)
				if tc.wantErr {
					require.Error(err)
				} else {
					require.NoError(err)

					assert.Equal(tc.wantDgst, dgst)
				}
			}
		})
	}
}

func TestEnforceIDKeyDigestMarshal(t *testing.T) {
	testCases := map[string]struct {
		input    EnforceIDKeyDigest
		wantJSON string
		wantYAML string
	}{
		"strict": {
			input:    StrictChecking,
			wantJSON: `"StrictChecking"`,
			wantYAML: "StrictChecking",
		},
		"maaFallback": {
			input:    MAAFallback,
			wantJSON: `"MAAFallback"`,
			wantYAML: "MAAFallback",
		},
		"warnOnly": {
			input:    WarnOnly,
			wantJSON: `"WarnOnly"`,
			wantYAML: "WarnOnly",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			{
				// YAML
				yaml, err := yaml.Marshal(tc.input)
				require.NoError(err)
				assert.YAMLEq(tc.wantYAML, string(yaml))
			}

			{
				// JSON
				json, err := json.Marshal(tc.input)
				require.NoError(err)
				assert.JSONEq(tc.wantJSON, string(json))
			}
		})
	}
}

func TestEnforceIDKeyDigestUnmarshal(t *testing.T) {
	testCases := map[string]struct {
		inputJSON string
		inputYAML string
		want      EnforceIDKeyDigest
		wantErr   bool
	}{
		"strict": {
			inputJSON: `"StrictChecking"`,
			inputYAML: "StrictChecking",
			want:      StrictChecking,
		},
		"maaFallback": {
			inputJSON: `"MAAFallback"`,
			inputYAML: "MAAFallback",
			want:      MAAFallback,
		},
		"warnOnly": {
			inputJSON: `"WarnOnly"`,
			inputYAML: "WarnOnly",
			want:      WarnOnly,
		},
		"legacyTrue": {
			inputJSON: `true`,
			inputYAML: "true",
			want:      StrictChecking,
		},
		"legacyFalse": {
			inputJSON: `false`,
			inputYAML: "false",
			want:      WarnOnly,
		},
		"invalid": {
			inputJSON: `"invalid"`,
			inputYAML: "invalid",
			wantErr:   true,
		},
		"invalidType": {
			inputJSON: `{"object": "invalid"}`,
			inputYAML: "object: invalid",
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			{
				// YAML
				var got EnforceIDKeyDigest
				err := yaml.Unmarshal([]byte(tc.inputYAML), &got)
				if tc.wantErr {
					assert.Error(err)
					return
				}

				require.NoError(err)
				assert.Equal(tc.want, got)
			}

			{
				// JSON
				var got EnforceIDKeyDigest
				err := json.Unmarshal([]byte(tc.inputJSON), &got)
				if tc.wantErr {
					assert.Error(err)
					return
				}

				require.NoError(err)
				assert.Equal(tc.want, got)
			}
		})
	}
}
