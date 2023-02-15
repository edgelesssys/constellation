/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package compatibility

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestFilterNewerVersion(t *testing.T) {
	imageList := []string{
		"v0.0.0",
		"v1.0.0",
		"v1.0.1",
		"v1.0.2",
		"v1.1.0",
	}

	testCases := map[string]struct {
		images     []string
		version    string
		wantImages []string
	}{
		"filters <= v1.0.0": {
			images:  imageList,
			version: "v1.0.0",
			wantImages: []string{
				"v1.0.1",
				"v1.0.2",
				"v1.1.0",
			},
		},
		"no compatible images": {
			images:  imageList,
			version: "v999.999.999",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			compatibleImages := FilterNewerVersion(tc.version, tc.images)
			assert.EqualValues(tc.wantImages, compatibleImages)
		})
	}
}

func TestNextMinorVersion(t *testing.T) {
	testCases := map[string]struct {
		version              string
		wantNextMinorVersion string
		wantErr              bool
	}{
		"gets next": {
			version:              "v1.0.0",
			wantNextMinorVersion: "v1.1",
		},
		"gets next from minor version": {
			version:              "v1.0",
			wantNextMinorVersion: "v1.1",
		},
		"empty version": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotNext, err := NextMinorVersion(tc.version)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.wantNextMinorVersion, gotNext)
		})
	}
}

func TestBinaryWith(t *testing.T) {
	testCases := map[string]struct {
		cli       string
		target    string
		wantError bool
	}{
		"success": {
			cli:    "v0.0.0",
			target: "v0.0.0",
		},
		"different major version": {
			cli:       "v1",
			target:    "v2",
			wantError: true,
		},
		"major version diff too large": {
			cli:       "v1.0",
			target:    "v1.2",
			wantError: true,
		},
		"cli has to be the newer version": {
			cli:       "v2.4.0",
			target:    "v2.5.0",
			wantError: true,
		},
		"pre prelease version ordering is correct": {
			cli:    "v2.5.0-pre",
			target: "v2.4.0",
		},
		"pre prelease versions are not forward compatible": {
			cli:       "v2.4.0",
			target:    "v2.5.0-pre",
			wantError: true,
		},
		"pre release versions match": {
			cli:    "v2.6.0-pre",
			target: "v2.6.0-pre",
		},
		"patch versions are forward compatible": {
			cli:    "v2.6.0",
			target: "v2.6.1",
		},
		"pseudo versions are not forward compatible": {
			cli:       "v2.5.0",
			target:    "v2.6.0-pre.0.20230125085856-aaaaaaaaaaaa",
			wantError: true,
		},
		"pseudo version is newer than first pre release": {
			cli:    "v2.6.0-pre",
			target: "v2.6.0-pre.0.20230125085856-aaaaaaaaaaaa",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			constants.VersionInfo = tc.cli
			err := BinaryWith(tc.target)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestIsValidUpgrade(t *testing.T) {
	testCases := map[string]struct {
		a         string
		b         string
		wantError bool
	}{
		"success": {
			a: "v0.0.0",
			b: "v0.1.0",
		},
		"different major version": {
			a:         "v1",
			b:         "v2",
			wantError: true,
		},
		"minor version diff too large": {
			a:         "v1.0",
			b:         "v1.2",
			wantError: true,
		},
		"b has to be the newer version": {
			a:         "v2.5.0",
			b:         "v2.4.0",
			wantError: true,
		},
		"pre prelease version ordering is correct": {
			a: "v2.4.0",
			b: "v2.5.0-pre",
		},
		"wrong pre release ordering creates error": {
			a:         "v2.5.0-pre",
			b:         "v2.4.0",
			wantError: true,
		},
		"pre release versions are equal": {
			a:         "v2.6.0-pre",
			b:         "v2.6.0-pre",
			wantError: true,
		},
		"pseudo version is newer than first pre release": {
			a:         "v2.6.0-pre.0.20230125085856-aaaaaaaaaaaa",
			b:         "v2.6.0-pre",
			wantError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := IsValidUpgrade(tc.a, tc.b)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestParseCanonicalSemver(t *testing.T) {
	testCases := map[string]struct {
		version   string
		major     int
		minor     int
		wantError bool
	}{
		"canonical input": {
			version: "v1.1.1",
			major:   1,
			minor:   1,
		},
		"vMAJOR.MINOR input": {
			version: "v1.1",
			major:   1,
			minor:   1,
		},
		"vMAJOR input": {
			version: "v1",
			major:   1,
			minor:   0,
		},
		"invalid (go)semver": {
			version:   "1.1", // valid semver, but invalid according to go's semver
			wantError: true,
		},
		"invalid (go)semver #2": {
			version:   "asdf",
			wantError: true,
		},
		"invalid (go)semver #3": {
			version:   "v1.1.1.1.1",
			wantError: true,
		},
		"pseudoversion": {
			version: "v2.6.0-pre.0.20230125085856-aaaaaaaaaaaa",
			major:   2,
			minor:   6,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			major, minor, err := parseCanonicalSemver(tc.version)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.major, major)
			assert.Equal(tc.minor, minor)
		})
	}
}
