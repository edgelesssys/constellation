/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package shortname

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestToParts(t *testing.T) {
	testCases := map[string]struct {
		shortname string
		wantParts [3]string
		wantErr   bool
	}{
		"only version": {
			shortname: "v1.2.3",
			wantParts: [3]string{"-", "stable", "v1.2.3"},
		},
		"stream and version": {
			shortname: "stream/nightly/v1.2.3",
			wantParts: [3]string{"-", "nightly", "v1.2.3"},
		},
		"full name": {
			shortname: "ref/feat-xyz/stream/nightly/v1.2.3",
			wantParts: [3]string{"feat-xyz", "nightly", "v1.2.3"},
		},
		"full name with extra slashes": {
			shortname: "ref/feat-xyz//stream/nightly/v1.2.3",
			wantErr:   true,
		},
		"invalid three part path": {
			shortname: "invalid/invalid/invalid",
			wantErr:   true,
		},
		"five part path": {
			shortname: "invalid/invalid/invalid/invalid/invalid",
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			ref, stream, version, err := ToParts(tc.shortname)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantParts, [3]string{ref, stream, version})
		})
	}
}

func TestFromParts(t *testing.T) {
	testCases := map[string]struct {
		ref, stream, version string
		wantShortname        string
	}{
		"only version": {
			ref:           "-",
			stream:        "stable",
			version:       "v1.2.3",
			wantShortname: "v1.2.3",
		},
		"stream and version": {
			ref:           "-",
			stream:        "nightly",
			version:       "v1.2.3",
			wantShortname: "stream/nightly/v1.2.3",
		},
		"full name": {
			ref:           "feat-xyz",
			stream:        "nightly",
			version:       "v1.2.3",
			wantShortname: "ref/feat-xyz/stream/nightly/v1.2.3",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotShortname := FromParts(tc.ref, tc.stream, tc.version)
			assert.Equal(tc.wantShortname, gotShortname)
		})
	}
}
