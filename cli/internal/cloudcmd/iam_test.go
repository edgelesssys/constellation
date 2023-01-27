/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/iamid"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIAMCreator(t *testing.T) {
	someErr := errors.New("failed")

	validGCPIAMConfig := GCPIAMConfig{
		Region:           "europe-west1",
		Zone:             "europe-west1-a",
		ProjectID:        "project-1234",
		ServiceAccountID: "const-test",
	}
	validGCPIAMOutput := terraform.IAMOutput{
		GCP: terraform.GCPIAMOutput{
			SaKey: "not_a_secret",
		},
	}
	validGCPIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.GCP,
		GCPOutput: iamid.GCPFile{
			ServiceAccountKey: "not_a_secret",
		},
	}

	validAzureIAMConfig := AzureIAMConfig{
		Region:           "westus",
		ServicePrincipal: "constell-test",
		ResourceGroup:    "constell-test",
	}
	validAzureIAMOutput := terraform.IAMOutput{
		Azure: terraform.AzureIAMOutput{
			SubscriptionID:               "test_subscription_id",
			TenantID:                     "test_tenant_id",
			ApplicationID:                "test_application_id",
			ApplicationClientSecretValue: "test_application_client_secret_value",
			UAMIID:                       "test_uami_id",
		},
	}
	validAzureIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.Azure,
		AzureOutput: iamid.AzureFile{
			SubscriptionID:               "test_subscription_id",
			TenantID:                     "test_tenant_id",
			ApplicationID:                "test_application_id",
			ApplicationClientSecretValue: "test_application_client_secret_value",
			UAMIID:                       "test_uami_id",
		},
	}

	validAWSIAMConfig := AWSIAMConfig{
		Region: "us-east-2",
		Prefix: "test",
	}
	validAWSIAMOutput := terraform.IAMOutput{
		AWS: terraform.AWSIAMOutput{
			WorkerNodeInstanceProfile:   "test_worker_node_instance_profile",
			ControlPlaneInstanceProfile: "test_control_plane_instance_profile",
		},
	}
	validAWSIAMIDFile := iamid.File{
		CloudProvider: cloudprovider.AWS,
		AWSOutput: iamid.AWSFile{
			ControlPlaneInstanceProfile: "test_control_plane_instance_profile",
			WorkerNodeInstanceProfile:   "test_worker_node_instance_profile",
		},
	}

	testCases := map[string]struct {
		tfClient       terraformClient
		newTfClientErr error
		config         *IAMConfig
		provider       cloudprovider.Provider
		wantIAMIDFile  iamid.File
		wantErr        bool
	}{
		"new terraform client err": {
			tfClient:       &stubTerraformClient{},
			newTfClientErr: someErr,
			wantErr:        true,
		},
		"create iam config err": {
			tfClient: &stubTerraformClient{iamOutputErr: someErr},
			wantErr:  true,
		},
		"gcp": {
			tfClient:      &stubTerraformClient{iamOutput: validGCPIAMOutput},
			wantIAMIDFile: validGCPIAMIDFile,
			provider:      cloudprovider.GCP,
			config:        &IAMConfig{GCP: validGCPIAMConfig},
		},
		"azure": {
			tfClient:      &stubTerraformClient{iamOutput: validAzureIAMOutput},
			wantIAMIDFile: validAzureIAMIDFile,
			provider:      cloudprovider.Azure,
			config:        &IAMConfig{Azure: validAzureIAMConfig},
		},
		"aws": {
			tfClient:      &stubTerraformClient{iamOutput: validAWSIAMOutput},
			wantIAMIDFile: validAWSIAMIDFile,
			provider:      cloudprovider.AWS,
			config:        &IAMConfig{AWS: validAWSIAMConfig},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			creator := &IAMCreator{
				out: &bytes.Buffer{},
				newTerraformClient: func(ctx context.Context) (terraformClient, error) {
					return tc.tfClient, tc.newTfClientErr
				},
			}

			idFile, err := creator.Create(context.Background(), tc.provider, tc.config)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.provider, idFile.CloudProvider)
				switch tc.provider {
				case cloudprovider.GCP:
					assert.Equal(tc.wantIAMIDFile.GCPOutput, idFile.GCPOutput)
				case cloudprovider.Azure:
					assert.Equal(tc.wantIAMIDFile.AzureOutput, idFile.AzureOutput)
				case cloudprovider.AWS:
					assert.Equal(tc.wantIAMIDFile.AWSOutput, idFile.AWSOutput)
				}
			}
		})
	}
}

