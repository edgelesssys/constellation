/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/stretchr/testify/assert"
)

func TestFetchMeasurements(t *testing.T) {
	measurements := `{
	"version": "v999.999.999",
	"ref": "-",
	"stream": "stable",
	"list": [
		{
			"csp": "GCP",
			"attestationVariant":"gcp-sev-es",
			"measurements": {
				"0": {
					"expected": "0000000000000000000000000000000000000000000000000000000000000000",
					"warnOnly":false
				},
				"1": {
					"expected": "1111111111111111111111111111111111111111111111111111111111111111",
					"warnOnly":false
				},
				"2": {
					"expected": "2222222222222222222222222222222222222222222222222222222222222222",
					"warnOnly":false
				},
				"3": {
					"expected": "3333333333333333333333333333333333333333333333333333333333333333",
					"warnOnly":false
				},
				"4": {
					"expected": "4444444444444444444444444444444444444444444444444444444444444444",
					"warnOnly":false
				},
				"5": {
					"expected": "5555555555555555555555555555555555555555555555555555555555555555",
					"warnOnly":false
				},
				"6": {
					"expected": "6666666666666666666666666666666666666666666666666666666666666666",
					"warnOnly":true
				}
			}
		}
	]
}
`
	signature := "placeholder-signature"

	client := newTestClient(func(req *http.Request) *http.Response {
		if req.URL.Path == "/constellation/v2/ref/-/stream/stable/v999.999.999/image/measurements.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(measurements)),
				Header:     make(http.Header),
			}
		}
		if req.URL.Path == "/constellation/v2/ref/-/stream/stable/v999.999.999/image/measurements.json.sig" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(signature)),
				Header:     make(http.Header),
			}
		}

		fmt.Println("unexpected request", req.URL.String())
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("Not found.")),
			Header:     make(http.Header),
		}
	})

	testCases := map[string]struct {
		cosign   cosignVerifierConstructor
		rekor    rekorVerifier
		noVerify bool
		wantErr  bool
		isErr    error
	}{
		"success": {
			cosign: newStubCosignVerifier,
			rekor:  singleUUIDVerifier(),
		},
		"success without cosign verify": {
			noVerify: true,
			cosign: func(_ []byte) (sigstore.Verifier, error) {
				return &stubCosignVerifier{
					verifyError: assert.AnError,
				}, nil
			},
			rekor: singleUUIDVerifier(),
		},
		"failing search results is ErrRekor": {
			cosign: newStubCosignVerifier,
			rekor: &stubRekorVerifier{
				SearchByHashUUIDs: []string{},
				SearchByHashError: assert.AnError,
			},
			wantErr: true,
			isErr:   ErrRekor,
		},
		"failing verify is ErrRekor": {
			cosign: newStubCosignVerifier,
			rekor: &stubRekorVerifier{
				SearchByHashUUIDs: []string{"11111111111111111111111111111111111111111111111111111111111111111111111111111111"},
				VerifyEntryError:  assert.AnError,
			},
			wantErr: true,
			isErr:   ErrRekor,
		},
		"signature verification failure": {
			cosign: func(_ []byte) (sigstore.Verifier, error) {
				return &stubCosignVerifier{
					verifyError: assert.AnError,
				}, nil
			},
			rekor:   singleUUIDVerifier(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			sut := NewVerifyFetcher(tc.cosign, tc.rekor, client)
			m, err := sut.FetchAndVerifyMeasurements(context.Background(), "v999.999.999", cloudprovider.GCP, variant.GCPSEVES{}, tc.noVerify)
			if tc.wantErr {
				assert.Error(err)
				if tc.isErr != nil {
					assert.ErrorIs(err, tc.isErr)
				}
				return
			}
			assert.NoError(err)
			// verify example measurements
			assert.Equal("6666666666666666666666666666666666666666666666666666666666666666", hex.EncodeToString(m[6].Expected))
			assert.Equal(WarnOnly, m[6].ValidationOpt)
		})
	}
}

// SubRekorVerifier is a stub for RekorVerifier.
type stubRekorVerifier struct {
	SearchByHashUUIDs []string
	SearchByHashError error
	VerifyEntryError  error
}

// SearchByHash returns the exported fields SearchByHashUUIDs, SearchByHashError.
func (v *stubRekorVerifier) SearchByHash(context.Context, string) ([]string, error) {
	return v.SearchByHashUUIDs, v.SearchByHashError
}

// VerifyEntry returns the exported field VerifyEntryError.
func (v *stubRekorVerifier) VerifyEntry(context.Context, string, string) error {
	return v.VerifyEntryError
}

type stubCosignVerifier struct {
	verifyError error
}

func newStubCosignVerifier(_ []byte) (sigstore.Verifier, error) {
	return &stubCosignVerifier{}, nil
}

func (v *stubCosignVerifier) VerifySignature(_, _ []byte) error {
	return v.verifyError
}

// singleUUIDVerifier constructs a RekorVerifier that returns a single UUID and no errors,
// and should work for most tests on the happy path.
func singleUUIDVerifier() *stubRekorVerifier {
	return &stubRekorVerifier{
		SearchByHashUUIDs: []string{"11111111111111111111111111111111111111111111111111111111111111111111111111111111"},
		SearchByHashError: nil,
		VerifyEntryError:  nil,
	}
}
