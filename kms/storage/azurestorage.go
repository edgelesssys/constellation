package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/edgelesssys/constellation/kms/config"
)

// AzureStorage is an implementation of the Storage interface, storing keys in the Azure Blob Store.
type AzureStorage struct {
	client azblob.ContainerClient
	opts   *AzureOpts
}

// AzureOpts are additional options to be used when interacting with the Azure API.
type AzureOpts struct {
	download *azblob.DownloadBlobOptions
	upload   *azblob.UploadBlockBlobOptions
	service  *azblob.ClientOptions
}

// NewAzureStorage initializes a storage client using Azure's Blob Storage: https://azure.microsoft.com/en-us/services/storage/blobs/
//
// A connections string is required to connect to the Storage Account, see https://docs.microsoft.com/en-us/azure/storage/common/storage-configure-connection-string
// If the container does not exists, a new one is created automatically.
// Connect options for the Client, Downloader and Uploader can be configured using opts.
func NewAzureStorage(ctx context.Context, connectionString, containerName string, opts *AzureOpts) (*AzureStorage, error) {
	if opts == nil {
		opts = &AzureOpts{}
	}
	service, err := azblob.NewServiceClientFromConnectionString(connectionString, opts.service)
	if err != nil {
		return nil, fmt.Errorf("creating storage client from connection string: %w", err)
	}
	client := service.NewContainerClient(containerName)

	// Try to create a new storage container, continue if it already exists
	_, err = client.Create(ctx, &azblob.CreateContainerOptions{
		Metadata: config.StorageTags,
	})
	if (err != nil) && !strings.Contains(err.Error(), string(azblob.StorageErrorCodeContainerAlreadyExists)) {
		return nil, fmt.Errorf("creating storage container: %w", err)
	}

	return &AzureStorage{client: client, opts: opts}, nil
}

// Get returns a DEK from from Azure Blob Storage by key ID.
func (s *AzureStorage) Get(ctx context.Context, keyID string) ([]byte, error) {
	client := s.client.NewBlockBlobClient(keyID)
	res, err := client.Download(ctx, s.opts.download)
	if err != nil {
		var storeErr *azblob.StorageError
		if errors.As(err, &storeErr) && (storeErr.ErrorCode == azblob.StorageErrorCodeBlobNotFound) {
			return nil, ErrDEKUnset
		}
		return nil, fmt.Errorf("downloading DEK from storage: %w", err)
	}

	key := &bytes.Buffer{}
	reader := res.Body(&azblob.RetryReaderOptions{MaxRetryRequests: 5, TreatEarlyCloseAsError: true})
	defer reader.Close()
	_, err = key.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("downloading DEK from storage: %w", err)
	}

	return key.Bytes(), nil
}

// Put saves a DEK to Azure Blob Storage by key ID.
func (s *AzureStorage) Put(ctx context.Context, keyID string, encDEK []byte) error {
	client := s.client.NewBlockBlobClient(keyID)
	if _, err := client.Upload(ctx, newNopCloser(bytes.NewReader(encDEK)), s.opts.upload); err != nil {
		return fmt.Errorf("uploading DEK to storage: %w", err)
	}
	return nil
}

// nopCloser is a wrapper for io.ReadSeeker implementing the Close method. This is required by the Azure SDK.
type nopCloser struct {
	io.ReadSeeker
}

func (n nopCloser) Close() error {
	return nil
}

// newNopCloser returns a ReadSeekCloser with a no-op close method wrapping the provided io.ReadSeeker.
func newNopCloser(rs io.ReadSeeker) io.ReadSeekCloser {
	return nopCloser{rs}
}
