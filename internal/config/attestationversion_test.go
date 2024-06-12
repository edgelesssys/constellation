/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/encoding"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestVersionMarshalYAML(t *testing.T) {
	testCasesUint8 := map[string]struct {
		sut  AttestationVersion[uint8]
		want string
	}{
		"version with latest writes latest": {
			sut: AttestationVersion[uint8]{
				Value:      1,
				WantLatest: true,
			},
			want: "latest\n",
		},
		"value 5 writes 5": {
			sut: AttestationVersion[uint8]{
				Value:      5,
				WantLatest: false,
			},
			want: "5\n",
		},
	}
	for name, tc := range testCasesUint8 {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			bt, err := yaml.Marshal(tc.sut)
			assert.NoError(err)
			assert.Equal(tc.want, string(bt))
		})
	}

	testCasesUint16 := map[string]struct {
		sut  AttestationVersion[uint16]
		want string
	}{
		"version with latest writes latest": {
			sut: AttestationVersion[uint16]{
				Value:      1,
				WantLatest: true,
			},
			want: "latest\n",
		},
		"value 5 writes 5": {
			sut: AttestationVersion[uint16]{
				Value:      5,
				WantLatest: false,
			},
			want: "5\n",
		},
	}
	for name, tc := range testCasesUint16 {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			bt, err := yaml.Marshal(tc.sut)
			assert.NoError(err)
			assert.Equal(tc.want, string(bt))
		})
	}

	testCasesHexBytes := map[string]struct {
		sut  AttestationVersion[encoding.HexBytes]
		want string
	}{
		"version with latest writes latest": {
			sut: AttestationVersion[encoding.HexBytes]{
				Value:      encoding.HexBytes(bytes.Repeat([]byte("0"), 16)),
				WantLatest: true,
			},
			want: "latest\n",
		},
		"value 5 writes 5": {
			sut: AttestationVersion[encoding.HexBytes]{
				Value:      encoding.HexBytes(bytes.Repeat([]byte("A"), 16)),
				WantLatest: false,
			},
			want: "\"41414141414141414141414141414141\"\n",
		},
	}
	for name, tc := range testCasesHexBytes {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			bt, err := yaml.Marshal(tc.sut)
			assert.NoError(err)
			assert.Equal(tc.want, string(bt))
		})
	}
}

func TestVersionUnmarshal(t *testing.T) {
	testCasesUint8 := map[string]struct {
		yamlData string
		jsonData string
		want     AttestationVersion[uint8]
		wantErr  bool
	}{
		"latest resolves to isLatest": {
			yamlData: "latest",
			jsonData: "\"latest\"",
			want: AttestationVersion[uint8]{
				Value:      0,
				WantLatest: true,
			},
			wantErr: false,
		},
		"1 resolves to value 1": {
			yamlData: "1",
			jsonData: "1",
			want: AttestationVersion[uint8]{
				Value:      1,
				WantLatest: false,
			},
			wantErr: false,
		},
		"max uint8+1 errors": {
			yamlData: "256",
			jsonData: "256",
			wantErr:  true,
		},
		"-1 errors": {
			yamlData: "-1",
			jsonData: "-1",
			wantErr:  true,
		},
		"0 resolves to value 0": {
			yamlData: "0",
			jsonData: "0",
			want: AttestationVersion[uint8]{
				Value:      0,
				WantLatest: false,
			},
		},
	}
	for name, tc := range testCasesUint8 {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			{
				var sut AttestationVersion[uint8]
				err := yaml.Unmarshal([]byte(tc.yamlData), &sut)
				if tc.wantErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(tc.want, sut)
				}
			}

			{
				var sut AttestationVersion[uint8]
				err := json.Unmarshal([]byte(tc.jsonData), &sut)
				if tc.wantErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(tc.want, sut)
				}
			}
		})
	}

	testCasesUint16 := map[string]struct {
		yamlData string
		jsonData string
		want     AttestationVersion[uint16]
		wantErr  bool
	}{
		"latest resolves to isLatest": {
			yamlData: "latest",
			jsonData: "\"latest\"",
			want: AttestationVersion[uint16]{
				Value:      0,
				WantLatest: true,
			},
			wantErr: false,
		},
		"1 resolves to value 1": {
			yamlData: "1",
			jsonData: "1",
			want: AttestationVersion[uint16]{
				Value:      1,
				WantLatest: false,
			},
			wantErr: false,
		},
		"max uint16+1 errors": {
			yamlData: "65536",
			jsonData: "65536",
			wantErr:  true,
		},
		"-1 errors": {
			yamlData: "-1",
			jsonData: "-1",
			wantErr:  true,
		},
		"0 resolves to value 0": {
			yamlData: "0",
			jsonData: "0",
			want: AttestationVersion[uint16]{
				Value:      0,
				WantLatest: false,
			},
		},
	}
	for name, tc := range testCasesUint16 {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			{
				var sut AttestationVersion[uint16]
				err := yaml.Unmarshal([]byte(tc.yamlData), &sut)
				if tc.wantErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(tc.want, sut)
				}
			}

			{
				var sut AttestationVersion[uint16]
				err := json.Unmarshal([]byte(tc.jsonData), &sut)
				if tc.wantErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(tc.want, sut)
				}
			}
		})
	}

	testCasesHexBytes := map[string]struct {
		yamlData string
		jsonData string
		want     AttestationVersion[encoding.HexBytes]
		wantErr  bool
	}{
		"latest resolves to isLatest": {
			yamlData: "latest",
			jsonData: "\"latest\"",
			want: AttestationVersion[encoding.HexBytes]{
				Value:      encoding.HexBytes(nil),
				WantLatest: true,
			},
			wantErr: false,
		},
		"hex string resolves to correctly": {
			yamlData: "41414141414141414141414141414141",
			jsonData: "\"41414141414141414141414141414141\"",
			want: AttestationVersion[encoding.HexBytes]{
				Value:      encoding.HexBytes(bytes.Repeat([]byte("A"), 16)),
				WantLatest: false,
			},
			wantErr: false,
		},
		"invalid hex string": {
			yamlData: "GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG",
			jsonData: "\"GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG\"",
			wantErr:  true,
		},
		"non hex data": {
			yamlData: "-15",
			jsonData: "-15",
			wantErr:  true,
		},
	}
	for name, tc := range testCasesHexBytes {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			{
				var sut AttestationVersion[encoding.HexBytes]
				err := yaml.Unmarshal([]byte(tc.yamlData), &sut)
				if tc.wantErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(tc.want, sut)
				}
			}

			{
				var sut AttestationVersion[encoding.HexBytes]
				err := json.Unmarshal([]byte(tc.jsonData), &sut)
				if tc.wantErr {
					assert.Error(err)
				} else {
					assert.NoError(err)
					assert.Equal(tc.want, sut)
				}
			}
		})
	}
}
