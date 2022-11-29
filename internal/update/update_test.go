/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package update

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestValidate(t *testing.T) {
	testCases := map[string]struct {
		listFunc     func() *VersionsList
		overrideFunc func(list *VersionsList)
		wantErr      bool
	}{
		"valid major list": {
			listFunc: majorList,
		},
		"valid minor list": {
			listFunc: minorList,
		},
		"invalid stream": {
			listFunc: majorList,
			overrideFunc: func(list *VersionsList) {
				list.Stream = "invalid"
			},
			wantErr: true,
		},
		"invalid granularity": {
			listFunc: majorList,
			overrideFunc: func(list *VersionsList) {
				list.Granularity = "invalid"
			},
			wantErr: true,
		},
		"invalid kind": {
			listFunc: majorList,
			overrideFunc: func(list *VersionsList) {
				list.Kind = "invalid"
			},
			wantErr: true,
		},
		"base ver is not semantic version": {
			listFunc: majorList,
			overrideFunc: func(list *VersionsList) {
				list.Base = "invalid"
			},
			wantErr: true,
		},
		"base ver does not reflect major granularity": {
			listFunc: majorList,
			overrideFunc: func(list *VersionsList) {
				list.Base = "v1.0"
			},
			wantErr: true,
		},
		"base ver does not reflect minor granularity": {
			listFunc: minorList,
			overrideFunc: func(list *VersionsList) {
				list.Base = "v1"
			},
			wantErr: true,
		},
		"version in list is not semantic version": {
			listFunc: majorList,
			overrideFunc: func(list *VersionsList) {
				list.Versions[0] = "invalid"
			},
			wantErr: true,
		},
		"version in list is not sub version of base": {
			listFunc: majorList,
			overrideFunc: func(list *VersionsList) {
				list.Versions[0] = "v2.1"
			},
			wantErr: true,
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
			err := list.validate()
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestList(t *testing.T) {
	majorListJSON, err := json.Marshal(majorList())
	require.NoError(t, err)
	minorListJSON, err := json.Marshal(minorList())
	require.NoError(t, err)
	inconsistentList := majorList()
	inconsistentList.Base = "v2"
	inconsistentListJSON, err := json.Marshal(inconsistentList)
	require.NoError(t, err)
	client := newTestClient(func(req *http.Request) *http.Response {
		switch req.URL.Path {
		case "/constellation/v1/updates/stable/major/v1/image.json":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(majorListJSON)),
				Header:     make(http.Header),
			}
		case "/constellation/v1/updates/stable/minor/v1.1/image.json":
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(minorListJSON)),
				Header:     make(http.Header),
			}
		case "/constellation/v1/updates/stable/major/v1/500.json": // 500 error
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("Server Error.")),
				Header:     make(http.Header),
			}
		case "/constellation/v1/updates/stable/major/v1/nojson.json": // invalid format
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("not json")),
				Header:     make(http.Header),
			}
		case "/constellation/v1/updates/stable/major/v2/image.json": // inconsistent list
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(inconsistentListJSON)),
				Header:     make(http.Header),
			}
		case "/constellation/v1/updates/stable/major/v3/image.json": // does not match requested version
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(minorListJSON)),
				Header:     make(http.Header),
			}
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("Not found.")),
			Header:     make(http.Header),
		}
	})

	testCases := map[string]struct {
		stream, granularity, base, kind string
		overrideFile                    string
		wantList                        VersionsList
		wantErr                         bool
	}{
		"major list fetched remotely": {
			wantList: *majorList(),
		},
		"minor list fetched remotely": {
			granularity: "minor",
			base:        "v1.1",
			wantList:    *minorList(),
		},
		"list does not exist": {
			stream:  "unknown",
			wantErr: true,
		},
		"unexpected error code": {
			kind:    "500",
			wantErr: true,
		},
		"invalid json returned": {
			kind:    "nojson",
			wantErr: true,
		},
		"invalid list returned": {
			base:    "v2",
			wantErr: true,
		},
		"response does not match request": {
			base:    "v3",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			stream := "stable"
			granularity := "major"
			base := "v1"
			kind := "image"
			if tc.stream != "" {
				stream = tc.stream
			}
			if tc.granularity != "" {
				granularity = tc.granularity
			}
			if tc.base != "" {
				base = tc.base
			}
			if tc.kind != "" {
				kind = tc.kind
			}

			fetcher := &VersionsFetcher{
				httpc: client,
			}
			list, err := fetcher.list(context.Background(), stream, granularity, base, kind)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantList, *list)
		})
	}
}

// roundTripFunc .
type roundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newTestClient returns *http.Client with Transport replaced to avoid making real calls.
func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func majorList() *VersionsList {
	return &VersionsList{
		Stream:      "stable",
		Granularity: "major",
		Base:        "v1",
		Kind:        "image",
		Versions: []string{
			"v1.0", "v1.1", "v1.2",
		},
	}
}

func minorList() *VersionsList {
	return &VersionsList{
		Stream:      "stable",
		Granularity: "minor",
		Base:        "v1.1",
		Kind:        "image",
		Versions: []string{
			"v1.1.0", "v1.1.1", "v1.1.2",
		},
	}
}
