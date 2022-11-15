/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestPrepareCluster(t *testing.T) {
	qemuVars := &QEMUVariables{
		CommonVariables: CommonVariables{
			Name:               "name",
			CountControlPlanes: 1,
			CountWorkers:       2,
			StateDiskSizeGB:    11,
		},
		CPUCount:         1,
		MemorySizeMiB:    1024,
		ImagePath:        "path",
		ImageFormat:      "format",
		MetadataAPIImage: "api",
	}

	testCases := map[string]struct {
		provider           cloudprovider.Provider
		vars               Variables
		fs                 afero.Fs
		partiallyExtracted bool
		wantErr            bool
	}{
		"qemu": {
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			fs:       afero.NewMemMapFs(),
			wantErr:  false,
		},
		"no vars": {
			provider: cloudprovider.QEMU,
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"continue on partially extracted": {
			provider:           cloudprovider.QEMU,
			vars:               qemuVars,
			fs:                 afero.NewMemMapFs(),
			partiallyExtracted: true,
			wantErr:            false,
		},
		"prepare workspace fails": {
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			fs:       afero.NewReadOnlyFs(afero.NewMemMapFs()),
			wantErr:  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			c := &Client{
				tf:         &stubTerraform{},
				file:       file.NewHandler(tc.fs),
				workingDir: constants.TerraformWorkingDir,
			}

			err := c.PrepareWorkspace(tc.provider, tc.vars)

			// Test case: Check if we can continue to create on an incomplete workspace.
			if tc.partiallyExtracted {
				require.NoError(c.file.Remove(filepath.Join(c.workingDir, "main.tf")))
				err = c.PrepareWorkspace(tc.provider, tc.vars)
			}

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestCreateCluster(t *testing.T) {
	someErr := errors.New("failed")
	newTestState := func() *tfjson.State {
		workingState := tfjson.State{
			Values: &tfjson.StateValues{
				Outputs: map[string]*tfjson.StateOutput{
					"ip": {
						Value: "192.0.2.100",
					},
				},
			},
		}
		return &workingState
	}
	qemuVars := &QEMUVariables{
		CommonVariables: CommonVariables{
			Name:               "name",
			CountControlPlanes: 1,
			CountWorkers:       2,
			StateDiskSizeGB:    11,
		},
		CPUCount:         1,
		MemorySizeMiB:    1024,
		ImagePath:        "path",
		ImageFormat:      "format",
		MetadataAPIImage: "api",
	}

	testCases := map[string]struct {
		provider cloudprovider.Provider
		vars     Variables
		tf       *stubTerraform
		fs       afero.Fs
		wantErr  bool
	}{
		"works": {
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{showState: newTestState()},
			fs:       afero.NewMemMapFs(),
		},
		"init fails": {
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{initErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"apply fails": {
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{applyErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"show fails": {
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{showErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"no ip": {
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{},
					},
				},
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"ip has wrong type": {
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{"ip": {Value: 42}},
					},
				},
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			c := &Client{
				tf:         tc.tf,
				file:       file.NewHandler(tc.fs),
				workingDir: constants.TerraformWorkingDir,
			}

			require.NoError(c.PrepareWorkspace(tc.provider, tc.vars))
			ip, err := c.CreateCluster(context.Background())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal("192.0.2.100", ip)
		})
	}
}

func TestDestroyInstances(t *testing.T) {
	testCases := map[string]struct {
		tf      *stubTerraform
		wantErr bool
	}{
		"works": {
			tf: &stubTerraform{},
		},
		"destroy fails": {
			tf: &stubTerraform{
				destroyErr: errors.New("error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &Client{
				tf: tc.tf,
			}

			err := c.DestroyCluster(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
		})
	}
}

func TestCleanupWorkspace(t *testing.T) {
	someContent := []byte("some content")

	testCases := map[string]struct {
		provider  cloudprovider.Provider
		prepareFS func(file.Handler) error
		wantErr   bool
	}{
		"files are cleaned up": {
			provider: cloudprovider.QEMU,
			prepareFS: func(f file.Handler) error {
				var err error
				err = multierr.Append(err, f.Write("terraform.tfvars", someContent))
				err = multierr.Append(err, f.Write("terraform.tfstate", someContent))
				return multierr.Append(err, f.Write("terraform.tfstate.backup", someContent))
			},
		},
		"no error if files do not exist": {
			provider:  cloudprovider.QEMU,
			prepareFS: func(f file.Handler) error { return nil },
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			file := file.NewHandler(afero.NewMemMapFs())
			require.NoError(tc.prepareFS(file))

			c := &Client{
				file:       file,
				tf:         &stubTerraform{},
				workingDir: constants.TerraformWorkingDir,
			}

			err := c.CleanUpWorkspace()
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			_, err = file.Stat(filepath.Join(c.workingDir, "terraform.tfvars"))
			assert.ErrorIs(err, fs.ErrNotExist)
			_, err = file.Stat(filepath.Join(c.workingDir, "terraform.tfstate"))
			assert.ErrorIs(err, fs.ErrNotExist)
			_, err = file.Stat(filepath.Join(c.workingDir, "terraform.tfstate.backup"))
			assert.ErrorIs(err, fs.ErrNotExist)
		})
	}
}

type stubTerraform struct {
	applyErr   error
	destroyErr error
	initErr    error
	showErr    error
	showState  *tfjson.State
}

func (s *stubTerraform) Apply(context.Context, ...tfexec.ApplyOption) error {
	return s.applyErr
}

func (s *stubTerraform) Destroy(context.Context, ...tfexec.DestroyOption) error {
	return s.destroyErr
}

func (s *stubTerraform) Init(context.Context, ...tfexec.InitOption) error {
	return s.initErr
}

func (s *stubTerraform) Show(context.Context, ...tfexec.ShowOption) (*tfjson.State, error) {
	return s.showState, s.showErr
}
