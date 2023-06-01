/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"errors"
	"strconv"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
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
	idFile := clusterid.File{IP: "192.0.2.1"}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		setupFs             func(*require.Assertions, cloudprovider.Provider) afero.Fs
		creator             *stubCloudCreator
		provider            cloudprovider.Provider
		yesFlag             bool
		controllerCountFlag *int
		workerCountFlag     *int
		configFlag          string
		stdin               string
		wantErr             bool
		wantAbort           bool
	}{
		"create": {
			setupFs:             fsWithDefaultConfig,
			creator:             &stubCloudCreator{id: idFile},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(2),
			yesFlag:             true,
		},
		"interactive": {
			setupFs:             fsWithDefaultConfig,
			creator:             &stubCloudCreator{id: idFile},
			provider:            cloudprovider.Azure,
			controllerCountFlag: intPtr(2),
			workerCountFlag:     intPtr(1),
			stdin:               "yes\n",
		},
		"interactive abort": {
			setupFs:             fsWithDefaultConfig,
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			stdin:               "no\n",
			wantAbort:           true,
		},
		"interactive error": {
			setupFs:             fsWithDefaultConfig,
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			stdin:               "foo\nfoo\nfoo\n",
			wantErr:             true,
		},
		"flag control-plane-count invalid": {
			setupFs:             fsWithDefaultConfig,
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(0),
			workerCountFlag:     intPtr(3),
			wantErr:             true,
		},
		"flag worker-count invalid": {
			setupFs:             fsWithDefaultConfig,
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(3),
			workerCountFlag:     intPtr(0),
			wantErr:             true,
		},
		"flag control-plane-count missing": {
			setupFs:         fsWithDefaultConfig,
			creator:         &stubCloudCreator{},
			provider:        cloudprovider.GCP,
			workerCountFlag: intPtr(3),
			wantErr:         true,
		},
		"flag worker-count missing": {
			setupFs:             fsWithDefaultConfig,
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(3),
			wantErr:             true,
		},
		"old adminConf in directory": {
			setupFs: func(require *require.Assertions, csp cloudprovider.Provider) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1}, file.OptNone))
				require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), csp)))
				return fs
			},
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			wantErr:             true,
		},
		"old masterSecret in directory": {
			setupFs: func(require *require.Assertions, csp cloudprovider.Provider) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.MasterSecretFilename, []byte{1}, file.OptNone))
				require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), csp)))
				return fs
			},
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			wantErr:             true,
		},
		"config does not exist": {
			setupFs:             fsWithDefaultConfig,
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			configFlag:          "/does/not/exist",
			wantErr:             true,
		},
		"create error": {
			setupFs:             fsWithDefaultConfig,
			creator:             &stubCloudCreator{createErr: someErr},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			wantErr:             true,
		},
		"write id file error": {
			setupFs: func(require *require.Assertions, csp cloudprovider.Provider) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, defaultConfigWithExpectedMeasurements(t, config.Default(), csp)))
				return afero.NewReadOnlyFs(fs)
			},
			creator:             &stubCloudCreator{},
			provider:            cloudprovider.GCP,
			controllerCountFlag: intPtr(1),
			workerCountFlag:     intPtr(1),
			yesFlag:             true,
			wantErr:             true,
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
			cmd.Flags().String("config", constants.ConfigFilename, "") // register persistent flag manually
			cmd.Flags().Bool("force", true, "")                        // register persistent flag manually
			cmd.Flags().String("tf-log", "NONE", "")                   // register persistent flag manually

			if tc.yesFlag {
				require.NoError(cmd.Flags().Set("yes", "true"))
			}
			if tc.configFlag != "" {
				require.NoError(cmd.Flags().Set("config", tc.configFlag))
			}
			if tc.controllerCountFlag != nil {
				require.NoError(cmd.Flags().Set("control-plane-nodes", strconv.Itoa(*tc.controllerCountFlag)))
			}
			if tc.workerCountFlag != nil {
				require.NoError(cmd.Flags().Set("worker-nodes", strconv.Itoa(*tc.workerCountFlag)))
			}

			fileHandler := file.NewHandler(tc.setupFs(require, tc.provider))
			c := &createCmd{log: logger.NewTest(t)}
			err := c.create(cmd, tc.creator, fileHandler, &nopSpinner{}, fakeConfigFetcher{})

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				if tc.wantAbort {
					assert.False(tc.creator.createCalled)
				} else {
					assert.True(tc.creator.createCalled)
					var gotIDFile clusterid.File
					require.NoError(fileHandler.ReadJSON(constants.ClusterIDsFileName, &gotIDFile))
					assert.Equal(gotIDFile, clusterid.File{
						IP:            idFile.IP,
						CloudProvider: tc.provider,
					})
				}
			}
		})
	}
}

func TestCheckDirClean(t *testing.T) {
	testCases := map[string]struct {
		fileHandler   file.Handler
		existingFiles []string
		wantErr       bool
	}{
		"no file exists": {
			fileHandler: file.NewHandler(afero.NewMemMapFs()),
		},
		"adminconf exists": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.AdminConfFilename},
			wantErr:       true,
		},
		"master secret exists": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.MasterSecretFilename},
			wantErr:       true,
		},
		"multiple exist": {
			fileHandler:   file.NewHandler(afero.NewMemMapFs()),
			existingFiles: []string{constants.AdminConfFilename, constants.MasterSecretFilename},
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			for _, f := range tc.existingFiles {
				require.NoError(tc.fileHandler.Write(f, []byte{1, 2, 3}, file.OptNone))
			}
			c := &createCmd{log: logger.NewTest(t)}
			err := c.checkDirClean(tc.fileHandler)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
