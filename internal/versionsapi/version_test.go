/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"fmt"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersionFromShortPath(t *testing.T) {
	testCases := map[string]struct {
		path    string
		kind    VersionKind
		wantVer Version
		wantErr bool
	}{
		"stable release image": {
			path: "v9.9.9",
			kind: VersionKindImage,
			wantVer: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
		},
		"release debug image": {
			path: "stream/debug/v9.9.9",
			kind: VersionKindImage,
			wantVer: Version{
				Ref:     ReleaseRef,
				Stream:  "debug",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
		},
		"unknown kind": {
			path:    "v9.9.9",
			kind:    VersionKindUnknown,
			wantErr: true,
		},
		"invalid path": {
			path:    "va.b.c",
			kind:    VersionKindImage,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ver, err := NewVersionFromShortPath(tc.path, tc.kind)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantVer, ver)
		})
	}
}

func TestVersionShortPath(t *testing.T) {
	testCases := map[string]struct {
		ver  Version
		want string
	}{
		"stable release image": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			want: "v9.9.9",
		},
		"release debug image": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "debug",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			want: "stream/debug/v9.9.9",
		},
		"branch image": {
			ver: Version{
				Ref:     "foo",
				Stream:  "debug",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			want: "ref/foo/stream/debug/v9.9.9",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			got := tc.ver.ShortPath()
			assert.Equal(tc.want, got)
		})
	}
}

func TestVersionValidate(t *testing.T) {
	testCases := map[string]struct {
		ver     Version
		wantErr bool
	}{
		"valid": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
		},
		"invalid ref": {
			ver: Version{
				Ref:     "foo/bar",
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			wantErr: true,
		},
		"invalid stream": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "foo/bar",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			wantErr: true,
		},
		"invalid version": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9/foo",
				Kind:    VersionKindImage,
			},
			wantErr: true,
		},
		"invalid kind": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindUnknown,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := tc.ver.Validate()
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestVersionMajor(t *testing.T) {
	testCases := map[string]string{
		"v9.9.9":     "v9",
		"v9.6.9-foo": "v9",
		"v7.9.9":     "v7",
	}

	for version, major := range testCases {
		t.Run(version, func(t *testing.T) {
			assert := assert.New(t)

			ver := Version{Version: version}
			assert.Equal(major, ver.Major())
		})
	}
}

func TestVersionMajorMinor(t *testing.T) {
	testCases := map[string]string{
		"v9.9.9":     "v9.9",
		"v9.6.9-foo": "v9.6",
		"v7.9.9":     "v7.9",
	}

	for version, major := range testCases {
		t.Run(version, func(t *testing.T) {
			assert := assert.New(t)

			ver := Version{Version: version}
			assert.Equal(major, ver.MajorMinor())
		})
	}
}

func TestVersionWithGranularity(t *testing.T) {
	testCases := []struct {
		ver  string
		gran Granularity
		want string
	}{
		{
			ver:  "v9.9.9",
			gran: GranularityMajor,
			want: "v9",
		},
		{
			ver:  "v9.9.9",
			gran: GranularityMinor,
			want: "v9.9",
		},
		{
			ver:  "v9.9.9",
			gran: GranularityPatch,
			want: "v9.9.9",
		},
		{
			ver:  "v9.9.9-foo",
			gran: GranularityMajor,
			want: "v9",
		},
		{
			ver:  "v9.9.9-foo",
			gran: GranularityPatch,
			want: "v9.9.9-foo",
		},
		{
			ver:  "v9.9.9-foo",
			gran: GranularityUnknown,
			want: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.ver, func(t *testing.T) {
			assert := assert.New(t)

			ver := Version{Version: tc.ver}
			assert.Equal(tc.want, ver.WithGranularity(tc.gran))
		})
	}
}

