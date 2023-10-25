/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	consemver "github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	fsWithDefaultConfigAndState := func(require *require.Assertions, provider cloudprovider.Provider) afero.Fs {
		fs := afero.NewMemMapFs()
		file := file.NewHandler(fs)
		require.NoError(file.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), provider)))
		stateFile := state.New()
		switch provider {
		case cloudprovider.GCP:
			stateFile.SetInfrastructure(state.Infrastructure{GCP: &state.GCP{}})
		case cloudprovider.Azure:
			stateFile.SetInfrastructure(state.Infrastructure{Azure: &state.Azure{}})
		}
		require.NoError(stateFile.WriteToFile(file, constants.StateFilename))
		return fs
	}
	fsWithoutState := func(require *require.Assertions, provider cloudprovider.Provider) afero.Fs {
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
		wantErr             bool
		wantAbort           bool
	}{
		"create": {
			setupFs:  fsWithDefaultConfigAndState,
			creator:  &stubCloudCreator{state: infraState},
			provider: cloudprovider.GCP,
			yesFlag:  true,
		},
		"interactive": {
			setupFs:  fsWithDefaultConfigAndState,
			creator:  &stubCloudCreator{state: infraState},
			provider: cloudprovider.Azure,
			stdin:    "yes\n",
		},
		"interactive abort": {
			setupFs:   fsWithDefaultConfigAndState,
			creator:   &stubCloudCreator{state: infraState},
			provider:  cloudprovider.GCP,
			stdin:     "no\n",
			wantAbort: true,
		},
		"interactive error": {
			setupFs:  fsWithDefaultConfigAndState,
			creator:  &stubCloudCreator{state: infraState},
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
			creator:  &stubCloudCreator{state: infraState},
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
			creator:  &stubCloudCreator{state: infraState},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"config does not exist": {
			setupFs:  func(a *require.Assertions, p cloudprovider.Provider) afero.Fs { return afero.NewMemMapFs() },
			creator:  &stubCloudCreator{state: infraState},
			provider: cloudprovider.GCP,
			yesFlag:  true,
			wantErr:  true,
		},
		"state file does not exist": {
			setupFs:  fsWithoutState,
			creator:  &stubCloudCreator{state: infraState},
			provider: cloudprovider.GCP,
			yesFlag:  true,
		},
		"create error": {
			setupFs:  fsWithDefaultConfigAndState,
			creator:  &stubCloudCreator{applyErr: assert.AnError},
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
			creator:  &stubCloudCreator{state: infraState},
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
			c := &createCmd{log: logger.NewTest(t), flags: createFlags{yes: tc.yesFlag}}
			err := c.create(cmd, tc.creator, fileHandler, &nopSpinner{}, stubAttestationFetcher{})

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				if tc.wantAbort {
					assert.False(tc.creator.planCalled)
					assert.False(tc.creator.applyCalled)
				} else {
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
		"terraform dir exists": {
			existingFiles: []string{constants.TerraformWorkingDir},
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
			c := &createCmd{log: logger.NewTest(t)}
			err := c.checkDirClean(fh)

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
