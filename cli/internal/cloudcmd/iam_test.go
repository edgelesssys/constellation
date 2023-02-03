/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/cli/internal/iamid"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	tfjson "github.com/hashicorp/terraform-json"
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
	newError := func() error {
		return errors.New("failed")
	}

	testCases := map[string]struct {
		tfClient       terraformClient
		newTfClientErr error
		wantErr        bool
	}{
		"new terraform client error": {
			tfClient:       &stubTerraformClient{},
			newTfClientErr: newError(),
			wantErr:        true,
		},
		"destroy error": {
			tfClient: &stubTerraformClient{destroyErr: newError()},
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

			err := destroyer.DestroyIAMConfiguration(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestGetTfstateSaKey(t *testing.T) {
	someError := errors.New("failed")
	destroyer := NewIAMDestroyer(context.Background())

	gcpFile := `
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
	`
	gcpFileB64 := base64.StdEncoding.EncodeToString([]byte(gcpFile))

	testCases := map[string]struct {
		cl             terraformClient
		wantValidSaKey bool
		wantErr        bool
	}{
		"valid": {
			cl: &stubTerraformClient{
				tfjsonState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{
							"sa_key": {
								Value: gcpFileB64,
							},
						},
					},
				},
			},
			wantValidSaKey: true,
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
			wantErr: true,
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
								Value: base64.StdEncoding.EncodeToString([]byte("asdf")),
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"not string": {
			cl: &stubTerraformClient{
				tfjsonState: &tfjson.State{
					Values: &tfjson.StateValues{
						Outputs: map[string]*tfjson.StateOutput{
							"sa_key": {
								Value: 1,
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			saKey, err := destroyer.getTfstateSaKey(context.Background(), tc.cl)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			if tc.wantValidSaKey {
				var saKeyComp gcpshared.ServiceAccountKey
				require.NoError(t, json.Unmarshal([]byte(gcpFile), &saKeyComp))

				assert.Equal(saKey, saKeyComp)
			}
		})
	}
}
