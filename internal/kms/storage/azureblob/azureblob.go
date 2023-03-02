/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package azureblob implements a storage backend for the KMS using Azure Blob Storage.
package azureblob

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/edgelesssys/constellation/v2/internal/kms/config"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
)

type azureBlobAPI interface {
	CreateContainer(context.Context, string, *container.CreateOptions) (azblob.CreateContainerResponse, error)
	DownloadStream(context.Context, string, string, *blob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error)
	UploadStream(context.Context, string, string, io.Reader, *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error)
}

// Storage is an implementation of the Storage interface, storing keys in the Azure Blob Store.
type Storage struct {
	client    azureBlobAPI
	container string
}

// New initializes a storage client using Azure's Blob Storage using the provided config.
//
// See the Azure docs for more information: https://azure.microsoft.com/en-us/services/storage/blobs/
func New(ctx context.Context, cfg uri.AzureBlobConfig) (*Storage, error) {
	var creds azcore.TokenCredential

	creds, err := azidentity.NewClientSecretCredential(cfg.TenantID, cfg.ClientID, cfg.ClientSecret, nil)
	if err != nil {
		// Fallback: try to load default credentials
		creds, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("invalid client-secret credentials. Trying to load default credentials: %w", err)
		}
	}

	client, err := azblob.NewClient(fmt.Sprintf("https://%s.blob.core.windows.net/", cfg.StorageAccount), creds, nil)
	if err != nil {
		return nil, fmt.Errorf("creating storage client: %w", err)
	}

	s := &Storage{
		client:    client,
		container: cfg.Container,
	}

	// Try to create a new storage container, continue if it already exists
	if err := s.createContainerOrContinue(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

// Get returns a DEK from from Azure Blob Storage by key ID.
func (s *Storage) Get(ctx context.Context, keyID string) ([]byte, error) {
	res, err := s.client.DownloadStream(ctx, s.container, keyID, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return nil, storage.ErrDEKUnset
		}
		return nil, fmt.Errorf("downloading DEK from storage: %w", err)
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

// Put saves a DEK to Azure Blob Storage by key ID.
func (s *Storage) Put(ctx context.Context, keyID string, encDEK []byte) error {
	if _, err := s.client.UploadStream(ctx, s.container, keyID, bytes.NewReader(encDEK), nil); err != nil {
		return fmt.Errorf("uploading DEK to storage: %w", err)
	}

	return nil
}

// createContainerOrContinue creates a new storage container if necessary, or continues if it already exists.
func (s *Storage) createContainerOrContinue(ctx context.Context) error {
	_, err := s.client.CreateContainer(ctx, s.container, &azblob.CreateContainerOptions{
		Metadata: config.StorageTags,
	})
	if (err == nil) || bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
		return nil
	}

	return fmt.Errorf("creating storage container: %w", err)
}
