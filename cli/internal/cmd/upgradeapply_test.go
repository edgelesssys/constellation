/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/cli/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestUpgradeApply(t *testing.T) {
	someErr := errors.New("some error")
	testCases := map[string]struct {
		upgrader stubUpgrader
		fetcher  stubImageFetcher
		wantErr  bool
		yesFlag  bool
		stdin    string
	}{
		"success": {
			upgrader: stubUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			yesFlag:  true,
		},
		"nodeVersion some error": {
			upgrader: stubUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: someErr,
			},
			wantErr: true,
			yesFlag: true,
		},
		"nodeVersion in progress error": {
			upgrader: stubUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: kubernetes.ErrInProgress,
			},
			yesFlag: true,
		},
		"helm other error": {
			upgrader: stubUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
				helmErr:       someErr,
			},
			wantErr: true,
			fetcher: stubImageFetcher{},
			yesFlag: true,
		},
		"check terraform error": {
			upgrader: stubUpgrader{
				currentConfig:     config.DefaultForAzureSEVSNP(),
				checkTerraformErr: someErr,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			yesFlag: true,
		},
		"abort": {
			upgrader: stubUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
				terraformDiff: true,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			stdin:   "no\n",
		},
		"clean terraform error": {
			upgrader: stubUpgrader{
				currentConfig:     config.DefaultForAzureSEVSNP(),
				cleanTerraformErr: someErr,
				terraformDiff:     true,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			stdin:   "no\n",
		},
		"plan terraform error": {
			upgrader: stubUpgrader{
				currentConfig:    config.DefaultForAzureSEVSNP(),
				planTerraformErr: someErr,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			yesFlag: true,
		},
		"apply terraform error": {
			upgrader: stubUpgrader{
				currentConfig:     config.DefaultForAzureSEVSNP(),
				applyTerraformErr: someErr,
				terraformDiff:     true,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			yesFlag: true,
		},
		"fetch reference error": {
			upgrader: stubUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			fetcher: stubImageFetcher{fetchReferenceErr: someErr},
			wantErr: true,
			yesFlag: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			cmd := newUpgradeApplyCmd()
			cmd.SetIn(bytes.NewBufferString(tc.stdin))
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")                        // register persistent flag manually
			cmd.Flags().String("tf-log", "DEBUG", "")                  // register persistent flag manually

			if tc.yesFlag {
				err := cmd.Flags().Set("yes", "true")
				require.NoError(err)
			}

			handler := file.NewHandler(afero.NewMemMapFs())
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.Azure)
			require.NoError(handler.WriteYAML(constants.ConfigFilename, cfg))
			require.NoError(handler.WriteJSON(constants.ClusterIDsFileName, clusterid.File{}))

			upgrader := upgradeApplyCmd{upgrader: tc.upgrader, log: logger.NewTest(t), imageFetcher: tc.fetcher, configFetcher: stubAttestationFetcher{}, migrationExecutor: &migrationExecutorPlaceholder{}}

			err := upgrader.upgradeApply(cmd, handler)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type migrationExecutorPlaceholder struct{}

func (d *migrationExecutorPlaceholder) applyMigration(_ *cobra.Command, _ file.Handler, _ upgrade.TfMigrationCmd, _ upgradeApplyFlags) error {
	return nil
}

type stubUpgrader struct {
	currentConfig     config.AttestationCfg
	nodeVersionErr    error
	helmErr           error
	terraformDiff     bool
	planTerraformErr  error
	checkTerraformErr error
	applyTerraformErr error
	cleanTerraformErr error
}

func (u stubUpgrader) GetUpgradeID() string {
	return "test-upgrade"
}

func (u stubUpgrader) UpgradeNodeVersion(_ context.Context, _ *config.Config, _ bool) error {
	return u.nodeVersionErr
}

func (u stubUpgrader) UpgradeHelmServices(_ context.Context, _ *config.Config, _ clusterid.File, _ time.Duration, _, _ bool) error {
	return u.helmErr
}

func (u stubUpgrader) UpdateAttestationConfig(_ context.Context, _ config.AttestationCfg) error {
	return nil
}

func (u stubUpgrader) GetClusterAttestationConfig(_ context.Context, _ variant.Variant) (config.AttestationCfg, *corev1.ConfigMap, error) {
	return u.currentConfig, &corev1.ConfigMap{}, nil
}

func (u stubUpgrader) CheckTerraformMigrations() error {
	return u.checkTerraformErr
}

func (u stubUpgrader) CleanUpTerraformMigrations() error {
	return u.cleanTerraformErr
}

func (u stubUpgrader) PlanTerraformMigrations(context.Context, upgrade.TerraformUpgradeOptions) (bool, error) {
	return u.terraformDiff, u.planTerraformErr
}

func (u stubUpgrader) ApplyTerraformMigrations(context.Context, upgrade.TerraformUpgradeOptions) error {
	return u.applyTerraformErr
}

func (u stubUpgrader) ExtendClusterConfigCertSANs(_ context.Context, _ []string) error {
	return nil
}

// AddManualStateMigration is not used in this test.
// TODO(AB#3248): remove this method together with the definition in the interfaces.
func (u stubUpgrader) AddManualStateMigration(_ terraform.StateMigration) {
	panic("unused")
}

type stubImageFetcher struct {
	fetchReferenceErr error
}

func (f stubImageFetcher) FetchReference(_ context.Context,
	_ cloudprovider.Provider, _ variant.Variant,
	_, _ string,
) (string, error) {
	return "", f.fetchReferenceErr
}
