/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/siderolabs/talos/pkg/machinery/config/encoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
)

func TestMarshal(t *testing.T) {
	testCases := map[string]struct {
		m        Measurement
		wantYAML string
		wantJSON string
	}{
		"measurement": {
			m: Measurement{
				Expected: []byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53},
			},
			wantYAML: "expected: \"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35\"\nwarnOnly: false",
			wantJSON: `{"expected":"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35","warnOnly":false}`,
		},
		"warn only": {
			m: Measurement{
				Expected:      []byte{1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				ValidationOpt: WarnOnly,
			},
			wantYAML: "expected: \"0102030400000000000000000000000000000000000000000000000000000000\"\nwarnOnly: true",
			wantJSON: `{"expected":"0102030400000000000000000000000000000000000000000000000000000000","warnOnly":true}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			{
				// YAML
				yaml, err := yaml.Marshal(tc.m)
				require.NoError(err)

				assert.YAMLEq(tc.wantYAML, string(yaml))
			}

			{
				// JSON
				json, err := json.Marshal(tc.m)
				require.NoError(err)

				assert.JSONEq(tc.wantJSON, string(json))
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	testCases := map[string]struct {
		inputYAML        string
		inputJSON        string
		wantMeasurements M
		wantErr          bool
	}{
		"valid measurements hex": {
			inputYAML: "2:\n expected: \"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35\"\n3:\n expected: \"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463fe607f2488ddf4f1006ef9e\"",
			inputJSON: `{"2":{"expected":"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35"},"3":{"expected":"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463fe607f2488ddf4f1006ef9e"}}`,
			wantMeasurements: M{
				2: {
					Expected: []byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53},
				},
				3: {
					Expected: []byte{213, 164, 73, 109, 33, 222, 201, 165, 37, 141, 219, 25, 198, 254, 181, 59, 180, 211, 192, 70, 63, 230, 7, 242, 72, 141, 223, 79, 16, 6, 239, 158},
				},
			},
		},
		"valid measurements hex 48 bytes": {
			inputYAML: "2:\n expected: \"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35fd5de9df350e3bc4410ac06bbfe5ccde\"\n3:\n expected: \"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463fe607f2488ddf4f1006ef9efd5de9df350e3bc4410ac06bbfe5ccde\"",
			inputJSON: `{"2":{"expected":"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35fd5de9df350e3bc4410ac06bbfe5ccde"},"3":{"expected":"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463fe607f2488ddf4f1006ef9efd5de9df350e3bc4410ac06bbfe5ccde"}}`,
			wantMeasurements: M{
				2: {
					Expected: []byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53, 253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222},
				},
				3: {
					Expected: []byte{213, 164, 73, 109, 33, 222, 201, 165, 37, 141, 219, 25, 198, 254, 181, 59, 180, 211, 192, 70, 63, 230, 7, 242, 72, 141, 223, 79, 16, 6, 239, 158, 253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222},
				},
			},
		},
		"empty bytes": {
			inputYAML: "2:\n expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n3:\n expected: \"0102030400000000000000000000000000000000000000000000000000000000\"",
			inputJSON: `{"2":{"expected":"0000000000000000000000000000000000000000000000000000000000000000"},"3":{"expected":"0102030400000000000000000000000000000000000000000000000000000000"}}`,
			wantMeasurements: M{
				2: {
					Expected: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				3: {
					Expected: []byte{1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
			},
		},
		"invalid base64": {
			inputYAML: "2:\n expected: \"This is not base64\"\n3:\n expected: \"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"",
			inputJSON: `{"2":{"expected":"This is not base64"},"3":{"expected":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}`,
			wantErr:   true,
		},
		"invalid length hex": {
			inputYAML: "2:\n expected: \"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef\"\n3:\n expected: \"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463f\"",
			inputJSON: `{"2":{"expected":"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef"},"3":{"expected":"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463f"}}`,
			wantErr:   true,
		},
		"mixed length hex": {
			inputYAML: "2:\n expected: \"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35fd5de9df350e3bc4410ac06bbfe5ccde\"\n3:\n expected: \"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463fe607f2488ddf4f1006ef9e\"",
			inputJSON: `{"2":{"expected":"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35fd5de9df350e3bc4410ac06bbfe5ccde"},"3":{"expected":"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463fe607f2488ddf4f1006ef9e"}}`,
			wantErr:   true,
		},
		"invalid length base64": {
			inputYAML: "2:\n expected: \"AA==\"\n3:\n expected: \"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==\"",
			inputJSON: `{"2":{"expected":"AA=="},"3":{"expected":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=="}}`,
			wantErr:   true,
		},
		"invalid format": {
			inputYAML: "1:\n expected:\n  someKey: 12\n  anotherKey: 34",
			inputJSON: `{"1":{"expected":{"someKey":12,"anotherKey":34}}}`,
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			{
				// YAML
				var m M
				err := yaml.Unmarshal([]byte(tc.inputYAML), &m)

				if tc.wantErr {
					assert.Error(err, "yaml.Unmarshal should have failed")
				} else {
					require.NoError(err, "yaml.Unmarshal failed")
					assert.Equal(tc.wantMeasurements, m)
				}
			}

			{
				// JSON
				var m M
				err := json.Unmarshal([]byte(tc.inputJSON), &m)

				if tc.wantErr {
					assert.Error(err, "json.Unmarshal should have failed")
				} else {
					require.NoError(err, "json.Unmarshal failed")
					assert.Equal(tc.wantMeasurements, m)
				}
			}
		})
	}
}

func TestEncodeM(t *testing.T) {
	testCases := map[string]struct {
		m    M
		want string
	}{
		"basic": {
			m: M{
				1: WithAllBytes(1, false, PCRMeasurementLength),
				2: WithAllBytes(2, WarnOnly, PCRMeasurementLength),
			},
			want: `1:
    expected: "0101010101010101010101010101010101010101010101010101010101010101"
    warnOnly: false
2:
    expected: "0202020202020202020202020202020202020202020202020202020202020202"
    warnOnly: true
`,
		},
		"output is sorted": {
			m: M{
				3:  WithAllBytes(0, false, PCRMeasurementLength),
				1:  WithAllBytes(0, false, PCRMeasurementLength),
				11: WithAllBytes(0, false, PCRMeasurementLength),
				2:  WithAllBytes(0, false, PCRMeasurementLength),
			},
			want: `1:
    expected: "0000000000000000000000000000000000000000000000000000000000000000"
    warnOnly: false
2:
    expected: "0000000000000000000000000000000000000000000000000000000000000000"
    warnOnly: false
3:
    expected: "0000000000000000000000000000000000000000000000000000000000000000"
    warnOnly: false
11:
    expected: "0000000000000000000000000000000000000000000000000000000000000000"
    warnOnly: false
`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			encoded, err := encoder.NewEncoder(tc.m).Encode()
			require.NoError(err)
			assert.Equal(tc.want, string(encoded))
		})
	}
}

func TestMeasurementsCopyFrom(t *testing.T) {
	testCases := map[string]struct {
		current          M
		newMeasurements  M
		wantMeasurements M
	}{
		"add to empty": {
			current: M{},
			newMeasurements: M{
				1: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
				3: WithAllBytes(0x02, WarnOnly, PCRMeasurementLength),
			},
			wantMeasurements: M{
				1: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
				3: WithAllBytes(0x02, WarnOnly, PCRMeasurementLength),
			},
		},
		"keep existing": {
			current: M{
				4: WithAllBytes(0x01, Enforce, PCRMeasurementLength),
				5: WithAllBytes(0x02, WarnOnly, PCRMeasurementLength),
			},
			newMeasurements: M{
				1: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
				3: WithAllBytes(0x02, WarnOnly, PCRMeasurementLength),
			},
			wantMeasurements: M{
				1: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
				3: WithAllBytes(0x02, WarnOnly, PCRMeasurementLength),
				4: WithAllBytes(0x01, Enforce, PCRMeasurementLength),
				5: WithAllBytes(0x02, WarnOnly, PCRMeasurementLength),
			},
		},
		"overwrite existing": {
			current: M{
				2: WithAllBytes(0x04, Enforce, PCRMeasurementLength),
				3: WithAllBytes(0x05, Enforce, PCRMeasurementLength),
			},
			newMeasurements: M{
				1: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
				3: WithAllBytes(0x02, WarnOnly, PCRMeasurementLength),
			},
			wantMeasurements: M{
				1: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
				3: WithAllBytes(0x02, WarnOnly, PCRMeasurementLength),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			tc.current.CopyFrom(tc.newMeasurements)
			assert.Equal(tc.wantMeasurements, tc.current)
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

func urlMustParse(raw string) *url.URL {
	parsed, _ := url.Parse(raw)
	return parsed
}

func TestMeasurementsFetchAndVerify(t *testing.T) {
	// Cosign private key used to sign the
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

	v1Test, err := versionsapi.NewVersion("-", "stable", "v1.0.0-test", versionsapi.VersionKindImage)
	require.NoError(t, err)
	v1AnotherImage, err := versionsapi.NewVersion("-", "stable", "v1.0.0-another-image", versionsapi.VersionKindImage)
	require.NoError(t, err)

	testCases := map[string]struct {
		measurements       string
		csp                cloudprovider.Provider
		attestationVariant variant.Variant
		imageVersion       versionsapi.Version
		measurementsStatus int
		signature          string
		signatureStatus    int
		wantMeasurements   M
		wantSHA            string
		wantError          bool
	}{
		"json measurements": {
			measurements:       `{"version":"v1.0.0-test","ref":"-","stream":"stable","list":[{"csp":"Unknown","attestationVariant":"dummy","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}]}`,
			csp:                cloudprovider.Unknown,
			imageVersion:       v1Test,
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIHuW2420EqN4Kj6OEaVMmufH7d01vyR1J+SWg8H4elyBAiEA1Ki5Hfq0iI70qpViYbrTFrd8e840NjtdAxGqJKiJgbA=",
			signatureStatus:    http.StatusOK,
			wantMeasurements: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
			},
			wantSHA: "7269a1e8c6a379b86af605f993352df1d4a289bbf79fe655fd78338bd7549d52",
		},
		"404 measurements": {
			measurements:       `{"version":"v1.0.0-test","ref":"-","stream":"stable","list":[{"csp":"Unknown","attestationVariant":"dummy","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}]}`,
			csp:                cloudprovider.Unknown,
			imageVersion:       v1Test,
			measurementsStatus: http.StatusNotFound,
			signature:          "MEUCIHuW2420EqN4Kj6OEaVMmufH7d01vyR1J+SWg8H4elyBAiEA1Ki5Hfq0iI70qpViYbrTFrd8e840NjtdAxGqJKiJgbA=",
			signatureStatus:    http.StatusOK,
			wantError:          true,
		},
		"404 signature": {
			measurements:       `{"version":"v1.0.0-test","ref":"-","stream":"stable","list":[{"csp":"Unknown","attestationVariant":"dummy","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}]}`,
			csp:                cloudprovider.Unknown,
			imageVersion:       v1Test,
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIHuW2420EqN4Kj6OEaVMmufH7d01vyR1J+SWg8H4elyBAiEA1Ki5Hfq0iI70qpViYbrTFrd8e840NjtdAxGqJKiJgbA=",
			signatureStatus:    http.StatusNotFound,
			wantError:          true,
		},
		"broken signature": {
			measurements:       `{"version":"v1.0.0-test","ref":"-","stream":"stable","list":[{"csp":"Unknown","attestationVariant":"dummy","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}]}`,
			csp:                cloudprovider.Unknown,
			imageVersion:       v1Test,
			measurementsStatus: http.StatusOK,
			signature:          "AAAAAAA1RR91pWPw1BMWXTSmTBHg/JtfKerbZNQ9PJTWDdW0sgIhANQbETJGb67qzQmMVmcq007VUFbHRMtYWKZeeyRf0gVa",
			signatureStatus:    http.StatusOK,
			wantError:          true,
		},
		"metadata CSP mismatch": {
			measurements:       `{"version":"v1.0.0-test","ref":"-","stream":"stable","list":[{"csp":"Unknown","attestationVariant":"dummy","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}]}`,
			csp:                cloudprovider.GCP,
			imageVersion:       v1Test,
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIHuW2420EqN4Kj6OEaVMmufH7d01vyR1J+SWg8H4elyBAiEA1Ki5Hfq0iI70qpViYbrTFrd8e840NjtdAxGqJKiJgbA=",
			signatureStatus:    http.StatusOK,
			wantError:          true,
		},
		"metadata image mismatch": {
			measurements:       `{"version":"v1.0.0-test","ref":"-","stream":"stable","list":[{"csp":"Unknown","attestationVariant":"dummy","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}]}`,
			csp:                cloudprovider.Unknown,
			imageVersion:       v1AnotherImage,
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIHuW2420EqN4Kj6OEaVMmufH7d01vyR1J+SWg8H4elyBAiEA1Ki5Hfq0iI70qpViYbrTFrd8e840NjtdAxGqJKiJgbA=",
			signatureStatus:    http.StatusOK,
			wantError:          true,
		},
		"not json": {
			measurements:       "This is some content to be signed!\n",
			csp:                cloudprovider.Unknown,
			imageVersion:       v1Test,
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIQCGA/lSu5qCJgNNvgMaTKJ9rj6vQMecUDaQo3ukaiAfUgIgWoxXRoDKLY9naN7YgxokM7r2fwnyYk3M2WKJJO1g6yo=",
			signatureStatus:    http.StatusOK,
			wantError:          true,
		},
	}

	measurementsURL := urlMustParse("https://somesite.com/yaml")
	signatureURL := urlMustParse("https://somesite.com/yaml.sig")

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			if tc.attestationVariant == nil {
				tc.attestationVariant = variant.Dummy{}
			}

			client := newTestClient(func(req *http.Request) *http.Response {
				if req.URL.String() == measurementsURL.String() {
					return &http.Response{
						StatusCode: tc.measurementsStatus,
						Body:       io.NopCloser(strings.NewReader(tc.measurements)),
						Header:     make(http.Header),
					}
				}
				if req.URL.String() == signatureURL.String() {
					return &http.Response{
						StatusCode: tc.signatureStatus,
						Body:       io.NopCloser(strings.NewReader(tc.signature)),
						Header:     make(http.Header),
					}
				}
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("Not found.")),
					Header:     make(http.Header),
				}
			})

			m := M{}

			verifier, err := sigstore.NewCosignVerifier(cosignPublicKey)
			require.NoError(err)

			hash, err := m.fetchAndVerify(
				context.Background(), client, verifier,
				measurementsURL, signatureURL,
				tc.imageVersion,
				tc.csp,
				tc.attestationVariant,
			)

			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.Equal(tc.wantSHA, hash)
			assert.NoError(err)
			assert.EqualValues(tc.wantMeasurements, m)
		})
	}
}

