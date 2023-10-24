/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"io"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanUpgrade(t *testing.T) {
	const (
		templateDir       = "templateDir"
		existingWorkspace = "existing"
		backupDir         = "backup"
	)
	fsWithWorkspace := func(require *require.Assertions) file.Handler {
		fs := file.NewHandler(afero.NewMemMapFs())
		require.NoError(fs.MkdirAll(existingWorkspace))
		return fs
	}

	testCases := map[string]struct {
		prepareFs func(require *require.Assertions) file.Handler
		tf        *stubUpgradePlanner
		wantDiff  bool
		wantErr   bool
	}{
		"success no diff": {
			prepareFs: fsWithWorkspace,
			tf:        &stubUpgradePlanner{},
		},
		"success diff": {
			prepareFs: fsWithWorkspace,
			tf: &stubUpgradePlanner{
				planDiff: true,
			},
			wantDiff: true,
		},
		"workspace does not exist": {
			prepareFs: func(require *require.Assertions) file.Handler {
				return file.NewHandler(afero.NewMemMapFs())
			},
			tf:      &stubUpgradePlanner{},
			wantErr: true,
		},
		"workspace not clean": {
			prepareFs: func(require *require.Assertions) file.Handler {
				fs := fsWithWorkspace(require)
				require.NoError(fs.MkdirAll(backupDir))
				return fs
			},
			tf:      &stubUpgradePlanner{},
			wantErr: true,
		},
		"prepare workspace error": {
			prepareFs: fsWithWorkspace,
			tf: &stubUpgradePlanner{
				prepareWorkspaceErr: assert.AnError,
			},
			wantErr: true,
		},
		"plan error": {
			prepareFs: fsWithWorkspace,
			tf: &stubUpgradePlanner{
				planErr: assert.AnError,
			},
			wantErr: true,
		},
		"show plan error": {
			prepareFs: fsWithWorkspace,
			tf: &stubUpgradePlanner{
				planDiff:    true,
				showPlanErr: assert.AnError,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			fs := tc.prepareFs(require.New(t))

			hasDiff, err := planUpgrade(
				context.Background(), tc.tf, fs, io.Discard, terraform.LogLevelDebug,
				&terraform.QEMUVariables{},
				templateDir, existingWorkspace, backupDir,
			)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantDiff, hasDiff)
		})
	}
}

func TestRestoreBackup(t *testing.T) {
	existingWorkspace := "foo"
	backupDir := "bar"

	testCases := map[string]struct {
		prepareFs func(require *require.Assertions) file.Handler
		wantErr   bool
	}{
		"success": {
			prepareFs: func(require *require.Assertions) file.Handler {
				fs := file.NewHandler(afero.NewMemMapFs())
				require.NoError(fs.MkdirAll(existingWorkspace))
				require.NoError(fs.MkdirAll(backupDir))
				return fs
			},
		},
		"existing workspace does not exist": {
			prepareFs: func(require *require.Assertions) file.Handler {
				fs := file.NewHandler(afero.NewMemMapFs())
				require.NoError(fs.MkdirAll(backupDir))
				return fs
			},
		},
		"backup dir does not exist": {
			prepareFs: func(require *require.Assertions) file.Handler {
				fs := file.NewHandler(afero.NewMemMapFs())
				require.NoError(fs.MkdirAll(existingWorkspace))
				return fs
			},
			wantErr: true,
		},
		"read only file system": {
			prepareFs: func(require *require.Assertions) file.Handler {
				memFS := afero.NewMemMapFs()
				fs := file.NewHandler(memFS)
				require.NoError(fs.MkdirAll(existingWorkspace))
				require.NoError(fs.MkdirAll(backupDir))
				return file.NewHandler(afero.NewReadOnlyFs(memFS))
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			fs := tc.prepareFs(require.New(t))

			err := restoreBackup(fs, existingWorkspace, backupDir)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestEnsureFileNotExist(t *testing.T) {
	testCases := map[string]struct {
		fs       file.Handler
		fileName string
		wantErr  bool
	}{
		"file does not exist": {
			fs:       file.NewHandler(afero.NewMemMapFs()),
			fileName: "foo",
		},
		"file exists": {
			fs: func() file.Handler {
				fs := file.NewHandler(afero.NewMemMapFs())
				err := fs.Write("foo", []byte{})
				require.NoError(t, err)
				return fs
			}(),
			fileName: "foo",
			wantErr:  true,
		},
		"directory exists": {
			fs: func() file.Handler {
				fs := file.NewHandler(afero.NewMemMapFs())
				err := fs.MkdirAll("foo/bar")
				require.NoError(t, err)
				return fs
			}(),
			fileName: "foo/bar",
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := ensureFileNotExist(tc.fs, tc.fileName)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type stubUpgradePlanner struct {
	prepareWorkspaceErr error
	planDiff            bool
	planErr             error
	showPlanErr         error
}

func (s *stubUpgradePlanner) PrepareWorkspace(_ string, _ terraform.Variables) error {
	return s.prepareWorkspaceErr
}

func (s *stubUpgradePlanner) Plan(_ context.Context, _ terraform.LogLevel) (bool, error) {
	return s.planDiff, s.planErr
}

func (s *stubUpgradePlanner) ShowPlan(_ context.Context, _ terraform.LogLevel, _ io.Writer) error {
	return s.showPlanErr
}
