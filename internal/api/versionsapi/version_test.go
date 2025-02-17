/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

func TestNewVersion(t *testing.T) {
	testCases := map[string]struct {
		ref     string
		stream  string
		version string
		kind    VersionKind
		wantVer Version
		wantErr bool
	}{
		"stable release image": {
			ref:     ReleaseRef,
			stream:  "stable",
			version: "v9.9.9",
			kind:    VersionKindImage,
			wantVer: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
		},
		"release debug image": {
			ref:     ReleaseRef,
			stream:  "debug",
			version: "v9.9.9",
			kind:    VersionKindImage,
			wantVer: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
		},
		"stable release cli": {
			ref:     ReleaseRef,
			stream:  "stable",
			version: "v9.9.9",
			kind:    VersionKindCLI,
			wantVer: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
		},
		"release debug cli": {
			ref:     ReleaseRef,
			stream:  "debug",
			version: "v9.9.9",
			kind:    VersionKindCLI,
			wantVer: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
		},
		"unknown kind": {
			ref:     ReleaseRef,
			stream:  "debug",
			version: "v9.9.9",
			kind:    VersionKindUnknown,
			wantErr: true,
		},
		"non-release ref as input": {
			ref:     "working-branch",
			stream:  "debug",
			version: "v9.9.9",
			kind:    VersionKindImage,
			wantVer: Version{
				ref:     "working-branch",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
		},
		"non-canonical ref as input": {
			ref:     "testing-1.23",
			stream:  "debug",
			version: "v9.9.9",
			kind:    VersionKindImage,
			wantVer: Version{
				ref:     "testing-1-23",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ver, err := NewVersion(tc.ref, tc.stream, tc.version, tc.kind)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantVer, ver)
		})
	}
}

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
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
		},
		"release debug image": {
			path: "stream/debug/v9.9.9",
			kind: VersionKindImage,
			wantVer: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
		},
		"stable release cli": {
			path: "v9.9.9",
			kind: VersionKindCLI,
			wantVer: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
		},
		"release debug cli": {
			path: "stream/debug/v9.9.9",
			kind: VersionKindCLI,
			wantVer: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
		},
		"unknown kind": {
			path:    "v9.9.9",
			kind:    VersionKindUnknown,
			wantErr: true,
		},
		"invalid path image": {
			path:    "va.b.c",
			kind:    VersionKindImage,
			wantErr: true,
		},
		"invalid path cli": {
			path:    "va.b.c",
			kind:    VersionKindCLI,
			wantErr: true,
		},
		"non-release ref as input": {
			path: "ref/working-branch/stream/debug/v9.9.9",
			kind: VersionKindImage,
			wantVer: Version{
				ref:     "working-branch",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
		},
		"non-canonical ref as input": {
			path: "ref/testing-1.23/stream/debug/v9.9.9",
			kind: VersionKindImage,
			wantVer: Version{
				ref:     "testing-1-23",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
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
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			want: "v9.9.9",
		},
		"release debug image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			want: "stream/debug/v9.9.9",
		},
		"branch image": {
			ver: Version{
				ref:     "foo",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			want: "ref/foo/stream/debug/v9.9.9",
		},
		"stable release cli": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
			want: "v9.9.9",
		},
		"release debug cli": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
			want: "stream/debug/v9.9.9",
		},
		"branch cli": {
			ver: Version{
				ref:     "foo",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindCLI,
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
		"valid image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
		},
		"invalid ref image": {
			ver: Version{
				ref:     "foo/bar",
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			wantErr: true,
		},
		"invalid stream image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "foo/bar",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			wantErr: true,
		},
		"invalid version image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9/foo",
				kind:    VersionKindImage,
			},
			wantErr: true,
		},
		"valid cli": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
		},
		"invalid ref cli": {
			ver: Version{
				ref:     "foo/bar",
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
			wantErr: true,
		},
		"invalid stream cli": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "foo/bar",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
			wantErr: true,
		},
		"invalid version cli": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9/foo",
				kind:    VersionKindCLI,
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

			ver := Version{version: version}
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

			ver := Version{version: version}
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

			ver := Version{version: tc.ver}
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
		"release image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			gran:     GranularityMajor,
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/versions/major/v9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/versions/major/v9/image.json",
		},
		"release with minor image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			gran:     GranularityMinor,
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/versions/minor/v9.9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/versions/minor/v9.9/image.json",
		},
		"release with patch image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			gran:     GranularityPatch,
			wantPath: "",
			wantURL:  "",
		},
		"release with unknown image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			gran:     GranularityUnknown,
			wantPath: "",
			wantURL:  "",
		},
		"release debug stream image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			gran:     GranularityMajor,
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/versions/major/v9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/versions/major/v9/image.json",
		},
		"release debug stream with minor image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			gran:     GranularityMinor,
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/versions/minor/v9.9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/versions/minor/v9.9/image.json",
		},
		"branch ref image": {
			ver: Version{
				ref:     "foo",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			gran:     GranularityMajor,
			wantPath: constants.CDNAPIPrefix + "/ref/foo/stream/debug/versions/major/v9/image.json",
			wantURL:  constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/foo/stream/debug/versions/major/v9/image.json",
		},
		"branch ref with minor image": {
			ver: Version{
				ref:     "foo",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
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
		ver                Version
		csp                cloudprovider.Provider
		wantMeasurementURL string
		wantSignatureURL   string
		wantErr            bool
	}{
		"nightly-feature": {
			ver: Version{
				ref:     "feat-some-feature",
				stream:  "nightly",
				version: "v2.6.0-pre.0.20230217095603-193dd48ca19f",
				kind:    VersionKindImage,
			},
			csp:                cloudprovider.GCP,
			wantMeasurementURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefixV2 + "/ref/feat-some-feature/stream/nightly/v2.6.0-pre.0.20230217095603-193dd48ca19f/image/measurements.json",
			wantSignatureURL:   constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefixV2 + "/ref/feat-some-feature/stream/nightly/v2.6.0-pre.0.20230217095603-193dd48ca19f/image/measurements.json.sig",
		},
		"fail for wrong kind": {
			ver: Version{
				kind: VersionKindCLI,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			measurementURL, signatureURL, err := MeasurementURL(tc.ver)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantMeasurementURL, measurementURL.String())
			assert.Equal(tc.wantSignatureURL, signatureURL.String())
		})
	}
}

