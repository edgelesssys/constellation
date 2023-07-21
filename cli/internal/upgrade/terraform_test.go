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
	upgrader := func(fileHandler file.Handler) *TerraformUpgrader {
		u, err := NewTerraformUpgrader(&stubTerraformClient{}, bytes.NewBuffer(nil), fileHandler)
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
		upgradeID string
		workspace file.Handler
		wantErr   bool
	}{
		"success": {
			upgradeID: "1234",
			workspace: workspace(nil),
		},
		"terraform backup dir already exists": {
			upgradeID: "1234",
			workspace: workspace([]string{filepath.Join(constants.UpgradeDir, "1234", constants.TerraformUpgradeBackupDir)}),
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			u := upgrader(tc.workspace)
			err := u.CheckTerraformMigrations(tc.upgradeID)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestPlanTerraformMigrations(t *testing.T) {
	upgrader := func(tf tfClient, fileHandler file.Handler) *TerraformUpgrader {
		u, err := NewTerraformUpgrader(tf, bytes.NewBuffer(nil), fileHandler)
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
		upgradeID string
		tf        tfClient
		workspace file.Handler
		want      bool
		wantErr   bool
	}{
		"success no diff": {
			upgradeID: "1234",
			tf:        &stubTerraformClient{},
			workspace: workspace([]string{constants.ClusterIDsFileName}),
		},
		"success diff": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				hasDiff: true,
			},
			workspace: workspace([]string{constants.ClusterIDsFileName}),
			want:      true,
		},
		"prepare workspace error": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				prepareWorkspaceErr: assert.AnError,
			},
			workspace: workspace([]string{constants.ClusterIDsFileName}),
			wantErr:   true,
		},
		"constellation-id.json does not exist": {
			upgradeID: "1234",
			tf:        &stubTerraformClient{},
			workspace: workspace(nil),
			wantErr:   true,
		},
		"constellation-id backup already exists": {
			upgradeID: "1234",
			tf:        &stubTerraformClient{},
			workspace: workspace([]string{filepath.Join(constants.UpgradeDir, "1234", constants.ClusterIDsFileName+".old")}),
			wantErr:   true,
		},
		"plan error": {
			tf: &stubTerraformClient{
				planErr: assert.AnError,
			},
			workspace: workspace([]string{constants.ClusterIDsFileName}),
			wantErr:   true,
		},
		"show plan error no diff": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				showErr: assert.AnError,
			},
			workspace: workspace([]string{constants.ClusterIDsFileName}),
		},
		"show plan error diff": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				showErr: assert.AnError,
				hasDiff: true,
			},
			workspace: workspace([]string{constants.ClusterIDsFileName}),
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			u := upgrader(tc.tf, tc.workspace)

			opts := TerraformUpgradeOptions{
				LogLevel: terraform.LogLevelDebug,
				CSP:      cloudprovider.Unknown,
				Vars:     &terraform.QEMUVariables{},
			}

			diff, err := u.PlanTerraformMigrations(context.Background(), opts, tc.upgradeID)
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
	upgrader := func(tf tfClient, fileHandler file.Handler) *TerraformUpgrader {
		u, err := NewTerraformUpgrader(tf, bytes.NewBuffer(nil), fileHandler)
		require.NoError(t, err)

		return u
	}

	fileHandler := func(upgradeID string, existingFiles ...string) file.Handler {
		fh := file.NewHandler(afero.NewMemMapFs())

		require.NoError(t,
			fh.Write(
				filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformUpgradeWorkingDir, "someFile"),
				[]byte("some content"),
			))
		for _, f := range existingFiles {
			require.NoError(t, fh.Write(f, []byte("some content")))
		}
		return fh
	}

	testCases := map[string]struct {
		upgradeID          string
		tf                 tfClient
		policyPatcher      stubPolicyPatcher
		fs                 file.Handler
		skipIDFileCreation bool // if true, do not create the constellation-id.json file
		wantErr            bool
	}{
		"success": {
			upgradeID:     "1234",
			tf:            &stubTerraformClient{},
			fs:            fileHandler("1234"),
			policyPatcher: stubPolicyPatcher{},
		},
		"create cluster error": {
			upgradeID: "1234",
			tf: &stubTerraformClient{
				CreateClusterErr: assert.AnError,
			},
			fs:            fileHandler("1234"),
			policyPatcher: stubPolicyPatcher{},
			wantErr:       true,
		},
		"constellation-id.json does not exist": {
			upgradeID:          "1234",
			tf:                 &stubTerraformClient{},
			fs:                 fileHandler("1234"),
			policyPatcher:      stubPolicyPatcher{},
			skipIDFileCreation: true,
			wantErr:            true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			u := upgrader(tc.tf, tc.fs)

			if !tc.skipIDFileCreation {
				require.NoError(
					tc.fs.Write(
						filepath.Join(constants.ClusterIDsFileName),
						[]byte("{}"),
					))
			}

			opts := TerraformUpgradeOptions{
				LogLevel: terraform.LogLevelDebug,
				CSP:      cloudprovider.Unknown,
				Vars:     &terraform.QEMUVariables{},
			}

			err := u.ApplyTerraformMigrations(context.Background(), opts, tc.upgradeID)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestCleanUpTerraformMigrations(t *testing.T) {
	upgrader := func(fileHandler file.Handler) *TerraformUpgrader {
		u, err := NewTerraformUpgrader(&stubTerraformClient{}, bytes.NewBuffer(nil), fileHandler)
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
		upgradeID      string
		workspaceFiles []string
		wantFiles      []string
		wantErr        bool
	}{
		"no files": {
			upgradeID:      "1234",
			workspaceFiles: nil,
			wantFiles:      []string{},
		},
		"clean backup dir": {
			upgradeID: "1234",
			workspaceFiles: []string{
				filepath.Join(constants.UpgradeDir, "1234", constants.TerraformUpgradeBackupDir),
			},
			wantFiles: []string{},
		},
		"clean working dir": {
			upgradeID: "1234",
			workspaceFiles: []string{
				filepath.Join(constants.UpgradeDir, "1234", constants.TerraformUpgradeWorkingDir),
			},
			wantFiles: []string{},
		},
		"clean all": {
			upgradeID: "1234",
			workspaceFiles: []string{
				filepath.Join(constants.UpgradeDir, "1234", constants.TerraformUpgradeBackupDir),
				filepath.Join(constants.UpgradeDir, "1234", constants.TerraformUpgradeWorkingDir),
				filepath.Join(constants.UpgradeDir, "1234", "abc"),
			},
			wantFiles: []string{},
		},
		"leave other files": {
			upgradeID: "1234",
			workspaceFiles: []string{
				filepath.Join(constants.UpgradeDir, "1234", constants.TerraformUpgradeBackupDir),
				filepath.Join(constants.UpgradeDir, "other"),
			},
			wantFiles: []string{
				filepath.Join(constants.UpgradeDir, "other"),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			workspace := workspace(tc.workspaceFiles)
			u := upgrader(workspace)

			err := u.CleanUpTerraformMigrations(tc.upgradeID)
			if tc.wantErr {
				require.Error(err)
				return
			}

			require.NoError(err)

			for _, haveFile := range tc.workspaceFiles {
				for _, wantFile := range tc.wantFiles {
					if haveFile == wantFile {
						_, err := workspace.Stat(wantFile)
						require.NoError(err, "file %s should exist", wantFile)
					} else {
						_, err := workspace.Stat(haveFile)
						require.Error(err, "file %s should not exist", haveFile)
					}
				}
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

func (u *stubTerraformClient) PrepareUpgradeWorkspace(string, string, string, string, terraform.Variables) error {
	return u.prepareWorkspaceErr
}

func (u *stubTerraformClient) ShowPlan(context.Context, terraform.LogLevel, string, io.Writer) error {
	return u.showErr
}

func (u *stubTerraformClient) Plan(context.Context, terraform.LogLevel, string) (bool, error) {
	return u.hasDiff, u.planErr
}

func (u *stubTerraformClient) CreateCluster(context.Context, terraform.LogLevel) (terraform.ApplyOutput, error) {
	return terraform.ApplyOutput{}, u.CreateClusterErr
}

type stubPolicyPatcher struct {
	patchErr error
}

func (p *stubPolicyPatcher) PatchPolicy(context.Context, string) error {
	return p.patchErr
}
