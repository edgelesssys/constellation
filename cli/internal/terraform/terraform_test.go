/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"context"
	"errors"
	"io/fs"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/multierr"
)

func TestCreateInstances(t *testing.T) {
	someErr := errors.New("error")
	getState := func() *tfjson.State {
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

	testCases := map[string]struct {
		provider cloudprovider.Provider
		input    CreateClusterInput
		loader   *stubLoader
		tf       *stubTerraform
		wantErr  bool
	}{
		"works": {
			provider: cloudprovider.QEMU,
			tf: &stubTerraform{
				showState: getState(),
			},
			loader: &stubLoader{},
		},
		"init fails": {
			provider: cloudprovider.QEMU,
			tf: &stubTerraform{
				initErr:   someErr,
				showState: getState(),
			},
			loader:  &stubLoader{},
			wantErr: true,
		},
		"apply fails": {
			provider: cloudprovider.QEMU,
			tf: &stubTerraform{
				applyErr:  someErr,
				showState: getState(),
			},
			loader:  &stubLoader{},
			wantErr: true,
		},
		"show fails": {
			provider: cloudprovider.QEMU,
			tf: &stubTerraform{
				showErr: someErr,
			},
			loader:  &stubLoader{},
			wantErr: true,
		},
		"no ip": {
			provider: cloudprovider.QEMU,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{},
					},
				},
			},
			loader:  &stubLoader{},
			wantErr: true,
		},
		"prepare workspace fails": {
			provider: cloudprovider.QEMU,
			tf: &stubTerraform{
				showState: getState(),
			},
			loader:  &stubLoader{prepareErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &Client{
				provider: tc.provider,
				tf:       tc.tf,
				file:     file.NewHandler(afero.NewMemMapFs()),
				loader:   tc.loader,
			}

			err := c.CreateCluster(context.Background(), "test", tc.input)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
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
				provider: cloudprovider.QEMU,
				tf:       tc.tf,
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
		loader    *stubLoader
		prepareFS func(file.Handler) error
		wantErr   bool
	}{
		"files are cleaned up": {
			provider: cloudprovider.QEMU,
			loader:   &stubLoader{},
			prepareFS: func(f file.Handler) error {
				var err error
				err = multierr.Append(err, f.Write("terraform.tfvars", someContent))
				err = multierr.Append(err, f.Write("terraform.tfstate", someContent))
				return multierr.Append(err, f.Write("terraform.tfstate.backup", someContent))
			},
		},
		"no error if files do not exist": {
			provider:  cloudprovider.QEMU,
			loader:    &stubLoader{},
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
				provider: tc.provider,
				loader:   tc.loader,
				file:     file,
				tf:       &stubTerraform{},
			}

			err := c.CleanUpWorkspace()
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			_, err = file.Stat("terraform.tfvars")
			assert.ErrorIs(err, fs.ErrNotExist)
			_, err = file.Stat("terraform.tfstate")
			assert.ErrorIs(err, fs.ErrNotExist)
			_, err = file.Stat("terraform.tfstate.backup")
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

type stubLoader struct {
	prepareErr error
	cleanupErr error
}

func (s *stubLoader) prepareWorkspace() error {
	return s.prepareErr
}

func (s *stubLoader) cleanUpWorkspace() error {
	return s.cleanupErr
}
