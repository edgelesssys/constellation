package gcp

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteGCEConf(t *testing.T) {
	config := "someConfig"

	testCases := map[string]struct {
		fs            afero.Afero
		expectedValue string
		expectErr     bool
	}{
		"write works": {
			fs: afero.Afero{
				Fs: afero.NewMemMapFs(),
			},
			expectedValue: config,
			expectErr:     false,
		},
		"write fails": {
			fs: afero.Afero{
				Fs: afero.NewReadOnlyFs(afero.NewMemMapFs()),
			},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			writer := Writer{
				fs: tc.fs,
			}
			err := writer.WriteGCEConf(config)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			value, err := tc.fs.ReadFile("/etc/gce.conf")
			assert.NoError(err)
			assert.Equal(tc.expectedValue, string(value))
		})
	}
}
