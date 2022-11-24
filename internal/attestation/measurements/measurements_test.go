/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMarshal(t *testing.T) {
	testCases := map[string]struct {
		m        Measurement
		wantYAML string
		wantJSON string
	}{
		"measurement": {
			m: Measurement{
				Expected: [32]byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53},
			},
			wantYAML: "expected: \"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35\"\nwarnOnly: false",
			wantJSON: `{"expected":"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35","warnOnly":false}`,
		},
		"warn only": {
			m: Measurement{
				Expected: [32]byte{1, 2, 3, 4}, // implicitly padded with 0s
				WarnOnly: true,
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
		"valid measurements base64": {
			inputYAML: "2:\n expected: \"/V3p3zUOO8RBCsBrv+XM3rk/U7nvUSOfdSzmnbxgDzU=\"\n3:\n expected: \"1aRJbSHeyaUljdsZxv61O7TTwEY/5gfySI3fTxAG754=\"",
			inputJSON: `{"2":{"expected":"/V3p3zUOO8RBCsBrv+XM3rk/U7nvUSOfdSzmnbxgDzU="},"3":{"expected":"1aRJbSHeyaUljdsZxv61O7TTwEY/5gfySI3fTxAG754="}}`,
			wantMeasurements: M{
				2: {
					Expected: [32]byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53},
				},
				3: {
					Expected: [32]byte{213, 164, 73, 109, 33, 222, 201, 165, 37, 141, 219, 25, 198, 254, 181, 59, 180, 211, 192, 70, 63, 230, 7, 242, 72, 141, 223, 79, 16, 6, 239, 158},
				},
			},
		},
		"valid measurements hex": {
			inputYAML: "2:\n expected: \"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35\"\n3:\n expected: \"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463fe607f2488ddf4f1006ef9e\"",
			inputJSON: `{"2":{"expected":"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef51239f752ce69dbc600f35"},"3":{"expected":"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463fe607f2488ddf4f1006ef9e"}}`,
			wantMeasurements: M{
				2: {
					Expected: [32]byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53},
				},
				3: {
					Expected: [32]byte{213, 164, 73, 109, 33, 222, 201, 165, 37, 141, 219, 25, 198, 254, 181, 59, 180, 211, 192, 70, 63, 230, 7, 242, 72, 141, 223, 79, 16, 6, 239, 158},
				},
			},
		},
		"empty bytes": {
			inputYAML: "2:\n expected: \"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"\n3:\n expected: \"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"",
			inputJSON: `{"2":{"expected":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="},"3":{"expected":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}`,
			wantMeasurements: M{
				2: {
					Expected: [32]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
				3: {
					Expected: [32]byte{1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				},
			},
		},
		"invalid base64": {
			inputYAML: "2:\n expected: \"This is not base64\"\n3:\n expected: \"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\"",
			inputJSON: `{"2":{"expected":"This is not base64"},"3":{"expected":"AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="}}`,
			wantErr:   true,
		},
		"legacy format": {
			inputYAML: "2: \"/V3p3zUOO8RBCsBrv+XM3rk/U7nvUSOfdSzmnbxgDzU=\"\n3: \"1aRJbSHeyaUljdsZxv61O7TTwEY/5gfySI3fTxAG754=\"",
			inputJSON: `{"2":"/V3p3zUOO8RBCsBrv+XM3rk/U7nvUSOfdSzmnbxgDzU=","3":"1aRJbSHeyaUljdsZxv61O7TTwEY/5gfySI3fTxAG754="}`,
			wantMeasurements: M{
				2: {
					Expected: [32]byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53},
				},
				3: {
					Expected: [32]byte{213, 164, 73, 109, 33, 222, 201, 165, 37, 141, 219, 25, 198, 254, 181, 59, 180, 211, 192, 70, 63, 230, 7, 242, 72, 141, 223, 79, 16, 6, 239, 158},
				},
			},
		},
		"invalid length hex": {
			inputYAML: "2:\n expected: \"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef\"\n3:\n expected: \"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463f\"",
			inputJSON: `{"2":{"expected":"fd5de9df350e3bc4410ac06bbfe5ccdeb93f53b9ef"},"3":{"expected":"d5a4496d21dec9a5258ddb19c6feb53bb4d3c0463f"}}`,
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

func TestMeasurementsCopyFrom(t *testing.T) {
	testCases := map[string]struct {
		current          M
		newMeasurements  M
		wantMeasurements M
	}{
		"add to empty": {
			current: M{},
			newMeasurements: M{
				1: WithAllBytes(0x00, true),
				2: WithAllBytes(0x01, true),
				3: WithAllBytes(0x02, true),
			},
			wantMeasurements: M{
				1: WithAllBytes(0x00, true),
				2: WithAllBytes(0x01, true),
				3: WithAllBytes(0x02, true),
			},
		},
		"keep existing": {
			current: M{
				4: WithAllBytes(0x01, false),
				5: WithAllBytes(0x02, true),
			},
			newMeasurements: M{
				1: WithAllBytes(0x00, true),
				2: WithAllBytes(0x01, true),
				3: WithAllBytes(0x02, true),
			},
			wantMeasurements: M{
				1: WithAllBytes(0x00, true),
				2: WithAllBytes(0x01, true),
				3: WithAllBytes(0x02, true),
				4: WithAllBytes(0x01, false),
				5: WithAllBytes(0x02, true),
			},
		},
		"overwrite existing": {
			current: M{
				2: WithAllBytes(0x04, false),
				3: WithAllBytes(0x05, false),
			},
			newMeasurements: M{
				1: WithAllBytes(0x00, true),
				2: WithAllBytes(0x01, true),
				3: WithAllBytes(0x02, true),
			},
			wantMeasurements: M{
				1: WithAllBytes(0x00, true),
				2: WithAllBytes(0x01, true),
				3: WithAllBytes(0x02, true),
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

	testCases := map[string]struct {
		measurements       string
		measurementsStatus int
		signature          string
		signatureStatus    int
		publicKey          []byte
		wantMeasurements   M
		wantSHA            string
		wantError          bool
	}{
		"simple": {
			measurements:       "0: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n",
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIQDcHS2bLls7OrLHpQKuiFGXhPrTcehPDwgVyERHl4V02wIgeIxK4J9oJpXWRBjokbog2lgifRXuJK8ljlAID26MbHk=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----"),
			wantMeasurements: M{
				0: WithAllBytes(0x00, false),
			},
			wantSHA: "4cd9d6ed8d9322150dff7738994c5e2fabff35f3bae6f5c993412d13249a5e87",
		},
		"json measurements": {
			measurements:       `{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}`,
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIQDh3nCgrdTiYWiV4NkiaZ6vxovj79Pk8V90mdWAnmCEOwIgMAVWAx5dW0saut+8X15SgtBEiKqEixYiSICSqqhxUMg=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----"),
			wantMeasurements: M{
				0: WithAllBytes(0x00, false),
			},
			wantSHA: "1da09758c89537946496358f80b892e508563fcbbc695c90b6c16bf158e69c11",
		},
		"yaml measurements": {
			measurements:       "0:\n expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n warnOnly: false\n",
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIFzQdwBS92aJjY0bcIag1uQRl42lUSBmmjEvO0tM/N0ZAiEAvuWaP744qYMw5uEmc7BY4mm4Ij3TEqAWFgxNhFkckp4=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----"),
			wantMeasurements: M{
				0: WithAllBytes(0x00, false),
			},
			wantSHA: "c651cd419fd536c63cfc5349ad44da140a09987465e31192660059d383413807",
		},
		"404 measurements": {
			measurements:       `{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}`,
			measurementsStatus: http.StatusNotFound,
			signature:          "MEUCIQDh3nCgrdTiYWiV4NkiaZ6vxovj79Pk8V90mdWAnmCEOwIgMAVWAx5dW0saut+8X15SgtBEiKqEixYiSICSqqhxUMg=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----"),
			wantError:          true,
		},
		"404 signature": {
			measurements:       `{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}`,
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIQDh3nCgrdTiYWiV4NkiaZ6vxovj79Pk8V90mdWAnmCEOwIgMAVWAx5dW0saut+8X15SgtBEiKqEixYiSICSqqhxUMg=",
			signatureStatus:    http.StatusNotFound,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----"),
			wantError:          true,
		},
		"broken signature": {
			measurements:       `{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}`,
			measurementsStatus: http.StatusOK,
			signature:          "AAAAAAAA3nCgrdTiYWiV4NkiaZ6vxovj79Pk8V90mdWAnmCEOwIgMAVWAx5dW0saut+8X15SgtBEiKqEixYiSICSqqhxUMg=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----"),
			wantError:          true,
		},
		"not yaml or json": {
			measurements:       "This is some content to be signed!\n",
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIQCGA/lSu5qCJgNNvgMaTKJ9rj6vQMecUDaQo3ukaiAfUgIgWoxXRoDKLY9naN7YgxokM7r2fwnyYk3M2WKJJO1g6yo=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----"),
			wantError:          true,
		},
	}

	measurementsURL := urlMustParse("https://somesite.com/measurements.yaml")
	signatureURL := urlMustParse("https://somesite.com/measurements.yaml.sig")

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

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
			hash, err := m.FetchAndVerify(context.Background(), client, measurementsURL, signatureURL, tc.publicKey)

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
				0: WithAllBytes(0x00, true),
				1: WithAllBytes(0x01, true),
			},
			want: map[uint32]struct{}{},
		},
		"all enforced": {
			input: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0x01, false),
			},
			want: map[uint32]struct{}{
				0: {},
				1: {},
			},
		},
		"mixed": {
			input: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0x01, true),
				2: WithAllBytes(0x02, false),
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
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0x01, false),
			},
			enforced: []uint32{},
			wantM: M{
				0: WithAllBytes(0x00, true),
				1: WithAllBytes(0x01, true),
			},
		},
		"all enforced measurements": {
			input: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0x01, false),
			},
			enforced: []uint32{0, 1},
			wantM: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0x01, false),
			},
		},
		"mixed": {
			input: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0x01, false),
				2: WithAllBytes(0x02, false),
				3: WithAllBytes(0x03, false),
			},
			enforced: []uint32{0, 2},
			wantM: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0x01, true),
				2: WithAllBytes(0x02, false),
				3: WithAllBytes(0x03, true),
			},
		},
		"warn only to enforced": {
			input: M{
				0: WithAllBytes(0x00, true),
				1: WithAllBytes(0x01, true),
			},
			enforced: []uint32{0, 1},
			wantM: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0x01, false),
			},
		},
		"more enforced than measurements": {
			input: M{
				0: WithAllBytes(0x00, true),
				1: WithAllBytes(0x01, true),
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
		warnOnly        bool
		wantMeasurement Measurement
	}{
		"0x00 warnOnly": {
			b:        0x00,
			warnOnly: true,
			wantMeasurement: Measurement{
				Expected: [32]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				WarnOnly: true,
			},
		},
		"0x00": {
			b:        0x00,
			warnOnly: false,
			wantMeasurement: Measurement{
				Expected: [32]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				WarnOnly: false,
			},
		},
		"0x01 warnOnly": {
			b:        0x01,
			warnOnly: true,
			wantMeasurement: Measurement{
				Expected: [32]byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
				WarnOnly: true,
			},
		},
		"0x01": {
			b:        0x01,
			warnOnly: false,
			wantMeasurement: Measurement{
				Expected: [32]byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
				WarnOnly: false,
			},
		},
		"0xFF warnOnly": {
			b:        0xFF,
			warnOnly: true,
			wantMeasurement: Measurement{
				Expected: [32]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
				WarnOnly: true,
			},
		},
		"0xFF": {
			b:        0xFF,
			warnOnly: false,
			wantMeasurement: Measurement{
				Expected: [32]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
				WarnOnly: false,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			measurement := WithAllBytes(tc.b, tc.warnOnly)
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
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0xFF, false),
			},
			other: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0xFF, false),
			},
			wantEqual: true,
		},
		"different number of elements": {
			given: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0xFF, false),
			},
			other: M{
				0: WithAllBytes(0x00, false),
			},
			wantEqual: false,
		},
		"different values": {
			given: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0xFF, false),
			},
			other: M{
				0: WithAllBytes(0xFF, false),
				1: WithAllBytes(0x00, false),
			},
			wantEqual: false,
		},
		"different warn settings": {
			given: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0xFF, false),
			},
			other: M{
				0: WithAllBytes(0x00, false),
				1: WithAllBytes(0xFF, true),
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
