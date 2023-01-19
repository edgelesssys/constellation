/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package aws implements a KMS backend for AWS KMS.
package aws

import (
	"context"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	kmsInterface "github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/internal"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
)

const (
	// DEKContext is used as the encryption context in AWS KMS.
	DEKContext = "aws:ebs:id"
)

// ClientAPI satisfies the Amazons KMS client's methods we need.
// This allows us to mock the actual client, see https://aws.github.io/aws-sdk-go-v2/docs/unit-testing/
type ClientAPI interface {
	CreateAlias(ctx context.Context, params *kms.CreateAliasInput, optFns ...func(*kms.Options)) (*kms.CreateAliasOutput, error)
	CreateKey(ctx context.Context, params *kms.CreateKeyInput, optFns ...func(*kms.Options)) (*kms.CreateKeyOutput, error)
	Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error)
	DeleteAlias(ctx context.Context, params *kms.DeleteAliasInput, optFns ...func(*kms.Options)) (*kms.DeleteAliasOutput, error)
	DescribeKey(ctx context.Context, params *kms.DescribeKeyInput, optFns ...func(*kms.Options)) (*kms.DescribeKeyOutput, error)
	Encrypt(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error)
	GenerateDataKey(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error)
	GenerateDataKeyWithoutPlaintext(ctx context.Context, params *kms.GenerateDataKeyWithoutPlaintextInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyWithoutPlaintextOutput, error)
	GetParametersForImport(ctx context.Context, params *kms.GetParametersForImportInput, optFns ...func(*kms.Options)) (*kms.GetParametersForImportOutput, error)
	ImportKeyMaterial(ctx context.Context, params *kms.ImportKeyMaterialInput, optFns ...func(*kms.Options)) (*kms.ImportKeyMaterialOutput, error)
	PutKeyPolicy(ctx context.Context, params *kms.PutKeyPolicyInput, optFns ...func(*kms.Options)) (*kms.PutKeyPolicyOutput, error)
	ScheduleKeyDeletion(ctx context.Context, params *kms.ScheduleKeyDeletionInput, optFns ...func(*kms.Options)) (*kms.ScheduleKeyDeletionOutput, error)
}

// KeyPolicyProducer allows to have callbacks for generating key policies at runtime.
type KeyPolicyProducer interface {
	// CreateKeyPolicy returns a key policy for a given key ID.
	CreateKeyPolicy(keyID string) (string, error)
}

// KMSClient implements the CloudKMS interface for AWS.
type KMSClient struct {
	awsClient      ClientAPI
	policyProducer KeyPolicyProducer
	storage        kmsInterface.Storage
	kekID          string

	kms *internal.KMSClient
}

// New creates and initializes a new KMSClient for AWS.
//
// The parameter client needs to be initialized with valid AWS credentials (https://aws.github.io/aws-sdk-go-v2/docs/getting-started).
// If storage is nil, the default MemMapStorage is used.
func New(ctx context.Context, policyProducer KeyPolicyProducer, store kmsInterface.Storage, kekID string, optFns ...func(*awsconfig.LoadOptions) error) (*KMSClient, error) {
	if store == nil {
		store = storage.NewMemMapStorage()
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, err
	}
	client := kms.NewFromConfig(cfg)

	return &KMSClient{
		awsClient:      client,
		policyProducer: policyProducer,
		storage:        store,
		kekID:          kekID,
	}, nil
}

// GetDEK returns the DEK for dekID and kekID from the KMS.
func (c *KMSClient) GetDEK(ctx context.Context, keyID string, dekSize int) ([]byte, error) {
	return c.kms.GetDEK(ctx, keyID, dekSize)
}

// Close is a no-op for AWS.
func (c *KMSClient) Close() {}
