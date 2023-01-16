/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
)

type stubGCPStorageAPI struct {
	newClientErr       error
	attrsErr           error
	createBucketErr    error
	createBucketCalled bool
	newReaderErr       error
	newReaderOutput    []byte
	writer             *stubWriteCloser
}

func (s *stubGCPStorageAPI) stubClientFactory(ctx context.Context, opts ...option.ClientOption) (gcpStorageAPI, error) {
	return s, s.newClientErr
}

func (s *stubGCPStorageAPI) Attrs(ctx context.Context, bucketName string) (*storage.BucketAttrs, error) {
	return &storage.BucketAttrs{}, s.attrsErr
}

func (s *stubGCPStorageAPI) Close() error {
	return nil
}

func (s *stubGCPStorageAPI) CreateBucket(ctx context.Context, bucketName, projectID string, attrs *storage.BucketAttrs) error {
	s.createBucketCalled = true
	return s.createBucketErr
}

func (s *stubGCPStorageAPI) NewWriter(ctx context.Context, bucketName, objectName string) io.WriteCloser {
	return s.writer
}

func (s *stubGCPStorageAPI) NewReader(ctx context.Context, bucketName, objectName string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.newReaderOutput)), s.newReaderErr
}

type stubWriteCloser struct {
	result   *[]byte
	writeErr error
	writeN   int
}

func (s stubWriteCloser) Write(p []byte) (int, error) {
	*s.result = p
	return s.writeN, s.writeErr
}

func (s stubWriteCloser) Close() error {
	return nil
}

func TestGCPGet(t *testing.T) {
	someErr := errors.New("error")

	testCases := map[string]struct {
		client     *stubGCPStorageAPI
		unsetError bool
		wantErr    bool
	}{
		"success": {
			client: &stubGCPStorageAPI{newReaderOutput: []byte("test-data")},
		},
		"creating client fails": {
			client:  &stubGCPStorageAPI{newClientErr: someErr},
			wantErr: true,
		},
		"NewReader fails": {
			client:  &stubGCPStorageAPI{newReaderErr: someErr},
			wantErr: true,
		},
		"ErrObjectNotExist error": {
			client:     &stubGCPStorageAPI{newReaderErr: storage.ErrObjectNotExist},
			unsetError: true,
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &GoogleCloudStorage{
				newClient:  tc.client.stubClientFactory,
				projectID:  "test",
				bucketName: "test",
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
				assert.Equal(tc.client.newReaderOutput, out)
			}
		})
	}
}

func TestGCPPut(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		client     *stubGCPStorageAPI
		unsetError bool
		wantErr    bool
	}{
		"success": {
			client: &stubGCPStorageAPI{
				writer: &stubWriteCloser{
					result: new([]byte),
				},
			},
		},
		"creating client fails": {
			client:  &stubGCPStorageAPI{newClientErr: someErr},
			wantErr: true,
		},
		"NewWriter fails": {
			client: &stubGCPStorageAPI{
				writer: &stubWriteCloser{
					result:   new([]byte),
					writeErr: someErr,
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &GoogleCloudStorage{
				newClient:  tc.client.stubClientFactory,
				projectID:  "test",
				bucketName: "test",
			}
			testData := []byte{0x1, 0x2, 0x3}

			err := client.Put(context.Background(), "test-key", testData)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				assert.Equal(testData, *tc.client.writer.result)
			}
		})
	}
}

func TestGCPCreateContainerOrContinue(t *testing.T) {
	someErr := errors.New("error")
	testCases := map[string]struct {
		client          *stubGCPStorageAPI
		createNewBucket bool
		wantErr         bool
	}{
		"success": {
			client: &stubGCPStorageAPI{},
		},
		"container does not exist": {
			client:          &stubGCPStorageAPI{attrsErr: storage.ErrBucketNotExist},
			createNewBucket: true,
		},
		"creating client fails": {
			client:  &stubGCPStorageAPI{newClientErr: someErr},
			wantErr: true,
		},
		"Attrs fails": {
			client:  &stubGCPStorageAPI{attrsErr: someErr},
			wantErr: true,
		},
		"CreateBucket fails": {
			client: &stubGCPStorageAPI{
				attrsErr:        storage.ErrBucketNotExist,
				createBucketErr: someErr,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &GoogleCloudStorage{
				newClient:  tc.client.stubClientFactory,
				projectID:  "test",
				bucketName: "test",
			}

			err := client.createContainerOrContinue(context.Background(), nil)
			if tc.wantErr {
				assert.Error(err)
			} else {
				assert.NoError(err)
				if tc.createNewBucket {
					assert.True(tc.client.createBucketCalled)
				}
			}
		})
	}
}
