/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package measurements

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalYAML(t *testing.T) {
	testCases := map[string]struct {
		measurements  M
		wantBase64Map map[uint32]b64Measurement
	}{
		"valid measurements": {
			measurements: M{
				2: Measurement{
					Expected: [32]byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53},
					WarnOnly: false,
				},
				3: Measurement{
					Expected: [32]byte{213, 164, 73, 109, 33, 222, 201, 165, 37, 141, 219, 25, 198, 254, 181, 59, 180, 211, 192, 70, 63, 230, 7, 242, 72, 141, 223, 79, 16, 6, 239, 158},
					WarnOnly: true,
				},
			},
			wantBase64Map: map[uint32]b64Measurement{
				2: {
					Expected: "/V3p3zUOO8RBCsBrv+XM3rk/U7nvUSOfdSzmnbxgDzU=",
					WarnOnly: false,
				},
				3: {
					Expected: "1aRJbSHeyaUljdsZxv61O7TTwEY/5gfySI3fTxAG754=",
					WarnOnly: true,
				},
			},
		},
		"omit bytes": {
			measurements: M{
				2: {
					Expected: [32]byte{}, // implicitly set to all 0s
					WarnOnly: true,
				},
				3: {
					Expected: [32]byte{1, 2, 3, 4}, // implicitly padded with 0s
					WarnOnly: true,
				},
			},
			wantBase64Map: map[uint32]b64Measurement{
				2: {
					Expected: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
					WarnOnly: true,
				},
				3: {
					Expected: "AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
					WarnOnly: true,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			base64Map, err := tc.measurements.MarshalYAML()
			require.NoError(err)

			assert.Equal(tc.wantBase64Map, base64Map)
		})
	}
}

func TestUnmarshalYAML(t *testing.T) {
	testCases := map[string]struct {
		inputBase64Map      map[uint32]b64Measurement
		forceUnmarshalError bool
		wantMeasurements    M
		wantErr             bool
	}{
		"valid measurements": {
			inputBase64Map: map[uint32]b64Measurement{
				2: {
					Expected: "/V3p3zUOO8RBCsBrv+XM3rk/U7nvUSOfdSzmnbxgDzU=",
				},
				3: {
					Expected: "1aRJbSHeyaUljdsZxv61O7TTwEY/5gfySI3fTxAG754=",
				},
			},
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
			inputBase64Map: map[uint32]b64Measurement{
				2: {
					Expected: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				},
				3: {
					Expected: "AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				},
			},
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
			inputBase64Map: map[uint32]b64Measurement{
				2: {
					Expected: "This is not base64",
				},
				3: {
					Expected: "AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				},
			},
			wantMeasurements: M{
				2: {
					Expected: [32]byte{},
				},
				3: {
					Expected: [32]byte{1, 2, 3, 4},
				},
			},
			wantErr: true,
		},
		"simulated unmarshal error": {
			inputBase64Map: map[uint32]b64Measurement{
				2: {
					Expected: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				},
				3: {
					Expected: "AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				},
			},
			forceUnmarshalError: true,
			wantMeasurements: M{
				2: {
					Expected: [32]byte{},
				},
				3: {
					Expected: [32]byte{1, 2, 3, 4},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var m M
			err := m.UnmarshalYAML(func(i any) error {
				if base64Map, ok := i.(map[uint32]b64Measurement); ok {
					for key, value := range tc.inputBase64Map {
						base64Map[key] = b64Measurement{
							Expected: value.Expected,
							WarnOnly: value.WarnOnly,
						}
					}
				}
				if tc.forceUnmarshalError {
					return errors.New("unmarshal error")
				}
				return nil
			})

			if tc.wantErr {
				assert.Error(err)
			} else {
				require.NoError(err)
				assert.Equal(tc.wantMeasurements, m)
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
			signature:          "MEUCIBs1g2/n0FsgPfJ+0uLD5TaunGhxwDcQcUGBroejKvg3AiEAzZtcLU9O6IiVhxB8tBS+ty6MXoPNwL8WRWMzyr35eKI=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----"),
			wantMeasurements: M{
				0: WithAllBytes(0x00, false),
			},
			wantSHA: "4cd9d6ed8d9322150dff7738994c5e2fabff35f3bae6f5c993412d13249a5e87",
		},
		"404 measurements": {
			measurements:       "0: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n",
			measurementsStatus: http.StatusNotFound,
			signature:          "MEUCIBs1g2/n0FsgPfJ+0uLD5TaunGhxwDcQcUGBroejKvg3AiEAzZtcLU9O6IiVhxB8tBS+ty6MXoPNwL8WRWMzyr35eKI=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----"),
			wantError:          true,
		},
		"404 signature": {
			measurements:       "0: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n",
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIBs1g2/n0FsgPfJ+0uLD5TaunGhxwDcQcUGBroejKvg3AiEAzZtcLU9O6IiVhxB8tBS+ty6MXoPNwL8WRWMzyr35eKI=",
			signatureStatus:    http.StatusNotFound,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----"),
			wantError:          true,
		},
		"broken signature": {
			measurements:       "0: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n",
			measurementsStatus: http.StatusOK,
			signature:          "AAAAAAs1g2/n0FsgPfJ+0uLD5TaunGhxwDcQcUGBroejKvg3AiEAzZtcLU9O6IiVhxB8tBS+ty6MXoPNwL8WRWMzyr35eKI=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----"),
			wantError:          true,
		},
		"not yaml": {
			measurements:       "This is some content to be signed!\n",
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIQDzMN3yaiO9sxLGAaSA9YD8rLwzvOaZKWa/bzkcjImUFAIgXLLGzClYUd1dGbuEiY3O/g/eiwQYlyxqLQalxjFmz+8=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAElWUhon39eAqzEC+/GP03oY4/MQg+\ngCDlEzkuOCybCHf+q766bve799L7Y5y5oRsHY1MrUCUwYF/tL7Sg7EYMsA==\n-----END PUBLIC KEY-----"),
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
