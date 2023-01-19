/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package azure implements KMS backends for Azure Key Vault and Azure managed HSM.
package azure

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/kms/kms"
	"github.com/edgelesssys/constellation/v2/internal/kms/kms/internal"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	azurekeyvault "github.com/hashicorp/go-kms-wrapping/wrappers/azurekeyvault/v2"
)

// KMSClient implements the CloudKMS interface for Azure Key Vault.
type KMSClient struct {
	kms *internal.KMSClient
}

// New initializes a KMS client for Azure Key Vault.
func New(ctx context.Context, store kms.Storage, cfg uri.AzureConfig) (*KMSClient, error) {
	if store == nil {
		store = storage.NewMemMapStorage()
	}

	wrapper := azurekeyvault.NewWrapper()
	if _, err := wrapper.SetConfig(
		ctx,
		azurekeyvault.WithTenantId(cfg.TenantID),
		azurekeyvault.WithClientId(cfg.ClientID),
		azurekeyvault.WithClientSecret(cfg.ClientSecret),
		azurekeyvault.WithResource(string(cfg.VaultType)),
		azurekeyvault.WithVaultName(cfg.VaultName),
		azurekeyvault.WithKeyName(cfg.KeyName),
	); err != nil {
		return nil, fmt.Errorf("setting Azure Key Vault config: %w", err)
	}

	return &KMSClient{
		kms: &internal.KMSClient{
			Storage: store,
			Wrapper: wrapper,
		},
	}, nil
}

// GetDEK fetches an encrypted Data Encryption Key from storage and decrypts it using a KEK stored in Google's KMS.
func (c *KMSClient) GetDEK(ctx context.Context, keyID string, dekSize int) ([]byte, error) {
	return c.kms.GetDEK(ctx, keyID, dekSize)
}

// Close is a no-op for Azure.
func (c *KMSClient) Close() {}
