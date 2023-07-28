/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListJSONPath(t *testing.T) {
	testCases := map[string]struct {
		list     List
		wantPath string
	}{
		"major list": {
			list: List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMajor,
				Base:        "v1",
				Kind:        VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/versions/major/v1/image.json",
		},
		"minor list": {
			list: List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMinor,
				Base:        "v1.1",
				Kind:        VersionKindImage,
			},
			wantPath: constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/versions/minor/v1.1/image.json",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.wantPath, tc.list.JSONPath())
		})
	}
}

func TestListURL(t *testing.T) {
	testCases := map[string]struct {
		list     List
		wantURL  string
		wantPath string
	}{
		"major list": {
			list: List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMajor,
				Base:        "v1",
				Kind:        VersionKindImage,
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/versions/major/v1/image.json",
		},
		"minor list": {
			list: List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMinor,
				Base:        "v1.1",
				Kind:        VersionKindImage,
			},
			wantURL: constants.CDNRepositoryURL + "/" + constants.CDNAPIPrefix + "/ref/test-ref/stream/nightly/versions/minor/v1.1/image.json",
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

func TestListValidate(t *testing.T) {
	majorList := func() *List {
		return &List{
			Ref:         "test-ref",
			Stream:      "nightly",
			Granularity: GranularityMajor,
			Base:        "v1",
			Kind:        VersionKindImage,
			Versions: []string{
				"v1.0", "v1.1", "v1.2",
			},
		}
	}
	minorList := func() *List {
		return &List{
			Ref:         "test-ref",
			Stream:      "nightly",
			Granularity: GranularityMinor,
			Base:        "v1.1",
			Kind:        VersionKindImage,
			Versions: []string{
				"v1.1.0", "v1.1.1", "v1.1.2",
			},
		}
	}

	testCases := map[string]struct {
		listFunc     func() *List
		overrideFunc func(list *List)
		wantErr      bool
	}{
		"valid major list": {
			listFunc: majorList,
		},
		"valid minor list": {
			listFunc: minorList,
		},
		"invalid ref": {
			listFunc:     majorList,
			overrideFunc: func(list *List) { list.Ref = "" },
			wantErr:      true,
		},
		"invalid stream": {
			listFunc:     majorList,
			overrideFunc: func(list *List) { list.Stream = "invalid" },
			wantErr:      true,
		},
		"invalid granularity": {
			listFunc:     majorList,
			overrideFunc: func(list *List) { list.Granularity = GranularityUnknown },
			wantErr:      true,
		},
		"invalid kind": {
			listFunc:     majorList,
			overrideFunc: func(list *List) { list.Kind = VersionKindUnknown },
			wantErr:      true,
		},
		"base ver is not semantic version": {
			listFunc:     majorList,
			overrideFunc: func(list *List) { list.Base = "invalid" },
			wantErr:      true,
		},
		"base ver does not reflect major granularity": {
			listFunc:     majorList,
			overrideFunc: func(list *List) { list.Base = "v1.0" },
			wantErr:      true,
		},
		"base ver does not reflect minor granularity": {
			listFunc:     minorList,
			overrideFunc: func(list *List) { list.Base = "v1" },
			wantErr:      true,
		},
		"version in list is not semantic version": {
			listFunc:     majorList,
			overrideFunc: func(list *List) { list.Versions[0] = "invalid" },
			wantErr:      true,
		},
		"version in list is not sub version of base": {
			listFunc:     majorList,
			overrideFunc: func(list *List) { list.Versions[0] = "v2.1" },
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			list := tc.listFunc()
			if tc.overrideFunc != nil {
				tc.overrideFunc(list)
			}

			err := list.Validate()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestListValidateRequest(t *testing.T) {
	majorListReq := func() *List {
		return &List{
			Ref:         "test-ref",
			Stream:      "nightly",
			Granularity: GranularityMajor,
			Base:        "v1",
			Kind:        VersionKindImage,
		}
	}
	minorListReq := func() *List {
		return &List{
			Ref:         "test-ref",
			Stream:      "nightly",
			Granularity: GranularityMinor,
			Base:        "v1.1",
			Kind:        VersionKindImage,
		}
	}

	testCases := map[string]struct {
		listFunc     func() *List
		overrideFunc func(list *List)
		wantErr      bool
	}{
		"valid major list": {
			listFunc: majorListReq,
		},
		"valid minor list": {
			listFunc: minorListReq,
		},
		"invalid ref": {
			listFunc:     majorListReq,
			overrideFunc: func(list *List) { list.Ref = "" },
			wantErr:      true,
		},
		"invalid stream": {
			listFunc:     majorListReq,
			overrideFunc: func(list *List) { list.Stream = "invalid" },
			wantErr:      true,
		},
		"invalid granularity": {
			listFunc:     majorListReq,
			overrideFunc: func(list *List) { list.Granularity = GranularityUnknown },
			wantErr:      true,
		},
		"invalid kind": {
			listFunc:     majorListReq,
			overrideFunc: func(list *List) { list.Kind = VersionKindUnknown },
			wantErr:      true,
		},
		"base ver is not semantic version": {
			listFunc:     majorListReq,
			overrideFunc: func(list *List) { list.Base = "invalid" },
			wantErr:      true,
		},
		"base ver does not reflect major granularity": {
			listFunc:     majorListReq,
			overrideFunc: func(list *List) { list.Base = "v1.0" },
			wantErr:      true,
		},
		"base ver does not reflect minor granularity": {
			listFunc:     minorListReq,
			overrideFunc: func(list *List) { list.Base = "v1" },
			wantErr:      true,
		},
		"version in list is not empty": {
			listFunc:     majorListReq,
			overrideFunc: func(list *List) { list.Versions = []string{"v1.1"} },
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			list := tc.listFunc()
			if tc.overrideFunc != nil {
				tc.overrideFunc(list)
			}

			err := list.ValidateRequest()

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestListContainer(t *testing.T) {
	testCases := map[string]struct {
		versions []string
		version  string
		want     bool
	}{
		"empty list": {
			versions: []string{},
			version:  "v1.1.1",
			want:     false,
		},
		"version not in list": {
			versions: []string{"v1.1.1"},
			version:  "v1.1.2",
			want:     false,
		},
		"version in list": {
			versions: []string{"v1.1.1", "v1.1.2"},
			version:  "v1.1.1",
			want:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			list := &List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMinor,
				Base:        "v1.1",
				Kind:        VersionKindImage,
				Versions:    tc.versions,
			}

			assert.Equal(tc.want, list.Contains(tc.version))
		})
	}
}

func TestListStructuredVersions(t *testing.T) {
	assert := assert.New(t)

	list := List{
		Ref:         "test-ref",
		Stream:      "nightly",
		Granularity: GranularityMinor,
		Base:        "v1.1",
		Kind:        VersionKindImage,
		Versions:    []string{"v1.1.1", "v1.1.2", "v1.1.3", "v1.1.4", "v1.1.5"},
	}

	versions := list.StructuredVersions()
	assert.Len(versions, 5)

	verStrs := make([]string, len(versions))
	for i, v := range versions {
		assert.Equal(list.Ref, v.Ref())
		assert.Equal(list.Stream, v.Stream())
		assert.Equal(list.Kind, v.Kind())
		verStrs[i] = v.version
	}

	assert.ElementsMatch(list.Versions, verStrs)
}
