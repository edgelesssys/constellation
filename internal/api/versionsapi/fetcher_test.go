/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestFetchVersionList(t *testing.T) {
	require := require.New(t)

	majorList := func() *List {
		return &List{
			Ref:         "test-ref",
			Stream:      "nightly",
			Granularity: GranularityMajor,
			Base:        "v1",
			Kind:        VersionKindImage,
			Versions:    []string{"v1.0", "v1.1", "v1.2"},
		}
	}
	minorList := func() *List {
		return &List{
			Ref:         "test-ref",
			Stream:      "nightly",
			Granularity: GranularityMinor,
			Base:        "v1.1",
			Kind:        VersionKindImage,
			Versions:    []string{"v1.1.0", "v1.1.1", "v1.1.2"},
		}
	}
	majorListJSON, err := json.Marshal(majorList())
	require.NoError(err)
	minorListJSON, err := json.Marshal(minorList())
	require.NoError(err)
	inconsistentList := majorList()
	inconsistentList.Base = "v2"
	inconsistentListJSON, err := json.Marshal(inconsistentList)
	require.NoError(err)

	testCases := map[string]struct {
		list       List
		serverPath string
		serverResp *http.Response
		wantList   List
		wantErr    bool
	}{
		"major list fetched": {
			list: List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMajor,
				Base:        "v1",
				Kind:        VersionKindImage,
			},
			serverPath: "/constellation/v1/ref/test-ref/stream/nightly/versions/major/v1/image.json",
			serverResp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(majorListJSON)),
			},
			wantList: *majorList(),
		},
		"minor list fetched": {
			list: List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMinor,
				Base:        "v1.1",
				Kind:        VersionKindImage,
			},
			serverPath: "/constellation/v1/ref/test-ref/stream/nightly/versions/minor/v1.1/image.json",
			serverResp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(minorListJSON)),
			},
			wantList: *minorList(),
		},
		"list does not exist": {
			list: List{
				Ref:         "another-ref",
				Stream:      "nightly",
				Granularity: GranularityMajor,
				Base:        "v1",
				Kind:        VersionKindImage,
			},
			wantErr: true,
		},
		"invalid list requested": {
			list: List{
				Ref:         "",
				Stream:      "unknown",
				Granularity: GranularityMajor,
				Base:        "v1",
				Kind:        VersionKindImage,
			},
			wantErr: true,
		},
		"unexpected error code": {
			list: List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMajor,
				Base:        "v1",
				Kind:        VersionKindImage,
			},
			serverPath: "/constellation/v1/ref/test-ref/stream/nightly/versions/major/v1/image.json",
			serverResp: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("Internal Server Error")),
			},
			wantErr: true,
		},
		"invalid json returned": {
			list: List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMajor,
				Base:        "v1",
				Kind:        VersionKindImage,
			},
			serverPath: "/constellation/v1/ref/test-ref/stream/nightly/versions/major/v1/image.json",
			serverResp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("invalid json")),
			},
			wantErr: true,
		},
		"invalid list returned": {
			list: List{
				Ref:         "test-ref",
				Stream:      "nightly",
				Granularity: GranularityMajor,
				Base:        "v2",
				Kind:        VersionKindImage,
			},
			serverPath: "/constellation/v1/ref/test-ref/stream/nightly/versions/major/v2/image.json",
			serverResp: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(inconsistentListJSON)),
			},
			wantErr: true,
		},
		// TODO(katexochen): Remove or find strategy to implement this check in a generic way
		// "response does not match request": {
		// 	list: List{
		// 		Ref:         "test-ref",
		// 		Stream:      "nightly",
		// 		Granularity: GranularityMajor,
		// 		Base:        "v3",
		// 		Kind:        VersionKindImage,
		// 	},
		// 	serverPath: "/constellation/v1/ref/test-ref/stream/nightly/versions/major/v3/image.json",
		// 	serverResp: &http.Response{
		// 		StatusCode: http.StatusOK,
		// 		Body:       io.NopCloser(bytes.NewBuffer(minorListJSON)),
		// 	},
		// 	wantErr: true,
		// },
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := newTestClient(func(req *http.Request) *http.Response {
				if req.URL.Path != tc.serverPath {
					return &http.Response{
						StatusCode: http.StatusNotFound,
						Body:       io.NopCloser(bytes.NewBufferString("Not found.")),
					}
				}
				return tc.serverResp
			})

			fetcher := Fetcher{client, constants.CDNRepositoryURL}

			list, err := fetcher.FetchVersionList(context.Background(), tc.list)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantList, list)
		})
	}
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newTestClient returns *http.Client with Transport replaced to avoid making real calls.
func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}
