package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
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
		wantFlags        *fetchMeasurementsFlags
		wantErr          bool
	}{
		"default": {
			wantFlags: &fetchMeasurementsFlags{
				measurementsURL: nil,
				signatureURL:    nil,
				config:          constants.ConfigFilename,
			},
		},
		"url": {
			urlFlag:          "https://some.other.url/with/path",
			signatureURLFlag: "https://some.other.url/with/path.sig",
			wantFlags: &fetchMeasurementsFlags{
				measurementsURL: urlMustParse("https://some.other.url/with/path"),
				signatureURL:    urlMustParse("https://some.other.url/with/path.sig"),
				config:          constants.ConfigFilename,
			},
		},
		"broken url": {
			urlFlag: "%notaurl%",
			wantErr: true,
		},
		"config": {
			configFlag: "someOtherConfig.yaml",
			wantFlags: &fetchMeasurementsFlags{
				config: "someOtherConfig.yaml",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := newConfigFetchMeasurementsCmd()
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persisten flag manually

			if tc.urlFlag != "" {
				require.NoError(cmd.Flags().Set("url", tc.urlFlag))
			}
			if tc.signatureURLFlag != "" {
				require.NoError(cmd.Flags().Set("signature-url", tc.signatureURLFlag))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}

			flags, err := parseFetchMeasurementsFlags(cmd)
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
	testCases := map[string]struct {
		conf                   *config.Config
		flags                  *fetchMeasurementsFlags
		wantMeasurementsURL    string
		wantMeasurementsSigURL string
	}{
		"both values nil": {
			conf: &config.Config{
				Provider: config.ProviderConfig{
					GCP: &config.GCPConfig{
						Image: "some/image/path/image-123456",
					},
				},
			},
			flags:                  &fetchMeasurementsFlags{},
			wantMeasurementsURL:    constants.S3PublicBucket + "some/image/path/image-123456/measurements.yaml",
			wantMeasurementsSigURL: constants.S3PublicBucket + "some/image/path/image-123456/measurements.yaml.sig",
		},
		"both set by user": {
			conf: &config.Config{},
			flags: &fetchMeasurementsFlags{
				measurementsURL: urlMustParse("get.my/measurements.yaml"),
				signatureURL:    urlMustParse("get.my/measurements.yaml.sig"),
			},
			wantMeasurementsURL:    "get.my/measurements.yaml",
			wantMeasurementsSigURL: "get.my/measurements.yaml.sig",
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
	assert := assert.New(t)
	require := require.New(t)

	measurements := `1: fPRxd3lV3uybnSVhcBmM6XLzcvMitXW78G0RRuQxYGc=
2: PUWM/lXMA+ofRD8VYr7sjfUcdeFKn8+acjShPxmOeWk=
3: PUWM/lXMA+ofRD8VYr7sjfUcdeFKn8+acjShPxmOeWk=
4: HaV5ivUAGzMxmKkfKjcG3wmW08MRUWr+vsfIMVQpOH0=
5: PemdXV59WnLLzPz0F4GGCTKm8KbHskPRvon1dtNw7oY=
7: 8dI/6SUmQ5sd8+bulPDpJ8ghs0UX0+fgLlW8kutAYKw=
8: XJ5IBWy6b6vqojkTsk/GLOWyfNUB2qaf58+JjMYiAB4=
9: Gw5gq8D1WXfz46sF/OKiWbkBssyt4ayGybzNyV9cUCQ=
`
	signature := "MEUCIFdJ5dH6HDywxQWTUh9Bw77wMrq0mNCUjMQGYP+6QsVmAiEAmazj/L7rFGA4/Gz8y+kI5h5E5cDgc3brihvXBKF6qZA="

	cmd := newConfigFetchMeasurementsCmd()
	cmd.Flags().String("config", constants.ConfigFilename, "") // register persisten flag manually
	fileHandler := file.NewHandler(afero.NewMemMapFs())

	gcpConfig := config.Default()
	gcpConfig.RemoveProviderExcept(cloudprovider.GCP)
	gcpConfig.Provider.GCP.Image = "projects/constellation-images/global/images/constellation-coreos-1658216163"

	err := fileHandler.WriteYAML(constants.ConfigFilename, gcpConfig, file.OptMkdirAll)
	require.NoError(err)

	client := newTestClient(func(req *http.Request) *http.Response {
		if req.URL.String() == "https://public-edgeless-constellation.s3.us-east-2.amazonaws.com/projects/constellation-images/global/images/constellation-coreos-1658216163/measurements.yaml" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(measurements)),
				Header:     make(http.Header),
			}
		}
		if req.URL.String() == "https://public-edgeless-constellation.s3.us-east-2.amazonaws.com/projects/constellation-images/global/images/constellation-coreos-1658216163/measurements.yaml.sig" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(signature)),
				Header:     make(http.Header),
			}
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("Not found.")),
			Header:     make(http.Header),
		}
	})

	assert.NoError(configFetchMeasurements(cmd, fileHandler, client))
}
