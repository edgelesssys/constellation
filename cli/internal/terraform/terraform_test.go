/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"context"
	"errors"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
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
		pathBase           string
		provider           cloudprovider.Provider
		vars               Variables
		fs                 afero.Fs
		partiallyExtracted bool
		wantErr            bool
	}{
		"qemu": {
			pathBase: "terraform",
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			fs:       afero.NewMemMapFs(),
			wantErr:  false,
		},
		"no vars": {
			pathBase: "terraform",
			provider: cloudprovider.QEMU,
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"continue on partially extracted": {
			pathBase:           "terraform",
			provider:           cloudprovider.QEMU,
			vars:               qemuVars,
			fs:                 afero.NewMemMapFs(),
			partiallyExtracted: true,
			wantErr:            false,
		},
		"prepare workspace fails": {
			pathBase: "terraform",
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

			path := path.Join(tc.pathBase, strings.ToLower(tc.provider.String()))
			err := c.PrepareWorkspace(path, tc.vars)

			// Test case: Check if we can continue to create on an incomplete workspace.
			if tc.partiallyExtracted {
				require.NoError(c.file.Remove(filepath.Join(c.workingDir, "main.tf")))
				err = c.PrepareWorkspace(path, tc.vars)
			}

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestPrepareIAM(t *testing.T) {
	gcpVars := &GCPIAMVariables{
		Project:          "const-1234",
		Region:           "europe-west1",
		Zone:             "europe-west1-a",
		ServiceAccountID: "const-test-case",
	}
	azureVars := &AzureIAMVariables{
		Region:           "westus",
		ServicePrincipal: "constell-test-sp",
		ResourceGroup:    "constell-test-rg",
	}
	awsVars := &AWSIAMVariables{
		Region: "eu-east-2a",
		Prefix: "test",
	}
	testCases := map[string]struct {
		pathBase           string
		provider           cloudprovider.Provider
		vars               Variables
		fs                 afero.Fs
		partiallyExtracted bool
		wantErr            bool
	}{
		"no vars": {
			pathBase: path.Join("terraform", "iam"),
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"invalid path": {
			pathBase: path.Join("abc", "123"),
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"gcp": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			fs:       afero.NewMemMapFs(),
			wantErr:  false,
		},
		"continue on partially extracted": {
			pathBase:           path.Join("terraform", "iam"),
			provider:           cloudprovider.GCP,
			vars:               gcpVars,
			fs:                 afero.NewMemMapFs(),
			partiallyExtracted: true,
			wantErr:            false,
		},
		"azure": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			fs:       afero.NewMemMapFs(),
			wantErr:  false,
		},
		"aws": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			fs:       afero.NewMemMapFs(),
			wantErr:  false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			c := &Client{
				tf:         &stubTerraform{},
				file:       file.NewHandler(tc.fs),
				workingDir: constants.TerraformIAMWorkingDir,
			}

			path := path.Join(tc.pathBase, strings.ToLower(tc.provider.String()))
			err := c.PrepareWorkspace(path, tc.vars)

			// Test case: Check if we can continue to create on an incomplete workspace.
			if tc.partiallyExtracted {
				require.NoError(c.file.Remove(filepath.Join(c.workingDir, "main.tf")))
				err = c.PrepareWorkspace(path, tc.vars)
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
					"initSecret": {
						Value: "initSecret",
					},
					"uid": {
						Value: "12345abc",
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
		pathBase string
		provider cloudprovider.Provider
		vars     Variables
		tf       *stubTerraform
		fs       afero.Fs
		wantErr  bool
	}{
		"works": {
			pathBase: "terraform",
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{showState: newTestState()},
			fs:       afero.NewMemMapFs(),
		},
		"init fails": {
			pathBase: "terraform",
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{initErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"apply fails": {
			pathBase: "terraform",
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{applyErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"show fails": {
			pathBase: "terraform",
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{showErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"no ip": {
			pathBase: "terraform",
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
			pathBase: "terraform",
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
		"no uid": {
			pathBase: "terraform",
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
		"uid has wrong type": {
			pathBase: "terraform",
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{"uid": {Value: 42}},
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

			path := path.Join(tc.pathBase, strings.ToLower(tc.provider.String()))
			require.NoError(c.PrepareWorkspace(path, tc.vars))
			tfOutput, err := c.CreateCluster(context.Background())

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal("192.0.2.100", tfOutput.IP)
			assert.Equal("initSecret", tfOutput.Secret)
			assert.Equal("12345abc", tfOutput.UID)
		})
	}
}

func TestCreateIAM(t *testing.T) {
	someErr := errors.New("failed")
	newTestState := func() *tfjson.State {
		workingState := tfjson.State{
			Values: &tfjson.StateValues{
				Outputs: map[string]*tfjson.StateOutput{
					"sa_key": {
						Value: "12345678_abcdefg",
					},
					"subscription_id": {
						Value: "test_subscription_id",
					},
					"tenant_id": {
						Value: "test_tenant_id",
					},
					"application_id": {
						Value: "test_application_id",
					},
					"uami_id": {
						Value: "test_uami_id",
					},
					"application_client_secret_value": {
						Value: "test_application_client_secret_value",
					},
					"control_plane_instance_profile": {
						Value: "test_control_plane_instance_profile",
					},
					"worker_nodes_instance_profile": {
						Value: "test_worker_nodes_instance_profile",
					},
				},
			},
		}
		return &workingState
	}
	gcpVars := &GCPIAMVariables{
		Project:          "const-1234",
		Region:           "europe-west1",
		Zone:             "europe-west1-a",
		ServiceAccountID: "const-test-case",
	}
	azureVars := &AzureIAMVariables{
		Region:           "westus",
		ServicePrincipal: "constell-test-sp",
		ResourceGroup:    "constell-test-rg",
	}
	awsVars := &AWSIAMVariables{
		Region: "eu-east-2a",
		Prefix: "test",
	}

	testCases := map[string]struct {
		pathBase string
		provider cloudprovider.Provider
		vars     Variables
		tf       *stubTerraform
		fs       afero.Fs
		wantErr  bool
		want     IAMOutput
	}{
		"gcp works": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{showState: newTestState()},
			fs:       afero.NewMemMapFs(),
			want:     IAMOutput{GCP: GCPIAMOutput{SaKey: "12345678_abcdefg"}},
		},
		"gcp init fails": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{initErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"gcp apply fails": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{applyErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"gcp show fails": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{showErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"gcp no sa_key": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
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
		"gcp sa_key has wrong type": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{"sa_key": {Value: 42}},
					},
				},
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"azure works": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			tf:       &stubTerraform{showState: newTestState()},
			fs:       afero.NewMemMapFs(),
			want: IAMOutput{Azure: AzureIAMOutput{
				SubscriptionID:               "test_subscription_id",
				TenantID:                     "test_tenant_id",
				ApplicationID:                "test_application_id",
				ApplicationClientSecretValue: "test_application_client_secret_value",
				UAMIID:                       "test_uami_id",
			}},
		},
		"azure init fails": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			tf:       &stubTerraform{initErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"azure apply fails": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			tf:       &stubTerraform{applyErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"azure show fails": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			tf:       &stubTerraform{showErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"azure no subscription_id": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
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
		"azure subscription_id has wrong type": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{"subscription_id": {Value: 42}},
					},
				},
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"aws works": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf:       &stubTerraform{showState: newTestState()},
			fs:       afero.NewMemMapFs(),
			want: IAMOutput{AWS: AWSIAMOutput{
				ControlPlaneInstanceProfile: "test_control_plane_instance_profile",
				WorkerNodeInstanceProfile:   "test_worker_nodes_instance_profile",
			}},
		},
		"aws init fails": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf:       &stubTerraform{initErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"aws apply fails": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf:       &stubTerraform{applyErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"aws show fails": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf:       &stubTerraform{showErr: someErr},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"aws no control_plane_instance_profile": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
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
		"azure control_plane_instance_profile has wrong type": {
			pathBase: path.Join("terraform", "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{"control_plane_instance_profile": {Value: 42}},
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
				workingDir: constants.TerraformIAMWorkingDir,
			}

			path := path.Join(tc.pathBase, strings.ToLower(tc.provider.String()))
			require.NoError(c.PrepareWorkspace(path, tc.vars))
			IAMoutput, err := c.CreateIAMConfig(context.Background(), tc.provider)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.want, IAMoutput)
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

			err := c.Destroy(context.Background())
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
