/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
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
		"default config is not valid": {
			cnf:        config.Default(),
			wantOutput: true,
			wantErr:    true,
		},
		"default Azure config is not valid": {
			cnf: func() *config.Config {
				cnf := config.Default()
				az := cnf.Provider.Azure
				cnf.Provider = config.ProviderConfig{}
				cnf.Provider.Azure = az
				return cnf
			}(),
			provider:   cloudprovider.Azure,
			wantOutput: true,
			wantErr:    true,
		},
		"default GCP config is not valid": {
			cnf: func() *config.Config {
				cnf := config.Default()
				gcp := cnf.Provider.GCP
				cnf.Provider = config.ProviderConfig{}
				cnf.Provider.GCP = gcp
				return cnf
			}(),
			provider:   cloudprovider.GCP,
			wantOutput: true,
			wantErr:    true,
		},
		"default QEMU config is valid": {
			cnf: func() *config.Config {
				cnf := config.Default()
				qemu := cnf.Provider.QEMU
				cnf.Provider = config.ProviderConfig{}
				cnf.Provider.QEMU = qemu
				return cnf
			}(),
			provider: cloudprovider.QEMU,
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
		"config without provider is not ok": {
			cnf: func() *config.Config {
				cnf := config.Default()
				cnf.Provider = config.ProviderConfig{}
				return cnf
			}(),
			wantErr: true,
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

			err := validateConfig(out, tc.cnf)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)

			assert.Equal(tc.wantOutput, out.Len() > 0)
		})
	}
}
