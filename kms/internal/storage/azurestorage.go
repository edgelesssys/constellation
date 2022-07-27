package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/edgelesssys/constellation/kms/internal/config"
)

type azureContainerAPI interface {
	Create(ctx context.Context, options *azblob.ContainerCreateOptions) (azblob.ContainerCreateResponse, error)
	NewBlockBlobClient(blobName string) (azureBlobAPI, error)
}

type azureBlobAPI interface {
	DownloadToWriterAt(ctx context.Context, offset int64, count int64, writer io.WriterAt, options azblob.DownloadOptions) error
	Upload(ctx context.Context, body io.ReadSeekCloser, options *azblob.BlockBlobUploadOptions) (azblob.BlockBlobUploadResponse, error)
}

type wrappedAzureClient struct {
	azblob.ContainerClient
}

func (c wrappedAzureClient) NewBlockBlobClient(blobName string) (azureBlobAPI, error) {
	return c.ContainerClient.NewBlockBlobClient(blobName)
}

// AzureStorage is an implementation of the Storage interface, storing keys in the Azure Blob Store.
type AzureStorage struct {
	newClient        func(ctx context.Context, connectionString, containerName string, opts *azblob.ClientOptions) (azureContainerAPI, error)
	connectionString string
	containerName    string
	opts             *AzureOpts
}

// AzureOpts are additional options to be used when interacting with the Azure API.
type AzureOpts struct {
	upload  *azblob.BlockBlobUploadOptions
	service *azblob.ClientOptions
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

	s := &AzureStorage{
		newClient:        azureContainerClientFactory,
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
func (s *AzureStorage) Get(ctx context.Context, keyID string) ([]byte, error) {
	client, err := s.newBlobClient(ctx, keyID)
	if err != nil {
		return nil, err
	}

	// the Azure SDK requires an io.WriterAt, the AWS SDK provides a utility function to create one from a byte slice
	keyBuffer := manager.NewWriteAtBuffer([]byte{})

	opts := azblob.DownloadOptions{
		RetryReaderOptionsPerBlock: azblob.RetryReaderOptions{
			MaxRetryRequests:       5,
			TreatEarlyCloseAsError: true,
		},
	}

	if err := client.DownloadToWriterAt(ctx, 0, 0, keyBuffer, opts); err != nil {
		var storeErr *azblob.StorageError
		if errors.As(err, &storeErr) && (storeErr.ErrorCode == azblob.StorageErrorCodeBlobNotFound) {
			return nil, ErrDEKUnset
		}
		return nil, fmt.Errorf("downloading DEK from storage: %w", err)
	}

	return keyBuffer.Bytes(), nil
}

// Put saves a DEK to Azure Blob Storage by key ID.
func (s *AzureStorage) Put(ctx context.Context, keyID string, encDEK []byte) error {
	client, err := s.newBlobClient(ctx, keyID)
	if err != nil {
		return err
	}

	if _, err := client.Upload(ctx, readSeekNopCloser{bytes.NewReader(encDEK)}, s.opts.upload); err != nil {
		return fmt.Errorf("uploading DEK to storage: %w", err)
	}
	return nil
}

// createContainerOrContinue creates a new storage container if necessary, or continues if it already exists.
func (s *AzureStorage) createContainerOrContinue(ctx context.Context) error {
	client, err := s.newClient(ctx, s.connectionString, s.containerName, s.opts.service)
	if err != nil {
		return err
	}

	var storeErr *azblob.StorageError
	_, err = client.Create(ctx, &azblob.ContainerCreateOptions{
		Metadata: config.StorageTags,
	})
	if (err == nil) || (errors.As(err, &storeErr) && (storeErr.ErrorCode == azblob.StorageErrorCodeContainerAlreadyExists)) {
		return nil
	}

	return fmt.Errorf("creating storage container: %w", err)
}

// newBlobClient is a convenience function to create BlockBlobClients.
func (s *AzureStorage) newBlobClient(ctx context.Context, blobName string) (azureBlobAPI, error) {
	c, err := s.newClient(ctx, s.connectionString, s.containerName, s.opts.service)
	if err != nil {
		return nil, err
	}
	return c.NewBlockBlobClient(blobName)
}

func azureContainerClientFactory(ctx context.Context, connectionString, containerName string, opts *azblob.ClientOptions) (azureContainerAPI, error) {
	service, err := azblob.NewServiceClientFromConnectionString(connectionString, opts)
	if err != nil {
		return nil, fmt.Errorf("creating storage client from connection string: %w", err)
	}

	containerClient, err := service.NewContainerClient(containerName)
	if err != nil {
		return nil, fmt.Errorf("creating storage container client: %w", err)
	}
	return &wrappedAzureClient{*containerClient}, err
}

// readSeekNopCloser is a wrapper for io.ReadSeeker implementing the Close method. This is required by the Azure SDK.
type readSeekNopCloser struct {
	io.ReadSeeker
}

func (n readSeekNopCloser) Close() error {
	return nil
}