func TestVersionArtifactPathURL(t *testing.T) {
	testCases := map[string]struct {
		ver      Version
		wantPath string
	}{
		"release image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/v9.9.9",
		},
		"release debug stream image": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/v9.9.9",
		},
		"branch ref image": {
			ver: Version{
				ref:     "foo",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/foo/stream/debug/v9.9.9",
		},
		"release cli": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "stable",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/stable/v9.9.9",
		},
		"release debug stream cli": {
			ver: Version{
				ref:     ReleaseRef,
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/" + ReleaseRef + "/stream/debug/v9.9.9",
		},
		"branch ref cli": {
			ver: Version{
				ref:     "foo",
				stream:  "debug",
				version: "v9.9.9",
				kind:    VersionKindCLI,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/foo/stream/debug/v9.9.9",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			path := tc.ver.ArtifactPath(APIV1)
			assert.Equal(tc.wantPath, path)
			url := tc.ver.ArtifactsURL(APIV1)
			assert.Equal(constants.CDNRepositoryURL+"/"+tc.wantPath, url)
		})
	}
}

func TestVersionKindUnMarshalJson(t *testing.T) {
	testCases := map[string]VersionKind{
		`"image"`:   VersionKindImage,
		`"cli"`:     VersionKindCLI,
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
		"cli":     VersionKindCLI,
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
