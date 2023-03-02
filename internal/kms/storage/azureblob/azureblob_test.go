/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azureblob

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/stretchr/testify/assert"
)

func TestAzureGet(t *testing.T) {
	testCases := map[string]struct {
		client     stubAzureBlobAPI
		unsetError bool
		wantErr    bool
	}{
		"success": {
			client: stubAzureBlobAPI{downloadData: []byte{0x1, 0x2, 0x3}},
		},
		"DownloadBuffer fails": {
			client:  stubAzureBlobAPI{downloadErr: errors.New("failed")},
			wantErr: true,
		},
		"BlobNotFound error": {
			client:     stubAzureBlobAPI{downloadErr: &azcore.ResponseError{ErrorCode: string(bloberror.BlobNotFound)}},
			unsetError: true,
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &Storage{
				client:    &tc.client,
				container: "test",
			}

			out, err := client.Get(context.Background(), "test-key")
			if tc.wantErr {
				assert.Error(err)

				if tc.unsetError {
					assert.ErrorIs(err, storage.ErrDEKUnset)
				} else {
					assert.False(errors.Is(err, storage.ErrDEKUnset))
				}
				return
			}
			assert.NoError(err)
			assert.Equal(tc.client.downloadData, out)
		})
	}
}

func TestAzurePut(t *testing.T) {
	testCases := map[string]struct {
		client  stubAzureBlobAPI
		wantErr bool
	}{
		"success": {
			client: stubAzureBlobAPI{},
		},
		"Upload fails": {
			client:  stubAzureBlobAPI{uploadErr: errors.New("failed")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			testData := []byte{0x1, 0x2, 0x3}

			client := &Storage{
				client:    &tc.client,
				container: "test",
			}

			err := client.Put(context.Background(), "test-key", testData)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(testData, tc.client.uploadData)
		})
	}
}

func TestCreateContainerOrContinue(t *testing.T) {
	testCases := map[string]struct {
		client  stubAzureBlobAPI
		wantErr bool
	}{
		"success": {
			client: stubAzureBlobAPI{},
		},
		"container already exists": {
			client: stubAzureBlobAPI{createErr: &azcore.ResponseError{ErrorCode: string(bloberror.ContainerAlreadyExists)}},
		},
		"CreateContainer fails": {
			client:  stubAzureBlobAPI{createErr: errors.New("failed")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &Storage{
				client:    &tc.client,
				container: "test",
			}

			err := client.createContainerOrContinue(context.Background())
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.True(tc.client.createCalled)
			}
		})
	}
}

type stubAzureBlobAPI struct {
	createErr    error
	createCalled bool
	downloadErr  error
	downloadData []byte
	uploadErr    error
	uploadData   []byte
}

func (s *stubAzureBlobAPI) CreateContainer(context.Context, string, *container.CreateOptions) (azblob.CreateContainerResponse, error) {
	s.createCalled = true
	return azblob.CreateContainerResponse{}, s.createErr
}

func (s *stubAzureBlobAPI) DownloadStream(context.Context, string, string, *blob.DownloadStreamOptions) (blob.DownloadStreamResponse, error) {
	res := blob.DownloadStreamResponse{}
	res.Body = io.NopCloser(bytes.NewReader(s.downloadData))
	return res, s.downloadErr
}

func (s *stubAzureBlobAPI) UploadStream(_ context.Context, _, _ string, data io.Reader, _ *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error) {
	uploadData, err := io.ReadAll(data)
	if err != nil {
		return azblob.UploadStreamResponse{}, err
	}
	s.uploadData = uploadData
	return azblob.UploadStreamResponse{}, s.uploadErr
}