func TestVersionListPathURL(t *testing.T) {
	testCases := map[string]struct {
		ver      Version
		gran     Granularity
		wantPath string
		wantURL  string
	}{
		"release": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			gran:     GranularityMajor,
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/versions/major/v9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/versions/major/v9/image.json",
		},
		"release with minor": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			gran:     GranularityMinor,
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/versions/minor/v9.9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/versions/minor/v9.9/image.json",
		},
		"release with patch": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			gran:     GranularityPatch,
			wantPath: "",
			wantURL:  "",
		},
		"release with unknown": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			gran:     GranularityUnknown,
			wantPath: "",
			wantURL:  "",
		},
		"release debug stream": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "debug",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			gran:     GranularityMajor,
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/versions/major/v9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/versions/major/v9/image.json",
		},
		"release debug stream with minor": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "debug",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			gran:     GranularityMinor,
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/versions/minor/v9.9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/versions/minor/v9.9/image.json",
		},
		"branch ref": {
			ver: Version{
				Ref:     "foo",
				Stream:  "debug",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			gran:     GranularityMajor,
			wantPath: constants.CDNAPIPrefix + "/ref/foo/stream/debug/versions/major/v9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/foo/stream/debug/versions/major/v9/image.json",
		},
		"branch ref with minor": {
			ver: Version{
				Ref:     "foo",
				Stream:  "debug",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			gran:     GranularityMinor,
			wantPath: constants.CDNAPIPrefix + "/ref/foo/stream/debug/versions/minor/v9.9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/foo/stream/debug/versions/minor/v9.9/image.json",
		},
	}

	for name, tc := range testCases {
		t.Run(fmt.Sprintf("%s url", name), func(t *testing.T) {
			assert := assert.New(t)

			url := tc.ver.ListURL(tc.gran)
			assert.Equal(tc.wantURL, url)
		})

		t.Run(fmt.Sprintf("%s path", name), func(t *testing.T) {
			assert := assert.New(t)

			path := tc.ver.ListPath(tc.gran)
			assert.Equal(tc.wantPath, path)
		})
	}
}

func TestVersionArtifactURL(t *testing.T) {
	testCases := map[string]struct {
		ver     Version
		csp     cloudprovider.Provider
		file    string
		wantURL string
	}{
		"nightly-feature": {
			ver: Version{
				Ref:     "feat-some-feature",
				Stream:  "nightly",
				Version: "v2.6.0-pre.0.20230217095603-193dd48ca19f",
				Kind:    VersionKindImage,
			},
			csp:     cloudprovider.GCP,
			file:    "measurements.json",
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/feat-some-feature/stream/nightly/v2.6.0-pre.0.20230217095603-193dd48ca19f/image/csp/gcp/measurements.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			artifactURL, err := tc.ver.ArtifactURL(tc.csp, tc.file)
			require.NoError(err)
			assert.Equal(tc.wantURL, artifactURL.String())
		})
	}
}

func TestVersionArtifactPathURL(t *testing.T) {
	testCases := map[string]struct {
		ver      Version
		wantPath string
	}{
		"release": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "stable",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/v9.9.9",
		},
		"release debug stream": {
			ver: Version{
				Ref:     ReleaseRef,
				Stream:  "debug",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/v9.9.9",
		},
		"branch ref": {
			ver: Version{
				Ref:     "foo",
				Stream:  "debug",
				Version: "v9.9.9",
				Kind:    VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/foo/stream/debug/v9.9.9",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			path := tc.ver.ArtifactPath()
			assert.Equal(tc.wantPath, path)
			url := tc.ver.ArtifactsURL()
			assert.Equal(constants.CDNRepositoryURL+"/"+tc.wantPath, url)
		})
	}
}

func TestVersionKindUnMarshalJson(t *testing.T) {
	testCases := map[string]VersionKind{
		`"image"`:   VersionKindImage,
		`"unknown"`: VersionKindUnknown,
	}

	for name, kind := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			data, err := kind.MarshalJSON()
			assert.NoError(err)
			assert.Equal(name, string(data))

			var gotKind VersionKind
			err = gotKind.UnmarshalJSON([]byte(name))
			assert.NoError(err)
			assert.Equal(kind, gotKind)
		})
	}
}

func TestVersionKindFromString(t *testing.T) {
	testCases := map[string]VersionKind{
		"image":   VersionKindImage,
		"unknown": VersionKindUnknown,
	}

	for name, kind := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			s := kind.String()
			assert.Equal(name, s)

			k := VersionKindFromString(name)
			assert.Equal(kind, k)
		})
	}
}

func TestGranularityUnMarschalJSON(t *testing.T) {
	testCases := map[string]Granularity{
		`"major"`:   GranularityMajor,
		`"minor"`:   GranularityMinor,
		`"patch"`:   GranularityPatch,
		`"unknown"`: GranularityUnknown,
	}

	for name, gran := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			data, err := gran.MarshalJSON()
			assert.NoError(err)
			assert.Equal(name, string(data))

			var gotGran Granularity
			err = gotGran.UnmarshalJSON([]byte(name))
			assert.NoError(err)
			assert.Equal(gran, gotGran)
		})
	}
}

func TestGranularityFromString(t *testing.T) {
	testCases := map[string]Granularity{
		"major":   GranularityMajor,
		"minor":   GranularityMinor,
		"patch":   GranularityPatch,
		"unknown": GranularityUnknown,
	}

	for name, gran := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			s := gran.String()
			assert.Equal(name, s)

			g := GranularityFromString(name)
			assert.Equal(gran, g)
		})
	}
}

