/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/helm"
	"github.com/edgelesssys/constellation/v2/internal/constellation/kubecmd"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
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
	fsWithStateFileAndTfState := func() file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		require.NoError(t, fh.MkdirAll(constants.TerraformWorkingDir))
		require.NoError(t, fh.WriteYAML(constants.StateFilename, defaultStateFile(cloudprovider.Azure)))
		return fh
	}

	testCases := map[string]struct {
		helmUpgrader      helmApplier
		kubeUpgrader      *stubKubernetesUpgrader
		fh                func() file.Handler
		fhAssertions      func(require *require.Assertions, assert *assert.Assertions, fh file.Handler)
		terraformUpgrader cloudApplier
		fetchImageErr     error
		wantErr           bool
		customK8sVersion  string
		flags             applyFlags
		stdin             string
	}{
		"success": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			helmUpgrader:      &stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
			fh:                fsWithStateFileAndTfState,
			fhAssertions: func(require *require.Assertions, assert *assert.Assertions, fh file.Handler) {
				gotState, err := state.ReadFromFile(fh, constants.StateFilename)
				require.NoError(err)
				assert.Equal("v1", gotState.Version)
				assert.Equal(defaultStateFile(cloudprovider.Azure), gotState)
			},
		},
		"id file and state file do not exist": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			helmUpgrader:      &stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
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
			helmUpgrader:      &stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			flags:             applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
			fh:                fsWithStateFileAndTfState,
		},
		"nodeVersion in progress error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig:  config.DefaultForAzureSEVSNP(),
				nodeVersionErr: kubecmd.ErrInProgress,
			},
			helmUpgrader:      &stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
			fh:                fsWithStateFileAndTfState,
		},
		"helm other error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubHelmApplier{err: assert.AnError},
			terraformUpgrader: &stubTerraformUpgrader{},
			wantErr:           true,
			flags:             applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
			fh:                fsWithStateFileAndTfState,
		},
		"abort": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{terraformDiff: true},
			wantErr:           true,
			stdin:             "no\n",
			fh:                fsWithStateFileAndTfState,
		},
		"abort, restore terraform err": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{terraformDiff: true, rollbackWorkspaceErr: assert.AnError},
			wantErr:           true,
			stdin:             "no\n",
			fh:                fsWithStateFileAndTfState,
		},
		"plan terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{planTerraformErr: assert.AnError},
			wantErr:           true,
			flags:             applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
			fh:                fsWithStateFileAndTfState,
		},
		"apply terraform error": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader: stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{
				applyTerraformErr: assert.AnError,
				terraformDiff:     true,
			},
			wantErr: true,
			flags:   applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
			fh:      fsWithStateFileAndTfState,
		},
		"outdated K8s patch version": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			customK8sVersion: func() string {
				v, err := semver.New(versions.SupportedK8sVersions()[0])
				require.NoError(t, err)
				return semver.NewFromInt(v.Major(), v.Minor(), v.Patch()-1, "").String()
			}(),
			flags: applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
			fh:    fsWithStateFileAndTfState,
		},
		"outdated K8s version": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			customK8sVersion:  "v1.20.0",
			flags:             applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
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
				skipPhases: newPhases(skipInfrastructurePhase, skipAttestationConfigPhase, skipCertSANsPhase, skipHelmPhase, skipK8sPhase, skipImagePhase),
				yes:        true,
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
				skipPhases: newPhases(skipInfrastructurePhase, skipAttestationConfigPhase, skipCertSANsPhase, skipHelmPhase, skipK8sPhase),
				yes:        true,
			},
			fh: fsWithStateFileAndTfState,
		},
		"no tf state, infra phase skipped": {
			kubeUpgrader: &stubKubernetesUpgrader{
				currentConfig: config.DefaultForAzureSEVSNP(),
			},
			helmUpgrader:      &stubHelmApplier{},
			terraformUpgrader: &mockTerraformUpgrader{},
			flags: applyFlags{
				yes:        true,
				skipPhases: newPhases(skipInfrastructurePhase),
			},
			fh: func() file.Handler {
				fh := file.NewHandler(afero.NewMemMapFs())
				require.NoError(t, fh.WriteYAML(constants.StateFilename, defaultStateFile(cloudprovider.Azure)))
				return fh
			},
		},
		"attempt to change attestation variant": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: &config.AzureTrustedLaunch{}},
			helmUpgrader:      &stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			flags:             applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
			fh:                fsWithStateFileAndTfState,
			wantErr:           true,
		},
		"image fetching fails": {
			kubeUpgrader:      &stubKubernetesUpgrader{currentConfig: config.DefaultForAzureSEVSNP()},
			helmUpgrader:      &stubHelmApplier{},
			terraformUpgrader: &stubTerraformUpgrader{},
			fetchImageErr:     assert.AnError,
			flags:             applyFlags{yes: true, skipPhases: skipPhases{skipInitPhase: struct{}{}}},
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
				fileHandler: fh,
				flags:       tc.flags,
				log:         logger.NewTest(t),
				spinner:     &nopSpinner{},
				merger:      &stubMerger{},
				newInfraApplier: func(_ context.Context) (cloudApplier, func(), error) {
					return tc.terraformUpgrader, func() {}, nil
				},
				applier: &stubConstellApplier{
					stubKubernetesUpgrader: tc.kubeUpgrader,
					helmApplier:            tc.helmUpgrader,
				},
				imageFetcher: &stubImageFetcher{fetchReferenceErr: tc.fetchImageErr},
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
	kubernetesVersionErr           error
	currentConfig                  config.AttestationCfg
	getClusterAttestationConfigErr error
	calledNodeUpgrade              bool
	calledKubernetesUpgrade        bool
	backupCRDsErr                  error
	backupCRDsCalled               bool
	backupCRsErr                   error
	backupCRsCalled                bool
}

