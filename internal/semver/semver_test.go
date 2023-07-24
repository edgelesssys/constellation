/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package semver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var (
	v1_18_0         = Semver{major: 1, minor: 18, patch: 0}
	v1_18_0Pre      = Semver{major: 1, minor: 18, patch: 0, prerelease: "pre"}
	v1_18_0PreExtra = Semver{major: 1, minor: 18, patch: 0, prerelease: "pre.1"}
	v1_19_0         = Semver{major: 1, minor: 19, patch: 0}
	v1_18_1         = Semver{major: 1, minor: 18, patch: 1}
	v1_20_0         = Semver{major: 1, minor: 20, patch: 0}
	v2_0_0          = Semver{major: 2, minor: 0, patch: 0}
)

func TestNewVersion(t *testing.T) {
	testCases := map[string]struct {
		version string
		want    Semver
		wantErr bool
	}{
		"valid version": {
			version: "v1.18.0",
			want: Semver{
				major: 1,
				minor: 18,
				patch: 0,
			},
			wantErr: false,
		},
		"valid version prerelease": {
			version: "v1.18.0-pre+yyyymmddhhmmss-abcdefabcdef",
			want: Semver{
				major:      1,
				minor:      18,
				patch:      0,
				prerelease: "pre",
			},
			wantErr: false,
		},
		"only prerelease": {version: "v-pre.0.yyyymmddhhmmss-abcdefabcdef", wantErr: true},
		"invalid version": {version: "v1.18. 0", wantErr: true},
		"add prefix": {
			version: "1.18.0",
			want: Semver{
				major: 1,
				minor: 18,
				patch: 0,
			},
			wantErr: false,
		},
		"only major.minor": {
			version: "v1.18",
			want: Semver{
				major: 1,
				minor: 18,
				patch: 0,
			},
			wantErr: false,
		},
		"only major": {
			version: "v1",
			want: Semver{
				major: 1,
				minor: 0,
				patch: 0,
			},
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ver, err := New(tc.version)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, ver)
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
		"prerelease": {
			version:    v1_18_0Pre,
			wantString: `"v1.18.0-pre"`,
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
		"prerelease": {
			version:    `"v1.18.0-pre"`,
			wantString: "v1.18.0-pre",
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
		"prerelease less than": {
			version1: v1_18_0Pre,
			version2: v1_18_0,
			want:     -1,
		},
		"prerelease greater than": {
			version1: v1_18_0,
			version2: v1_18_0Pre,
			want:     1,
		},
		"prerelease equal": {
			version1: v1_18_0Pre,
			version2: v1_18_0Pre,
			want:     0,
		},
		"prerelease extra less than": {
			version1: v1_18_0Pre,
			version2: v1_18_0PreExtra,
			want:     -1,
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
		wantErr  bool
	}{
		"equal": {
			version1: v1_18_0,
			version2: v1_18_0,
			wantErr:  false,
		},
		"patch less than": {
			version1: v1_18_0,
			version2: v1_18_1,
			wantErr:  true,
		},
		"minor less then": {
			version1: v1_18_0,
			version2: v1_19_0,
			wantErr:  true,
		},
		"minor too big drift": {
			version1: v1_18_0,
			version2: v1_20_0,
			wantErr:  false,
		},
		"major too big drift": {
			version1: v1_18_0,
			version2: v2_0_0,
			wantErr:  false,
		},
		"greater than": {
			version1: v1_18_1,
			version2: v1_18_0,
			wantErr:  false,
		},
		"prerelease less than": {
			version1: v1_18_0Pre,
			version2: v1_18_0,
			wantErr:  true,
		},
		"prerelease greater than": {
			version1: v1_18_0,
			version2: v1_18_0Pre,
			wantErr:  false,
		},
		"prerelease equal": {
			version1: v1_18_0Pre,
			version2: v1_18_0Pre,
			wantErr:  false,
		},
		"prerelease extra": {
			version1: v1_18_0Pre,
			version2: v1_18_0PreExtra,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.wantErr, tc.version2.IsUpgradeTo(tc.version1) == nil)
		})
	}
}

func TestNextminor(t *testing.T) {
	testCases := map[string]struct {
		version Semver
		want    string
	}{
		"valid version": {
			version: v1_18_0,
			want:    "v1.19",
		},
		"prerelease": {
			version: v1_18_0Pre,
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

func TestVersionMarshalYAML(t *testing.T) {
	testCases := map[string]struct {
		version Semver
		want    string
	}{
		"simple": {
			version: Semver{
				major:      1,
				minor:      18,
				patch:      0,
				prerelease: "",
			},
			want: "v1.18.0\n",
		},
		"with prerelease": {
			version: Semver{
				major:      1,
				minor:      18,
				patch:      0,
				prerelease: "pre",
			},
			want: "v1.18.0-pre\n",
		},
		"empty semver": {
			version: Semver{},
			want:    "v0.0.0\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			marshalled, err := yaml.Marshal(tc.version)

			require.NoError(t, err)
			require.Equal(t, tc.want, string(marshalled))
		})
	}
}

func TestVersionUnmarshalYAML(t *testing.T) {
	testCases := map[string]struct {
		version   []byte
		want      Semver
		wantError bool
	}{
		"simple unmarshal works": {
			version: []byte("v1.18.0"),
			want: Semver{
				major:      1,
				minor:      18,
				patch:      0,
				prerelease: "",
			},
		},
		"empty string": {
			version: []byte(""),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var new Semver
			err := yaml.Unmarshal(tc.version, &new)
			if tc.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want.Compare(new), 0, fmt.Sprintf("expected %s, got %s", tc.want, new))
		})
	}
}
