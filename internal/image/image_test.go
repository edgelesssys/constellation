/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package image

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestGetReference(t *testing.T) {
	testCases := map[string]struct {
		csp, variant  string
		wantReference string
		wantErr       bool
	}{
		"reference exists": {
			csp:           "someCSP",
			variant:       "someVariant",
			wantReference: "someReference",
		},
		"csp does not exist": {
			csp:     "nonExistingCSP",
			variant: "someVariant",
			wantErr: true,
		},
		"variant does not exist": {
			csp:     "someCSP",
			variant: "nonExistingVariant",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			lut := &imageLookupTable{
				"someCSP": {
					"someVariant": "someReference",
				},
			}
			reference, err := lut.getReference(tc.csp, tc.variant)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantReference, reference)
		})
	}
}

func TestGetReferenceOnNil(t *testing.T) {
	assert := assert.New(t)

	var lut *imageLookupTable
	_, err := lut.getReference("someCSP", "someVariant")
	assert.Error(err)
}

func TestVariant(t *testing.T) {
	testCases := map[string]struct {
		csp         cloudprovider.Provider
		config      *config.Config
		wantVariant string
		wantErr     bool
	}{
		"AWS region": {
			csp: cloudprovider.AWS,
			config: &config.Config{Image: "someImage", Provider: config.ProviderConfig{
				AWS: &config.AWSConfig{Region: "someRegion"},
			}},
			wantVariant: "someRegion",
		},
		"Azure cvm": {
			csp: cloudprovider.Azure,
			config: &config.Config{
				Image: "someImage", Provider: config.ProviderConfig{
					Azure: &config.AzureConfig{ConfidentialVM: func() *bool { b := true; return &b }()},
				},
			},
			wantVariant: "cvm",
		},
		"Azure trustedlaunch": {
			csp: cloudprovider.Azure,
			config: &config.Config{
				Image: "someImage", Provider: config.ProviderConfig{
					Azure: &config.AzureConfig{ConfidentialVM: func() *bool { b := false; return &b }()},
				},
			},
			wantVariant: "trustedlaunch",
		},
		"GCP": {
			csp: cloudprovider.GCP,
			config: &config.Config{Image: "someImage", Provider: config.ProviderConfig{
				GCP: &config.GCPConfig{},
			}},
			wantVariant: "sev-es",
		},
		"QEMU": {
			csp: cloudprovider.QEMU,
			config: &config.Config{Image: "someImage", Provider: config.ProviderConfig{
				QEMU: &config.QEMUConfig{},
			}},
			wantVariant: "default",
		},
		"invalid": {
			csp:     cloudprovider.Provider(9999),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			vari, err := variant(tc.csp, tc.config)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantVariant, vari)
		})
	}
}

func TestFetchReference(t *testing.T) {
	imageVersionUID := "someImageVersionUID"
	client := newTestClient(func(req *http.Request) *http.Response {
		if req.URL.String() == "https://cdn.confidential.cloud/constellation/v1/images/someImageVersionUID.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(lut)),
				Header:     make(http.Header),
			}
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("Not found.")),
			Header:     make(http.Header),
		}
	})

	testCases := map[string]struct {
		config        *config.Config
		overrideFile  string
		wantReference string
		wantErr       bool
	}{
		"reference fetched remotely": {
			config: &config.Config{Image: imageVersionUID, Provider: config.ProviderConfig{
				QEMU: &config.QEMUConfig{},
			}},
			wantReference: "someReference",
		},
		"reference fetched locally": {
			config: &config.Config{Image: imageVersionUID, Provider: config.ProviderConfig{
				QEMU: &config.QEMUConfig{},
			}},
			overrideFile:  `{"qemu":{"default":"localOverrideReference"}}`,
			wantReference: "localOverrideReference",
		},
		"lut is invalid": {
			config: &config.Config{Image: imageVersionUID, Provider: config.ProviderConfig{
				QEMU: &config.QEMUConfig{},
			}},
			overrideFile: `{`,
			wantErr:      true,
		},
		"image version does not exist": {
			config: &config.Config{Image: "nonExistingImageVersionUID", Provider: config.ProviderConfig{
				QEMU: &config.QEMUConfig{},
			}},
			wantErr: true,
		},
		"invalid config": {
			config:  &config.Config{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fetcher := &Fetcher{
				httpc: client,
				fs:    newImageVersionStubFs(t, imageVersionUID, tc.overrideFile),
			}
			reference, err := fetcher.FetchReference(context.Background(), tc.config)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantReference, reference)
		})
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// roundTripFunc .
type roundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newTestClient returns *http.Client with Transport replaced to avoid making real calls.
func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func newImageVersionStubFs(t *testing.T, imageVersionUID string, overrideFile string) *afero.Afero {
	fs := afero.NewMemMapFs()
	if overrideFile != "" {
		must(t, afero.WriteFile(fs, imageVersionUID+".json", []byte(overrideFile), os.ModePerm))
	}
	return &afero.Afero{Fs: fs}
}

const lut = `{"qemu":{"default":"someReference"}}`
