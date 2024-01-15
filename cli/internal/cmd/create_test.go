/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	consemver "github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	fsWithDefaultConfig := func(require *require.Assertions, provider cloudprovider.Provider) afero.Fs {
		fs := afero.NewMemMapFs()
		file := file.NewHandler(fs)
		require.NoError(file.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), provider)))
		return fs
	}
	infraState := state.Infrastructure{ClusterEndpoint: "192.0.2.1"}

	testCases := map[string]struct {
		setupFs             func(*require.Assertions, cloudprovider.Provider) afero.Fs
		creator             *stubCloudCreator
		provider            cloudprovider.Provider
		yesFlag             bool
		controllerCountFlag *int
		workerCountFlag     *int
		stdin               string
		getCreatorErr       error
		wantErr             bool
		wantAbort           bool
	}{
		"create": {
			setupFs: fsWithDefaultConfig,
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				workspaceIsEmpty: true,
			},
			provider: cloudprovider.GCP,
			yesFlag:  true,
		},
		"interactive": {
			setupFs: fsWithDefaultConfig,
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				workspaceIsEmpty: true,
			},
			provider: cloudprovider.Azure,
			stdin:    "yes\n",
		},
		"interactive abort": {
			setupFs: fsWithDefaultConfig,
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				workspaceIsEmpty: true,
			},
			provider:  cloudprovider.GCP,
			stdin:     "no\n",
			wantAbort: true,
			wantErr:   true,
		},
		"interactive error": {
			setupFs: fsWithDefaultConfig,
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				workspaceIsEmpty: true,
			},
			provider: cloudprovider.GCP,
			stdin:    "foo\nfoo\nfoo\n",
			wantErr:  true,
		},
		"old adminConf in directory": {
			setupFs: func(require *require.Assertions, csp cloudprovider.Provider) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1}, file.OptNone))
				require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), csp)))
				return fs
			},
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				workspaceIsEmpty: true,
			},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"old masterSecret in directory": {
			setupFs: func(require *require.Assertions, csp cloudprovider.Provider) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.MasterSecretFilename, []byte{1}, file.OptNone))
				require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), csp)))
				return fs
			},
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				workspaceIsEmpty: true,
			},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"config does not exist": {
			setupFs: func(a *require.Assertions, p cloudprovider.Provider) afero.Fs { return afero.NewMemMapFs() },
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				workspaceIsEmpty: true,
			},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"state file exist (but is empty)": {
			setupFs: func(r *require.Assertions, csp cloudprovider.Provider) afero.Fs {
				fs := afero.NewMemMapFs()
				file := file.NewHandler(fs)
				r.NoError(file.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), csp)))
				r.NoError(file.WriteYAML(constants.StateFilename, state.New()))
				return fs
			},
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				workspaceIsEmpty: true,
			},
			provider: cloudprovider.GCP,
			yesFlag:  true,
		},
		"create error": {
			setupFs:  fsWithDefaultConfig,
			creator:  &stubCloudCreator{applyErr: assert.AnError, planDiff: true, workspaceIsEmpty: true},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"write state file error": {
			setupFs: func(require *require.Assertions, csp cloudprovider.Provider) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), csp)))
				return afero.NewReadOnlyFs(fs)
			},
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				workspaceIsEmpty: true,
			},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"check dir clean error": {
			setupFs: fsWithDefaultConfig,
			creator: &stubCloudCreator{
				state:               infraState,
				planDiff:            true,
				workspaceIsEmptyErr: assert.AnError,
			},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"get creator error": {
			setupFs: fsWithDefaultConfig,
			creator: &stubCloudCreator{
				state:               infraState,
				planDiff:            true,
				workspaceIsEmptyErr: assert.AnError,
			},
			provider:      cloudprovider.GCP,
			yesFlag:       true,
			getCreatorErr: assert.AnError,
			wantErr:       true,
		},
		"plan error": {
			setupFs: fsWithDefaultConfig,
			creator: &stubCloudCreator{
				state:            infraState,
				planDiff:         true,
				planErr:          assert.AnError,
				workspaceIsEmpty: true,
			},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewCreateCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetIn(bytes.NewBufferString(tc.stdin))

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider))

			a := &applyCmd{
				fileHandler: fileHandler,
				flags: applyFlags{
					yes:        tc.yesFlag,
					skipPhases: newPhases(skipInitPhase, skipAttestationConfigPhase, skipCertSANsPhase, skipHelmPhase, skipImagePhase, skipK8sPhase),
				},

				log:     slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				spinner: &nopSpinner{},

				newInfraApplier: func(_ context.Context) (cloudApplier, func(), error) {
					return tc.creator, func() {}, tc.getCreatorErr
				},

				applier: &stubConstellApplier{},
			}

			err := a.apply(cmd, stubAttestationFetcher{}, "create")

			if tc.wantErr {
				assert.Error(err)
				if tc.wantAbort {
					assert.True(tc.creator.planCalled)
					assert.False(tc.creator.applyCalled)
				}
			} else {
				assert.NoError(err)

				assert.True(tc.creator.planCalled)
				assert.True(tc.creator.applyCalled)

				var gotState state.State
				expectedState := state.Infrastructure{
					ClusterEndpoint:   "192.0.2.1",
					APIServerCertSANs: []string{},
					InitSecret:        []byte{},
				}
				require.NoError(fileHandler.ReadYAML(constants.StateFilename, &gotState))
				assert.Equal("v1", gotState.Version)
				assert.Equal(expectedState, gotState.Infrastructure)

			}
		})
	}
}

