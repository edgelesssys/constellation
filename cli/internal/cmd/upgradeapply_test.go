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
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/cli/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeApply(t *testing.T) {
	someErr := errors.New("some error")
	testCases := map[string]struct {
		upgrader                 *stubUpgrader
		fetcher                  stubImageFetcher
		wantErr                  bool
		yesFlag                  bool
		dontWantJoinConfigBackup bool
		stdin                    string
	}{
		"success": {
			upgrader: &stubUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			yesFlag:  true,
		},
		"nodeVersion some error": {
			upgrader: &stubUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: someErr,
			},
			wantErr: true,
			yesFlag: true,
		},
		"nodeVersion in progress error": {
			upgrader: &stubUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: kubernetes.ErrInProgress,
			},
			yesFlag: true,
		},
		"helm other error": {
			upgrader: &stubUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
				helmErr:       someErr,
			},
			wantErr: true,
			fetcher: stubImageFetcher{},
			yesFlag: true,
		},
		"check terraform error": {
			upgrader: &stubUpgrader{
				currentConfig:     config.DefaultForAzureSEVSNP(),
				checkTerraformErr: someErr,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			yesFlag: true,
		},
		"abort": {
			upgrader: &stubUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
				terraformDiff: true,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			stdin:   "no\n",
		},
		"clean terraform error": {
			upgrader: &stubUpgrader{
				currentConfig:     config.DefaultForAzureSEVSNP(),
				cleanTerraformErr: someErr,
				terraformDiff:     true,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			stdin:   "no\n",
		},
		"plan terraform error": {
			upgrader: &stubUpgrader{
				currentConfig:    config.DefaultForAzureSEVSNP(),
				planTerraformErr: someErr,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			yesFlag: true,
		},
		"apply terraform error": {
			upgrader: &stubUpgrader{
				currentConfig:     config.DefaultForAzureSEVSNP(),
				applyTerraformErr: someErr,
				terraformDiff:     true,
			},
			fetcher: stubImageFetcher{},
			wantErr: true,
			yesFlag: true,
		},
		"fetch reference error": {
			upgrader: &stubUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			fetcher: stubImageFetcher{fetchReferenceErr: someErr},
			wantErr: true,
			yesFlag: true,
		},
		"do no backup join-config when remote attestation config is the same": {
			upgrader: &stubUpgrader{
				currentConfig: fakeAzureAttestationConfigFromCluster(context.Background(), t, cloudprovider.Azure),
			},
			fetcher:                  stubImageFetcher{},
			yesFlag:                  true,
			dontWantJoinConfigBackup: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			cmd := newUpgradeApplyCmd()
			cmd.SetIn(bytes.NewBufferString(tc.stdin))
			cmd.Flags().String("workspace", "", "")   // register persistent flag manually
			cmd.Flags().Bool("force", true, "")       // register persistent flag manually
			cmd.Flags().String("tf-log", "DEBUG", "") // register persistent flag manually

			if tc.yesFlag {
				err := cmd.Flags().Set("yes", "true")
				require.NoError(err)
			}

			handler := file.NewHandler(afero.NewMemMapFs())

			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.Azure)

			require.NoError(handler.WriteYAML(constants.ConfigFilename, cfg))
			require.NoError(handler.WriteJSON(constants.ClusterIDsFilename, clusterid.File{}))
			require.NoError(handler.WriteJSON(constants.MasterSecretFilename, uri.MasterSecret{}))

			upgrader := upgradeApplyCmd{upgrader: tc.upgrader, log: logger.NewTest(t), imageFetcher: tc.fetcher, configFetcher: stubAttestationFetcher{}, clusterShower: &stubShowCluster{}, fileHandler: handler}

			err := upgrader.upgradeApply(cmd)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(!tc.dontWantJoinConfigBackup, tc.upgrader.backupWasCalled)
		})
	}
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
	backupWasCalled   bool
}

func (u stubUpgrader) GetUpgradeID() string {
	return "test-upgrade"
}

func (u *stubUpgrader) BackupConfigMap(_ context.Context, _ string) error {
	u.backupWasCalled = true
	return nil
}

func (u stubUpgrader) UpgradeNodeVersion(_ context.Context, _ *config.Config, _ bool) error {
	return u.nodeVersionErr
}

func (u stubUpgrader) UpgradeHelmServices(_ context.Context, _ *config.Config, _ clusterid.File, _ time.Duration, _, _, _ bool, _ helm.WaitMode, _ uri.MasterSecret, _ string, _ versions.ValidK8sVersion, _ terraform.ApplyOutput) error {
	return u.helmErr
}

func (u stubUpgrader) UpdateAttestationConfig(_ context.Context, _ config.AttestationCfg) error {
	return nil
}

func (u stubUpgrader) GetClusterAttestationConfig(_ context.Context, _ variant.Variant) (config.AttestationCfg, error) {
	return u.currentConfig, nil
}

func (u stubUpgrader) CheckTerraformMigrations(_ string) error {
	return u.checkTerraformErr
}

func (u stubUpgrader) CleanUpTerraformMigrations(_ string) error {
	return u.cleanTerraformErr
}

func (u stubUpgrader) PlanTerraformMigrations(context.Context, upgrade.TerraformUpgradeOptions) (bool, error) {
	return u.terraformDiff, u.planTerraformErr
}

func (u stubUpgrader) ApplyTerraformMigrations(context.Context, upgrade.TerraformUpgradeOptions) (terraform.ApplyOutput, error) {
	return terraform.ApplyOutput{}, u.applyTerraformErr
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

func fakeAzureAttestationConfigFromCluster(ctx context.Context, t *testing.T, provider cloudprovider.Provider) config.AttestationCfg {
	cpCfg := defaultConfigWithExpectedMeasurements(t, config.Default(), provider)
	// the cluster attestation config needs to have real version numbers that are translated from "latest" as defined in config.Default()
	err := cpCfg.Attestation.AzureSEVSNP.FetchAndSetLatestVersionNumbers(ctx, stubAttestationFetcher{}, time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	return cpCfg.GetAttestationConfig()
}
