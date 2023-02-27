/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestImageInfoJSONPath(t *testing.T) {
	testCases := map[string]struct {
		info     ImageInfo
		wantPath string
	}{
		"image info": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantPath: constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/v1.0.0/image/info.json",
		},
		"image info release": {
			info: ImageInfo{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v1.0.0",
			},
			wantPath: constants.CDNAPIPrefix + "/ref/-/stream/stable/v1.0.0/image/info.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.wantPath, tc.info.JSONPath())
		})
	}
}

func TestImageInfoURL(t *testing.T) {
	testCases := map[string]struct {
		info     ImageInfo
		wantURL  string
		wantPath string
	}{
		"image info": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/v1.0.0/image/info.json",
		},
		"image info release": {
			info: ImageInfo{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v1.0.0",
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/-/stream/stable/v1.0.0/image/info.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			url, err := tc.info.URL()
			assert.NoError(err)
			assert.Equal(tc.wantURL, url)
		})
	}
}

func TestImageInfoValidate(t *testing.T) {
	testCases := map[string]struct {
		info    ImageInfo
		wantErr bool
	}{
		"valid image info": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				AWS:     map[string]string{"key": "value", "key2": "value2"},
				GCP:     map[string]string{"key": "value", "key2": "value2"},
				Azure:   map[string]string{"key": "value", "key2": "value2"},
				QEMU:    map[string]string{"key": "value", "key2": "value2"},
			},
		},
		"invalid ref": {
			info: ImageInfo{
				Ref:     "",
				Stream:  "nightly",
				Version: "v1.0.0",
				AWS:     map[string]string{"key": "value", "key2": "value2"},
				GCP:     map[string]string{"key": "value", "key2": "value2"},
				Azure:   map[string]string{"key": "value", "key2": "value2"},
				QEMU:    map[string]string{"key": "value", "key2": "value2"},
			},
			wantErr: true,
		},
		"invalid stream": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "",
				Version: "v1.0.0",
				AWS:     map[string]string{"key": "value", "key2": "value2"},
				GCP:     map[string]string{"key": "value", "key2": "value2"},
				Azure:   map[string]string{"key": "value", "key2": "value2"},
				QEMU:    map[string]string{"key": "value", "key2": "value2"},
			},
			wantErr: true,
		},
		"invalid version": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "",
				AWS:     map[string]string{"key": "value", "key2": "value2"},
				GCP:     map[string]string{"key": "value", "key2": "value2"},
				Azure:   map[string]string{"key": "value", "key2": "value2"},
				QEMU:    map[string]string{"key": "value", "key2": "value2"},
			},
			wantErr: true,
		},
		"no provider": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"multiple errors": {
			info: ImageInfo{
				Ref:     "",
				Stream:  "",
				Version: "",
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := tc.info.Validate()

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestImageInfoValidateRequest(t *testing.T) {
	testCases := map[string]struct {
		info    ImageInfo
		wantErr bool
	}{
		"valid image info": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
		},
		"invalid ref": {
			info: ImageInfo{
				Ref:     "",
				Stream:  "nightly",
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"invalid stream": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "",
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"invalid version": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "",
			},
			wantErr: true,
		},
		"invalid gcp": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				GCP:     map[string]string{"key": "value", "key2": "value2"},
			},
			wantErr: true,
		},
		"invalid azure": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				Azure:   map[string]string{"key": "value", "key2": "value2"},
			},
			wantErr: true,
		},
		"invalid aws": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				AWS:     map[string]string{"key": "value", "key2": "value2"},
			},
			wantErr: true,
		},
		"invalid qemu": {
			info: ImageInfo{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				QEMU:    map[string]string{"key": "value", "key2": "value2"},
			},
			wantErr: true,
		},
		"invalid openstack": {
			info: ImageInfo{
				Ref:       "test-ref",
				Stream:    "nightly",
				Version:   "v1.0.0",
				OpenStack: map[string]string{"key": "value", "key2": "value2"},
			},
			wantErr: true,
		},
		"multiple errors": {
			info: ImageInfo{
				Ref:     "",
				Stream:  "",
				Version: "",
				AWS:     map[string]string{"key": "value", "key2": "value2"},
				GCP:     map[string]string{"key": "value", "key2": "value2"},
				Azure:   map[string]string{"key": "value", "key2": "value2"},
				QEMU:    map[string]string{"key": "value", "key2": "value2"},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			err := tc.info.ValidateRequest()

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