func TestCheckDirClean(t *testing.T) {
	testCases := map[string]struct {
		existingFiles []string
		wantErr       bool
	}{
		"no file exists": {},
		"adminconf exists": {
			existingFiles: []string{constants.AdminConfFilename},
			wantErr:       true,
		},
		"master secret exists": {
			existingFiles: []string{constants.MasterSecretFilename},
			wantErr:       true,
		},
		"multiple exist": {
			existingFiles: []string{constants.AdminConfFilename, constants.MasterSecretFilename},
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fh := file.NewHandler(afero.NewMemMapFs())
			for _, f := range tc.existingFiles {
				require.NoError(fh.Write(f, []byte{1, 2, 3}, file.OptNone))
			}
			a := &applyCmd{log: slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)), fileHandler: fh}
			err := a.checkInitFilesClean()

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestValidateCLIandConstellationVersionCompatibility(t *testing.T) {
	testCases := map[string]struct {
		imageVersion        string
		microServiceVersion consemver.Semver
		cliVersion          consemver.Semver
		wantErr             bool
	}{
		"empty": {
			imageVersion:        "",
			microServiceVersion: consemver.Semver{},
			cliVersion:          consemver.Semver{},
			wantErr:             true,
		},
		"invalid when image < CLI": {
			imageVersion:        "v2.7.1",
			microServiceVersion: consemver.NewFromInt(2, 8, 0, ""),
			cliVersion:          consemver.NewFromInt(2, 8, 0, ""),
			wantErr:             true,
		},
		"invalid when microservice < CLI": {
			imageVersion:        "v2.8.0",
			microServiceVersion: consemver.NewFromInt(2, 7, 1, ""),
			cliVersion:          consemver.NewFromInt(2, 8, 0, ""),
			wantErr:             true,
		},
		"valid release version": {
			imageVersion:        "v2.9.0",
			microServiceVersion: consemver.NewFromInt(2, 9, 0, ""),
			cliVersion:          consemver.NewFromInt(2, 9, 0, ""),
		},
		"valid pre-version": {
			imageVersion:        "ref/main/stream/nightly/v2.9.0-pre.0.20230626150512-0a36ce61719f",
			microServiceVersion: consemver.NewFromInt(2, 9, 0, "pre.0.20230626150512-0a36ce61719f"),
			cliVersion:          consemver.NewFromInt(2, 9, 0, "pre.0.20230626150512-0a36ce61719f"),
		},
		"image version suffix need not be equal to CLI version": {
			imageVersion:        "ref/main/stream/nightly/v2.9.0-pre.0.19990626150512-9z36ce61799z",
			microServiceVersion: consemver.NewFromInt(2, 9, 0, "pre.0.20230626150512-0a36ce61719f"),
			cliVersion:          consemver.NewFromInt(2, 9, 0, "pre.0.20230626150512-0a36ce61719f"),
		},
		"image version can have different patch version": {
			imageVersion:        "ref/main/stream/nightly/v2.9.1-pre.0.19990626150512-9z36ce61799z",
			microServiceVersion: consemver.NewFromInt(2, 9, 0, "pre.0.20230626150512-0a36ce61719f"),
			cliVersion:          consemver.NewFromInt(2, 9, 0, "pre.0.20230626150512-0a36ce61719f"),
		},
		"microService version suffix must be equal to CLI version": {
			imageVersion:        "ref/main/stream/nightly/v2.9.0-pre.0.20230626150512-0a36ce61719f",
			microServiceVersion: consemver.NewFromInt(2, 9, 0, "pre.0.19990626150512-9z36ce61799z"),
			cliVersion:          consemver.NewFromInt(2, 9, 0, "pre.0.20230626150512-0a36ce61719f"),
			wantErr:             true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			err := validateCLIandConstellationVersionAreEqual(tc.cliVersion, tc.imageVersion, tc.microServiceVersion)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}
