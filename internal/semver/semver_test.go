/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package semver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	v1_18_0 = Semver{Major: 1, Minor: 18, Patch: 0}
	v1_19_0 = Semver{Major: 1, Minor: 19, Patch: 0}
	v1_18_1 = Semver{Major: 1, Minor: 18, Patch: 1}
	v1_20_0 = Semver{Major: 1, Minor: 20, Patch: 0}
	v2_0_0  = Semver{Major: 2, Minor: 0, Patch: 0}
)

func TestNewVersion(t *testing.T) {
	testCases := map[string]struct {
		version string
		wantErr bool
	}{
		"valid version":    {"v1.18.0", false},
		"invalid version":  {"v1.18. 0", true},
		"add prefix":       {"1.18.0", false},
		"only major.minor": {"v1.18", false},
		"only major":       {"v1", false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err := New(tc.version)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestJSONMarshal(t *testing.T) {
	testCases := map[string]struct {
		version    Semver
		wantString string
		wantErr    bool
	}{
		"valid version": {
			version:    v1_18_0,
			wantString: `"v1.18.0"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			b, err := tc.version.MarshalJSON()
			if tc.wantErr {
				require.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantString, string(b))
			}
		})
	}
}

func TestJSONUnmarshal(t *testing.T) {
	testCases := map[string]struct {
		version    string
		wantString string
		wantErr    bool
	}{
		"valid version": {
			version:    `"v1.18.0"`,
			wantString: "v1.18.0",
		},
		"invalid version": {
			version: `"v1. 18.0"`,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var version Semver
			err := version.UnmarshalJSON([]byte(tc.version))
			if tc.wantErr {
				require.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantString, version.String())
			}
		})
	}
}

func TestComparison(t *testing.T) {
	testCases := map[string]struct {
		version1 Semver
		version2 Semver
		want     int
	}{
		"equal": {
			version1: v1_18_0,
			version2: v1_18_0,
			want:     0,
		},
		"less than": {
			version1: v1_18_0,
			version2: v1_18_1,
			want:     -1,
		},
		"greater than": {
			version1: v1_18_1,
			version2: v1_18_0,
			want:     1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.want, tc.version1.Compare(tc.version2))
		})
	}
}

func TestCanUpgrade(t *testing.T) {
	testCases := map[string]struct {
		version1 Semver
		version2 Semver
		want     bool
		wantErr  bool
	}{
		"equal": {
			version1: v1_18_0,
			version2: v1_18_0,
			want:     false,
		},
		"patch less than": {
			version1: v1_18_0,
			version2: v1_18_1,
			want:     true,
		},
		"minor less then": {
			version1: v1_18_0,
			version2: v1_19_0,
			want:     true,
		},
		"minor too big drift": {
			version1: v1_18_0,
			version2: v1_20_0,
			want:     false,
		},
		"major too big drift": {
			version1: v1_18_0,
			version2: v2_0_0,
			want:     false,
		},
		"greater than": {
			version1: v1_18_1,
			version2: v1_18_0,
			want:     false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.want, tc.version2.IsUpgradeTo(tc.version1))
		})
	}
}

func TestNextMinor(t *testing.T) {
	testCases := map[string]struct {
		version Semver
		want    string
	}{
		"valid version": {
			version: v1_18_0,
			want:    "v1.19",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(tc.want, tc.version.NextMinor())
		})
	}
}
