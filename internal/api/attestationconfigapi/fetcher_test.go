/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package attestationconfigapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testCfg = AzureSEVSNPVersionAPI{
	AzureSEVSNPVersion: AzureSEVSNPVersion{
		Microcode:  93,
		TEE:        0,
		SNP:        6,
		Bootloader: 2,
	},
}

func TestFetchLatestAzureSEVSNPVersion(t *testing.T) {
	testcases := map[string]struct {
		signature []byte
		wantErr   bool
		want      AzureSEVSNPVersionAPI
	}{
		"get version with valid signature": {
			signature: []byte("MEQCIBPEbYg89MIQuaGStLhKGLGMKvKFoYCaAniDLwoIwulqAiB+rj7KMaMOMGxmUsjI7KheCXSNM8NzN+tuDw6AywI75A=="), // signed with release key
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
			fetcher := NewFetcherWithClient(client)
			res, err := fetcher.FetchAzureSEVSNPVersionLatest(context.Background())

			assert := assert.New(t)
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
