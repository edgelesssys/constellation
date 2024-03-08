/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package openstack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderID(t *testing.T) {
	someErr := errors.New("failed")

	type testCase struct {
		cache      any
		cacheTime  time.Time
		newClient  httpClientJSONCreateFunc
		wantResult string
		wantCall   bool
		wantErr    bool
	}

	newTestCases := func(mResp1, mResp2 any, expect1, expect2 string) map[string]testCase {
		return map[string]testCase{
			"cached": {
				cache:      mResp1,
				cacheTime:  time.Now(),
				wantResult: expect1,
				wantCall:   false,
			},
			"from http": {
				newClient:  newStubHTTPClientJSONFunc(mResp1, 200, nil),
				wantResult: expect1,
				wantCall:   true,
			},
			"cache outdated": {
				cache:      mResp1,
				cacheTime:  time.Now().AddDate(0, 0, -1),
				newClient:  newStubHTTPClientJSONFunc(mResp2, 200, nil),
				wantResult: expect2,
				wantCall:   true,
			},
			"cache empty": {
				cacheTime:  time.Now(),
				newClient:  newStubHTTPClientJSONFunc(mResp1, 200, nil),
				wantResult: expect1,
				wantCall:   true,
			},
			"http error": {
				newClient: newStubHTTPClientJSONFunc(metadataResponse{}, 200, someErr),
				wantCall:  true,
				wantErr:   true,
			},
			"http empty response": {
				newClient: newStubHTTPClientJSONFunc(metadataResponse{}, 200, nil),
				wantCall:  true,
				wantErr:   true,
			},
		}
	}

	testUnits := map[string]struct {
		method    func(c *imdsClient, ctx context.Context) (string, error)
		testCases map[string]testCase
	}{
		"providerID": {
			method: (*imdsClient).providerID,
			testCases: newTestCases(
				metadataResponse{UUID: "uuid1"},
				metadataResponse{UUID: "uuid2"},
				"uuid1", "uuid2",
			),
		},
		"name": {
			method: (*imdsClient).name,
			testCases: newTestCases(
				metadataResponse{Name: "name1"},
				metadataResponse{Name: "name2"},
				"name1", "name2",
			),
		},
		"projectID": {
			method: (*imdsClient).projectID,
			testCases: newTestCases(
				metadataResponse{ProjectID: "projectID1"},
				metadataResponse{ProjectID: "projectID2"},
				"projectID1", "projectID2",
			),
		},
		"uid": {
			method: (*imdsClient).uid,
			testCases: newTestCases(
				metadataResponse{Tags: metadataTags{UID: "uid1"}},
				metadataResponse{Tags: metadataTags{UID: "uid2"}},
				"uid1", "uid2",
			),
		},
		"initSecretHash": {
			method: (*imdsClient).initSecretHash,
			testCases: newTestCases(
				metadataResponse{Tags: metadataTags{InitSecretHash: "hash1"}},
				metadataResponse{Tags: metadataTags{InitSecretHash: "hash2"}},
				"hash1", "hash2",
			),
		},
		"authURL": {
			method: (*imdsClient).authURL,
			testCases: newTestCases(
				userDataResponse{AuthURL: "authURL1"},
				userDataResponse{AuthURL: "authURL2"},
				"authURL1", "authURL2",
			),
		},
		"userDomainName": {
			method: (*imdsClient).userDomainName,
			testCases: newTestCases(
				userDataResponse{UserDomainName: "userDomainName1"},
				userDataResponse{UserDomainName: "userDomainName2"},
				"userDomainName1", "userDomainName2",
			),
		},
		"username": {
			method: (*imdsClient).username,
			testCases: newTestCases(
				userDataResponse{Username: "username1"},
				userDataResponse{Username: "username2"},
				"username1", "username2",
			),
		},
		"password": {
			method: (*imdsClient).password,
			testCases: newTestCases(
				userDataResponse{Password: "password1"},
				userDataResponse{Password: "password2"},
				"password1", "password2",
			),
		},
	}

	for name, tu := range testUnits {
		t.Run(name, func(t *testing.T) {
			for name, tc := range tu.testCases {
				t.Run(name, func(t *testing.T) {
					assert := assert.New(t)
					require := require.New(t)

					var client *stubHTTPClientJSON
					if tc.newClient != nil {
						client = tc.newClient(require)
					}
					var cache metadataResponse
					var userDataCache userDataResponse
					if _, ok := tc.cache.(metadataResponse); ok {
						cache = tc.cache.(metadataResponse)
					} else if _, ok := tc.cache.(userDataResponse); ok {
						userDataCache = tc.cache.(userDataResponse)
					}
					imds := &imdsClient{
						client:        client,
						cache:         cache,
						userDataCache: userDataCache,
						cacheTime:     tc.cacheTime,
					}

					result, err := tu.method(imds, context.Background())

					if tc.wantErr {
						assert.Error(err)
					} else {
						assert.NoError(err)
						assert.Equal(tc.wantResult, result)
						if client != nil {
							assert.Equal(tc.wantCall, client.called)
						}
					}
				})
			}
		})
	}
}

