package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsCloudProvider(t *testing.T) {
	testCases := map[string]struct {
		pos     int
		args    []string
		wantErr bool
	}{
		"gcp":     {0, []string{"gcp"}, false},
		"azure":   {1, []string{"foo", "azure"}, false},
		"foo":     {0, []string{"foo"}, true},
		"empty":   {0, []string{""}, true},
		"unknown": {0, []string{"unknown"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testCmd := &cobra.Command{Args: isCloudProvider(tc.pos)}

			err := testCmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestValidateEndpoint(t *testing.T) {
	testCases := map[string]struct {
		endpoint    string
		defaultPort int
		wantResult  string
		wantErr     bool
	}{
		"ip and port": {
			endpoint:    "192.0.2.1:2",
			defaultPort: 3,
			wantResult:  "192.0.2.1:2",
		},
		"hostname and port": {
			endpoint:    "foo:2",
			defaultPort: 3,
			wantResult:  "foo:2",
		},
		"ip": {
			endpoint:    "192.0.2.1",
			defaultPort: 3,
			wantResult:  "192.0.2.1:3",
		},
		"hostname": {
			endpoint:    "foo",
			defaultPort: 3,
			wantResult:  "foo:3",
		},
		"empty endpoint": {
			endpoint:    "",
			defaultPort: 3,
			wantErr:     true,
		},
		"invalid endpoint": {
			endpoint:    "foo:2:2",
			defaultPort: 3,
			wantErr:     true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			res, err := validateEndpoint(tc.endpoint, tc.defaultPort)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.wantResult, res)
		})
	}
}