func TestDestroyIAMUser(t *testing.T) {
	someError := errors.New("failed")

	testCases := map[string]struct {
		tfClient       terraformClient
		newTfClientErr error
		wantErr        bool
	}{
		"new terraform client error": {
			tfClient:       &stubTerraformClient{},
			newTfClientErr: someError,
			wantErr:        true,
		},
		"destroy error": {
			tfClient: &stubTerraformClient{destroyErr: someError},
			wantErr:  true,
		},
		"destroy": {
			tfClient: &stubTerraformClient{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			destroyer := &IAMDestroyer{
				newTerraformClient: func(ctx context.Context) (terraformClient, error) {
					return tc.tfClient, tc.newTfClientErr
				},
			}

			err := destroyer.DestroyIAMUser(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestDeleteGCPKeyFile(t *testing.T) {
	require := require.New(t)
	someError := errors.New("failed")
	destroyer := NewIAMDestroyer(context.Background())

	testFsValid := file.NewHandler(afero.NewMemMapFs())
	require.NoError(testFsValid.Write(constants.GCPServiceAccountKeyFile, []byte(`
	{
		"auth_provider_x509_cert_url": "",
		"auth_uri": "",
		"client_email": "",
		"client_id": "",
		"client_x509_cert_url": "",
		"private_key": "",
		"private_key_id": "",
		"project_id": "",
		"token_uri": "",
		"type": ""
	}
	`)))
	testFsInvalid := file.NewHandler(afero.NewMemMapFs())
	require.NoError(testFsInvalid.Write(constants.GCPServiceAccountKeyFile, []byte(`
	asdf
	`)))

	validTfClient := &stubTerraformClient{
		tfjsonState: &tfjson.State{
			Values: &tfjson.StateValues{
				Outputs: map[string]*tfjson.StateOutput{
					"sa_key": {
						Value: "ewoJCSJhdXRoX3Byb3ZpZGVyX3g1MDlfY2VydF91cmwiOiAiIiwKCQkiYXV0aF91cmkiOiAiIiwKCQkiY2xpZW50X2VtYWlsIjogIiIsCgkJImNsaWVudF9pZCI6ICIiLAoJCSJjbGllbnRfeDUwOV9jZXJ0X3VybCI6ICIiLAoJCSJwcml2YXRlX2tleSI6ICIiLAoJCSJwcml2YXRlX2tleV9pZCI6ICIiLAoJCSJwcm9qZWN0X2lkIjogIiIsCgkJInRva2VuX3VyaSI6ICIiLAoJCSJ0eXBlIjogIiIKCX0=",
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		fsHandler   file.Handler
		cl          terraformClient
		wantErr     bool
		wantDeleted bool
	}{
		"valid delete": {
			fsHandler:   testFsValid,
			cl:          validTfClient,
			wantDeleted: true,
		},
		"show error": {
			cl: &stubTerraformClient{
				showErr: someError,
			},
			wantErr: true,
		},
		"nil tfstate values": {
			cl: &stubTerraformClient{
				tfjsonState: &tfjson.State{
					Values: nil,
				},
			},
			wantErr: true,
		},
		"no key": {
			cl: &stubTerraformClient{
				tfjsonState: &tfjson.State{
					Values: &tfjson.StateValues{},
				},
			},
		},
		"invalid base64": {
			cl: &stubTerraformClient{
				tfjsonState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{
							"sa_key": {
								Value: "iamnotvalid",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"valid base64 invalid json": {
			cl: &stubTerraformClient{
				tfjsonState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{
							"sa_key": {
								Value: "YXNkZg==",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"invalid gcp file": {
			cl:        validTfClient,
			fsHandler: testFsInvalid,
			wantErr:   true,
		},
		"not same": {
			fsHandler: testFsValid,
			cl: &stubTerraformClient{
				tfjsonState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{
							"sa_key": {
								Value: "ewoJCSJhdXRoX3Byb3ZpZGVyX3g1MDlfY2VydF91cmwiOiAiIiwKCQkiYXV0aF91cmkiOiAiIiwKCQkiY2xpZW50X2VtYWlsIjogIiIsCgkJImNsaWVudF9pZCI6ICIiLAoJCSJjbGllbnRfeDUwOV9jZXJ0X3VybCI6ICIiLAoJCSJwcml2YXRlX2tleSI6ICJOT1RUSEVTQU1FIiwKCQkicHJpdmF0ZV9rZXlfaWQiOiAiIiwKCQkicHJvamVjdF9pZCI6ICIiLAoJCSJ0b2tlbl91cmkiOiAiIiwKCQkidHlwZSI6ICIiCgl9",
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			deleted, err := destroyer.deleteGCPKeyFile(context.Background(), tc.fsHandler, tc.cl)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			if tc.wantDeleted {
				assert.True(deleted)
				require.NoError(testFsValid.Write(constants.GCPServiceAccountKeyFile, []byte(`
			{
				"auth_provider_x509_cert_url": "",
				"auth_uri": "",
				"client_email": "",
				"client_id": "",
				"client_x509_cert_url": "",
				"private_key": "",
				"private_key_id": "",
				"project_id": "",
				"token_uri": "",
				"type": ""
			}
			`)))
			} else {
				assert.False(deleted)
			}
		})
	}
}
