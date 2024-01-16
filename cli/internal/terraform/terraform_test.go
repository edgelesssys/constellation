/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestPrepareCluster(t *testing.T) {
	qemuVars := &QEMUVariables{
		Name: "name",
		NodeGroups: map[string]QEMUNodeGroup{
			"control-plane": {
				Role:       role.ControlPlane.TFString(),
				DiskSize:   30,
				CPUCount:   1,
				MemorySize: 1024,
			},
		},
		Machine: "q35",
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
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			fs:       afero.NewMemMapFs(),
			wantErr:  false,
		},
		"no vars": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"continue on partially extracted": {
			pathBase:           constants.TerraformEmbeddedDir,
			provider:           cloudprovider.QEMU,
			vars:               qemuVars,
			fs:                 afero.NewMemMapFs(),
			partiallyExtracted: true,
			wantErr:            false,
		},
		"prepare workspace fails": {
			pathBase: constants.TerraformEmbeddedDir,
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
				workingDir: "unittest",
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
		Location:      "westus",
		ResourceGroup: "constell-test-rg",
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
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"invalid path": {
			pathBase: path.Join("abc", "123"),
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"gcp": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			fs:       afero.NewMemMapFs(),
			wantErr:  false,
		},
		"continue on partially extracted": {
			pathBase:           path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider:           cloudprovider.GCP,
			vars:               gcpVars,
			fs:                 afero.NewMemMapFs(),
			partiallyExtracted: true,
			wantErr:            false,
		},
		"azure": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			fs:       afero.NewMemMapFs(),
			wantErr:  false,
		},
		"aws": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
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
	newQEMUState := func() *tfjson.State {
		workingState := tfjson.State{
			Values: &tfjson.StateValues{
				Outputs: map[string]*tfjson.StateOutput{
					"out_of_cluster_endpoint": {
						Value: "192.0.2.100",
					},
					"in_cluster_endpoint": {
						Value: "192.0.2.101",
					},
					"init_secret": {
						Value: "initSecret",
					},
					"uid": {
						Value: "12345abc",
					},
					"api_server_cert_sans": {
						Value: []any{"192.0.2.100"},
					},
					"name": {
						Value: "constell-12345abc",
					},
					"ip_cidr_node": {
						Value: "192.0.2.103/32",
					},
				},
			},
		}
		return &workingState
	}
	newAzureState := func() *tfjson.State {
		workingState := tfjson.State{
			Values: &tfjson.StateValues{
				Outputs: map[string]*tfjson.StateOutput{
					"out_of_cluster_endpoint": {
						Value: "192.0.2.100",
					},
					"in_cluster_endpoint": {
						Value: "192.0.2.101",
					},
					"init_secret": {
						Value: "initSecret",
					},
					"uid": {
						Value: "12345abc",
					},
					"attestation_url": {
						Value: "https://12345.neu.attest.azure.net",
					},
					"api_server_cert_sans": {
						Value: []any{"192.0.2.100"},
					},
					"user_assigned_identity_client_id": {
						Value: "test_uami_id",
					},
					"resource_group": {
						Value: "test_rg",
					},
					"subscription_id": {
						Value: "test_subscription_id",
					},
					"network_security_group_name": {
						Value: "test_nsg_name",
					},
					"loadbalancer_name": {
						Value: "test_lb_name",
					},
					"name": {
						Value: "constell-12345abc",
					},
					"ip_cidr_node": {
						Value: "192.0.2.103/32",
					},
				},
			},
		}
		return &workingState
	}
	qemuVars := &QEMUVariables{
		Name: "name",
		NodeGroups: map[string]QEMUNodeGroup{
			"control-plane": {
				Role:       role.ControlPlane.TFString(),
				DiskSize:   11,
				CPUCount:   1,
				MemorySize: 1024,
			},
		},
		Machine:          "q35",
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
		// expectedAttestationURL is the expected attestation URL to be returned by
		// the Terraform client. It is declared in the test case because it is
		// provider-specific.
		expectedAttestationURL string
		wantErr                bool
	}{
		"works": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{showState: newQEMUState()},
			fs:       afero.NewMemMapFs(),
		},
		"init fails": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{initErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"apply fails": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{applyErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"show fails": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{showErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"set log fails": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{setLogErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"set log path fails": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf:       &stubTerraform{setLogPathErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"no ip": {
			pathBase: constants.TerraformEmbeddedDir,
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
			pathBase: constants.TerraformEmbeddedDir,
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
			pathBase: constants.TerraformEmbeddedDir,
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
			pathBase: constants.TerraformEmbeddedDir,
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
		"name has wrong type": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.QEMU,
			vars:     qemuVars,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{"name": {Value: 42}},
					},
				},
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"working attestation url": {
			pathBase:               constants.TerraformEmbeddedDir,
			provider:               cloudprovider.Azure,
			vars:                   qemuVars, // works for mocking azure vars
			tf:                     &stubTerraform{showState: newAzureState()},
			fs:                     afero.NewMemMapFs(),
			expectedAttestationURL: "https://12345.neu.attest.azure.net",
		},
		"no attestation url": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.Azure,
			vars:     qemuVars, // works for mocking azure vars
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
		"attestation url has wrong type": {
			pathBase: constants.TerraformEmbeddedDir,
			provider: cloudprovider.Azure,
			vars:     qemuVars, // works for mocking azure vars
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{"attestation_url": {Value: 42}},
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
				workingDir: "unittest",
			}

			path := path.Join(tc.pathBase, strings.ToLower(tc.provider.String()))
			require.NoError(c.PrepareWorkspace(path, tc.vars))
			infraState, err := c.ApplyCluster(context.Background(), tc.provider, LogLevelDebug)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal("192.0.2.100", infraState.ClusterEndpoint)
			assert.Equal(state.HexBytes("initSecret"), infraState.InitSecret)
			assert.Equal("12345abc", infraState.UID)
			assert.Equal("192.0.2.101", infraState.InClusterEndpoint)
			assert.Equal("192.0.2.103/32", infraState.IPCidrNode)
			if tc.provider == cloudprovider.Azure {
				assert.Equal(tc.expectedAttestationURL, infraState.Azure.AttestationURL)
			}
		})
	}
}

func TestCreateIAM(t *testing.T) {
	newTestState := func() *tfjson.State {
		workingState := tfjson.State{
			Values: &tfjson.StateValues{
				Outputs: map[string]*tfjson.StateOutput{
					"service_account_key": {
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
					"iam_instance_profile_name_control_plane": {
						Value: "test_iam_instance_profile_name_control_plane",
					},
					"iam_instance_profile_name_worker_nodes": {
						Value: "test_iam_instance_profile_name_worker_nodes",
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
		Location:      "westus",
		ResourceGroup: "constell-test-rg",
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
		"set log fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{setLogErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"set log path fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{setLogPathErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"gcp works": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{showState: newTestState()},
			fs:       afero.NewMemMapFs(),
			want:     IAMOutput{GCP: GCPIAMOutput{SaKey: "12345678_abcdefg"}},
		},
		"gcp init fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{initErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"gcp apply fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{applyErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"gcp show fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf:       &stubTerraform{showErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"gcp no service_account_key": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
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
		"gcp service_account_key has wrong type": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.GCP,
			vars:     gcpVars,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{"service_account_key": {Value: 42}},
					},
				},
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"azure works": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			tf:       &stubTerraform{showState: newTestState()},
			fs:       afero.NewMemMapFs(),
			want: IAMOutput{Azure: AzureIAMOutput{
				SubscriptionID: "test_subscription_id",
				TenantID:       "test_tenant_id",
				UAMIID:         "test_uami_id",
			}},
		},
		"azure init fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			tf:       &stubTerraform{initErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"azure apply fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			tf:       &stubTerraform{applyErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"azure show fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.Azure,
			vars:     azureVars,
			tf:       &stubTerraform{showErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"azure no subscription_id": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
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
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
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
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf:       &stubTerraform{showState: newTestState()},
			fs:       afero.NewMemMapFs(),
			want: IAMOutput{AWS: AWSIAMOutput{
				ControlPlaneInstanceProfile: "test_iam_instance_profile_name_control_plane",
				WorkerNodeInstanceProfile:   "test_iam_instance_profile_name_worker_nodes",
			}},
		},
		"aws init fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf:       &stubTerraform{initErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"aws apply fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf:       &stubTerraform{applyErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"aws show fails": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf:       &stubTerraform{showErr: assert.AnError},
			fs:       afero.NewMemMapFs(),
			wantErr:  true,
		},
		"aws no iam_instance_profile_name_control_plane": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
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
		"azure iam_instance_profile_name_control_plane has wrong type": {
			pathBase: path.Join(constants.TerraformEmbeddedDir, "iam"),
			provider: cloudprovider.AWS,
			vars:     awsVars,
			tf: &stubTerraform{
				showState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{"iam_instance_profile_name_control_plane": {Value: 42}},
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
			IAMoutput, err := c.ApplyIAM(context.Background(), tc.provider, LogLevelDebug)

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
			tf:      &stubTerraform{destroyErr: assert.AnError},
			wantErr: true,
		},
		"setLog fails": {
			tf:      &stubTerraform{setLogErr: assert.AnError},
			wantErr: true,
		},
		"setLogPath fails": {
			tf:      &stubTerraform{setLogPathErr: assert.AnError},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &Client{
				tf: tc.tf,
			}

			err := c.Destroy(context.Background(), LogLevelDebug)
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
				if err := f.Write("terraform.tfvars", someContent); err != nil {
					return err
				}
				if err := f.Write("terraform.tfstate", someContent); err != nil {
					return err
				}
				return f.Write("terraform.tfstate.backup", someContent)
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
				workingDir: "unittest",
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

func TestParseLogLevel(t *testing.T) {
	testCases := map[string]struct {
		level   string
		want    LogLevel
		wantErr bool
	}{
		"json": {
			level: "json",
			want:  LogLevelJSON,
		},
		"trace": {
			level: "trace",
			want:  LogLevelTrace,
		},
		"debug": {
			level: "debug",
			want:  LogLevelDebug,
		},
		"info": {
			level: "info",
			want:  LogLevelInfo,
		},
		"warn": {
			level: "warn",
			want:  LogLevelWarn,
		},
		"error": {
			level: "error",
			want:  LogLevelError,
		},
		"none": {
			level: "none",
			want:  LogLevelNone,
		},
		"unknown": {
			level:   "unknown",
			wantErr: true,
		},
		"empty": {
			level:   "",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			level, err := ParseLogLevel(tc.level)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.want, level)
		})
	}
}

func TestLogLevelString(t *testing.T) {
	testCases := map[string]struct {
		level LogLevel
		want  string
	}{
		"json": {
			level: LogLevelJSON,
			want:  "JSON",
		},
		"trace": {
			level: LogLevelTrace,
			want:  "TRACE",
		},
		"debug": {
			level: LogLevelDebug,
			want:  "DEBUG",
		},
		"info": {
			level: LogLevelInfo,
			want:  "INFO",
		},
		"warn": {
			level: LogLevelWarn,
			want:  "WARN",
		},
		"error": {
			level: LogLevelError,
			want:  "ERROR",
		},
		"none": {
			level: LogLevelNone,
			want:  "",
		},
		"invalid int": {
			level: LogLevel(-1),
			want:  "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			assert.Equal(tc.want, tc.level.String())
		})
	}
}

func TestPlan(t *testing.T) {
	someError := errors.New("some error")

	testCases := map[string]struct {
		pathBase string
		tf       *stubTerraform
		fs       afero.Fs
		wantErr  bool
	}{
		"plan succeeds": {
			pathBase: constants.TerraformEmbeddedDir,
			tf:       &stubTerraform{},
			fs:       afero.NewMemMapFs(),
		},
		"set log path fails": {
			pathBase: constants.TerraformEmbeddedDir,
			tf: &stubTerraform{
				setLogPathErr: someError,
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"set log fails": {
			pathBase: constants.TerraformEmbeddedDir,
			tf: &stubTerraform{
				setLogErr: someError,
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"plan fails": {
			pathBase: constants.TerraformEmbeddedDir,
			tf: &stubTerraform{
				planJSONErr: someError,
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"init fails": {
			pathBase: constants.TerraformEmbeddedDir,
			tf: &stubTerraform{
				initErr: someError,
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			c := &Client{
				file:       file.NewHandler(tc.fs),
				tf:         tc.tf,
				workingDir: tc.pathBase,
			}

			_, err := c.Plan(context.Background(), LogLevelDebug)
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestShowPlan(t *testing.T) {
	someError := errors.New("some error")
	testCases := map[string]struct {
		pathBase string
		tf       *stubTerraform
		fs       afero.Fs
		wantErr  bool
	}{
		"show plan succeeds": {
			pathBase: constants.TerraformEmbeddedDir,
			tf:       &stubTerraform{},
			fs:       afero.NewMemMapFs(),
		},
		"set log path fails": {
			pathBase: constants.TerraformEmbeddedDir,
			tf: &stubTerraform{
				setLogPathErr: someError,
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"set log fails": {
			pathBase: constants.TerraformEmbeddedDir,
			tf: &stubTerraform{
				setLogErr: someError,
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
		"show plan file fails": {
			pathBase: constants.TerraformEmbeddedDir,
			tf: &stubTerraform{
				showPlanFileErr: someError,
			},
			fs:      afero.NewMemMapFs(),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			c := &Client{
				file:       file.NewHandler(tc.fs),
				tf:         tc.tf,
				workingDir: tc.pathBase,
			}

			err := c.ShowPlan(context.Background(), LogLevelDebug, bytes.NewBuffer(nil))
			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
			}
		})
	}
}

func TestShowIAM(t *testing.T) {
	testCases := map[string]struct {
		tf      *stubTerraform
		csp     cloudprovider.Provider
		wantErr bool
	}{
		"GCP success": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"service_account_key": "key",
				}),
			},
			csp: cloudprovider.GCP,
		},
		"GCP wrong data type": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"service_account_key": map[string]any{},
				}),
			},
			csp:     cloudprovider.GCP,
			wantErr: true,
		},
		"GCP missing key": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{}),
			},
			csp:     cloudprovider.GCP,
			wantErr: true,
		},
		"Azure success": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"subscription_id": "sub",
					"tenant_id":       "tenant",
					"uami_id":         "uami",
				}),
			},
			csp: cloudprovider.Azure,
		},
		"Azure wrong data type subscription_id": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"subscription_id": map[string]any{},
					"tenant_id":       "tenant",
					"uami_id":         "uami",
				}),
			},
			csp:     cloudprovider.Azure,
			wantErr: true,
		},
		"Azure wrong data type tenant_id": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"subscription_id": "sub",
					"tenant_id":       map[string]any{},
					"uami_id":         "uami",
				}),
			},
			csp:     cloudprovider.Azure,
			wantErr: true,
		},
		"Azure wrong data type uami_id": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"subscription_id": "sub",
					"tenant_id":       "tenant",
					"uami_id":         map[string]any{},
				}),
			},
			csp:     cloudprovider.Azure,
			wantErr: true,
		},
		"Azure missing uami_id": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"subscription_id": "sub",
					"tenant_id":       "tenant",
				}),
			},
			csp:     cloudprovider.Azure,
			wantErr: true,
		},
		"Azure missing tenant_id": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"subscription_id": "sub",
					"uami_id":         "uami",
				}),
			},
			csp:     cloudprovider.Azure,
			wantErr: true,
		},
		"Azure missing subscription_id": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"tenant_id": "tenant",
					"uami_id":   "uami",
				}),
			},
			csp:     cloudprovider.Azure,
			wantErr: true,
		},
		"AWS success": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"iam_instance_profile_name_control_plane": "profile",
					"iam_instance_profile_name_worker_nodes":  "profile",
				}),
			},
			csp: cloudprovider.AWS,
		},
		"AWS wrong data type iam_instance_profile_name_control_plane": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"iam_instance_profile_name_control_plane": map[string]any{},
					"iam_instance_profile_name_worker_nodes":  "profile",
				}),
			},
			csp:     cloudprovider.AWS,
			wantErr: true,
		},
		"AWS wrong data type iam_instance_profile_name_worker_nodes": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"iam_instance_profile_name_control_plane": "profile",
					"iam_instance_profile_name_worker_nodes":  map[string]any{},
				}),
			},
			csp:     cloudprovider.AWS,
			wantErr: true,
		},
		"AWS missing iam_instance_profile_name_control_plane": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"iam_instance_profile_name_worker_nodes": "profile",
				}),
			},
			csp:     cloudprovider.AWS,
			wantErr: true,
		},
		"AWS missing iam_instance_profile_name_worker_nodes": {
			tf: &stubTerraform{
				showState: getTfjsonState(map[string]any{
					"iam_instance_profile_name_control_plane": "profile",
				}),
			},
			csp:     cloudprovider.AWS,
			wantErr: true,
		},
		"Show fails": {
			tf: &stubTerraform{
				showErr: assert.AnError,
			},
			csp:     cloudprovider.AWS,
			wantErr: true,
		},
		"Show returns state with nil Value": {
			tf: &stubTerraform{
				showState: &tfjson.State{},
			},
			csp:     cloudprovider.AWS,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			c := &Client{
				tf: tc.tf,
			}

			_, err := c.ShowIAM(context.Background(), tc.csp)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

type stubTerraform struct {
	applyErr        error
	destroyErr      error
	initErr         error
	showErr         error
	setLogErr       error
	setLogPathErr   error
	planJSONErr     error
	showPlanFileErr error
	stateMvErr      error
	showState       *tfjson.State
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

func (s *stubTerraform) Plan(context.Context, ...tfexec.PlanOption) (bool, error) {
	return false, s.planJSONErr
}

func (s *stubTerraform) ShowPlanFileRaw(context.Context, string, ...tfexec.ShowOption) (string, error) {
	return "", s.showPlanFileErr
}

func (s *stubTerraform) SetLog(_ string) error {
	return s.setLogErr
}

func (s *stubTerraform) SetLogPath(_ string) error {
	return s.setLogPathErr
}

func (s *stubTerraform) StateMv(_ context.Context, _, _ string, _ ...tfexec.StateMvCmdOption) error {
	return s.stateMvErr
}

func getTfjsonState(values map[string]any) *tfjson.State {
	state := tfjson.State{
		Values: &tfjson.StateValues{
			Outputs: map[string]*tfjson.StateOutput{},
		},
	}
	for k, v := range values {
		state.Values.Outputs[k] = &tfjson.StateOutput{Value: v}
	}
	return &state
}
