/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/state"
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
	setupFs := func(require *require.Assertions, state state.ConstellationState) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1, 2}, file.OptNone))
		require.NoError(fileHandler.Write(constants.WGQuickConfigFilename, []byte{1, 2}, file.OptNone))
		require.NoError(fileHandler.Write(constants.ClusterIDsFileName, []byte{1, 2}, file.OptNone))
		require.NoError(fileHandler.WriteJSON(constants.StateFilename, state, file.OptNone))
		return fs
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		state      state.ConstellationState
		setupFs    func(*require.Assertions, state.ConstellationState) afero.Fs
		terminator spyCloudTerminator
		wantErr    bool
	}{
		"success": {
			state:      state.ConstellationState{CloudProvider: "gcp"},
			setupFs:    setupFs,
			terminator: &stubCloudTerminator{},
		},
		"files to remove do not exist": {
			state: state.ConstellationState{CloudProvider: "gcp"},
			setupFs: func(require *require.Assertions, state state.ConstellationState) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.WriteJSON(constants.StateFilename, state, file.OptNone))
				return fs
			},
			terminator: &stubCloudTerminator{},
		},
		"terminate error": {
			state:      state.ConstellationState{CloudProvider: "gcp"},
			setupFs:    setupFs,
			terminator: &stubCloudTerminator{terminateErr: someErr},
			wantErr:    true,
		},
		"missing state file": {
			state: state.ConstellationState{CloudProvider: "gcp"},
			setupFs: func(require *require.Assertions, state state.ConstellationState) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1, 2}, file.OptNone))
				require.NoError(fileHandler.Write(constants.WGQuickConfigFilename, []byte{1, 2}, file.OptNone))
				return fs
			},
			terminator: &stubCloudTerminator{},
			wantErr:    true,
		},
		"remove file fails": {
			state: state.ConstellationState{CloudProvider: "gcp"},
			setupFs: func(require *require.Assertions, state state.ConstellationState) afero.Fs {
				fs := setupFs(require, state)
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
			fileHandler := file.NewHandler(tc.setupFs(require, tc.state))

			err := terminate(cmd, tc.terminator, fileHandler)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.True(tc.terminator.Called())
				_, err := fileHandler.Stat(constants.StateFilename)
				assert.Error(err)
				_, err = fileHandler.Stat(constants.AdminConfFilename)
				assert.Error(err)
				_, err = fileHandler.Stat(constants.WGQuickConfigFilename)
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
