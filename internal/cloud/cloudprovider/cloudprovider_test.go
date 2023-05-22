/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudprovider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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
		"openstack": {
			input: OpenStack,
			want:  []byte("\"OpenStack\""),
		},
		"qemu": {
			input: QEMU,
			want:  []byte("\"QEMU\""),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			b, err := json.Marshal(tc.input)

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
		"openstack": {
			input: []byte("\"openstack\""),
			want:  OpenStack,
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
			err := json.Unmarshal(tc.input, &p)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, p)
			}
		})
	}
}

func TestMarshalYAML(t *testing.T) {
	testCases := map[string]struct {
		input Provider
		want  []byte
	}{
		"unknown": {
			input: Unknown,
			want:  []byte("Unknown\n"),
		},
		"aws": {
			input: AWS,
			want:  []byte("AWS\n"),
		},
		"azure": {
			input: Azure,
			want:  []byte("Azure\n"),
		},
		"gcp": {
			input: GCP,
			want:  []byte("GCP\n"),
		},
		"openstack": {
			input: OpenStack,
			want:  []byte("OpenStack\n"),
		},
		"qemu": {
			input: QEMU,
			want:  []byte("QEMU\n"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			b, err := yaml.Marshal(tc.input)

			assert.NoError(err)
			assert.Equal(tc.want, b)
		})
	}
}

func TestUnmarshalYAML(t *testing.T) {
	testCases := map[string]struct {
		input   []byte
		want    Provider
		wantErr bool
	}{
		"empty": {
			input:   []byte("foo: bar\n"),
			wantErr: true,
		},
		"unknown": {
			input: []byte("unknown\n"),
			want:  Unknown,
		},
		"aws": {
			input: []byte("aws\n"),
			want:  AWS,
		},
		"azure": {
			input: []byte("azure\n"),
			want:  Azure,
		},
		"gcp": {
			input: []byte("gcp\n"),
			want:  GCP,
		},
		"openstack": {
			input: []byte("openstack\n"),
			want:  OpenStack,
		},
		"qemu": {
			input: []byte("qemu\n"),
			want:  QEMU,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			var p Provider
			err := yaml.Unmarshal(tc.input, &p)

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
		"openstack": {
			input: "openstack",
			want:  OpenStack,
		},
		"stackit": {
			input: "stackit",
			want:  OpenStack,
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
