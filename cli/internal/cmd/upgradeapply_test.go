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

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/helm"
	"github.com/edgelesssys/constellation/v2/cli/internal/kubecmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
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
	defaultState := state.New().
		SetInfrastructure(state.Infrastructure{
			APIServerCertSANs: []string{},
			UID:               "uid",
			Name:              "kubernetes-uid", // default test cfg uses "kubernetes" prefix
			InitSecret:        []byte{0x42},
		}).
		SetClusterValues(state.ClusterValues{MeasurementSalt: []byte{0x41}})
	fsWithStateFileAndTfState := func() file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		require.NoError(t, fh.MkdirAll(constants.TerraformWorkingDir))
		require.NoError(t, fh.WriteYAML(constants.StateFilename, defaultState))
		return fh
	}

	testCases := map[string]struct {
		helmUpgrader      helmApplier
		kubeUpgrader      *stubKubernetesUpgrader
		fh                func() file.Handler
		fhAssertions      func(require *require.Assertions, assert *assert.Assertions, fh file.Handler)
		terraformUpgrader cloudApplier
		wantErr           bool
		customK8sVersion  string
		flags             applyFlags
		stdin             string
	}{
		"success": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             applyFlags{yes: true},
			fh:                fsWithStateFileAndTfState,
			fhAssertions: func(require *require.Assertions, assert *assert.Assertions, fh file.Handler) {
				gotState, err := state.ReadFromFile(fh, constants.StateFilename)
				require.NoError(err)
				assert.Equal("v1", gotState.Version)
				assert.Equal(defaultState, gotState)
			},
		},
		"id file and state file do not exist": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             applyFlags{yes: true},
			fh: func() file.Handler {
				return file.NewHandler(afero.NewMemMapFs())
			},
			wantErr: true,
		},
		"nodeVersion some error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: assert.AnError,
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			flags:             applyFlags{yes: true},
			fh:                fsWithStateFileAndTfState,
		},
		"nodeVersion in progress error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: kubecmd.ErrInProgress,
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             applyFlags{yes: true},
			fh:                fsWithStateFileAndTfState,
		},
		"helm other error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubApplier{err: assert.AnError},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			flags:             applyFlags{yes: true},
			fh:                fsWithStateFileAndTfState,
		},
		"abort": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{terraformDiff: true},
			wantErr:           true,
			stdin:             "no\n",
			fh:                fsWithStateFileAndTfState,
		},
		"abort, restore terraform err": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{terraformDiff: true, rollbackWorkspaceErr: assert.AnError},
			wantErr:           true,
			stdin:             "no\n",
			fh:                fsWithStateFileAndTfState,
		},
		"plan terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{planTerraformErr: assert.AnError},
			wantErr:           true,
			flags:             applyFlags{yes: true},
			fh:                fsWithStateFileAndTfState,
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
			flags:   applyFlags{yes: true},
			fh:      fsWithStateFileAndTfState,
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
			flags: applyFlags{yes: true},
			fh:    fsWithStateFileAndTfState,
		},
		"outdated K8s version": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			customK8sVersion:  "v1.20.0",
			flags:             applyFlags{yes: true},
			wantErr:           true,
			fh:                fsWithStateFileAndTfState,
		},
		"skip all upgrade phases": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &mockApplier{}, // mocks ensure that no methods are called
			terraformUpgrader: &mockTerraformUpgrader{},
			flags: applyFlags{
				skipPhases: skipPhases{
					skipInfrastructurePhase: struct{}{}, skipHelmPhase: struct{}{},
					skipK8sPhase: struct{}{}, skipImagePhase: struct{}{},
				},
				yes: true,
			},
			fh: fsWithStateFileAndTfState,
		},
		"skip all phases except node upgrade": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &mockApplier{}, // mocks ensure that no methods are called
			terraformUpgrader: &mockTerraformUpgrader{},
			flags: applyFlags{
				skipPhases: skipPhases{
					skipInfrastructurePhase: struct{}{}, skipHelmPhase: struct{}{},
					skipK8sPhase: struct{}{},
				},
				yes: true,
			},
			fh: fsWithStateFileAndTfState,
		},
		"no tf state, skip infrastructure upgrade": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubApplier{},
			terraformUpgrader: &mockTerraformUpgrader{},
			flags: applyFlags{
				yes: true,
			},
			fh: func() file.Handler {
				fh := file.NewHandler(afero.NewMemMapFs())
				require.NoError(t, fh.WriteYAML(constants.StateFilename, defaultState))
				return fh
			},
		},
		"attempt to change attestation variant": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: &config.AzureTrustedLaunch{}},
			helmUpgrader:      stubApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             applyFlags{yes: true},
			fh:                fsWithStateFileAndTfState,
			wantErr:           true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			cmd := newUpgradeApplyCmd()
			cmd.SetIn(bytes.NewBufferString(tc.stdin))

			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), cloudprovider.Azure)
			if tc.customK8sVersion != "" {
				cfg.KubernetesVersion = versions.ValidK8sVersion(tc.customK8sVersion)
			}
			fh := tc.fh()
			require.NoError(fh.Write(constants.AdminConfFilename, []byte{}))
			require.NoError(fh.WriteYAML(constants.ConfigFilename, cfg))
			require.NoError(fh.WriteJSON(constants.MasterSecretFilename, uri.MasterSecret{}))

			upgrader := &applyCmd{
				fileHandler:  fh,
				flags:        tc.flags,
				log:          logger.NewTest(t),
				spinner:      &nopSpinner{},
				merger:       &stubMerger{},
				quotaChecker: &stubLicenseClient{},
				newHelmClient: func(string, debugLog) (helmApplier, error) {
					return tc.helmUpgrader, nil
				},
				newKubeUpgrader: func(_ io.Writer, _ string, _ debugLog) (kubernetesUpgrader, error) {
					return tc.kubeUpgrader, nil
				},
				newInfraApplier: func(ctx context.Context) (cloudApplier, func(), error) {
					return tc.terraformUpgrader, func() {}, nil
				},
			}
			err := upgrader.apply(cmd, stubAttestationFetcher{}, "test")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(!tc.flags.skipPhases.contains(skipImagePhase), tc.kubeUpgrader.calledNodeUpgrade,
				"incorrect node upgrade skipping behavior")

			if tc.fhAssertions != nil {
				tc.fhAssertions(require, assert, fh)
			}
		})
	}
}