func (u *stubKubernetesUpgrader) BackupCRDs(_ context.Context, _ file.Handler, _ string) ([]apiextensionsv1.CustomResourceDefinition, error) {
	u.backupCRDsCalled = true
	return []apiextensionsv1.CustomResourceDefinition{}, u.backupCRDsErr
}

func (u *stubKubernetesUpgrader) BackupCRs(_ context.Context, _ file.Handler, _ []apiextensionsv1.CustomResourceDefinition, _ string) error {
	u.backupCRsCalled = true
	return u.backupCRsErr
}

func (u *stubKubernetesUpgrader) UpgradeNodeImage(_ context.Context, _ semver.Semver, _ string, _ bool) error {
	u.calledNodeUpgrade = true
	return u.nodeVersionErr
}

func (u *stubKubernetesUpgrader) UpgradeKubernetesVersion(_ context.Context, _ versions.ValidK8sVersion, _ bool) error {
	u.calledKubernetesUpgrade = true
	return u.kubernetesVersionErr
}

func (u *stubKubernetesUpgrader) ApplyJoinConfig(_ context.Context, _ config.AttestationCfg, _ []byte) error {
	return nil
}

func (u *stubKubernetesUpgrader) GetClusterAttestationConfig(_ context.Context, _ variant.Variant) (config.AttestationCfg, error) {
	return u.currentConfig, u.getClusterAttestationConfigErr
}

func (u *stubKubernetesUpgrader) ExtendClusterConfigCertSANs(_ context.Context, _, _ string, _ []string) error {
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

func (u stubTerraformUpgrader) Apply(_ context.Context, _ cloudprovider.Provider, _ variant.Variant, _ cloudcmd.RollbackBehavior) (state.Infrastructure, error) {
	return state.Infrastructure{}, u.applyTerraformErr
}

func (u stubTerraformUpgrader) RestoreWorkspace() error {
	return u.rollbackWorkspaceErr
}

func (u stubTerraformUpgrader) WorkingDirIsEmpty() (bool, error) {
	return false, nil
}

type mockTerraformUpgrader struct {
	mock.Mock
}

func (m *mockTerraformUpgrader) Plan(ctx context.Context, conf *config.Config) (bool, error) {
	args := m.Called(ctx, conf)
	return args.Bool(0), args.Error(1)
}

func (m *mockTerraformUpgrader) Apply(ctx context.Context, provider cloudprovider.Provider, variant variant.Variant, rollback cloudcmd.RollbackBehavior) (state.Infrastructure, error) {
	args := m.Called(ctx, provider, variant, rollback)
	return args.Get(0).(state.Infrastructure), args.Error(1)
}

func (m *mockTerraformUpgrader) RestoreWorkspace() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockTerraformUpgrader) WorkingDirIsEmpty() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

type mockApplier struct {
	mock.Mock
}

func (m *mockApplier) AnnotateCoreDNSResources(_ context.Context) error {
	return nil
}

func (m *mockApplier) CleanupCoreDNSResources(_ context.Context) error {
	return nil
}

func (m *mockApplier) PrepareHelmCharts(
	helmOpts helm.Options, stateFile *state.State, str string, masterSecret uri.MasterSecret,
) (helm.Applier, bool, error) {
	args := m.Called(helmOpts, stateFile, helmOpts, str, masterSecret)
	return args.Get(0).(helm.Applier), args.Bool(1), args.Error(2)
}

type stubImageFetcher struct {
	reference         string
	fetchReferenceErr error
}

func (f *stubImageFetcher) FetchReference(_ context.Context,
	_ cloudprovider.Provider, _ variant.Variant,
	_, _ string, _ bool,
) (string, error) {
	return f.reference, f.fetchReferenceErr
}
