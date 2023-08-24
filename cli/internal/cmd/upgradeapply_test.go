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

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestUpgradeApply(t *testing.T) {
	testCases := map[string]struct {
		helmUpgrader      stubApplier
		kubeUpgrader      *stubKubernetesUpgrader
		terraformUpgrader *stubTerraformUpgrader
		wantErr           bool
		flags             upgradeApplyFlags
		stdin             string
	}{
		"success": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             upgradeApplyFlags{yes: true},
		},
		"nodeVersion some error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: assert.AnError,
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			flags:             upgradeApplyFlags{yes: true},
		},
		"nodeVersion in progress error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: kubecmd.ErrInProgress,
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             upgradeApplyFlags{yes: true},
		},
		"helm other error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubApplier{err: assert.AnError},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			flags:             upgradeApplyFlags{yes: true},
		},
		"abort": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{terraformDiff: true},
			wantErr:           true,
			stdin:             "no\n",
		},
		"plan terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{planTerraformErr: assert.AnError},
			wantErr:           true,
			flags:             upgradeApplyFlags{yes: true},
		},
		"apply terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader: stubApplier{},
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
			require.NoError(handler.WriteJSON(constants.ClusterIDsFilename, clusterid.File{MeasurementSalt: []byte("measurementSalt")}))
			require.NoError(handler.WriteJSON(constants.MasterSecretFilename, uri.MasterSecret{}))

			upgrader := upgradeApplyCmd{
				kubeUpgrader:    tc.kubeUpgrader,
				helmApplier:     tc.helmUpgrader,
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

type stubKubernetesUpgrader struct {
	nodeVersionErr error
	currentConfig  config.AttestationCfg
}

func (u stubKubernetesUpgrader) BackupCRDs(_ context.Context, _ string) ([]apiextensionsv1.CustomResourceDefinition, error) {
	return []apiextensionsv1.CustomResourceDefinition{}, nil
}

func (u stubKubernetesUpgrader) BackupCRs(_ context.Context, _ []apiextensionsv1.CustomResourceDefinition, _ string) error {
	return nil
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