type stubKubernetesUpgrader struct {
	nodeVersionErr                 error
	currentConfig                  config.AttestationCfg
	getClusterAttestationConfigErr error
	calledNodeUpgrade              bool
	backupCRDsErr                  error
	backupCRDsCalled               bool
	backupCRsErr                   error
	backupCRsCalled                bool
}

func (u *stubKubernetesUpgrader) BackupCRDs(_ context.Context, _ string) ([]apiextensionsv1.CustomResourceDefinition, error) {
	u.backupCRDsCalled = true
	return []apiextensionsv1.CustomResourceDefinition{}, u.backupCRDsErr
}

func (u *stubKubernetesUpgrader) BackupCRs(_ context.Context, _ []apiextensionsv1.CustomResourceDefinition, _ string) error {
	u.backupCRsCalled = true
	return u.backupCRsErr
}

func (u *stubKubernetesUpgrader) UpgradeNodeVersion(_ context.Context, _ *config.Config, _, _, _ bool) error {
	u.calledNodeUpgrade = true
	return u.nodeVersionErr
}

func (u *stubKubernetesUpgrader) ApplyJoinConfig(_ context.Context, _ config.AttestationCfg, _ []byte) error {
	return nil
}

func (u *stubKubernetesUpgrader) GetClusterAttestationConfig(_ context.Context, _ variant.Variant) (config.AttestationCfg, error) {
	return u.currentConfig, u.getClusterAttestationConfigErr
}

func (u *stubKubernetesUpgrader) ExtendClusterConfigCertSANs(_ context.Context, _ []string) error {
	return nil
}

type stubTerraformUpgrader struct {
	terraformDiff        bool
	planTerraformErr     error
	applyTerraformErr    error
	rollbackWorkspaceErr error
}

func (u stubTerraformUpgrader) Plan(_ context.Context, _ *config.Config) (bool, error) {
	return u.terraformDiff, u.planTerraformErr
}

func (u stubTerraformUpgrader) Apply(_ context.Context, _ cloudprovider.Provider, _ cloudcmd.RollbackBehavior) (state.Infrastructure, error) {
	return state.Infrastructure{}, u.applyTerraformErr
}

func (u stubTerraformUpgrader) RestoreWorkspace() error {
	return u.rollbackWorkspaceErr
}

type mockTerraformUpgrader struct {
	mock.Mock
}

func (m *mockTerraformUpgrader) Plan(ctx context.Context, conf *config.Config) (bool, error) {
	args := m.Called(ctx, conf)
	return args.Bool(0), args.Error(1)
}

func (m *mockTerraformUpgrader) Apply(ctx context.Context, provider cloudprovider.Provider, rollback cloudcmd.RollbackBehavior) (state.Infrastructure, error) {
	args := m.Called(ctx, provider, rollback)
	return args.Get(0).(state.Infrastructure), args.Error(1)
}

func (m *mockTerraformUpgrader) RestoreWorkspace() error {
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
