/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestVersionMarshalYAML(t *testing.T) {
	tests := map[string]struct {
		sut  AttestationVersion
		want string
	}{
		"isLatest resolves to latest": {
			sut: AttestationVersion{
				Value:      1,
				WantLatest: true,
			},
			want: "latest\n",
		},
		"value 5 resolves to 5": {
			sut: AttestationVersion{
				Value:      5,
				WantLatest: false,
			},
			want: "5\n",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			bt, err := yaml.Marshal(tc.sut)
			require.NoError(err)
			require.Equal(tc.want, string(bt))
		})
	}
}

func TestVersionUnmarshalYAML(t *testing.T) {
	tests := map[string]struct {
		sut     string
		want    AttestationVersion
		wantErr bool
	}{
		"latest resolves to isLatest": {
			sut: "latest",
			want: AttestationVersion{
				Value:      0,
				WantLatest: true,
			},
			wantErr: false,
		},
		"1 resolves to value 1": {
			sut: "1",
			want: AttestationVersion{
				Value:      1,
				WantLatest: false,
			},
			wantErr: false,
		},
		"max uint8+1 errors": {
			sut:     "256",
			wantErr: true,
		},
		"-1 errors": {
			sut:     "-1",
			wantErr: true,
		},
		"2.6 errors": {
			sut:     "2.6",
			wantErr: true,
		},
		"2.0 errors": {
			sut:     "2.0",
			wantErr: true,
		},
		"hex format is invalid": {
			sut:     "0x10",
			wantErr: true,
		},
		"octal format is invalid": {
			sut:     "010",
			wantErr: true,
		},
		"0 resolves to value 0": {
			sut: "0",
			want: AttestationVersion{
				Value:      0,
				WantLatest: false,
			},
		},
		"00 errors": {
			sut:     "00",
			wantErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			var sut AttestationVersion
			err := yaml.Unmarshal([]byte(tc.sut), &sut)
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			require.Equal(tc.want, sut)
		})
	}
}

func TestVersionUnmarshalJSON(t *testing.T) {
	tests := map[string]struct {
		sut     string
		want    AttestationVersion
		wantErr bool
	}{
		"latest resolves to isLatest": {
			sut: `"latest"`,
			want: AttestationVersion{
				Value:      0,
				WantLatest: true,
			},
		},
		"1 resolves to value 1": {
			sut: "1",
			want: AttestationVersion{
				Value:      1,
				WantLatest: false,
			},
		},
		"quoted number resolves to value": {
			sut: `"1"`,
			want: AttestationVersion{
				Value:      1,
				WantLatest: false,
			},
		},
		"quoted float errors": {
			sut:     `"1.0"`,
			wantErr: true,
		},
		"max uint8+1 errors": {
			sut:     "256",
			wantErr: true,
		},
		"-1 errors": {
			sut:     "-1",
			wantErr: true,
		},
		"2.6 errors": {
			sut:     "2.6",
			wantErr: true,
		},
		"2.0 errors": {
			sut:     "2.0",
			wantErr: true,
		},
		"hex format is invalid": {
			sut:     "0x10",
			wantErr: true,
		},
		"octal format is invalid": {
			sut:     "010",
			wantErr: true,
		},
		"0 resolves to value 0": {
			sut: "0",
			want: AttestationVersion{
				Value:      0,
				WantLatest: false,
			},
		},
		"quoted 0 resolves to value 0": {
			sut: `"0"`,
			want: AttestationVersion{
				Value:      0,
				WantLatest: false,
			},
		},
		"00 errors": {
			sut:     "00",
			wantErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			var sut AttestationVersion
			err := json.Unmarshal([]byte(tc.sut), &sut)
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			require.Equal(tc.want, sut)
		})
	}
}
