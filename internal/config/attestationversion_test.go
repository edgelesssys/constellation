/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestVersionMarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		sut  AttestationVersion
		want string
	}{
		{
			name: "isLatest resolves to latest",
			sut: AttestationVersion{
				Value:      1,
				WantLatest: true,
			},
			want: "latest\n",
		},
		{
			name: "value 5 resolves to 5",
			sut: AttestationVersion{
				Value:      5,
				WantLatest: false,
			},
			want: "5\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt, err := yaml.Marshal(tt.sut)
			require.NoError(t, err)
			require.Equal(t, tt.want, string(bt))
		})
	}
}

func TestVersionUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		sut     string
		want    AttestationVersion
		wantErr bool
	}{
		{
			name: "latest resolves to isLatest",
			sut:  "latest",
			want: AttestationVersion{
				Value:      0,
				WantLatest: true,
			},
			wantErr: false,
		},
		{
			name: "1 resolves to value 1",
			sut:  "1",
			want: AttestationVersion{
				Value:      1,
				WantLatest: false,
			},
			wantErr: false,
		},
		{
			name:    "max uint8+1 errors",
			sut:     "256",
			wantErr: true,
		},
		{
			name:    "-1 errors",
			sut:     "-1",
			wantErr: true,
		},
		{
			name:    "2.6 errors",
			sut:     "2.6",
			wantErr: true,
		},
		{
			name:    "2.0 errors",
			sut:     "2.0",
			wantErr: true,
		},
		{
			name:    "hex format is invalid",
			sut:     "0x10",
			wantErr: true,
		},
		{
			name:    "octal format is invalid",
			sut:     "010",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sut AttestationVersion
			err := yaml.Unmarshal([]byte(tt.sut), &sut)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, sut)
		})
	}
}
