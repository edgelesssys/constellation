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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	client := &http.Client{
		Transport: &fakeConfigAPIHandler{},
	}
	fetcher := NewConfigAPIFetcherWithClient(client)
	res, err := fetcher.FetchLatestAzureSEVSNPVersion(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint8(2), res.Bootloader)
}

type fakeConfigAPIHandler struct{}

// RoundTrip resolves the request and returns a dummy response.
func (f *fakeConfigAPIHandler) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/constellation/v1/attestation/azure-sev-snp/list" {
		res := &http.Response{}
		data := []string{"2021-01-01-01-01.json"}
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
		bt, err := json.Marshal(configapi.AzureSEVSNPVersion{
			Microcode:  93,
			TEE:        0,
			SNP:        6,
			Bootloader: 2,
		})
		if err != nil {
			return nil, err
		}
		res.Body = io.NopCloser(bytes.NewReader(bt))
		res.StatusCode = http.StatusOK
		return res, nil

	}
	return nil, errors.New("no endpoint found")
}
