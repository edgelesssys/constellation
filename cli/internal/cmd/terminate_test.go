/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
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
	setupFs := func(require *require.Assertions, stateFile *state.State) afero.Fs {
		fs := afero.NewMemMapFs()
		fileHandler := file.NewHandler(fs)
		require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1, 2}, file.OptNone))
		require.NoError(stateFile.WriteToFile(fileHandler, constants.StateFilename))
		return fs
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		stateFile  *state.State
		yesFlag    bool
		stdin      string
		setupFs    func(*require.Assertions, *state.State) afero.Fs
		terminator spyCloudTerminator
		wantErr    bool
		wantAbort  bool
	}{
		"success": {
			stateFile:  state.New(),
			setupFs:    setupFs,
			terminator: &stubCloudTerminator{},
			yesFlag:    true,
		},
		"interactive": {
			stateFile:  state.New(),
			setupFs:    setupFs,
			terminator: &stubCloudTerminator{},
			stdin:      "yes\n",
		},
		"interactive abort": {
			stateFile:  state.New(),
			setupFs:    setupFs,
			terminator: &stubCloudTerminator{},
			stdin:      "no\n",
			wantAbort:  true,
		},
		"files to remove do not exist": {
			stateFile: state.New(),
			setupFs: func(require *require.Assertions, stateFile *state.State) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(stateFile.WriteToFile(fileHandler, constants.StateFilename))
				return fs
			},
			terminator: &stubCloudTerminator{},
			yesFlag:    true,
		},
		"terminate error": {
			stateFile:  state.New(),
			setupFs:    setupFs,
			terminator: &stubCloudTerminator{terminateErr: someErr},
			yesFlag:    true,
			wantErr:    true,
		},
		"missing id file does not error": {
			stateFile: state.New(),
			setupFs: func(require *require.Assertions, stateFile *state.State) afero.Fs {
				fs := afero.NewMemMapFs()
				fileHandler := file.NewHandler(fs)
				require.NoError(fileHandler.Write(constants.AdminConfFilename, []byte{1, 2}, file.OptNone))
				return fs
			},
			terminator: &stubCloudTerminator{},
			yesFlag:    true,
		},
		"remove file fails": {
			stateFile: state.New(),
			setupFs: func(require *require.Assertions, stateFile *state.State) afero.Fs {
				fs := setupFs(require, stateFile)
				return afero.NewReadOnlyFs(fs)
			},
			terminator: &stubCloudTerminator{},
			yesFlag:    true,
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
			cmd.SetIn(bytes.NewBufferString(tc.stdin))

			require.NotNil(tc.setupFs)
			fileHandler := file.NewHandler(tc.setupFs(require, tc.stateFile))

			tCmd := &terminateCmd{
        log:         slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
				fileHandler: fileHandler,
				flags: terminateFlags{
					yes: tc.yesFlag,
				},
			}
			err := tCmd.terminate(cmd, tc.terminator, &nopSpinner{})

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				if tc.wantAbort {
					assert.False(tc.terminator.Called())
				} else {
					assert.True(tc.terminator.Called())
					_, err = fileHandler.Stat(constants.AdminConfFilename)
					assert.Error(err)
					_, err = fileHandler.Stat(constants.StateFilename)
					assert.Error(err)
				}
			}
		})
	}
}

type spyCloudTerminator interface {
	cloudTerminator
	Called() bool
}
