package gcp

import (
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareInstance(t *testing.T) {
	err := errors.New("some err")
	vpnIP := "192.0.2.0"
	instance := core.Instance{
		Name:       "someInstance",
		ProviderID: "gce://someProjectID/someZone/someInstance",
		IPs:        []string{"192.0.2.0"},
	}

	testCases := map[string]struct {
		writer         stubWriter
		expectErr      bool
		expectedConfig string
	}{
		"prepare works": {
			expectedConfig: `[global]
project-id = someProjectID
use-metadata-server = false
`,
		},
		"GCE conf write error is detected": {
			writer:    stubWriter{writeErr: err},
			expectErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			ccm := CloudControllerManager{writer: &tc.writer}
			err := ccm.PrepareInstance(instance, vpnIP)

			if tc.expectErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch([]string{tc.expectedConfig}, tc.writer.configs)
		})
	}
}

func TestTrivialCCMFunctions(t *testing.T) {
	assert := assert.New(t)
	cloud := CloudControllerManager{}

	assert.NotEmpty(cloud.Image())
	assert.NotEmpty(cloud.Path())
	assert.NotEmpty(cloud.Name())
	assert.True(cloud.Supported())
}

type stubWriter struct {
	writeErr error

	configs []string
}

func (s *stubWriter) WriteGCEConf(config string) error {
	s.configs = append(s.configs, config)
	return s.writeErr
}
