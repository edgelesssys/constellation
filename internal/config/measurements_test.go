package config

import (
	"errors"
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
