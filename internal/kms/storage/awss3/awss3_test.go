/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package awss3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/kms/storage"
	"github.com/stretchr/testify/assert"
)

type stubAWSS3StorageClient struct {
	getObjectOutputData []byte
	getObjectErr        error
	putObjectErr        error
	savedObject         []byte
	createBucketCalled  bool
	createBucketErr     error
}

func (s *stubAWSS3StorageClient) GetObject(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(s.getObjectOutputData)),
	}, s.getObjectErr
}

func (s *stubAWSS3StorageClient) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	out, err := io.ReadAll(params.Body)
	if err != nil {
		panic(err)
	}
	s.savedObject = out
	return &s3.PutObjectOutput{}, s.putObjectErr
}

func (s *stubAWSS3StorageClient) CreateBucket(_ context.Context, _ *s3.CreateBucketInput, _ ...func(*s3.Options)) (*s3.CreateBucketOutput, error) {
	s.createBucketCalled = true
	return &s3.CreateBucketOutput{}, s.createBucketErr
}

func TestAWSS3Get(t *testing.T) {
	testCases := map[string]struct {
		client     *stubAWSS3StorageClient
		unsetError bool
		wantErr    bool
	}{
		"Get successful": {
			client: &stubAWSS3StorageClient{getObjectOutputData: []byte("test-data")},
		},
		"GetObject fails": {
			client:  &stubAWSS3StorageClient{getObjectErr: errors.New("error")},
			wantErr: true,
		},
		"GetObject fails with NoSuchKey": {
			client:     &stubAWSS3StorageClient{getObjectErr: &types.NoSuchKey{}},
			wantErr:    true,
			unsetError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			store := &Storage{
				client: tc.client,
			}

			out, err := store.Get(context.Background(), "test-key")
			if tc.wantErr {
				assert.Error(err)

				if tc.unsetError {
					assert.ErrorIs(err, storage.ErrDEKUnset)
				} else {
					assert.False(errors.Is(err, storage.ErrDEKUnset))
				}

			} else {
				assert.NoError(err)
				assert.Equal(tc.client.getObjectOutputData, out)
			}
		})
	}
}

func TestAWSS3Put(t *testing.T) {
	testCases := map[string]struct {
		client  *stubAWSS3StorageClient
		wantErr bool
	}{
		"Put successful": {
			client: &stubAWSS3StorageClient{},
		},
		"PutObject fails": {
			client:  &stubAWSS3StorageClient{putObjectErr: errors.New("error")},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			store := &Storage{
				client: tc.client,
			}

			testData := []byte{0x1, 0x2, 0x3}

			err := store.Put(context.Background(), "test-key", testData)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(testData, tc.client.savedObject)
			}
		})
	}
}

func TestAWSS3CreateBucket(t *testing.T) {
	testCases := map[string]struct {
		client  *stubAWSS3StorageClient
		wantErr bool
	}{
		"CreateBucket successful": {
			client: &stubAWSS3StorageClient{},
		},
		"CreateBucket fails": {
			client:  &stubAWSS3StorageClient{createBucketErr: errors.New("error")},
			wantErr: true,
		},
		"CreateBucket fails with BucketAlreadyExists": {
			client:  &stubAWSS3StorageClient{createBucketErr: &types.BucketAlreadyExists{}},
			wantErr: false,
		},
		"CreateBucket fails with BucketAlreadyOwnedByYou": {
			client:  &stubAWSS3StorageClient{createBucketErr: &types.BucketAlreadyOwnedByYou{}},
			wantErr: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			store := &Storage{
				client: tc.client,
			}

			err := store.createBucket(context.Background(), "test-bucket", "test-region")
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.True(tc.client.createBucketCalled)
			}
		})
	}
}
