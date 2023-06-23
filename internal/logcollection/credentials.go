/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package logcollection

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"strings"

	gcpsecretmanager "cloud.google.com/go/secretmanager/apiv1"
	gcpsecretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssecretmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	gaxv2 "github.com/googleapis/gax-go/v2"
)

// Credentials contains the Credentials for an OpenSearch instance.
type Credentials struct {
	Username string
	Password string
}

// NewCloudCredentialGetter returns a new credentialGetter for the given cloud provider.
// qemuPassword is the password for the qemu user on the QEMU cloud provider. If another CSP is used,
// it can be ignored.
func NewCloudCredentialGetter(ctx context.Context, provider cloudprovider.Provider, qemuPassword string) (*credentialGetter, error) {
	switch provider {
	case cloudprovider.GCP:
		getter, err := newGCPCloudCredentialGetter(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating GCP cloud credential getter: %w", err)
		}
		return &credentialGetter{getter}, nil
	case cloudprovider.Azure:
		getter, err := newAzureCloudCredentialGetter()
		if err != nil {
			return nil, fmt.Errorf("creating Azure cloud credential getter: %w", err)
		}
		return &credentialGetter{getter}, nil
	case cloudprovider.AWS:
		getter, err := newAWSCloudCredentialGetter(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating AWS cloud credential getter: %w", err)
		}
		return &credentialGetter{getter}, nil
	case cloudprovider.QEMU:
		getter, err := newQemuCloudCredentialGetter(qemuPassword)
		if err != nil {
			return nil, fmt.Errorf("creating QEMU cloud credential getter: %w", err)
		}
		return &credentialGetter{getter}, nil
	default:
		return nil, fmt.Errorf("cloud provider not supported")
	}
}

// credentialGetter is a wrapper around the cloud provider specific credential getters.
type credentialGetter struct {
	openseachCredsGetter
}

type openseachCredsGetter interface {
	GetOpensearchCredentials(ctx context.Context) (Credentials, error)
	io.Closer
}

type gcpCloudCredentialGetter struct {
	secretsAPI gcpSecretManagerAPI
}

func newGCPCloudCredentialGetter(ctx context.Context) (*gcpCloudCredentialGetter, error) {
	client, err := gcpsecretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating secretmanager client: %w", err)
	}
	return &gcpCloudCredentialGetter{secretsAPI: client}, nil
}

func (g *gcpCloudCredentialGetter) GetOpensearchCredentials(ctx context.Context) (Credentials, error) {
	const secretName = "projects/796962942582/secrets/e2e-logs-OpenSearch-password/versions/1"
	const username = "cluster-instance-gcp"

	req := &gcpsecretmanagerpb.AccessSecretVersionRequest{Name: secretName}
	resp, err := g.secretsAPI.AccessSecretVersion(ctx, req)
	if err != nil {
		return Credentials{}, fmt.Errorf("accessing secret version: %w", err)
	}

	if resp.Payload == nil || len(resp.Payload.Data) == 0 {
		return Credentials{}, errors.New("response payload is empty")
	}

	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(resp.Payload.Data, crc32c))
	if checksum != *resp.Payload.DataCrc32C {
		return Credentials{}, errors.New("data corruption of secret detected")
	}

	return Credentials{
		Username: username,
		Password: string(resp.Payload.Data),
	}, nil
}

func (g *gcpCloudCredentialGetter) Close() error {
	return g.secretsAPI.Close()
}

type gcpSecretManagerAPI interface {
	AccessSecretVersion(ctx context.Context, req *gcpsecretmanagerpb.AccessSecretVersionRequest,
		opts ...gaxv2.CallOption,
	) (*gcpsecretmanagerpb.AccessSecretVersionResponse, error)
	io.Closer
}

type azureCloudCredentialGetter struct {
	secretsAPI azureSecretsAPI
}

func newAzureCloudCredentialGetter() (*azureCloudCredentialGetter, error) {
	const vaultURI = "https://opensearch-creds.vault.azure.net"

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("creating default azure credential: %w", err)
	}

	client, err := azsecrets.NewClient(vaultURI, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating Azure secrets client: %w", err)
	}

	return &azureCloudCredentialGetter{secretsAPI: client}, nil
}

func (a *azureCloudCredentialGetter) GetOpensearchCredentials(ctx context.Context) (Credentials, error) {
	const secretName = "opensearch-password"
	const username = "cluster-instance-azure"
	const version = "" // An empty string version gets the latest version of the secret.

	resp, err := a.secretsAPI.GetSecret(ctx, secretName, version, nil)
	if err != nil {
		return Credentials{}, fmt.Errorf("getting secret: %w", err)
	}

	if resp.Value == nil {
		return Credentials{}, errors.New("response value is empty")
	}

	return Credentials{
		Username: username,
		Password: *resp.Value,
	}, nil
}

func (a *azureCloudCredentialGetter) Close() error {
	return nil
}

type azureSecretsAPI interface {
	GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions,
	) (azsecrets.GetSecretResponse, error)
}

type awsCloudCredentialGetter struct {
	secretmanager awsSecretManagerAPI
}

func newAWSCloudCredentialGetter(ctx context.Context) (*awsCloudCredentialGetter, error) {
	const region = "eu-central-1"

	config, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading default AWS config: %w", err)
	}

	client := awssecretmanager.NewFromConfig(config)

	return &awsCloudCredentialGetter{secretmanager: client}, nil
}

func (a *awsCloudCredentialGetter) GetOpensearchCredentials(ctx context.Context) (Credentials, error) {
	const username = "cluster-instance-aws"
	secertName := "opensearch-password"

	req := &awssecretmanager.GetSecretValueInput{SecretId: &secertName}
	resp, err := a.secretmanager.GetSecretValue(ctx, req)
	if err != nil {
		return Credentials{}, fmt.Errorf("getting secret value: %w", err)
	}

	if resp.SecretString == nil {
		return Credentials{}, errors.New("response secret string is empty")
	}

	password := strings.TrimPrefix(*resp.SecretString, "{\"password\":\"")
	password = strings.TrimSuffix(password, "\"}")

	return Credentials{
		Username: username,
		Password: password,
	}, nil
}

func (a *awsCloudCredentialGetter) Close() error {
	return nil
}

type awsSecretManagerAPI interface {
	GetSecretValue(ctx context.Context, params *awssecretmanager.GetSecretValueInput,
		optFns ...func(*awssecretmanager.Options),
	) (*awssecretmanager.GetSecretValueOutput, error)
}

type qemuCloudCredentialGetter struct {
	creds Credentials
}

func newQemuCloudCredentialGetter(password string) (*qemuCloudCredentialGetter, error) {
	const username = "cluster-instance-qemu"

	return &qemuCloudCredentialGetter{
		creds: Credentials{
			Username: username,
			Password: password,
		},
	}, nil
}

func (q *qemuCloudCredentialGetter) GetOpensearchCredentials(_ context.Context) (Credentials, error) {
	return q.creds, nil
}

func (q *qemuCloudCredentialGetter) Close() error {
	return nil
}
