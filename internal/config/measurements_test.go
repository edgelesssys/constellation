package config

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalYAML(t *testing.T) {
	testCases := map[string]struct {
		measurements  Measurements
		wantBase64Map map[uint32]string
	}{
		"valid measurements": {
			measurements: Measurements{
				2: []byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53},
				3: []byte{213, 164, 73, 109, 33, 222, 201, 165, 37, 141, 219, 25, 198, 254, 181, 59, 180, 211, 192, 70, 63, 230, 7, 242, 72, 141, 223, 79, 16, 6, 239, 158},
			},
			wantBase64Map: map[uint32]string{
				2: "/V3p3zUOO8RBCsBrv+XM3rk/U7nvUSOfdSzmnbxgDzU=",
				3: "1aRJbSHeyaUljdsZxv61O7TTwEY/5gfySI3fTxAG754=",
			},
		},
		"omit bytes": {
			measurements: Measurements{
				2: []byte{},
				3: []byte{1, 2, 3, 4},
			},
			wantBase64Map: map[uint32]string{
				2: "",
				3: "AQIDBA==",
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
		inputBase64Map      map[uint32]string
		forceUnmarshalError bool
		wantMeasurements    Measurements
		wantErr             bool
	}{
		"valid measurements": {
			inputBase64Map: map[uint32]string{
				2: "/V3p3zUOO8RBCsBrv+XM3rk/U7nvUSOfdSzmnbxgDzU=",
				3: "1aRJbSHeyaUljdsZxv61O7TTwEY/5gfySI3fTxAG754=",
			},
			wantMeasurements: Measurements{
				2: []byte{253, 93, 233, 223, 53, 14, 59, 196, 65, 10, 192, 107, 191, 229, 204, 222, 185, 63, 83, 185, 239, 81, 35, 159, 117, 44, 230, 157, 188, 96, 15, 53},
				3: []byte{213, 164, 73, 109, 33, 222, 201, 165, 37, 141, 219, 25, 198, 254, 181, 59, 180, 211, 192, 70, 63, 230, 7, 242, 72, 141, 223, 79, 16, 6, 239, 158},
			},
		},
		"empty bytes": {
			inputBase64Map: map[uint32]string{
				2: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				3: "AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			},
			wantMeasurements: Measurements{
				2: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				3: []byte{1, 2, 3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
		},
		"invalid base64": {
			inputBase64Map: map[uint32]string{
				2: "This is not base64",
				3: "AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			},
			wantMeasurements: Measurements{
				2: []byte{},
				3: []byte{1, 2, 3, 4},
			},
			wantErr: true,
		},
		"simulated unmarshal error": {
			inputBase64Map: map[uint32]string{
				2: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				3: "AQIDBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			},
			forceUnmarshalError: true,
			wantMeasurements: Measurements{
				2: []byte{},
				3: []byte{1, 2, 3, 4},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			var m Measurements
			err := m.UnmarshalYAML(func(i interface{}) error {
				if base64Map, ok := i.(map[uint32]string); ok {
					for key, value := range tc.inputBase64Map {
						base64Map[key] = value
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
		current          Measurements
		newMeasurements  Measurements
		wantMeasurements Measurements
	}{
		"add to empty": {
			current: Measurements{},
			newMeasurements: Measurements{
				1: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				2: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				3: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
			wantMeasurements: Measurements{
				1: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				2: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				3: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
		},
		"keep existing": {
			current: Measurements{
				4: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				5: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
			newMeasurements: Measurements{
				1: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				2: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				3: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
			wantMeasurements: Measurements{
				1: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				2: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				3: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
				4: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				5: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
		},
		"overwrite existing": {
			current: Measurements{
				2: []byte{4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4},
				3: []byte{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
			},
			newMeasurements: Measurements{
				1: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				2: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				3: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
			},
			wantMeasurements: Measurements{
				1: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
				2: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
				3: []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
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
		wantMeasurements   Measurements
		wantError          bool
	}{
		"simple": {
			measurements:       "0: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n",
			measurementsStatus: http.StatusOK,
			signature:          "MEUCIBs1g2/n0FsgPfJ+0uLD5TaunGhxwDcQcUGBroejKvg3AiEAzZtcLU9O6IiVhxB8tBS+ty6MXoPNwL8WRWMzyr35eKI=",
			signatureStatus:    http.StatusOK,
			publicKey:          []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----"),
			wantMeasurements: Measurements{
				0: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
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
						Body:       ioutil.NopCloser(strings.NewReader(tc.measurements)),
						Header:     make(http.Header),
					}
				}
				if req.URL.String() == signatureURL.String() {
					return &http.Response{
						StatusCode: tc.signatureStatus,
						Body:       ioutil.NopCloser(strings.NewReader(tc.signature)),
						Header:     make(http.Header),
					}
				}
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(strings.NewReader("Not found.")),
					Header:     make(http.Header),
				}
			})

			m := Measurements{}
			err := m.FetchAndVerify(context.Background(), client, measurementsURL, signatureURL, tc.publicKey)

			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.EqualValues(tc.wantMeasurements, m)
		})
	}
}
