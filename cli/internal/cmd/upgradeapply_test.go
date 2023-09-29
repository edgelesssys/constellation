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
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestUpgradeApply(t *testing.T) {
	testCases := map[string]struct {
		helmUpgrader         helmApplier
		kubeUpgrader         *stubKubernetesUpgrader
		terraformUpgrader    clusterUpgrader
		infrastructureShower *stubShowInfrastructure
		wantErr              bool
		customK8sVersion     string
		flags                upgradeApplyFlags
		stdin                string
	}{
		"success": {
			kubeUpgrader:         &stubKubernetesUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			helmUpgrader:         stubApplier{},
			terraformUpgrader:    &stubTerraformUpgrader{},
			flags:                upgradeApplyFlags{yes: true},
			infrastructureShower: &stubShowInfrastructure{},
		},
		"nodeVersion some error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: assert.AnError,
			},
			helmUpgrader:         stubApplier{},
			terraformUpgrader:    &stubTerraformUpgrader{},
			wantErr:              true,
			flags:                upgradeApplyFlags{yes: true},
			infrastructureShower: &stubShowInfrastructure{},
		},
		"nodeVersion in progress error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: kubecmd.ErrInProgress,
			},
			helmUpgrader:         stubApplier{},
			terraformUpgrader:    &stubTerraformUpgrader{},
			flags:                upgradeApplyFlags{yes: true},
			infrastructureShower: &stubShowInfrastructure{},
		},
		"helm other error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:         stubApplier{err: assert.AnError},
			terraformUpgrader:    &stubTerraformUpgrader{},
			wantErr:              true,
			flags:                upgradeApplyFlags{yes: true},
			infrastructureShower: &stubShowInfrastructure{},
		},
		"abort": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:         stubApplier{},
			terraformUpgrader:    &stubTerraformUpgrader{terraformDiff: true},
			wantErr:              true,
			stdin:                "no\n",
			infrastructureShower: &stubShowInfrastructure{},
		},
		"abort, restore terraform err": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:         stubApplier{},
			terraformUpgrader:    &stubTerraformUpgrader{terraformDiff: true, rollbackWorkspaceErr: assert.AnError},
			wantErr:              true,
			stdin:                "no\n",
			infrastructureShower: &stubShowInfrastructure{},
		},
		"plan terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:         stubApplier{},
			terraformUpgrader:    &stubTerraformUpgrader{planTerraformErr: assert.AnError},
			wantErr:              true,
			flags:                upgradeApplyFlags{yes: true},
			infrastructureShower: &stubShowInfrastructure{},
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
			wantErr:              true,
			flags:                upgradeApplyFlags{yes: true},
			infrastructureShower: &stubShowInfrastructure{},
		},
		"outdated K8s patch version": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			customK8sVersion: func() string {
				v, err := semver.New(versions.SupportedK8sVersions()[0])
				require.NoError(t, err)
				return semver.NewFromInt(v.Major(), v.Minor(), v.Patch()-1, "").String()
			}(),
			flags:                upgradeApplyFlags{yes: true},
			infrastructureShower: &stubShowInfrastructure{},
		},
		"outdated K8s version": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:         stubApplier{},
			terraformUpgrader:    &stubTerraformUpgrader{},
			customK8sVersion:     "v1.20.0",
			flags:                upgradeApplyFlags{yes: true},
			wantErr:              true,
			infrastructureShower: &stubShowInfrastructure{},
		},
		"skip all upgrade phases": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &mockApplier{}, // mocks ensure that no methods are called
			terraformUpgrader: &mockTerraformUpgrader{},
			flags: upgradeApplyFlags{
				skipPhases: []skipPhase{skipInfrastructurePhase, skipHelmPhase, skipK8sPhase, skipImagePhase},
				yes:        true,
			},
			infrastructureShower: &stubShowInfrastructure{},
		},
		"show state err": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags: upgradeApplyFlags{
				skipPhases: []skipPhase{skipInfrastructurePhase},
				yes:        true,
			},
			infrastructureShower: &stubShowInfrastructure{
				showInfraErr: assert.AnError,
			},
			wantErr: true,
		},
		"skip all phases except node upgrade": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &mockApplier{}, // mocks ensure that no methods are called
			terraformUpgrader: &mockTerraformUpgrader{},
			flags: upgradeApplyFlags{
				skipPhases: []skipPhase{skipInfrastructurePhase, skipHelmPhase, skipK8sPhase},
				yes:        true,
			},
			infrastructureShower: &stubShowInfrastructure{},
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
			if tc.customK8sVersion != "" {
				cfg.KubernetesVersion = versions.ValidK8sVersion(tc.customK8sVersion)
			}
			require.NoError(handler.WriteYAML(constants.ConfigFilename, cfg))
			require.NoError(handler.WriteJSON(constants.ClusterIDsFilename, clusterid.File{
				MeasurementSalt: []byte{0x41},
				UID:             "uid",
			}))
			require.NoError(handler.WriteJSON(constants.MasterSecretFilename, uri.MasterSecret{}))

			upgrader := upgradeApplyCmd{
				kubeUpgrader:    tc.kubeUpgrader,
				helmApplier:     tc.helmUpgrader,
				clusterUpgrader: tc.terraformUpgrader,
				log:             logger.NewTest(t),
				configFetcher:   stubAttestationFetcher{},
				clusterShower:   tc.infrastructureShower,
				fileHandler:     handler,
			}

			err := upgrader.upgradeApply(cmd, "test", tc.flags)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(!tc.flags.skipPhases.contains(skipImagePhase), tc.kubeUpgrader.calledNodeUpgrade,
				"incorrect node upgrade skipping behavior")

			expectedState := state.New().
				SetInfrastructure(state.Infrastructure{
					APIServerCertSANs: []string{},
					UID:               "uid",
				}).
				SetClusterValues(state.ClusterValues{MeasurementSalt: []byte{0x41}})
			gotState, err := state.ReadFromFile(handler, constants.StateFilename)
			require.NoError(err)
			assert.Equal("v1", gotState.Version)
			assert.Equal(expectedState, gotState)
		})
	}
}

