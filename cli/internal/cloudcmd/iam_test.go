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
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
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

func TestDeleteGCPServiceAccountKeyFile(t *testing.T) {
	dummyTfstate := `
	 {
	 	"version": 4,
	 	"terraform_version": "1.3.6",
	 	"serial": 8,
	 	"lineage": "",
	 	"outputs": {
	 	  "sa_key": {
	 		"value": "ewoJCSJhdXRoX3Byb3ZpZGVyX3g1MDlfY2VydF91cmwiOiAiIiwKCQkiYXV0aF91cmkiOiAiIiwKCQkiY2xpZW50X2VtYWlsIjogIiIsCgkJImNsaWVudF9pZCI6ICIiLAoJCSJjbGllbnRfeDUwOV9jZXJ0X3VybCI6ICIiLAoJCSJwcml2YXRlX2tleSI6ICJJQU1BTkVYQU1QTEUiLAoJCSJwcml2YXRlX2tleV9pZCI6ICIiLAoJCSJwcm9qZWN0X2lkIjogIiIsCgkJInRva2VuX3VyaSI6ICIiLAoJCSJ0eXBlIjogIiIKCX0=",
	 		"type": "string",
	 		"sensitive": true
	 	  }
	 	},
	 	"resources": [
	 	  {
	 		"mode": "managed",
	 		"type": "google_project_iam_binding",
	 		"name": "iam_service_account_user_role",
	 		"provider": "provider[\"registry.terraform.io/hashicorp/google\"]",
	 		"instances": [
	 		  {
	 			"schema_version": 0,
	 			"attributes": {
	 			  "condition": [],
	 			  "etag": "",
	 			  "id": "",
	 			  "members": [
	 				""
	 			  ],
	 			  "project": "",
	 			  "role": ""
	 			},
	 			"sensitive_attributes": [],
	 			"private": "",
	 			"dependencies": [
	 			  "google_service_account.service_account"
	 			]
	 		  }
	 		]
	 	  }
	 	],
	 	"check_results": null
	   }
	 `
	dummyTfstateNoKey := `
	 {
	 	"version": 4,
	 	"terraform_version": "1.3.6",
	 	"serial": 8,
	 	"lineage": "",
	 	"outputs": {},
	 	"resources": [
	 	  {
	 		"mode": "managed",
	 		"type": "google_project_iam_binding",
	 		"name": "iam_service_account_user_role",
	 		"provider": "provider[\"registry.terraform.io/hashicorp/google\"]",
	 		"instances": [
	 		  {
	 			"schema_version": 0,
	 			"attributes": {
	 			  "condition": [],
	 			  "etag": "",
	 			  "id": "",
	 			  "members": [
	 				""
	 			  ],
	 			  "project": "",
	 			  "role": ""
	 			},
	 			"sensitive_attributes": [],
	 			"private": "",
	 			"dependencies": [
	 			  "google_service_account.service_account"
	 			]
	 		  }
	 		]
	 	  }
	 	],
	 	"check_results": null
	   }
	 `

	dummyGCPKeyFile := `
	  {
	  	"auth_provider_x509_cert_url": "",
	  	"auth_uri": "",
	  	"client_email": "",
	  	"client_id": "",
	  	"client_x509_cert_url": "",
	  	"private_key": "IAMANEXAMPLE",
	  	"private_key_id": "",
	  	"project_id": "",
	  	"token_uri": "",
	  	"type": ""
	  }
	  `
	dummyInvalidGCPKeyFile := `
	 {
	 	"auth_provider_x509_cert_url": "",
	 	"auth_uri": "",
	 	"client_email": "",
	 	"client_id": "",
	 	"client_x509_cert_url": "",
	 	"private_key": "IAMWRONG",
	 	"private_key_id": "",
	 	"project_id": "",
	 	"token_uri": "",
	 	"type": ""
	 }
	 `

	fsWithGCPFile := file.NewHandler(afero.NewMemMapFs())
	fsWithoutGCPFile := file.NewHandler(afero.NewMemMapFs())
	fsWithUnmatchingConfig := file.NewHandler(afero.NewMemMapFs())
	fsNoKey := file.NewHandler(afero.NewMemMapFs())

	// tfstate setup
	fsNoKey.MkdirAll(constants.TerraformIAMWorkingDir)
	fsWithGCPFile.MkdirAll(constants.TerraformIAMWorkingDir)
	fsWithoutGCPFile.MkdirAll(constants.TerraformIAMWorkingDir)
	fsWithUnmatchingConfig.MkdirAll(constants.TerraformIAMWorkingDir)
	fsNoKey.Write(constants.TerraformIAMWorkingDir+"/terraform.tfstate", []byte(dummyTfstateNoKey))
	fsWithGCPFile.Write(constants.TerraformIAMWorkingDir+"/terraform.tfstate", []byte(dummyTfstate))
	fsWithoutGCPFile.Write(constants.TerraformIAMWorkingDir+"/terraform.tfstate", []byte(dummyTfstate))
	fsWithUnmatchingConfig.Write(constants.TerraformIAMWorkingDir+"/terraform.tfstate", []byte(dummyTfstate))

	// keyfile setup
	fsNoKey.Write(constants.GCPServiceAccountKeyFile, []byte(dummyGCPKeyFile))
	fsWithGCPFile.Write(constants.GCPServiceAccountKeyFile, []byte(dummyGCPKeyFile))
	fsWithUnmatchingConfig.Write(constants.GCPServiceAccountKeyFile, []byte(dummyInvalidGCPKeyFile))

	testCases := map[string]struct {
		fsHandler  file.Handler
		wantErr    bool
		wantDelete bool
	}{
		"no key": {
			fsHandler: fsNoKey,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			destroyer := NewIAMDestroyer(context.Background())
			deleted, err := destroyer.DeleteGCPServiceAccountKeyFile(context.Background(), tc.fsHandler)

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
			}

			if tc.wantDelete {
				assert.True(deleted)
			} else {
				assert.False(deleted)
			}
		})
	}
}
