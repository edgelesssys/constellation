/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgrade

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckTerraformMigrations(t *testing.T) {
	upgrader := func() *TerraformUpgrader {
		u, err := NewTerraformUpgrader(&stubTerraformClient{}, bytes.NewBuffer(nil))
		require.NoError(t, err)

		return u
	}

	workspace := func(existingFiles []string) file.Handler {
		fs := afero.NewMemMapFs()
		for _, f := range existingFiles {
			require.NoError(t, afero.WriteFile(fs, f, []byte{}, 0o644))
		}

		return file.NewHandler(fs)
	}

	testCases := map[string]struct {
		workspace file.Handler
		wantErr   bool
	}{
		"success": {
			workspace: workspace(nil),
		},
		"migration output file already exists": {
			workspace: workspace([]string{constants.TerraformMigrationOutputFile}),
			wantErr:   true,
		},
		"terraform backup dir already exists": {
			workspace: workspace([]string{filepath.Join(constants.UpgradeDir, constants.TerraformUpgradeBackupDir)}),
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			u := upgrader()
			err := u.CheckTerraformMigrations(tc.workspace)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestPlanTerraformMigrations(t *testing.T) {
	upgrader := func(tf tfClient) *TerraformUpgrader {
		u, err := NewTerraformUpgrader(tf, bytes.NewBuffer(nil))
		require.NoError(t, err)

		return u
	}

	testCases := map[string]struct {
		tf      tfClient
		want    bool
		wantErr bool
	}{
		"success no diff": {
			tf: &stubTerraformClient{},
		},
		"success diff": {
			tf: &stubTerraformClient{
				hasDiff: true,
			},
			want: true,
		},
		"prepare workspace error": {
			tf: &stubTerraformClient{
				prepareWorkspaceErr: assert.AnError,
			},
			wantErr: true,
		},
		"plan error": {
			tf: &stubTerraformClient{
				planErr: assert.AnError,
			},
			wantErr: true,
		},
		"show plan error no diff": {
			tf: &stubTerraformClient{
				showErr: assert.AnError,
			},
		},
		"show plan error diff": {
			tf: &stubTerraformClient{
				showErr: assert.AnError,
				hasDiff: true,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			u := upgrader(tc.tf)

			opts := TerraformUpgradeOptions{
				LogLevel: terraform.LogLevelDebug,
				CSP:      cloudprovider.Unknown,
				Vars:     &terraform.QEMUVariables{},
			}

			diff, err := u.PlanTerraformMigrations(context.Background(), opts)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Equal(tc.want, diff)
			}
		})
	}
}

func TestApplyTerraformMigrations(t *testing.T) {
	upgrader := func(tf tfClient) *TerraformUpgrader {
		u, err := NewTerraformUpgrader(tf, bytes.NewBuffer(nil))
		require.NoError(t, err)

		return u
	}

	fileHandler := func(existingFiles ...string) file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())
		require.NoError(t,
			fh.Write(
				filepath.Join(constants.UpgradeDir, constants.TerraformUpgradeWorkingDir, "someFile"),
				[]byte("some content"),
			))
		for _, f := range existingFiles {
			require.NoError(t, fh.Write(f, []byte("some content")))
		}
		return fh
	}

	testCases := map[string]struct {
		tf             tfClient
		fs             file.Handler
		outputFileName string
		wantErr        bool
	}{
		"success": {
			tf:             &stubTerraformClient{},
			fs:             fileHandler(),
			outputFileName: "test.json",
		},
		"create cluster error": {
			tf: &stubTerraformClient{
				CreateClusterErr: assert.AnError,
			},
			fs:             fileHandler(),
			outputFileName: "test.json",
			wantErr:        true,
		},
		"empty file name": {
			tf:             &stubTerraformClient{},
			fs:             fileHandler(),
			outputFileName: "",
			wantErr:        true,
		},
		"file already exists": {
			tf:             &stubTerraformClient{},
			fs:             fileHandler("test.json"),
			outputFileName: "test.json",
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			u := upgrader(tc.tf)

			opts := TerraformUpgradeOptions{
				LogLevel:   terraform.LogLevelDebug,
				CSP:        cloudprovider.Unknown,
				Vars:       &terraform.QEMUVariables{},
				OutputFile: tc.outputFileName,
			}

			err := u.ApplyTerraformMigrations(context.Background(), tc.fs, opts)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestCleanUpTerraformMigrations(t *testing.T) {
	upgrader := func() *TerraformUpgrader {
		u, err := NewTerraformUpgrader(&stubTerraformClient{}, bytes.NewBuffer(nil))
		require.NoError(t, err)

		return u
	}

	workspace := func(existingFiles []string) file.Handler {
		fs := afero.NewMemMapFs()
		for _, f := range existingFiles {
			require.NoError(t, afero.WriteFile(fs, f, []byte{}, 0o644))
		}

		return file.NewHandler(fs)
	}

	testCases := map[string]struct {
		workspace file.Handler
		wantFiles []string
		wantErr   bool
	}{
		"no files": {
			workspace: workspace(nil),
			wantFiles: []string{},
		},
		"clean backup dir": {
			workspace: workspace([]string{
				filepath.Join(constants.UpgradeDir, constants.TerraformUpgradeBackupDir),
			}),
			wantFiles: []string{},
		},
		"clean working dir": {
			workspace: workspace([]string{
				filepath.Join(constants.UpgradeDir, constants.TerraformUpgradeWorkingDir),
			}),
			wantFiles: []string{},
		},
		"clean backup dir leave other files": {
			workspace: workspace([]string{
				filepath.Join(constants.UpgradeDir, constants.TerraformUpgradeBackupDir),
				filepath.Join(constants.UpgradeDir, "someFile"),
			}),
			wantFiles: []string{
				filepath.Join(constants.UpgradeDir, "someFile"),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			u := upgrader()
			err := u.CleanUpTerraformMigrations(tc.workspace)
			if tc.wantErr {
				require.Error(err)
				return
			}

			require.NoError(err)

			for _, f := range tc.wantFiles {
				_, err := tc.workspace.Stat(f)
				require.NoError(err, "file %s should exist", f)
			}
		})
	}
}

type stubTerraformClient struct {
	hasDiff             bool
	prepareWorkspaceErr error
	showErr             error
	planErr             error
	CreateClusterErr    error
}

func (u *stubTerraformClient) PrepareUpgradeWorkspace(string, string, string, terraform.Variables) error {
	return u.prepareWorkspaceErr
}

func (u *stubTerraformClient) ShowPlan(context.Context, terraform.LogLevel, string, io.Writer) error {
	return u.showErr
}

func (u *stubTerraformClient) Plan(context.Context, terraform.LogLevel, string, ...string) (bool, error) {
	return u.hasDiff, u.planErr
}

func (u *stubTerraformClient) CreateCluster(context.Context, terraform.LogLevel, ...string) (terraform.CreateOutput, error) {
	return terraform.CreateOutput{}, u.CreateClusterErr
}
