/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCfg = configapi.AzureSEVSNPVersion{
	Microcode:  93,
	TEE:        0,
	SNP:        6,
	Bootloader: 2,
}

func TestFetchLatestAzureSEVSNPVersion(t *testing.T) {
	testcases := map[string]struct {
		signature []byte
		wantErr   bool
		want      configapi.AzureSEVSNPVersion
	}{
		"get version with valid signature": {
			signature: []byte("MEUCIQDNn6wiSh9Nz9mtU9RvxvfkH3fNDFGeqopjTIRoBNkyrAIgSsKgdYNQXvPevaLWmmpnj/9WcgrltAQ+KfI+bQfklAo="),
			want:      testCfg,
		},
		"fail with invalid signature": {
			signature: []byte("invalid"),
			wantErr:   true,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			client := &http.Client{
				Transport: &fakeConfigAPIHandler{
					signature: tc.signature,
				},
			}
			require := require.New(t)
			version, err := versionsapi.NewVersionFromShortPath("stream/debug/v9.9.9", versionsapi.VersionKindImage)
			require.NoError(err)
			fetcher, err := NewConfigAPIFetcherWithClient(client, version)
			require.NoError(err)

			assert := assert.New(t)
			res, err := fetcher.FetchLatestAzureSEVSNPVersion(context.Background())
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(testCfg, res)
			}
		})
	}
}

type fakeConfigAPIHandler struct {
	signature []byte
}

// RoundTrip resolves the request and returns a dummy response.
func (f *fakeConfigAPIHandler) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/constellation/v1/attestation/azure-sev-snp/list" {
		res := &http.Response{}
		data := []string{"2021-01-01-01-01.json", "2019-01-01-01-02.json"} // return multiple versions to check that latest version is correctly selected
		bt, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(bt))
		res.Header = http.Header{}
		res.Header.Set("Content-Type", "application/json")
		res.StatusCode = http.StatusOK
		return res, nil
	} else if req.URL.Path == "/constellation/v1/attestation/azure-sev-snp/2021-01-01-01-01.json" {
		res := &http.Response{}
		bt, err := json.Marshal(testCfg)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(bt))
		res.StatusCode = http.StatusOK
		return res, nil

	} else if req.URL.Path == "/constellation/v1/attestation/azure-sev-snp/2021-01-01-01-01.json.sig" {
		res := &http.Response{}
		res.Body = io.NopCloser(bytes.NewReader(f.signature))
		res.StatusCode = http.StatusOK
		return res, nil

	}
	return nil, errors.New("no endpoint found")
}
