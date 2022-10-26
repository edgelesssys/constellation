/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminateCmdArgumentValidation(t *testing.T) {
	testCases := map[string]struct {
		args    []string
		wantErr bool
	}{
		"no args":         {[]string{}, false},
		"some args":       {[]string{"hello", "test"}, true},
		"some other args": {[]string{"12", "2"}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cmd := NewTerminateCmd()
			err := cmd.ValidateArgs(tc.args)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestTerminate(t *testing.T) {
	setupFs := func(require *require.Assertions, idFile clusterid.File) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1, 2}, file.OptNone))
		require.NoError(fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptNone))
		return fs
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		idFile     clusterid.File
		setupFs    func(*require.Assertions, clusterid.File) afero.Fs
		terminator spyCloudTerminator
		wantErr    bool
	}{
		"success": {
			idFile:     clusterid.File{CloudProvider: cloudprovider.GCP},
			setupFs:    setupFs,
			terminator: &stubCloudTerminator{},
		},
		"files to remove do not exist": {
			idFile: clusterid.File{CloudProvider: cloudprovider.GCP},
			setupFs: func(require *require.Assertions, idFile clusterid.File) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile, file.OptNone))
				return fs
			},
			terminator: &stubCloudTerminator{},
		},
		"terminate error": {
			idFile:     clusterid.File{CloudProvider: cloudprovider.GCP},
			setupFs:    setupFs,
			terminator: &stubCloudTerminator{terminateErr: someErr},
			wantErr:    true,
		},
		"missing id file does not error": {
			idFile: clusterid.File{CloudProvider: cloudprovider.GCP},
			setupFs: func(require *require.Assertions, idFile clusterid.File) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1, 2}, file.OptNone))
				return fs
			},
			terminator: &stubCloudTerminator{},
		},
		"remove file fails": {
			idFile: clusterid.File{CloudProvider: cloudprovider.GCP},
			setupFs: func(require *require.Assertions, idFile clusterid.File) afero.Fs {
				fs := setupFs(require, idFile)
				return afero.NewReadOnlyFs(fs)
			},
			terminator: &stubCloudTerminator{},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cmd := NewTerminateCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			require.NotNil(tc.setupFs)
			fileHandler := file.NewHandler(tc.setupFs(require, tc.idFile))

			err := terminate(cmd, tc.terminator, fileHandler, nopSpinner{})

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.True(tc.terminator.Called())
				_, err = fileHandler.Stat(constants.AdminConfFilename)
				assert.Error(err)
				_, err = fileHandler.Stat(constants.ClusterIDsFileName)
				assert.Error(err)
			}
		})
	}
}

type spyCloudTerminator interface {
	cloudTerminator
	Called() bool
}
