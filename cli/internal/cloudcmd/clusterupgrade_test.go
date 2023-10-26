/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanClusterUpgrade(t *testing.T) {
	setUpFilesystem := func(existingFiles []string) file.Handler {
		fs := file.NewHandler(afero.NewMemMapFs())
		require.NoError(t, fs.MkdirAll("test"))
		for _, f := range existingFiles {
			require.NoError(t, fs.Write(f, []byte{}))
		}
		return fs
	}

	testCases := map[string]struct {
		upgradeID string
		tf        *tfClusterUpgradeStub
		fs        file.Handler
		want      bool
		wantErr   bool
	}{
		"success no diff": {
			upgradeID: "1234",
			tf:        &tfClusterUpgradeStub{},
			fs:        setUpFilesystem([]string{}),
		},
		"success diff": {
			upgradeID: "1234",
			tf: &tfClusterUpgradeStub{
				planDiff: true,
			},
			fs:   setUpFilesystem([]string{}),
			want: true,
		},
		"prepare workspace error": {
			upgradeID: "1234",
			tf: &tfClusterUpgradeStub{
				prepareWorkspaceErr: assert.AnError,
			},
			fs:      setUpFilesystem([]string{}),
			wantErr: true,
		},
		"plan error": {
			tf: &tfClusterUpgradeStub{
				planErr: assert.AnError,
			},
			fs:      setUpFilesystem([]string{}),
			wantErr: true,
		},
		"show plan error no diff": {
			upgradeID: "1234",
			tf: &tfClusterUpgradeStub{
				showErr: assert.AnError,
			},
			fs: setUpFilesystem([]string{}),
		},
		"show plan error diff": {
			upgradeID: "1234",
			tf: &tfClusterUpgradeStub{
				showErr:  assert.AnError,
				planDiff: true,
			},
			fs:      setUpFilesystem([]string{}),
			wantErr: true,
		},
		"workspace not clean": {
			upgradeID: "1234",
			tf:        &tfClusterUpgradeStub{},
			fs:        setUpFilesystem([]string{filepath.Join(constants.UpgradeDir, "1234", constants.TerraformUpgradeBackupDir)}),
			wantErr:   true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			u := &ClusterUpgrader{
				tf:                tc.tf,
				policyPatcher:     stubPolicyPatcher{},
				fileHandler:       tc.fs,
				upgradeWorkspace:  filepath.Join(constants.UpgradeDir, tc.upgradeID),
				existingWorkspace: "test",
				logLevel:          terraform.LogLevelDebug,
			}

			diff, err := u.PlanClusterUpgrade(context.Background(), io.Discard, &terraform.QEMUVariables{}, cloudprovider.Unknown)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				require.Equal(tc.want, diff)
			}
		})
	}
}

func TestApplyClusterUpgrade(t *testing.T) {
	setUpFilesystem := func(upgradeID string, existingFiles ...string) file.Handler {
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
		upgradeID     string
		tf            *tfClusterUpgradeStub
		policyPatcher stubPolicyPatcher
		fs            file.Handler
		wantErr       bool
	}{
		"success": {
			upgradeID:     "1234",
			tf:            &tfClusterUpgradeStub{},
			fs:            setUpFilesystem("1234"),
			policyPatcher: stubPolicyPatcher{},
		},
		"apply error": {
			upgradeID: "1234",
			tf: &tfClusterUpgradeStub{
				applyErr: assert.AnError,
			},
			fs:            setUpFilesystem("1234"),
			policyPatcher: stubPolicyPatcher{},
			wantErr:       true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := require.New(t)

			tc.tf.file = tc.fs
			u := &ClusterUpgrader{
				tf:                tc.tf,
				policyPatcher:     stubPolicyPatcher{},
				fileHandler:       tc.fs,
				upgradeWorkspace:  filepath.Join(constants.UpgradeDir, tc.upgradeID),
				existingWorkspace: "test",
				logLevel:          terraform.LogLevelDebug,
			}

			_, err := u.ApplyClusterUpgrade(context.Background(), cloudprovider.Unknown)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

type tfClusterUpgradeStub struct {
	file                file.Handler
	applyErr            error
	planErr             error
	planDiff            bool
	showErr             error
	prepareWorkspaceErr error
}

func (t *tfClusterUpgradeStub) Plan(_ context.Context, _ terraform.LogLevel) (bool, error) {
	return t.planDiff, t.planErr
}

func (t *tfClusterUpgradeStub) ShowPlan(_ context.Context, _ terraform.LogLevel, _ io.Writer) error {
	return t.showErr
}

func (t *tfClusterUpgradeStub) ApplyCluster(_ context.Context, _ cloudprovider.Provider, _ terraform.LogLevel) (state.Infrastructure, error) {
	return state.Infrastructure{}, t.applyErr
}

func (t *tfClusterUpgradeStub) PrepareWorkspace(_ string, _ terraform.Variables) error {
	return t.prepareWorkspaceErr
}
