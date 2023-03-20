/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package gcp implements a KMS backend for Google Cloud KMS.

The following permissions are required for the service account used to authenticate with GCP:
  - cloudkms.cryptoKeys.get
  - cloudkms.cryptoKeys.encrypt
  - cloudkms.cryptoKeys.decrypt
*/
package gcp

import (
	"context"
	"errors"
	"fmt"
	"io"

	kmsInterface "github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/internal"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	gcpckms "github.com/hashicorp/go-kms-wrapping/wrappers/gcpckms/v2"
)

// KMSClient implements the CloudKMS interface for Google Cloud Platform.
type KMSClient struct {
	kms    *internal.KMSClient
	client io.Closer
}

// New initializes a KMS client for Google Cloud Platform.
func New(_ context.Context, store kmsInterface.Storage, cfg uri.GCPConfig) (*KMSClient, error) {
	if store == nil {
		return nil, errors.New("no storage backend provided for KMS")
	}

	wrapper := gcpckms.NewWrapper()
	if _, err := wrapper.SetConfig(
		context.Background(),
		gcpckms.WithProject(cfg.ProjectID),
		gcpckms.WithRegion(cfg.Location),
		gcpckms.WithKeyRing(cfg.KeyRing),
		gcpckms.WithCryptoKey(cfg.KeyName),
		gcpckms.WithCredentials(cfg.CredentialsPath),
	); err != nil {
		return nil, fmt.Errorf("setting GCP KMS config: %w", err)
	}

	return &KMSClient{
		kms: &internal.KMSClient{
			Storage: store,
			Wrapper: wrapper,
		},
		client: wrapper.Client(),
	}, nil
}

// GetDEK fetches an encrypted Data Encryption Key from storage and decrypts it using a KEK stored in Google's KMS.
func (c *KMSClient) GetDEK(ctx context.Context, keyID string, dekSize int) ([]byte, error) {
	return c.kms.GetDEK(ctx, keyID, dekSize)
}

// Close closes the KMS client.
func (c *KMSClient) Close() {
	_ = c.client.Close()
}
