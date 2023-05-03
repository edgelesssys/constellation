/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestUpgradeApply(t *testing.T) {
	someErr := errors.New("some error")
	testCases := map[string]struct {
		upgrader stubUpgrader
		wantErr  bool
	}{
		"success": {
			upgrader: stubUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
		},
		"nodeVersion some error": {
			upgrader: stubUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: someErr,
			},
			wantErr: true,
		},
		"nodeVersion in progress error": {
			upgrader: stubUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: kubernetes.ErrInProgress,
			},
		},
		"helm other error": {
			upgrader: stubUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
				helmErr:       someErr,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			cmd := newUpgradeApplyCmd()
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")                        // register persistent flag manually
			cmd.Flags().String("tf-log", "DEBUG", "")                  // register persistent flag manually

			err := cmd.Flags().Set("yes", "true")
			require.NoError(err)

			handler := file.NewHandler(afero.NewMemMapFs())
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.Azure)
			require.NoError(handler.WriteYAML(constants.ConfigFilename, cfg))
			require.NoError(handler.WriteJSON(constants.ClusterIDsFileName, clusterid.File{}))

			upgrader := upgradeApplyCmd{upgrader: tc.upgrader, log: logger.NewTest(t)}
			err = upgrader.upgradeApply(cmd, handler)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type stubUpgrader struct {
	currentConfig    config.AttestationCfg
	nodeVersionErr   error
	helmErr          error
	planTerraformErr error
}

func (u stubUpgrader) UpgradeNodeVersion(_ context.Context, _ *config.Config) error {
	return u.nodeVersionErr
}

func (u stubUpgrader) UpgradeHelmServices(_ context.Context, _ *config.Config, _ time.Duration, _ bool) error {
	return u.helmErr
}

func (u stubUpgrader) UpdateAttestationConfig(_ context.Context, _ config.AttestationCfg) error {
	return nil
}

func (u stubUpgrader) GetClusterAttestationConfig(_ context.Context, _ variant.Variant) (config.AttestationCfg, *corev1.ConfigMap, error) {
	return u.currentConfig, &corev1.ConfigMap{}, nil
}

func (u stubUpgrader) PlanTerraformMigrations(context.Context, terraform.LogLevel, cloudprovider.Provider, terraform.Variables) (bool, error) {
	return false, u.planTerraformErr
}
