/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalJSON(t *testing.T) {
	testCases := map[string]struct {
		input Provider
		want  []byte
	}{
		"unknown": {
			input: Unknown,
			want:  []byte("\"Unknown\""),
		},
		"aws": {
			input: AWS,
			want:  []byte("\"AWS\""),
		},
		"azure": {
			input: Azure,
			want:  []byte("\"Azure\""),
		},
		"gcp": {
			input: GCP,
			want:  []byte("\"GCP\""),
		},
		"qemu": {
			input: QEMU,
			want:  []byte("\"QEMU\""),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			b, err := tc.input.MarshalJSON()

			assert.NoError(err)
			assert.Equal(tc.want, b)
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	testCases := map[string]struct {
		input   []byte
		want    Provider
		wantErr bool
	}{
		"empty": {
			input:   []byte{},
			wantErr: true,
		},
		"unknown": {
			input: []byte("\"unknown\""),
			want:  Unknown,
		},
		"aws": {
			input: []byte("\"aws\""),
			want:  AWS,
		},
		"azure": {
			input: []byte("\"azure\""),
			want:  Azure,
		},
		"gcp": {
			input: []byte("\"gcp\""),
			want:  GCP,
		},
		"qemu": {
			input: []byte("\"qemu\""),
			want:  QEMU,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			var p Provider
			err := p.UnmarshalJSON(tc.input)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, p)
			}
		})
	}
}

func TestFromString(t *testing.T) {
	testCases := map[string]struct {
		input string
		want  Provider
	}{
		"empty": {
			input: "",
			want:  Unknown,
		},
		"unknown": {
			input: "unknown",
			want:  Unknown,
		},
		"aws": {
			input: "aws",
			want:  AWS,
		},
		"azure": {
			input: "azure",
			want:  Azure,
		},
		"gcp": {
			input: "gcp",
			want:  GCP,
		},
		"qemu": {
			input: "qemu",
			want:  QEMU,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			p := FromString(tc.input)

			assert.Equal(tc.want, p)
		})
	}
}