func TestGetEnforced(t *testing.T) {
	testCases := map[string]struct {
		input M
		want  map[uint32]struct{}
	}{
		"only warnings": {
			input: M{
				0: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				1: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
			},
			want: map[uint32]struct{}{},
		},
		"all enforced": {
			input: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x01, Enforce, PCRMeasurementLength),
			},
			want: map[uint32]struct{}{
				0: {},
				1: {},
			},
		},
		"mixed": {
			input: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x02, Enforce, PCRMeasurementLength),
			},
			want: map[uint32]struct{}{
				0: {},
				2: {},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			got := tc.input.GetEnforced()
			enforced := map[uint32]struct{}{}
			for _, id := range got {
				enforced[id] = struct{}{}
			}
			assert.Equal(tc.want, enforced)
		})
	}
}

func TestSetEnforced(t *testing.T) {
	testCases := map[string]struct {
		input    M
		enforced []uint32
		wantM    M
		wantErr  bool
	}{
		"no enforced measurements": {
			input: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x01, Enforce, PCRMeasurementLength),
			},
			enforced: []uint32{},
			wantM: M{
				0: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				1: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
			},
		},
		"all enforced measurements": {
			input: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x01, Enforce, PCRMeasurementLength),
			},
			enforced: []uint32{0, 1},
			wantM: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x01, Enforce, PCRMeasurementLength),
			},
		},
		"mixed": {
			input: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x01, Enforce, PCRMeasurementLength),
				2: WithAllBytes(0x02, Enforce, PCRMeasurementLength),
				3: WithAllBytes(0x03, Enforce, PCRMeasurementLength),
			},
			enforced: []uint32{0, 2},
			wantM: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x02, Enforce, PCRMeasurementLength),
				3: WithAllBytes(0x03, WarnOnly, PCRMeasurementLength),
			},
		},
		"warn only to enforced": {
			input: M{
				0: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				1: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
			},
			enforced: []uint32{0, 1},
			wantM: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x01, Enforce, PCRMeasurementLength),
			},
		},
		"more enforced than measurements": {
			input: M{
				0: WithAllBytes(0x00, WarnOnly, PCRMeasurementLength),
				1: WithAllBytes(0x01, WarnOnly, PCRMeasurementLength),
			},
			enforced: []uint32{0, 1, 2},
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := tc.input.SetEnforced(tc.enforced)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.True(tc.input.EqualTo(tc.wantM))
		})
	}
}

