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

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/assert"
)

func TestFetchLatestSEVSNPVersion(t *testing.T) {
	latestStr := "2023-06-11-14-09.json"
	olderStr := "2019-01-01-01-01.json"
	testcases := map[string]struct {
		fetcherVersions []string
		timeAtTest      time.Time
		wantErr         bool
		attestation     variant.Variant
		expectedVersion func() SEVSNPVersionAPI
		olderVersion    func() SEVSNPVersionAPI
		latestVersion   func() SEVSNPVersionAPI
	}{
		"get latest version azure": {
			fetcherVersions: []string{latestStr, olderStr},
			attestation:     variant.AzureSEVSNP{},
			expectedVersion: func() SEVSNPVersionAPI { tmp := latestVersion; tmp.Variant = variant.AzureSEVSNP{}; return tmp },
			olderVersion:    func() SEVSNPVersionAPI { tmp := olderVersion; tmp.Variant = variant.AzureSEVSNP{}; return tmp },
			latestVersion:   func() SEVSNPVersionAPI { tmp := latestVersion; tmp.Variant = variant.AzureSEVSNP{}; return tmp },
		},
		"get latest version aws": {
			fetcherVersions: []string{latestStr, olderStr},
			attestation:     variant.AWSSEVSNP{},
			expectedVersion: func() SEVSNPVersionAPI { tmp := latestVersion; tmp.Variant = variant.AWSSEVSNP{}; return tmp },
			olderVersion:    func() SEVSNPVersionAPI { tmp := olderVersion; tmp.Variant = variant.AWSSEVSNP{}; return tmp },
			latestVersion:   func() SEVSNPVersionAPI { tmp := latestVersion; tmp.Variant = variant.AWSSEVSNP{}; return tmp },
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			client := &http.Client{
				Transport: &fakeConfigAPIHandler{
					attestation:   tc.attestation,
					versions:      tc.fetcherVersions,
					latestDate:    latestStr,
					latestVersion: tc.latestVersion(),
					olderDate:     olderStr,
					olderVersion:  tc.olderVersion(),
				},
			}
			fetcher := newFetcherWithClientAndVerifier(client, dummyVerifier{}, constants.CDNRepositoryURL)
			res, err := fetcher.FetchLatestVersion(context.Background(), tc.attestation)
			assert := assert.New(t)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.expectedVersion(), res)
			}
		})
	}
}

var latestVersion = SEVSNPVersionAPI{
	SEVSNPVersion: SEVSNPVersion{
		Microcode:  93,
		TEE:        0,
		SNP:        6,
		Bootloader: 2,
	},
}

var olderVersion = SEVSNPVersionAPI{
	SEVSNPVersion: SEVSNPVersion{
		Microcode:  1,
		TEE:        0,
		SNP:        1,
		Bootloader: 1,
	},
}

type fakeConfigAPIHandler struct {
	attestation   variant.Variant
	versions      []string
	latestDate    string
	latestVersion SEVSNPVersionAPI
	olderDate     string
	olderVersion  SEVSNPVersionAPI
}

// RoundTrip resolves the request and returns a dummy response.
func (f *fakeConfigAPIHandler) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path == fmt.Sprintf("/constellation/v1/attestation/%s/list", f.attestation.String()) {
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
	} else if req.URL.Path == fmt.Sprintf("/constellation/v1/attestation/%s/%s", f.attestation.String(), f.latestDate) {
		res := &http.Response{}
		bt, err := json.Marshal(f.latestVersion)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(bt))
		res.StatusCode = http.StatusOK
		return res, nil

	} else if req.URL.Path == fmt.Sprintf("/constellation/v1/attestation/%s/%s", f.attestation.String(), f.olderDate) {
		res := &http.Response{}
		bt, err := json.Marshal(f.olderVersion)
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(bt))
		res.StatusCode = http.StatusOK
		return res, nil
	} else if req.URL.Path == fmt.Sprintf("/constellation/v1/attestation/%s/%s.sig", f.attestation.String(), f.latestDate) {
		res := &http.Response{}
		res.Body = io.NopCloser(bytes.NewReader([]byte("null")))
		res.StatusCode = http.StatusOK
		return res, nil

	} else if req.URL.Path == fmt.Sprintf("/constellation/v1/attestation/%s/%s.sig", f.attestation.String(), f.olderDate) {
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