func TestCanonicalRef(t *testing.T) {
	testCases := map[string]string{
		"feat/foo": "feat-foo",
		"feat-foo": "feat-foo",
		"feat$foo": "feat-foo",
		"3234":     "3234",
		"feat foo": "feat-foo",
		"/../":     "----",
		ReleaseRef: ReleaseRef,
		".":        "",
	}

	for ref, want := range testCases {
		t.Run(ref, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(want, CanonicalizeRef(ref))
		})
	}
}

func TestValidateRef(t *testing.T) {
	testCases := map[string]bool{
		"feat/foo":            true,
		"feat-foo":            false,
		"feat$foo":            true,
		"3234":                false,
		"feat foo":            true,
		"refs-heads-feat-foo": true,
		"":                    true,
	}

	for ref, wantErr := range testCases {
		t.Run(ref, func(t *testing.T) {
			assert := assert.New(t)
			err := ValidateRef(ref)
			if !wantErr {
				assert.NoError(err)
			} else {
				assert.Error(err)
			}
		})
	}
}

func TestValidateStream(t *testing.T) {
	testCases := []struct {
		branch  string
		stream  string
		wantErr bool
	}{
		{branch: "-", stream: "stable", wantErr: false},
		{branch: "-", stream: "debug", wantErr: false},
		{branch: "-", stream: "nightly", wantErr: true},
		{branch: "-", stream: "console", wantErr: false},
		{branch: "main", stream: "stable", wantErr: true},
		{branch: "main", stream: "debug", wantErr: false},
		{branch: "main", stream: "nightly", wantErr: false},
		{branch: "main", stream: "console", wantErr: false},
		{branch: "foo-branch", stream: "nightly", wantErr: false},
		{branch: "foo-branch", stream: "console", wantErr: false},
		{branch: "foo-branch", stream: "debug", wantErr: false},
		{branch: "foo-branch", stream: "stable", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.branch+"+"+tc.stream, func(t *testing.T) {
			assert := assert.New(t)

			err := ValidateStream(tc.branch, tc.stream)
			if !tc.wantErr {
				assert.NoError(err)
			} else {
				assert.Error(err)
			}
		})
	}
}

func TestShortPath(t *testing.T) {
	testCases := map[string]struct {
		ref     string
		stream  string
		version string
	}{
		"v9.9.9": {
			ref:     ReleaseRef,
			stream:  "stable",
			version: "v9.9.9",
		},
		"v9.9.9-foo": {
			ref:     ReleaseRef,
			stream:  "stable",
			version: "v9.9.9-foo",
		},
		"stream/debug/v9.9.9": {
			ref:     ReleaseRef,
			stream:  "debug",
			version: "v9.9.9",
		},
		"ref/foo/stream/debug/v9.9.9": {
			ref:     "foo",
			stream:  "debug",
			version: "v9.9.9",
		},
		"ref/foo/stream/stable/v9.9.9": {
			ref:     "foo",
			stream:  "stable",
			version: "v9.9.9",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			path := shortPath(tc.ref, tc.stream, tc.version)
			assert.Equal(name, path)
		})
	}
}

func TestParseShortPath(t *testing.T) {
	testCases := map[string]struct {
		wantRef     string
		wantStream  string
		wantVersion string
		wantErr     bool
	}{
		"v9.9.9": {
			wantRef:     ReleaseRef,
			wantStream:  "stable",
			wantVersion: "v9.9.9",
		},
		"v9.9.9-foo": {
			wantRef:     ReleaseRef,
			wantStream:  "stable",
			wantVersion: "v9.9.9-foo",
		},
		"stream/debug/v9.9.9": {
			wantRef:     ReleaseRef,
			wantStream:  "debug",
			wantVersion: "v9.9.9",
		},
		"ref/foo/stream/debug/v9.9.9": {
			wantRef:     "foo",
			wantStream:  "debug",
			wantVersion: "v9.9.9",
		},
		"v9.9.9-foo/bar": {
			wantErr: true,
		},
		"ref/foo/stream/debug/va.b.9": {
			wantErr: true,
		},
		"stream/debug/va.b.9": {
			wantErr: true,
		},
		"ref/foo/stream/bar/v9.9.9": {
			wantErr: true,
		},
		"stream/bar/v9.9.9": {
			wantErr: true,
		},
		"ref/refs-heads-bar/stream/debug/v9.9.9": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ref, stream, version, err := parseShortPath(name)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantRef, ref)
				assert.Equal(tc.wantStream, stream)
				assert.Equal(tc.wantVersion, version)
			}
		})
	}
}
