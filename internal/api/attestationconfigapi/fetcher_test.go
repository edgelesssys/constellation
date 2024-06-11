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
	latestVersionSNP := VersionAPIEntry{
		SEVSNPVersion: SEVSNPVersion{
			Microcode:  93,
			TEE:        0,
			SNP:        6,
			Bootloader: 2,
		},
	}
	olderVersionSNP := VersionAPIEntry{
		SEVSNPVersion: SEVSNPVersion{
			Microcode:  1,
			TEE:        0,
			SNP:        1,
			Bootloader: 1,
		},
	}
	latestVersionTDX := VersionAPIEntry{
		TDXVersion: TDXVersion{
			QESVN:      2,
			PCESVN:     3,
			TEETCBSVN:  [16]byte{4},
			QEVendorID: [16]byte{5},
			XFAM:       [8]byte{6},
		},
	}
	olderVersionTDX := VersionAPIEntry{
		TDXVersion: TDXVersion{
			QESVN:      1,
			PCESVN:     2,
			TEETCBSVN:  [16]byte{3},
			QEVendorID: [16]byte{4},
			XFAM:       [8]byte{5},
		},
	}

	latestStr := "2023-06-11-14-09.json"
	olderStr := "2019-01-01-01-01.json"
	testCases := map[string]struct {
		fetcherVersions []string
		timeAtTest      time.Time
		wantErr         bool
		attestation     variant.Variant
		expectedVersion VersionAPIEntry
		olderVersion    VersionAPIEntry
		latestVersion   VersionAPIEntry
	}{
		"get latest version azure-sev-snp": {
			fetcherVersions: []string{latestStr, olderStr},
			attestation:     variant.AzureSEVSNP{},
			expectedVersion: func() VersionAPIEntry { tmp := latestVersionSNP; tmp.Variant = variant.AzureSEVSNP{}; return tmp }(),
			olderVersion:    func() VersionAPIEntry { tmp := olderVersionSNP; tmp.Variant = variant.AzureSEVSNP{}; return tmp }(),
			latestVersion:   func() VersionAPIEntry { tmp := latestVersionSNP; tmp.Variant = variant.AzureSEVSNP{}; return tmp }(),
		},
		"get latest version aws-sev-snp": {
			fetcherVersions: []string{latestStr, olderStr},
			attestation:     variant.AWSSEVSNP{},
			expectedVersion: func() VersionAPIEntry { tmp := latestVersionSNP; tmp.Variant = variant.AWSSEVSNP{}; return tmp }(),
			olderVersion:    func() VersionAPIEntry { tmp := olderVersionSNP; tmp.Variant = variant.AWSSEVSNP{}; return tmp }(),
			latestVersion:   func() VersionAPIEntry { tmp := latestVersionSNP; tmp.Variant = variant.AWSSEVSNP{}; return tmp }(),
		},
		"get latest version azure-tdx": {
			fetcherVersions: []string{latestStr, olderStr},
			attestation:     variant.AzureTDX{},
			expectedVersion: func() VersionAPIEntry { tmp := latestVersionTDX; tmp.Variant = variant.AzureTDX{}; return tmp }(),
			olderVersion:    func() VersionAPIEntry { tmp := olderVersionTDX; tmp.Variant = variant.AzureTDX{}; return tmp }(),
			latestVersion:   func() VersionAPIEntry { tmp := latestVersionTDX; tmp.Variant = variant.AzureTDX{}; return tmp }(),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			client := &http.Client{
				Transport: &fakeConfigAPIHandler{
					attestation:   tc.attestation,
					versions:      tc.fetcherVersions,
					latestDate:    latestStr,
					latestVersion: tc.latestVersion,
					olderDate:     olderStr,
					olderVersion:  tc.olderVersion,
				},
			}
			fetcher := newFetcherWithClientAndVerifier(client, stubVerifier{}, constants.CDNRepositoryURL)
			res, err := fetcher.FetchLatestVersion(context.Background(), tc.attestation)
			assert := assert.New(t)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.expectedVersion, res)
			}
		})
	}
}

type fakeConfigAPIHandler struct {
	attestation   variant.Variant
	versions      []string
	latestDate    string
	latestVersion VersionAPIEntry
	olderDate     string
	olderVersion  VersionAPIEntry
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

type stubVerifier struct{}

func (s stubVerifier) VerifySignature(_, _ []byte) error {
	return nil
}