func TestRole(t *testing.T) {
	someErr := errors.New("failed")
	mResp1 := metadataResponse{Tags: metadataTags{Role: "control-plane"}}
	mResp2 := metadataResponse{Tags: metadataTags{Role: "worker"}}
	expect1 := role.ControlPlane
	expect2 := role.Worker

	testCases := map[string]struct {
		cache      metadataResponse
		cacheTime  time.Time
		newClient  httpClientJSONCreateFunc
		wantResult role.Role
		wantCall   bool
		wantErr    bool
	}{
		"cached": {
			cache:      mResp1,
			cacheTime:  time.Now(),
			wantResult: expect1,
			wantCall:   false,
		},
		"from http": {
			newClient:  newStubHTTPClientJSONFunc(mResp1, 200, nil),
			wantResult: expect1,
			wantCall:   true,
		},
		"cache outdated": {
			cache:      mResp1,
			cacheTime:  time.Now().AddDate(0, 0, -1),
			newClient:  newStubHTTPClientJSONFunc(mResp2, 200, nil),
			wantResult: expect2,
			wantCall:   true,
		},
		"cache empty": {
			cacheTime:  time.Now(),
			newClient:  newStubHTTPClientJSONFunc(mResp1, 200, nil),
			wantResult: expect1,
			wantCall:   true,
		},
		"http error": {
			newClient: newStubHTTPClientJSONFunc(metadataResponse{}, 200, someErr),
			wantCall:  true,
			wantErr:   true,
		},
		"http status code 500": {
			newClient: newStubHTTPClientJSONFunc(metadataResponse{}, 500, nil),
			wantCall:  true,
			wantErr:   true,
		},
		"http empty response": {
			newClient: newStubHTTPClientJSONFunc(metadataResponse{}, 200, nil),
			wantCall:  true,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var client *stubHTTPClientJSON
			if tc.newClient != nil {
				client = tc.newClient(require)
			}
			imds := &imdsClient{
				client:    client,
				cache:     tc.cache,
				cacheTime: tc.cacheTime,
			}

			result, err := imds.role(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantResult, result)
				if client != nil {
					assert.Equal(tc.wantCall, client.called)
				}
			}
		})
	}
}

func TestVPCIP(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		cache      string
		cacheTime  time.Time
		client     *stubHTTPClient
		wantResult string
		wantCall   bool
		wantErr    bool
	}{
		"cached": {
			cache:      "192.0.2.1",
			cacheTime:  time.Now(),
			wantResult: "192.0.2.1",
			wantCall:   false,
		},
		"from http": {
			client:     &stubHTTPClient{response: "192.0.2.1"},
			wantResult: "192.0.2.1",
			wantCall:   true,
		},
		"cache outdated": {
			cache:      "192.0.2.1",
			cacheTime:  time.Now().AddDate(0, 0, -1),
			client:     &stubHTTPClient{response: "192.0.2.2"},
			wantResult: "192.0.2.2",
			wantCall:   true,
		},
		"cache empty": {
			cacheTime:  time.Now(),
			client:     &stubHTTPClient{response: "192.0.2.1"},
			wantResult: "192.0.2.1",
			wantCall:   true,
		},
		"http error": {
			client:   &stubHTTPClient{err: someErr},
			wantCall: true,
			wantErr:  true,
		},
		"http empty response": {
			client:   &stubHTTPClient{},
			wantCall: true,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			imds := &imdsClient{
				client:         tc.client,
				vpcIPCache:     tc.cache,
				vpcIPCacheTime: tc.cacheTime,
			}

			result, err := imds.vpcIP(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantResult, result)
				if tc.client != nil {
					assert.Equal(tc.wantCall, tc.client.called)
				}
			}
		})
	}
}

func TestTimeForUpdate(t *testing.T) {
	testCases := map[string]struct {
		cacheTime time.Time
		want      bool
	}{
		"cache outdated": {
			cacheTime: time.Now().AddDate(-1, 0, -1),
			want:      true,
		},
		"cache not outdated": {
			cacheTime: time.Now(),
			want:      false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			imds := &imdsClient{cacheTime: tc.cacheTime}

			assert.Equal(tc.want, imds.timeForUpdate(tc.cacheTime))
		})
	}
}

type httpClientJSONCreateFunc func(r *require.Assertions) *stubHTTPClientJSON

type stubHTTPClientJSON struct {
	require  *require.Assertions
	response any
	code     int
	err      error
	called   bool
}

func newStubHTTPClientJSONFunc(response any, statusCode int, err error) httpClientJSONCreateFunc {
	return func(r *require.Assertions) *stubHTTPClientJSON {
		return &stubHTTPClientJSON{
			response: response,
			code:     statusCode,
			err:      err,
			require:  r,
		}
	}
}

func (c *stubHTTPClientJSON) Do(_ *http.Request) (*http.Response, error) {
	c.called = true
	body, err := json.Marshal(c.response)
	c.require.NoError(err)
	code := 200
	if c.code != 0 {
		code = c.code
	}
	return &http.Response{StatusCode: code, Status: http.StatusText(code), Body: io.NopCloser(bytes.NewReader(body))}, c.err
}

type stubHTTPClient struct {
	response string
	code     int
	err      error
	called   bool
}

func (c *stubHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	c.called = true
	code := 200
	if c.code != 0 {
		code = c.code
	}

	return &http.Response{StatusCode: code, Status: http.StatusText(code), Body: io.NopCloser(strings.NewReader(c.response))}, c.err
}
