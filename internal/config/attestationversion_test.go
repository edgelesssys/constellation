/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
				Value:    1,
				IsLatest: true,
			},
			want: "latest\n",
		},
		{
			name: "value 5 resolves to 5",
			sut: AttestationVersion{
				Value:    5,
				IsLatest: false,
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
		name      string
		expected  AttestationVersion
		yamlInput string
		wantErr   bool
	}{
		// TODO(elchead): activate latest logic for next release AB#3036
		{
			name:      "latest value is not allowed",
			expected:  AttestationVersion{},
			yamlInput: "latest\n",
			wantErr:   true,
		},
		{
			name: "value 5 resolves to 5",
			expected: AttestationVersion{
				Value: 5,
			},
			yamlInput: "5\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := AttestationVersion{}
			err := yaml.Unmarshal([]byte(tt.yamlInput), &sut)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.IsLatest, sut.IsLatest)
		})
	}
}
