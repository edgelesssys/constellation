/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	versionsapi "github.com/edgelesssys/constellation/v2/internal/api/versions"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func urlMustParse(raw string) *url.URL {
	parsed, _ := url.Parse(raw)
	return parsed
}

func TestParseFetchMeasurementsFlags(t *testing.T) {
	testCases := map[string]struct {
		urlFlag          string
		signatureURLFlag string
		configFlag       string
		forceFlag        bool
		wantFlags        *fetchMeasurementsFlags
		wantErr          bool
	}{
		"default": {
			wantFlags: &fetchMeasurementsFlags{
				measurementsURL: nil,
				signatureURL:    nil,
				configPath:      constants.ConfigFilename,
			},
		},
		"url": {
			urlFlag:          "https://some.other.url/with/path",
			signatureURLFlag: "https://some.other.url/with/path.sig",
			wantFlags: &fetchMeasurementsFlags{
				measurementsURL: urlMustParse("https://some.other.url/with/path"),
				signatureURL:    urlMustParse("https://some.other.url/with/path.sig"),
				configPath:      constants.ConfigFilename,
			},
		},
		"broken url": {
			urlFlag: "%notaurl%",
			wantErr: true,
		},
		"config": {
			configFlag: "someOtherConfig.yaml",
			wantFlags: &fetchMeasurementsFlags{
				configPath: "someOtherConfig.yaml",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newConfigFetchMeasurementsCmd()
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("force", false, "")                       // register persistent flag manually

			if tc.urlFlag != "" {
				require.NoError(cmd.Flags().Set("url", tc.urlFlag))
			}
			if tc.signatureURLFlag != "" {
				require.NoError(cmd.Flags().Set("signature-url", tc.signatureURLFlag))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}
			cfm := &configFetchMeasurementsCmd{log: logger.NewTest(t)}
			flags, err := cfm.parseFetchMeasurementsFlags(cmd)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantFlags, flags)
		})
	}
}

func TestUpdateURLs(t *testing.T) {
	ver := versionsapi.Version{
		Ref:     "foo",
		Stream:  "nightly",
		Version: "v7.7.7",
		Kind:    versionsapi.VersionKindImage,
	}
	testCases := map[string]struct {
		conf                   *config.Config
		flags                  *fetchMeasurementsFlags
		wantMeasurementsURL    string
		wantMeasurementsSigURL string
	}{
		"both values nil": {
			conf: &config.Config{
				Image: ver.ShortPath(),
				Provider: config.ProviderConfig{
					GCP: &config.GCPConfig{},
				},
			},
			flags:                  &fetchMeasurementsFlags{},
			wantMeasurementsURL:    ver.ArtifactsURL(versionsapi.APIV2) + "/image/measurements.json",
			wantMeasurementsSigURL: ver.ArtifactsURL(versionsapi.APIV2) + "/image/measurements.json.sig",
		},
		"both set by user": {
			conf: &config.Config{
				Image: ver.ShortPath(),
			},
			flags: &fetchMeasurementsFlags{
				measurementsURL: urlMustParse("get.my/measurements.json"),
				signatureURL:    urlMustParse("get.my/measurements.json.sig"),
			},
			wantMeasurementsURL:    "get.my/measurements.json",
			wantMeasurementsSigURL: "get.my/measurements.json.sig",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := tc.flags.updateURLs(tc.conf)
			assert.NoError(err)
			assert.Equal(tc.wantMeasurementsURL, tc.flags.measurementsURL.String())
		})
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

func TestConfigFetchMeasurements(t *testing.T) {
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
					"warnOnly":false
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
		cosign       cosignVerifier
		rekor        rekorVerifier
		insecureFlag bool
		wantErr      bool
	}{
		"success": {
			cosign: &stubCosignVerifier{},
			rekor:  singleUUIDVerifier(),
		},
		"success without cosign": {
			insecureFlag: true,
			cosign: &stubCosignVerifier{
				verifyError: assert.AnError,
			},
			rekor: singleUUIDVerifier(),
		},
		"failing search should not result in error": {
			cosign: &stubCosignVerifier{},
			rekor: &stubRekorVerifier{
				SearchByHashUUIDs: []string{},
				SearchByHashError: assert.AnError,
			},
		},
		"failing verify should not result in error": {
			cosign: &stubCosignVerifier{},
			rekor: &stubRekorVerifier{
				SearchByHashUUIDs: []string{"11111111111111111111111111111111111111111111111111111111111111111111111111111111"},
				VerifyEntryError:  assert.AnError,
			},
		},
		"signature verification failure": {
			cosign: &stubCosignVerifier{
				verifyError: assert.AnError,
			},
			rekor:   singleUUIDVerifier(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newConfigFetchMeasurementsCmd()
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")                        // register persistent flag manually
			require.NoError(cmd.Flags().Set("insecure", strconv.FormatBool(tc.insecureFlag)))
			fileHandler := file.NewHandler(afero.NewMemMapFs())

			gcpConfig := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.GCP)
			gcpConfig.Image = "v999.999.999"

			err := fileHandler.WriteYAML(constants.ConfigFilename, gcpConfig, file.OptMkdirAll)
			require.NoError(err)
			cfm := &configFetchMeasurementsCmd{canFetchMeasurements: true, log: logger.NewTest(t)}

			err = cfm.configFetchMeasurements(cmd, tc.cosign, tc.rekor, fileHandler, stubAttestationFetcher{}, client)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

type stubAttestationFetcher struct{}

func (f stubAttestationFetcher) FetchAzureSEVSNPVersionList(_ context.Context, _ attestationconfig.AzureSEVSNPVersionList) (attestationconfig.AzureSEVSNPVersionList, error) {
	return attestationconfig.AzureSEVSNPVersionList(
		[]string{},
	), nil
}

func (f stubAttestationFetcher) FetchAzureSEVSNPVersion(_ context.Context, _ attestationconfig.AzureSEVSNPVersionAPI) (attestationconfig.AzureSEVSNPVersionAPI, error) {
	return attestationconfig.AzureSEVSNPVersionAPI{
		AzureSEVSNPVersion: testCfg,
	}, nil
}

func (f stubAttestationFetcher) FetchAzureSEVSNPVersionLatest(_ context.Context) (attestationconfig.AzureSEVSNPVersionAPI, error) {
	return attestationconfig.AzureSEVSNPVersionAPI{
		AzureSEVSNPVersion: testCfg,
	}, nil
}

var testCfg = attestationconfig.AzureSEVSNPVersion{
	Microcode:  93,
	TEE:        0,
	SNP:        6,
	Bootloader: 2,
}
