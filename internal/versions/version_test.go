/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionValidation(t *testing.T) {
	testCases := map[string]struct {
		version string
		wantErr bool
	}{
		"valid version":    {"v1.18.0", false},
		"invalid version":  {"v1.18. 0", true},
		"missing prefix":   {"1.18.0", true},
		"only major.minor": {"v1.18", false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			_, err := NewVersion(tc.version)
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
		version    string
		wantString string
		wantErr    bool
	}{
		"valid version": {
			version:    "v1.18.0",
			wantString: `"v1.18.0"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			version, err := NewVersion(tc.version)
			require.NoError(err)

			b, err := version.MarshalJSON()
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

			var version Version
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
		version1 string
		version2 string
		want     int
		wantErr  bool
	}{
		"equal": {
			version1: "v1.18.0",
			version2: "v1.18.0",
			want:     0,
		},
		"less than": {
			version1: "v1.18.0",
			version2: "v1.18.1",
			want:     -1,
		},
		"greater than": {
			version1: "v1.18.1",
			version2: "v1.18.0",
			want:     1,
		},
		"invalid version": {
			version1: "v1.18.0",
			version2: "v1.18. ",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			v1, err := NewVersion(tc.version1)
			require.NoError(err)

			v2, err := NewVersion(tc.version2)
			if tc.wantErr {
				require.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, v1.Compare(v2))
			}
		})
	}
}
