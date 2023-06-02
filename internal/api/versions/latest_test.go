/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestLatestJSONPath(t *testing.T) {
	testCases := map[string]struct {
		list     Latest
		wantPath string
	}{
		"latest list": {
			list: Latest{
				Ref:    "test-ref",
				Stream: "nightly",
				Kind:   VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/versions/latest/image.json",
		},
		"latest list release": {
			list: Latest{
				Ref:    ReleaseRef,
				Stream: "stable",
				Kind:   VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/-/stream/stable/versions/latest/image.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.wantPath, tc.list.JSONPath())
		})
	}
}

func TestLatestURL(t *testing.T) {
	testCases := map[string]struct {
		list     Latest
		wantURL  string
		wantPath string
	}{
		"latest list": {
			list: Latest{
				Ref:    "test-ref",
				Stream: "nightly",
				Kind:   VersionKindImage,
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/versions/latest/image.json",
		},
		"latest list release": {
			list: Latest{
				Ref:    ReleaseRef,
				Stream: "stable",
				Kind:   VersionKindImage,
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/-/stream/stable/versions/latest/image.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			url, err := tc.list.URL()
			assert.NoError(err)
			assert.Equal(tc.wantURL, url)
		})
	}
}

func TestLatestValidate(t *testing.T) {
	testCases := map[string]struct {
		list    Latest
		wantErr bool
	}{
		"valid": {
			list: Latest{
				Ref:     "test-ref",
				Stream:  "nightly",
				Kind:    VersionKindImage,
				Version: "v1.0.0",
			},
		},
		"invalid ref": {
			list: Latest{
				Ref:     "",
				Stream:  "nightly",
				Kind:    VersionKindImage,
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"invalid stream": {
			list: Latest{
				Ref:     "test-ref",
				Stream:  "invalid-stream",
				Kind:    VersionKindImage,
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"invalid kind": {
			list: Latest{
				Ref:     "test-ref",
				Stream:  "nightly",
				Kind:    VersionKindUnknown,
				Version: "v1.0.0",
			},
			wantErr: true,
		},
		"invalid version": {
			list: Latest{
				Ref:     "test-ref",
				Stream:  "nightly",
				Kind:    VersionKindImage,
				Version: "",
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if tc.wantErr {
				assert.Error(t, tc.list.Validate())
			} else {
				assert.NoError(t, tc.list.Validate())
			}
		})
	}
}

func TestLatestValidateRequest(t *testing.T) {
	testCases := map[string]struct {
		list    Latest
		wantErr bool
	}{
		"valid": {
			list: Latest{
				Ref:    "test-ref",
				Stream: "nightly",
				Kind:   VersionKindImage,
			},
		},
		"invalid ref": {
			list: Latest{
				Ref:    "",
				Stream: "nightly",
				Kind:   VersionKindImage,
			},
			wantErr: true,
		},
		"invalid stream": {
			list: Latest{
				Ref:    "test-ref",
				Stream: "invalid-stream",
				Kind:   VersionKindImage,
			},
			wantErr: true,
		},
		"invalid kind": {
			list: Latest{
				Ref:    "test-ref",
				Stream: "nightly",
				Kind:   VersionKindUnknown,
			},
			wantErr: true,
		},
		"version not empty": {
			list: Latest{
				Ref:     "test-ref",
				Stream:  "nightly",
				Kind:    VersionKindImage,
				Version: "v1.1.1",
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if tc.wantErr {
				assert.Error(t, tc.list.ValidateRequest())
			} else {
				assert.NoError(t, tc.list.ValidateRequest())
			}
		})
	}
}

func TestLatestShortPath(t *testing.T) {
	assert := assert.New(t)

	latest := Latest{
		Ref:     "test-ref",
		Stream:  "nightly",
		Kind:    VersionKindImage,
		Version: "v1.0.0",
	}

	assert.Equal(shortPath(latest.Ref, latest.Stream, latest.Version), latest.ShortPath())
}
