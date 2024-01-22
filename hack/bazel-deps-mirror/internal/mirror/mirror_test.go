/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package mirror

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"testing"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestMirrorURL(t *testing.T) {
	testCases := map[string]struct {
		hash    string
		wantURL string
		wantErr bool
	}{
		"empty hash": {
			hash:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantURL: "https://example.com/constellation/cas/sha256/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		"other hash": {
			hash:    "0000000000000000000000000000000000000000000000000000000000000000",
			wantURL: "https://example.com/constellation/cas/sha256/0000000000000000000000000000000000000000000000000000000000000000",
		},
		"invalid hash": {
			hash:    "\x00",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m := Maintainer{
				mirrorBaseURL: "https://example.com/",
			}
			url, err := m.MirrorURL(tc.hash)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantURL, url)
		})
	}
}

func TestMirror(t *testing.T) {
	testCases := map[string]struct {
		unauthenticated bool
		hash            string
		data            []byte
		upstreamURL     string
		statusCode      int
		failUpload      bool
		wantErr         bool
	}{
		"cannot upload in unauthenticated mode": {
			unauthenticated: true,
			hash:            "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			data:            []byte(""),
			upstreamURL:     "https://example.com/empty",
			statusCode:      http.StatusOK,
			wantErr:         true,
		},
		"http error": {
			hash:        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			data:        []byte(""),
			upstreamURL: "https://example.com/empty",
			statusCode:  http.StatusNotFound,
			wantErr:     true,
		},
		"hash mismatch": {
			hash:        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			data:        []byte("evil"),
			upstreamURL: "https://example.com/empty",
			statusCode:  http.StatusOK,
			wantErr:     true,
		},
		"upload error": {
			hash:        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			data:        []byte(""),
			upstreamURL: "https://example.com/empty",
			statusCode:  http.StatusOK,
			failUpload:  true,
			wantErr:     true,
		},
		"success": {
			hash:        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			data:        []byte(""),
			upstreamURL: "https://example.com/empty",
			statusCode:  http.StatusOK,
		},
		"success with different hash": {
			hash:        "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			data:        []byte("foo"),
			upstreamURL: "https://example.com/foo",
			statusCode:  http.StatusOK,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m := Maintainer{
				httpClient: &http.Client{
					Transport: &stubUpstream{
						statusCode: tc.statusCode,
						body:       tc.data,
					},
				},
				uploadClient: &stubUploadClient{
					uploadErr: func() error {
						if tc.failUpload {
							return assert.AnError
						}
						return nil
					}(),
				},
				unauthenticated: tc.unauthenticated,
				log:             logger.NewTest(t),
			}
			err := m.Mirror(context.Background(), tc.hash, []string{tc.upstreamURL})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLearn(t *testing.T) {
	testCases := map[string]struct {
		wantHash           string
		upstreamResponse   []byte
		upstreamStatusCode int
		wantErr            bool
	}{
		"http error": {
			upstreamResponse:   []byte("foo"), // ignored
			upstreamStatusCode: http.StatusNotFound,
			wantErr:            true,
		},
		"success": {
			wantHash:           "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			upstreamResponse:   []byte("foo"),
			upstreamStatusCode: http.StatusOK,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			m := Maintainer{
				unauthenticated: true,
				httpClient: &http.Client{
					Transport: &stubUpstream{
						statusCode: tc.upstreamStatusCode,
						body:       tc.upstreamResponse,
					},
				},
				log: logger.NewTest(t),
			}
			gotHash, err := m.Learn(context.Background(), []string{"https://example.com/foo"})
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.wantHash, gotHash)
		})
	}
}

func TestCheck(t *testing.T) {
	testCases := map[string]struct {
		hash                      string
		unauthenticatedResponse   []byte
		unauthenticatedStatusCode int
		authenticatedResponse     *s3.GetObjectAttributesOutput
		authenticatedErr          error
		wantErr                   bool
	}{
		"unauthenticated mode, http error": {
			hash:                      "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			unauthenticatedResponse:   []byte("foo"), // ignored
			unauthenticatedStatusCode: http.StatusNotFound,
			wantErr:                   true,
		},
		"unauthenticated mode, hash mismatch": {
			hash:                      "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			unauthenticatedResponse:   []byte("foo"),
			unauthenticatedStatusCode: http.StatusOK,
			wantErr:                   true,
		},
		"unauthenticated mode, success": {
			hash:                      "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			unauthenticatedResponse:   []byte("foo"),
			unauthenticatedStatusCode: http.StatusOK,
		},
		"authenticated mode, get attributes fails": {
			hash:             "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			authenticatedErr: assert.AnError,
			wantErr:          true,
		},
		"authenticated mode, hash mismatch": {
			hash: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			authenticatedResponse: &s3.GetObjectAttributesOutput{
				Checksum: &types.Checksum{
					ChecksumSHA256: toPtr("tcH7Lvxta0Z0wv3MSM4BtDo7fAN2PAwzVd4Ame4PjHM="),
				},
				ObjectParts: &types.GetObjectAttributesParts{
					TotalPartsCount: toPtr(int32(1)),
				},
			},
			wantErr: true,
		},
		"authenticated mode, success": {
			hash: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			authenticatedResponse: &s3.GetObjectAttributesOutput{
				Checksum: &types.Checksum{
					ChecksumSHA256: toPtr("LCa0a2j/xo/5m0U8HTBBNBNCLXBkg7+g+YpeiGJm564="),
				},
				ObjectParts: &types.GetObjectAttributesParts{
					TotalPartsCount: toPtr(int32(1)),
				},
			},
		},
		"authenticated mode, fallback to unauthenticated": {
			hash: "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
			authenticatedResponse: &s3.GetObjectAttributesOutput{
				ObjectParts: &types.GetObjectAttributesParts{
					TotalPartsCount: toPtr(int32(2)),
				},
			},
			unauthenticatedResponse:   []byte("foo"),
			unauthenticatedStatusCode: http.StatusOK,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			m := Maintainer{
				unauthenticated: (tc.authenticatedResponse == nil),
				httpClient: &http.Client{
					Transport: &stubUpstream{
						statusCode: tc.unauthenticatedStatusCode,
						body:       tc.unauthenticatedResponse,
					},
				},
				objectStorageClient: &stubObjectStorageClient{
					response: tc.authenticatedResponse,
					err:      tc.authenticatedErr,
				},
				log: logger.NewTest(t),
			}
			err := m.Check(context.Background(), tc.hash)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// stubUpstream implements http.RoundTripper and returns a canned response.
type stubUpstream struct {
	statusCode int
	body       []byte
}

func (s *stubUpstream) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Printf("stubUpstream: %s %s -> %q\n", req.Method, req.URL, string(s.body))
	return &http.Response{
		StatusCode: s.statusCode,
		Body:       io.NopCloser(bytes.NewReader(s.body)),
	}, nil
}

type stubUploadClient struct {
	uploadErr    error
	uploadedData []byte
}

func (s *stubUploadClient) Upload(
	_ context.Context, input *s3.PutObjectInput,
	_ ...func(*s3manager.Uploader),
) (*s3manager.UploadOutput, error) {
	var err error
	s.uploadedData, err = io.ReadAll(input.Body)
	if err != nil {
		panic(err)
	}
	return nil, s.uploadErr
}

func toPtr[T any](v T) *T {
	return &v
}

type stubObjectStorageClient struct {
	response *s3.GetObjectAttributesOutput
	err      error
}

func (s *stubObjectStorageClient) GetObjectAttributes(
	_ context.Context, _ *s3.GetObjectAttributesInput, _ ...func(*s3.Options),
) (*s3.GetObjectAttributesOutput, error) {
	return s.response, s.err
}