func TestWithAllBytes(t *testing.T) {
	testCases := map[string]struct {
		b               byte
		warnOnly        MeasurementValidationOption
		wantMeasurement Measurement
	}{
		"0x00 warnOnly": {
			b:        0x00,
			warnOnly: WarnOnly,
			wantMeasurement: Measurement{
				Expected:      []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				ValidationOpt: WarnOnly,
			},
		},
		"0x00": {
			b:        0x00,
			warnOnly: Enforce,
			wantMeasurement: Measurement{
				Expected:      []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				ValidationOpt: Enforce,
			},
		},
		"0x01 warnOnly": {
			b:        0x01,
			warnOnly: WarnOnly,
			wantMeasurement: Measurement{
				Expected:      []byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
				ValidationOpt: WarnOnly,
			},
		},
		"0x01": {
			b:        0x01,
			warnOnly: Enforce,
			wantMeasurement: Measurement{
				Expected:      []byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
				ValidationOpt: Enforce,
			},
		},
		"0xFF warnOnly": {
			b:        0xFF,
			warnOnly: WarnOnly,
			wantMeasurement: Measurement{
				Expected:      []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
				ValidationOpt: WarnOnly,
			},
		},
		"0xFF": {
			b:        0xFF,
			warnOnly: Enforce,
			wantMeasurement: Measurement{
				Expected:      []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
				ValidationOpt: Enforce,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			measurement := WithAllBytes(tc.b, tc.warnOnly, PCRMeasurementLength)
			assert.Equal(tc.wantMeasurement, measurement)
		})
	}
}