func TestUpgradeApplyFlagsForSkipPhases(t *testing.T) {
	cmd := newUpgradeApplyCmd()
	cmd.Flags().String("workspace", "", "")  // register persistent flag manually
	cmd.Flags().Bool("force", true, "")      // register persistent flag manually
	cmd.Flags().String("tf-log", "NONE", "") // register persistent flag manually
	require.NoError(t, cmd.Flags().Set("skip-phases", "infrastructure,helm,k8s,image"))
	result, err := parseUpgradeApplyFlags(cmd)
	if err != nil {
		t.Fatalf("Error while parsing flags: %v", err)
	}
	assert.ElementsMatch(t, []skipPhase{skipInfrastructurePhase, skipHelmPhase, skipK8sPhase, skipImagePhase}, result.skipPhases)
}

type stubKubernetesUpgrader struct {
	nodeVersionErr    error
	currentConfig     config.AttestationCfg
	calledNodeUpgrade bool
}

func (u *stubKubernetesUpgrader) BackupCRDs(_ context.Context, _ string) ([]apiextensionsv1.CustomResourceDefinition, error) {
	return []apiextensionsv1.CustomResourceDefinition{}, nil
}

func (u *stubKubernetesUpgrader) BackupCRs(_ context.Context, _ []apiextensionsv1.CustomResourceDefinition, _ string) error {
	return nil
}

func (u *stubKubernetesUpgrader) UpgradeNodeVersion(_ context.Context, _ *config.Config, _, _, _ bool) error {
	u.calledNodeUpgrade = true
	return u.nodeVersionErr
}

func (u *stubKubernetesUpgrader) ApplyJoinConfig(_ context.Context, _ config.AttestationCfg, _ []byte) error {
	return nil
}

func (u *stubKubernetesUpgrader) GetClusterAttestationConfig(_ context.Context, _ variant.Variant) (config.AttestationCfg, error) {
	return u.currentConfig, nil
}

func (u *stubKubernetesUpgrader) ExtendClusterConfigCertSANs(_ context.Context, _ []string) error {
	return nil
}

// TODO(v2.11): Remove this function after v2.11 is released.
func (u *stubKubernetesUpgrader) RemoveAttestationConfigHelmManagement(_ context.Context) error {
	return nil
}

// TODO(v2.12): Remove this function.
func (u *stubKubernetesUpgrader) RemoveHelmKeepAnnotation(_ context.Context) error {
	return nil
}

type stubTerraformUpgrader struct {
	terraformDiff        bool
	planTerraformErr     error
	applyTerraformErr    error
	rollbackWorkspaceErr error
}

func (u stubTerraformUpgrader) PlanClusterUpgrade(_ context.Context, _ io.Writer, _ terraform.Variables, _ cloudprovider.Provider) (bool, error) {
	return u.terraformDiff, u.planTerraformErr
}

func (u stubTerraformUpgrader) ApplyClusterUpgrade(_ context.Context, _ cloudprovider.Provider) (state.Infrastructure, error) {
	return state.Infrastructure{}, u.applyTerraformErr
}

func (u stubTerraformUpgrader) RestoreClusterWorkspace() error {
	return u.rollbackWorkspaceErr
}

type mockTerraformUpgrader struct {
	mock.Mock
}

func (m *mockTerraformUpgrader) PlanClusterUpgrade(ctx context.Context, w io.Writer, variables terraform.Variables, provider cloudprovider.Provider) (bool, error) {
	args := m.Called(ctx, w, variables, provider)
	return args.Bool(0), args.Error(1)
}

func (m *mockTerraformUpgrader) ApplyClusterUpgrade(ctx context.Context, provider cloudprovider.Provider) (state.Infrastructure, error) {
	args := m.Called(ctx, provider)
	return args.Get(0).(state.Infrastructure), args.Error(1)
}

func (m *mockTerraformUpgrader) RestoreClusterWorkspace() error {
	args := m.Called()
	return args.Error(0)
}

type mockApplier struct {
	mock.Mock
}

func (m *mockApplier) PrepareApply(cfg *config.Config, stateFile *state.State,
	helmOpts helm.Options, str string, masterSecret uri.MasterSecret,
) (helm.Applier, bool, error) {
	args := m.Called(cfg, stateFile, helmOpts, str, masterSecret)
	return args.Get(0).(helm.Applier), args.Bool(1), args.Error(2)
}
