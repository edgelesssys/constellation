/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
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
		helmUpgrader      *stubHelmUpgrader
		kubeUpgrader      *stubKubernetesUpgrader
		terraformUpgrader *stubTerraformUpgrader
		flags             upgradeApplyFlags
		wantErr           bool
		stdin             string
	}{
		"success": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             upgradeApplyFlags{yes: true},
		},
		"nodeVersion some error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: assert.AnError,
			},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			flags:             upgradeApplyFlags{yes: true},
		},
		"nodeVersion in progress error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: kubecmd.ErrInProgress,
			},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             upgradeApplyFlags{yes: true},
		},
		"helm other error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubHelmUpgrader{err: assert.AnError},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			flags:             upgradeApplyFlags{yes: true},
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
		"plan terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubHelmUpgrader{},
			terraformUpgrader: &stubTerraformUpgrader{planTerraformErr: assert.AnError},
			wantErr:           true,
			flags:             upgradeApplyFlags{yes: true},
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
			flags:   upgradeApplyFlags{yes: true},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			cmd := newUpgradeApplyCmd()
			cmd.SetIn(bytes.NewBufferString(tc.stdin))

			handler := file.NewHandler(afero.NewMemMapFs())

			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.Azure)

			require.NoError(handler.WriteYAML(constants.ConfigFilename, cfg))
			require.NoError(handler.WriteJSON(constants.ClusterIDsFilename, clusterid.File{}))
			require.NoError(handler.WriteJSON(constants.MasterSecretFilename, uri.MasterSecret{}))

			upgrader := upgradeApplyCmd{
				kubeUpgrader:    tc.kubeUpgrader,
				helmUpgrader:    tc.helmUpgrader,
				clusterUpgrader: tc.terraformUpgrader,
				log:             logger.NewTest(t),
				configFetcher:   stubAttestationFetcher{},
				clusterShower:   &stubShowCluster{},
				fileHandler:     handler,
			}

			err := upgrader.upgradeApply(cmd, "test", tc.flags)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
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
	nodeVersionErr error
	currentConfig  config.AttestationCfg
}

func (u stubKubernetesUpgrader) UpgradeNodeVersion(_ context.Context, _ *config.Config, _ bool) error {
	return u.nodeVersionErr
}

func (u stubKubernetesUpgrader) ApplyJoinConfig(_ context.Context, _ config.AttestationCfg, _ []byte) error {
	return nil
}

func (u stubKubernetesUpgrader) GetClusterAttestationConfig(_ context.Context, _ variant.Variant) (config.AttestationCfg, error) {
	return u.currentConfig, nil
}

func (u stubKubernetesUpgrader) ExtendClusterConfigCertSANs(_ context.Context, _ []string) error {
	return nil
}

// TODO(v2.11): Remove this function.
func (u stubKubernetesUpgrader) RemoveAttestationConfigHelmManagement(_ context.Context) error {
	return nil
}

// TODO(v2.12): Remove this function.
func (u stubKubernetesUpgrader) RemoveHelmKeepAnnotation(_ context.Context) error {
	return nil
}

type stubTerraformUpgrader struct {
	terraformDiff     bool
	planTerraformErr  error
	applyTerraformErr error
}

func (u stubTerraformUpgrader) PlanClusterUpgrade(_ context.Context, _ io.Writer, _ terraform.Variables, _ cloudprovider.Provider) (bool, error) {
	return u.terraformDiff, u.planTerraformErr
}

func (u stubTerraformUpgrader) ApplyClusterUpgrade(_ context.Context, _ cloudprovider.Provider) (terraform.ApplyOutput, error) {
	return terraform.ApplyOutput{}, u.applyTerraformErr
}
