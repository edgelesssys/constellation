package cmd

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/cli/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	testCases := map[string]struct {
		cnf        *config.Config
		provider   cloudprovider.Provider
		wantOutput bool
		wantErr    bool
	}{
		"default config is valid": {
			cnf: config.Default(),
		},
		"config with an error": {
			cnf: func() *config.Config {
				cnf := config.Default()
				cnf.Version = "v0"
				return cnf
			}(),
			wantOutput: true,
			wantErr:    true,
		},
		"config without provider is ok if no provider required": {
			cnf: func() *config.Config {
				cnf := config.Default()
				cnf.Provider = config.ProviderConfig{}
				return cnf
			}(),
		},
		"config with only required provider": {
			cnf: func() *config.Config {
				cnf := config.Default()
				az := cnf.Provider.Azure
				cnf.Provider = config.ProviderConfig{}
				cnf.Provider.Azure = az
				return cnf
			}(),
			provider: cloudprovider.Azure,
		},
		"config without required provider": {
			cnf: func() *config.Config {
				cnf := config.Default()
				cnf.Provider.Azure = nil
				return cnf
			}(),
			provider: cloudprovider.Azure,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			out := &bytes.Buffer{}

			err := validateConfig(out, tc.cnf, tc.provider)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(tc.wantOutput, out.Len() > 0)
		})
	}
}