func TestEqualTo(t *testing.T) {
	testCases := map[string]struct {
		given     M
		other     M
		wantEqual bool
	}{
		"same values": {
			given: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0xFF, Enforce, PCRMeasurementLength),
			},
			other: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0xFF, Enforce, PCRMeasurementLength),
			},
			wantEqual: true,
		},
		"different number of elements": {
			given: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0xFF, Enforce, PCRMeasurementLength),
			},
			other: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
			},
			wantEqual: false,
		},
		"different values": {
			given: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0xFF, Enforce, PCRMeasurementLength),
			},
			other: M{
				0: WithAllBytes(0xFF, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
			},
			wantEqual: false,
		},
		"different warn settings": {
			given: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0xFF, Enforce, PCRMeasurementLength),
			},
			other: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0xFF, WarnOnly, PCRMeasurementLength),
			},
			wantEqual: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			if tc.wantEqual {
				assert.True(tc.given.EqualTo(tc.other))
			} else {
				assert.False(tc.given.EqualTo(tc.other))
			}
		})
	}
}

func TestMergeImageMeasurementsV2(t *testing.T) {
	testCases := map[string]struct {
		measurements     []ImageMeasurementsV2
		wantMeasurements ImageMeasurementsV2
		wantErr          bool
	}{
		"only one element": {
			measurements: []ImageMeasurementsV2{
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageMeasurementsV2Entry{
						{
							CSP:                cloudprovider.AWS,
							AttestationVariant: "aws-nitro-tpm",
							Measurements: M{
								0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
							},
						},
					},
				},
			},
			wantMeasurements: ImageMeasurementsV2{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				List: []ImageMeasurementsV2Entry{
					{
						CSP:                cloudprovider.AWS,
						AttestationVariant: "aws-nitro-tpm",
						Measurements: M{
							0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
						},
					},
				},
			},
		},
		"valid measurements": {
			measurements: []ImageMeasurementsV2{
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageMeasurementsV2Entry{
						{
							CSP:                cloudprovider.AWS,
							AttestationVariant: "aws-nitro-tpm",
							Measurements: M{
								0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
							},
						},
					},
				},
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageMeasurementsV2Entry{
						{
							CSP:                cloudprovider.GCP,
							AttestationVariant: "gcp-sev-es",
							Measurements: M{
								1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
							},
						},
					},
				},
			},
			wantMeasurements: ImageMeasurementsV2{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				List: []ImageMeasurementsV2Entry{
					{
						CSP:                cloudprovider.AWS,
						AttestationVariant: "aws-nitro-tpm",
						Measurements: M{
							0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
						},
					},
					{
						CSP:                cloudprovider.GCP,
						AttestationVariant: "gcp-sev-es",
						Measurements: M{
							1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
						},
					},
				},
			},
		},
		"sorting": {
			measurements: []ImageMeasurementsV2{
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageMeasurementsV2Entry{
						{
							CSP:                cloudprovider.GCP,
							AttestationVariant: "gcp-sev-es",
							Measurements: M{
								1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
							},
						},
					},
				},
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageMeasurementsV2Entry{
						{
							CSP:                cloudprovider.AWS,
							AttestationVariant: "aws-nitro-tpm",
							Measurements: M{
								0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
							},
						},
					},
				},
			},
			wantMeasurements: ImageMeasurementsV2{
				Ref:     "test-ref",
				Stream:  "nightly",
				Version: "v1.0.0",
				List: []ImageMeasurementsV2Entry{
					{
						CSP:                cloudprovider.AWS,
						AttestationVariant: "aws-nitro-tpm",
						Measurements: M{
							0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
						},
					},
					{
						CSP:                cloudprovider.GCP,
						AttestationVariant: "gcp-sev-es",
						Measurements: M{
							1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
						},
					},
				},
			},
		},
		"mismatch in base info": {
			measurements: []ImageMeasurementsV2{
				{
					Ref:     "test-ref",
					Stream:  "nightly",
					Version: "v1.0.0",
					List: []ImageMeasurementsV2Entry{
						{
							CSP:                cloudprovider.AWS,
							AttestationVariant: "aws-nitro-tpm",
							Measurements: M{
								0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
							},
						},
					},
				},
				{
					Ref:     "other-ref",
					Stream:  "stable",
					Version: "v2.0.0",
					List: []ImageMeasurementsV2Entry{
						{
							CSP:                cloudprovider.GCP,
							AttestationVariant: "gcp-sev-es",
							Measurements: M{
								1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"empty list": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			gotMeasurements, err := MergeImageMeasurementsV2(tc.measurements...)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantMeasurements, gotMeasurements)
		})
	}
}

func TestMeasurementsCompare(t *testing.T) {
	testCases := map[string]struct {
		expected     M
		actual       map[uint32][]byte
		wantErrs     int
		wantWarnings int
	}{
		"no errors": {
			expected: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
			},
			actual: map[uint32][]byte{
				0: bytes.Repeat([]byte{0x00}, PCRMeasurementLength),
				1: bytes.Repeat([]byte{0x11}, PCRMeasurementLength),
			},
			wantErrs:     0,
			wantWarnings: 0,
		},
		"no errors, with warnings": {
			expected: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x11, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x22, WarnOnly, PCRMeasurementLength),
			},
			actual: map[uint32][]byte{
				0: bytes.Repeat([]byte{0x00}, PCRMeasurementLength),
				1: bytes.Repeat([]byte{0xFF}, PCRMeasurementLength),
				2: bytes.Repeat([]byte{0xFF}, PCRMeasurementLength),
			},
			wantErrs:     0,
			wantWarnings: 2,
		},
		"with errors, no warnings": {
			expected: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
				2: WithAllBytes(0x22, Enforce, PCRMeasurementLength),
			},
			actual: map[uint32][]byte{
				0: bytes.Repeat([]byte{0x00}, PCRMeasurementLength),
				1: bytes.Repeat([]byte{0xFF}, PCRMeasurementLength),
				2: bytes.Repeat([]byte{0xFF}, PCRMeasurementLength),
			},
			wantErrs:     2,
			wantWarnings: 0,
		},
		"with errors and warnings": {
			expected: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x11, WarnOnly, PCRMeasurementLength),
				2: WithAllBytes(0x22, Enforce, PCRMeasurementLength),
			},

			actual: map[uint32][]byte{
				0: bytes.Repeat([]byte{0x00}, PCRMeasurementLength),
				1: bytes.Repeat([]byte{0xFF}, PCRMeasurementLength),
				2: bytes.Repeat([]byte{0xFF}, PCRMeasurementLength),
			},
			wantErrs:     1,
			wantWarnings: 1,
		},
		"extra measurements don't cause errors": {
			expected: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
			},
			actual: map[uint32][]byte{
				0: bytes.Repeat([]byte{0x00}, PCRMeasurementLength),
				1: bytes.Repeat([]byte{0x11}, PCRMeasurementLength),
				2: bytes.Repeat([]byte{0x22}, PCRMeasurementLength),
			},
			wantErrs:     0,
			wantWarnings: 0,
		},
		"missing measurements cause errors": {
			expected: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
				2: WithAllBytes(0x22, Enforce, PCRMeasurementLength),
			},
			actual: map[uint32][]byte{
				0: bytes.Repeat([]byte{0x00}, PCRMeasurementLength),
				1: bytes.Repeat([]byte{0x11}, PCRMeasurementLength),
			},
			wantErrs:     1,
			wantWarnings: 0,
		},
		"missing measurements cause warnings": {
			expected: M{
				0: WithAllBytes(0x00, Enforce, PCRMeasurementLength),
				1: WithAllBytes(0x11, Enforce, PCRMeasurementLength),
				2: WithAllBytes(0x22, WarnOnly, PCRMeasurementLength),
			},
			actual: map[uint32][]byte{
				0: bytes.Repeat([]byte{0x00}, PCRMeasurementLength),
				1: bytes.Repeat([]byte{0x11}, PCRMeasurementLength),
			},
			wantErrs:     0,
			wantWarnings: 1,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotWarnings, gotErrs := tc.expected.Compare(tc.actual)
			assert.Equal(tc.wantErrs, len(gotErrs))
			assert.Equal(tc.wantWarnings, len(gotWarnings))
		})
	}
}

func TestRekorErrCheck(t *testing.T) {
	err := newRekorErr(errors.New("test"))
	_, ok := err.(ErrRekor)
	assert.True(t, ok)
}
