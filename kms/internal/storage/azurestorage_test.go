package storage

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/stretchr/testify/assert"
)

type stubAzureContainerAPI struct {
	newClientErr error
	createErr    error
	createCalled *bool
	blockBlobAPI stubAzureBlockBlobAPI
}

func newStubClientFactory(stub stubAzureContainerAPI) func(ctx context.Context, connectionString, containerName string, opts *azblob.ClientOptions) (azureContainerAPI, error) {
	return func(ctx context.Context, connectionString, containerName string, opts *azblob.ClientOptions) (azureContainerAPI, error) {
		return stub, stub.newClientErr
	}
}

func (s stubAzureContainerAPI) Create(ctx context.Context, options *azblob.ContainerCreateOptions) (azblob.ContainerCreateResponse, error) {
	*s.createCalled = true
	return azblob.ContainerCreateResponse{}, s.createErr
}

func (s stubAzureContainerAPI) NewBlockBlobClient(blobName string) (azureBlobAPI, error) {
	return s.blockBlobAPI, nil
}

type stubAzureBlockBlobAPI struct {
	downloadBlobToWriterAtErr  error
	downloadBlobToWriterOutput []byte
	uploadErr                  error
	uploadData                 chan []byte
}

func (s stubAzureBlockBlobAPI) DownloadToWriterAt(ctx context.Context, offset int64, count int64, writer io.WriterAt, o azblob.DownloadOptions) error {
	if _, err := writer.WriteAt(s.downloadBlobToWriterOutput, 0); err != nil {
		panic(err)
	}
	return s.downloadBlobToWriterAtErr
}

func (s stubAzureBlockBlobAPI) Upload(ctx context.Context, body io.ReadSeekCloser, options *azblob.BlockBlobUploadOptions) (azblob.BlockBlobUploadResponse, error) {
	res, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}
	s.uploadData <- res
	return azblob.BlockBlobUploadResponse{}, s.uploadErr
}

func TestAzureGet(t *testing.T) {
	someErr := errors.New("error")

	testCases := map[string]struct {
		client     stubAzureContainerAPI
		unsetError bool
		wantErr    bool
	}{
		"success": {
			client: stubAzureContainerAPI{
				blockBlobAPI: stubAzureBlockBlobAPI{downloadBlobToWriterOutput: []byte("test-data")},
			},
		},
		"creating client fails": {
			client:  stubAzureContainerAPI{newClientErr: someErr},
			wantErr: true,
		},
		"DownloadBlobToBuffer fails": {
			client: stubAzureContainerAPI{
				blockBlobAPI: stubAzureBlockBlobAPI{downloadBlobToWriterAtErr: someErr},
			},
			wantErr: true,
		},
		"BlobNotFound error": {
			client: stubAzureContainerAPI{
				blockBlobAPI: stubAzureBlockBlobAPI{
					downloadBlobToWriterAtErr: &azblob.StorageError{
						ErrorCode: azblob.StorageErrorCodeBlobNotFound,
					},
				},
			},
			unsetError: true,
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &AzureStorage{
				newClient:        newStubClientFactory(tc.client),
				connectionString: "test",
				containerName:    "test",
				opts:             &AzureOpts{},
			}

			out, err := client.Get(context.Background(), "test-key")
			if tc.wantErr {
				assert.Error(err)

				if tc.unsetError {
					assert.ErrorIs(err, ErrDEKUnset)
				} else {
					assert.False(errors.Is(err, ErrDEKUnset))
				}

			} else {
				assert.NoError(err)
				assert.Equal(tc.client.blockBlobAPI.downloadBlobToWriterOutput, out)
			}
		})
	}
}

func TestAzurePut(t *testing.T) {
	someErr := errors.New("error")

	testCases := map[string]struct {
		client  stubAzureContainerAPI
		wantErr bool
	}{
		"success": {
			client: stubAzureContainerAPI{},
		},
		"creating client fails": {
			client:  stubAzureContainerAPI{newClientErr: someErr},
			wantErr: true,
		},
		"Upload fails": {
			client: stubAzureContainerAPI{
				blockBlobAPI: stubAzureBlockBlobAPI{uploadErr: someErr},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testData := []byte{0x1, 0x2, 0x3}
			tc.client.blockBlobAPI.uploadData = make(chan []byte, len(testData))

			client := &AzureStorage{
				newClient:        newStubClientFactory(tc.client),
				connectionString: "test",
				containerName:    "test",
				opts:             &AzureOpts{},
			}

			err := client.Put(context.Background(), "test-key", testData)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(testData, <-tc.client.blockBlobAPI.uploadData)
			}
		})
	}
}

func TestCreateContainerOrContinue(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		client  stubAzureContainerAPI
		wantErr bool
	}{
		"success": {
			client: stubAzureContainerAPI{},
		},
		"container already exists": {
			client: stubAzureContainerAPI{createErr: &azblob.StorageError{ErrorCode: azblob.StorageErrorCodeContainerAlreadyExists}},
		},
		"creating client fails": {
			client:  stubAzureContainerAPI{newClientErr: someErr},
			wantErr: true,
		},
		"Create fails": {
			client:  stubAzureContainerAPI{createErr: someErr},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			tc.client.createCalled = new(bool)
			client := &AzureStorage{
				newClient:        newStubClientFactory(tc.client),
				connectionString: "test",
				containerName:    "test",
				opts:             &AzureOpts{},
			}

			err := client.createContainerOrContinue(context.Background())
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.True(*tc.client.createCalled)
			}
		})
	}
}
