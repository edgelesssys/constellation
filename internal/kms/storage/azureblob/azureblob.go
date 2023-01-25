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

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/edgelesssys/constellation/v2/internal/kms/config"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
)

type azureBlobAPI interface {
	CreateContainer(context.Context, string, *container.CreateOptions) (azblob.CreateContainerResponse, error)
	DownloadStream(context.Context, string, string, *blob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error)
	UploadStream(context.Context, string, string, io.Reader, *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error)
}

// Storage is an implementation of the Storage interface, storing keys in the Azure Blob Store.
type Storage struct {
	client           azureBlobAPI
	connectionString string
	containerName    string
	opts             *AzureOpts
}

// AzureOpts are additional options to be used when interacting with the Azure API.
type AzureOpts struct {
	service  *azblob.ClientOptions
	download *azblob.DownloadStreamOptions
	upload   *azblob.UploadStreamOptions
}

// New initializes a storage client using Azure's Blob Storage: https://azure.microsoft.com/en-us/services/storage/blobs/
//
// A connections string is required to connect to the Storage Account, see https://docs.microsoft.com/en-us/azure/storage/common/storage-configure-connection-string
// If the container does not exists, a new one is created automatically.
// Connect options for the Client, Downloader and Uploader can be configured using opts.
func New(ctx context.Context, connectionString, containerName string, opts *AzureOpts) (*Storage, error) {
	if opts == nil {
		opts = &AzureOpts{}
	}

	client, err := azblob.NewClientFromConnectionString(connectionString, opts.service)
	if err != nil {
		return nil, fmt.Errorf("creating storage client from connection string: %w", err)
	}

	s := &Storage{
		client:           client,
		connectionString: connectionString,
		containerName:    containerName,
		opts:             opts,
	}

	// Try to create a new storage container, continue if it already exists
	if err := s.createContainerOrContinue(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

// Get returns a DEK from from Azure Blob Storage by key ID.
func (s *Storage) Get(ctx context.Context, keyID string) ([]byte, error) {
	res, err := s.client.DownloadStream(ctx, s.containerName, keyID, s.opts.download)
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
	if _, err := s.client.UploadStream(ctx, s.containerName, keyID, bytes.NewReader(encDEK), s.opts.upload); err != nil {
		return fmt.Errorf("uploading DEK to storage: %w", err)
	}

	return nil
}

// createContainerOrContinue creates a new storage container if necessary, or continues if it already exists.
func (s *Storage) createContainerOrContinue(ctx context.Context) error {
	_, err := s.client.CreateContainer(ctx, s.containerName, &azblob.CreateContainerOptions{
		Metadata: config.StorageTags,
	})
	if (err == nil) || bloberror.HasCode(err, bloberror.ContainerAlreadyExists) {
		return nil
	}

	return fmt.Errorf("creating storage container: %w", err)
}
