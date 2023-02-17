/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
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
			wantMeasurementsURL:    ver.ArtifactsURL() + "/image/csp/gcp/measurements.json",
			wantMeasurementsSigURL: ver.ArtifactsURL() + "/image/csp/gcp/measurements.json.sig",
		},
		"both set by user": {
			conf: &config.Config{},
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
	// Cosign private key used to sign the measurements.
	// Generated with: cosign generate-key-pair
	// Password left empty.
	//
	// -----BEGIN ENCRYPTED COSIGN PRIVATE KEY-----
	// eyJrZGYiOnsibmFtZSI6InNjcnlwdCIsInBhcmFtcyI6eyJOIjozMjc2OCwiciI6
	// OCwicCI6MX0sInNhbHQiOiJlRHVYMWRQMGtIWVRnK0xkbjcxM0tjbFVJaU92eFVX
	// VXgvNi9BbitFVk5BPSJ9LCJjaXBoZXIiOnsibmFtZSI6Im5hY2wvc2VjcmV0Ym94
	// Iiwibm9uY2UiOiJwaWhLL2txNmFXa2hqSVVHR3RVUzhTVkdHTDNIWWp4TCJ9LCJj
	// aXBoZXJ0ZXh0Ijoidm81SHVWRVFWcUZ2WFlQTTVPaTVaWHM5a255bndZU2dvcyth
	// VklIeHcrOGFPamNZNEtvVjVmL3lHRHR0K3BHV2toanJPR1FLOWdBbmtsazFpQ0c5
	// a2czUXpPQTZsU2JRaHgvZlowRVRZQ0hLeElncEdPRVRyTDlDenZDemhPZXVSOXJ6
	// TDcvRjBBVy9vUDVqZXR3dmJMNmQxOEhjck9kWE8yVmYxY2w0YzNLZjVRcnFSZzlN
	// dlRxQWFsNXJCNHNpY1JaMVhpUUJjb0YwNHc9PSJ9
	// -----END ENCRYPTED COSIGN PRIVATE KEY-----

	cosignPublicKey := []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----")

	measurements := `{
	"csp": "gcp",
	"image": "v999.999.999",
	"measurements": {
		"0": "0000000000000000000000000000000000000000000000000000000000000000",
		"1": "1111111111111111111111111111111111111111111111111111111111111111",
		"2": "2222222222222222222222222222222222222222222222222222222222222222",
		"3": "3333333333333333333333333333333333333333333333333333333333333333",
		"4": "4444444444444444444444444444444444444444444444444444444444444444",
		"5": "5555555555555555555555555555555555555555555555555555555555555555",
		"6": "6666666666666666666666666666666666666666666666666666666666666666"
	}
}
`
	signature := "MEYCIQDRAQNK2NjHJBGrnw3HQAyBsXMCmVCptBdgA6VZ3IlyiAIhAPG42waF1aFZq7dnjP3b2jsMNUtaKYDQQSazW1AX8jgF"

	client := newTestClient(func(req *http.Request) *http.Response {
		if req.URL.Path == "/constellation/v1/ref/-/stream/stable/v999.999.999/image/csp/gcp/measurements.json" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(measurements)),
				Header:     make(http.Header),
			}
		}
		if req.URL.Path == "/constellation/v1/ref/-/stream/stable/v999.999.999/image/csp/gcp/measurements.json.sig" {
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
		verifier rekorVerifier
	}{
		"success": {
			verifier: singleUUIDVerifier(),
		},
		"failing search should not result in error": {
			verifier: &stubRekorVerifier{
				SearchByHashUUIDs: []string{},
				SearchByHashError: errors.New("some error"),
			},
		},
		"failing verify should not result in error": {
			verifier: &stubRekorVerifier{
				SearchByHashUUIDs: []string{"11111111111111111111111111111111111111111111111111111111111111111111111111111111"},
				VerifyEntryError:  errors.New("some error"),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newConfigFetchMeasurementsCmd()
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")                        // register persistent flag manually
			fileHandler := file.NewHandler(afero.NewMemMapFs())

			gcpConfig := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.GCP)
			gcpConfig.Image = "v999.999.999"
			constants.VersionInfo = "v999.999.999"

			err := fileHandler.WriteYAML(constants.ConfigFilename, gcpConfig, file.OptMkdirAll)
			require.NoError(err)
			cfm := &configFetchMeasurementsCmd{log: logger.NewTest(t)}

			assert.NoError(cfm.configFetchMeasurements(cmd, tc.verifier, cosignPublicKey, fileHandler, client))
		})
	}
}
