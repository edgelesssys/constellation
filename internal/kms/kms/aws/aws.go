/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package aws implements a KMS backend for AWS KMS.
package aws

import (
	"context"
	"errors"
	"fmt"

	kmsInterface "github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/internal"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	wrapping "github.com/hashicorp/go-kms-wrapping/v2"
	awskms "github.com/hashicorp/go-kms-wrapping/wrappers/awskms/v2"
)

// KMSClient implements the CloudKMS interface for AWS.
type KMSClient struct {
	kms *internal.KMSClient
}

// New creates and initializes a new KMSClient for AWS.
func New(ctx context.Context, store kmsInterface.Storage, cfg uri.AWSConfig) (*KMSClient, error) {
	if store == nil {
		return nil, errors.New("no storage backend provided for KMS")
	}

	wrapper := awskms.NewWrapper()
	if _, err := wrapper.SetConfig(
		ctx,
		wrapping.WithKeyId(cfg.KeyName),
		awskms.WithRegion(cfg.Region),
		awskms.WithAccessKey(cfg.AccessKeyID),
		awskms.WithSecretKey(cfg.AccessKey),
	); err != nil {
		return nil, fmt.Errorf("setting AWS KMS config: %w", err)
	}
	return &KMSClient{
		kms: &internal.KMSClient{
			Storage: store,
			Wrapper: wrapper,
		},
	}, nil
}

// GetDEK fetches an encrypted Data Encryption Key from storage and decrypts it using a KEK stored in AWS KMS.
func (c *KMSClient) GetDEK(ctx context.Context, keyID string, dekSize int) ([]byte, error) {
	return c.kms.GetDEK(ctx, keyID, dekSize)
}

// Close is a no-op for AWS.
func (c *KMSClient) Close() {}
