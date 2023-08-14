/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
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
	testCases := map[string]struct {
		helmUpgrader             *stubHelmUpgrader
		kubeUpgrader             *stubKubernetesUpgrader
		terraformUpgrader        *stubTerraformUpgrader
		wantErr                  bool
		yesFlag                  bool
		dontWantJoinConfigBackup bool
		stdin                    string
	}{
		"success": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{},
			yesFlag:           true,
		},
		"nodeVersion some error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: assert.AnError,
			},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			yesFlag:           true,
		},
		"nodeVersion in progress error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: kubecmd.ErrInProgress,
			},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{},
			yesFlag:           true,
		},
		"helm other error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubHelmUpgrader{err: assert.AnError},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			yesFlag:           true,
		},
		"check terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{checkTerraformErr: assert.AnError},
			wantErr:           true,
			yesFlag:           true,
		},
		"abort": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{terraformDiff: true},
			wantErr:           true,
			stdin:             "no\n",
		},
		"clean terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader: &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{
				cleanTerraformErr: assert.AnError,
				terraformDiff:     true,
			},
			wantErr: true,
			stdin:   "no\n",
		},
		"plan terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{planTerraformErr: assert.AnError},
			wantErr:           true,
			yesFlag:           true,
		},
		"apply terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader: &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{
				applyTerraformErr: assert.AnError,
				terraformDiff:     true,
			},
			wantErr: true,
			yesFlag: true,
		},
		"do no backup join-config when remote attestation config is the same": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: fakeAzureAttestationConfigFromCluster(context.Background(), t, cloudprovider.Azure),
			},
			helmUpgrader:             &stubHelmUpgrader{},
			terraformUpgrader:        &stubTerraformUpgrader{},
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

			upgrader := upgradeApplyCmd{
				kubeUpgrader:      tc.kubeUpgrader,
				helmUpgrader:      tc.helmUpgrader,
				terraformUpgrader: tc.terraformUpgrader,
				log:               logger.NewTest(t),
				configFetcher:     stubAttestationFetcher{},
				clusterShower:     &stubShowCluster{},
				fileHandler:       handler,
			}

			err := upgrader.upgradeApply(cmd)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(!tc.dontWantJoinConfigBackup, tc.kubeUpgrader.backupWasCalled)
		})
	}
}

type stubHelmUpgrader struct {
	err error
}

func (u stubHelmUpgrader) Upgrade(
	_ context.Context, _ *config.Config, _ clusterid.File, _ time.Duration, _, _ bool, _ string, _ bool,
	_ helm.WaitMode, _ uri.MasterSecret, _ string, _ versions.ValidK8sVersion, _ terraform.ApplyOutput,
) error {
	return u.err
}

type stubKubernetesUpgrader struct {
	backupWasCalled bool
	nodeVersionErr  error
	currentConfig   config.AttestationCfg
}

func (u stubKubernetesUpgrader) GetMeasurementSalt(_ context.Context) ([]byte, error) {
	return []byte{}, nil
}

func (u *stubKubernetesUpgrader) BackupConfigMap(_ context.Context, _ string) error {
	u.backupWasCalled = true
	return nil
}

func (u stubKubernetesUpgrader) UpgradeNodeVersion(_ context.Context, _ *config.Config, _ bool) error {
	return u.nodeVersionErr
}

func (u stubKubernetesUpgrader) UpdateAttestationConfig(_ context.Context, _ config.AttestationCfg) error {
	return nil
}

func (u stubKubernetesUpgrader) GetClusterAttestationConfig(_ context.Context, _ variant.Variant) (config.AttestationCfg, error) {
	return u.currentConfig, nil
}

func (u stubKubernetesUpgrader) ExtendClusterConfigCertSANs(_ context.Context, _ []string) error {
	return nil
}

type stubTerraformUpgrader struct {
	terraformDiff     bool
	planTerraformErr  error
	checkTerraformErr error
	applyTerraformErr error
	cleanTerraformErr error
}

func (u stubTerraformUpgrader) CheckTerraformMigrations(_ string) error {
	return u.checkTerraformErr
}

func (u stubTerraformUpgrader) CleanUpTerraformMigrations(_ string) error {
	return u.cleanTerraformErr
}

func (u stubTerraformUpgrader) PlanTerraformMigrations(context.Context, upgrade.TerraformUpgradeOptions) (bool, error) {
	return u.terraformDiff, u.planTerraformErr
}

func (u stubTerraformUpgrader) ApplyTerraformMigrations(context.Context, upgrade.TerraformUpgradeOptions) (terraform.ApplyOutput, error) {
	return terraform.ApplyOutput{}, u.applyTerraformErr
}

func (u stubTerraformUpgrader) UpgradeID() string {
	return "test-upgrade"
}

func fakeAzureAttestationConfigFromCluster(ctx context.Context, t *testing.T, provider cloudprovider.Provider) config.AttestationCfg {
	cpCfg := defaultConfigWithExpectedMeasurements(t, config.Default(), provider)
	// the cluster attestation config needs to have real version numbers that are translated from "latest" as defined in config.Default()
	err := cpCfg.Attestation.AzureSEVSNP.FetchAndSetLatestVersionNumbers(ctx, stubAttestationFetcher{}, time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	return cpCfg.GetAttestationConfig()
}
