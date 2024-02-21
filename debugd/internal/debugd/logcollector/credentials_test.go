/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logcollector

import (
	"context"
	"errors"
	"hash/crc32"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	awssecretmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestGetOpensearchCredentialsGCP(t *testing.T) {
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	someErr := errors.New("failed")

	testCases := map[string]struct {
		gcpAPI    gcpSecretManagerAPI
		wantCreds credentials
		wantErr   bool
	}{
		"gcp success": {
			gcpAPI: stubGCPSecretManagerAPI{
				assessSecretVersionResp: &secretmanagerpb.AccessSecretVersionResponse{
					Name: "cluster-instance-gcp",
					Payload: &secretmanagerpb.SecretPayload{
						Data:       []byte("e2e-logs-OpenSearch-password"),
						DataCrc32C: ptr(int64(crc32.Checksum([]byte("e2e-logs-OpenSearch-password"), crc32c))),
					},
				},
			},
			wantCreds: credentials{
				Username: "cluster-instance-gcp",
				Password: "e2e-logs-OpenSearch-password",
			},
		},
		"gcp access secret version error": {
			gcpAPI:  stubGCPSecretManagerAPI{accessSecretVersionErr: someErr},
			wantErr: true,
		},
		"gcp invalid checksum": {
			gcpAPI: stubGCPSecretManagerAPI{
				assessSecretVersionResp: &secretmanagerpb.AccessSecretVersionResponse{
					Name: "cluster-instance-gcp",
					Payload: &secretmanagerpb.SecretPayload{
						Data:       []byte("e2e-logs-OpenSearch-password"),
						DataCrc32C: ptr(int64(crc32.Checksum([]byte("e2e-logs-OpenSearch-password"), crc32c)) + 1),
					},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			g := &gcpCloudCredentialGetter{secretsAPI: tc.gcpAPI}

			gotCreds, err := g.GetOpensearchCredentials(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantCreds, gotCreds)
			}
		})
	}
}

type stubGCPSecretManagerAPI struct {
	assessSecretVersionResp *secretmanagerpb.AccessSecretVersionResponse
	accessSecretVersionErr  error
}

func (s stubGCPSecretManagerAPI) AccessSecretVersion(_ context.Context, _ *secretmanagerpb.AccessSecretVersionRequest,
	_ ...gax.CallOption,
) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return s.assessSecretVersionResp, s.accessSecretVersionErr
}

func (s stubGCPSecretManagerAPI) Close() error {
	return nil
}

func TestGetOpensearchCredentialsAzure(t *testing.T) {
	testCases := map[string]struct {
		azureAPI  azureSecretsAPI
		wantCreds credentials
		wantErr   bool
	}{
		"azure success": {
			azureAPI: stubAzureSecretsAPI{
				getSecretResp: azsecrets.GetSecretResponse{
					Secret: azsecrets.Secret{
						Value: ptr("test-password"),
					},
				},
			},
			wantCreds: credentials{
				Username: "cluster-instance-azure",
				Password: "test-password",
			},
		},
		"azure get secret error": {
			azureAPI: stubAzureSecretsAPI{
				getSecretErr: errors.New("failed"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			a := &azureCloudCredentialGetter{secretsAPI: tc.azureAPI}

			gotCreds, err := a.GetOpensearchCredentials(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantCreds, gotCreds)
			}
		})
	}
}

type stubAzureSecretsAPI struct {
	getSecretResp azsecrets.GetSecretResponse
	getSecretErr  error
}

func (s stubAzureSecretsAPI) GetSecret(_ context.Context, _, _ string, _ *azsecrets.GetSecretOptions,
) (azsecrets.GetSecretResponse, error) {
	return s.getSecretResp, s.getSecretErr
}

func (s stubAzureSecretsAPI) Close() error {
	return nil
}

func TestGetOpensearchCredentialsAWS(t *testing.T) {
	testCases := map[string]struct {
		awsAPI    awsSecretManagerAPI
		wantCreds credentials
		wantErr   bool
	}{
		"aws success": {
			awsAPI: stubAWSSecretsAPI{
				getSecretValueResp: &awssecretmanager.GetSecretValueOutput{
					SecretString: ptr("test-password"),
				},
			},
			wantCreds: credentials{
				Username: "cluster-instance-aws",
				Password: "test-password",
			},
		},
		"aws get secret value error": {
			awsAPI: stubAWSSecretsAPI{
				getSecretValueErr: errors.New("failed"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			a := &awsCloudCredentialGetter{secretmanager: tc.awsAPI}

			gotCreds, err := a.GetOpensearchCredentials(context.Background())

			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.wantCreds, gotCreds)
			}
		})
	}
}

type stubAWSSecretsAPI struct {
	getSecretValueResp *awssecretmanager.GetSecretValueOutput
	getSecretValueErr  error
}

func (s stubAWSSecretsAPI) GetSecretValue(_ context.Context, _ *awssecretmanager.GetSecretValueInput,
	_ ...func(*awssecretmanager.Options),
) (*awssecretmanager.GetSecretValueOutput, error) {
	return s.getSecretValueResp, s.getSecretValueErr
}

func (s stubAWSSecretsAPI) Close() error {
	return nil
}

func ptr[T any](v T) *T {
	return &v
}
