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
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchLatestAzureSEVSNPVersion(t *testing.T) {
	latestStr := "2023-06-11-14-09.json"
	olderStr := "2019-01-01-01-01.json"
	testcases := map[string]struct {
		fetcherVersions []string
		timeAtTest      time.Time
		wantErr         bool
		want            AzureSEVSNPVersionAPI
	}{
		"get latest version if older than 2 weeks": {
			fetcherVersions: []string{latestStr, olderStr},
			want:            latestVersion,
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			client := &http.Client{
				Transport: &fakeConfigAPIHandler{
					versions:      tc.fetcherVersions,
					latestVersion: latestStr,
					olderVersion:  olderStr,
				},
			}
			fetcher := newFetcherWithClientAndVerifier(client, dummyVerifier{})
			res, err := fetcher.FetchAzureSEVSNPVersionLatest(context.Background())
			assert := assert.New(t)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.want, res)
			}
		})
	}
}

var latestVersion = AzureSEVSNPVersionAPI{
	AzureSEVSNPVersion: AzureSEVSNPVersion{
		Microcode:  93,
		TEE:        0,
		SNP:        6,
		Bootloader: 2,
	},
}

var olderVersion = AzureSEVSNPVersionAPI{
	AzureSEVSNPVersion: AzureSEVSNPVersion{
		Microcode:  1,
		TEE:        0,
		SNP:        1,
		Bootloader: 1,
	},
}

type fakeConfigAPIHandler struct {
	versions      []string
	latestVersion string
	olderVersion  string
}

// RoundTrip resolves the request and returns a dummy response.
func (f *fakeConfigAPIHandler) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/constellation/v1/attestation/azure-sev-snp/list" {
		res := &http.Response{}
		bt, err := json.Marshal(f.versions)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(bt))
		res.Header = http.Header{}
		res.Header.Set("Content-Type", "application/json")
		res.StatusCode = http.StatusOK
		return res, nil
	} else if req.URL.Path == fmt.Sprintf("/constellation/v1/attestation/azure-sev-snp/%s", f.latestVersion) {
		res := &http.Response{}
		bt, err := json.Marshal(latestVersion)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(bt))
		res.StatusCode = http.StatusOK
		return res, nil

	} else if req.URL.Path == fmt.Sprintf("/constellation/v1/attestation/azure-sev-snp/%s", f.olderVersion) {
		res := &http.Response{}
		bt, err := json.Marshal(olderVersion)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(bt))
		res.StatusCode = http.StatusOK
		return res, nil
	} else if req.URL.Path == fmt.Sprintf("/constellation/v1/attestation/azure-sev-snp/%s.sig", f.latestVersion) {
		res := &http.Response{}
		res.Body = io.NopCloser(bytes.NewReader([]byte("null")))
		res.StatusCode = http.StatusOK
		return res, nil

	} else if req.URL.Path == fmt.Sprintf("/constellation/v1/attestation/azure-sev-snp/%s.sig", f.olderVersion) {
		res := &http.Response{}
		res.Body = io.NopCloser(bytes.NewReader([]byte("null")))
		res.StatusCode = http.StatusOK
		return res, nil

	}
	return nil, errors.New("no endpoint found")
}

type dummyVerifier struct{}

func (s dummyVerifier) VerifySignature(_, _ []byte) error {
	return nil
}
