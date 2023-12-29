/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package staticupload

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"))
}

func TestUpload(t *testing.T) {
	newInput := func() *s3.PutObjectInput {
		return &s3.PutObjectInput{
			Bucket: ptr("test-bucket"),
			Key:    ptr("test-key"),
		}
	}

	testCases := map[string]struct {
		in                           *s3.PutObjectInput
		cacheInvalidationStrategy    CacheInvalidationStrategy
		cacheInvalidationWaitTimeout time.Duration
		uploadFails                  bool
		invalidationFails            bool
		wantInvalidations            int
		wantCacheInvalidationErr     bool
		wantErr                      bool
		wantDirtyKeys                []string
		wantInvalidationIDs          []string
	}{
		"eager invalidation": {
			in:                           newInput(),
			cacheInvalidationStrategy:    CacheInvalidateEager,
			cacheInvalidationWaitTimeout: time.Microsecond,
			wantInvalidations:            1,
			wantInvalidationIDs:          []string{"test-invalidation-id-1"},
		},
		"lazy invalidation": {
			in:                           newInput(),
			cacheInvalidationStrategy:    CacheInvalidateBatchOnFlush,
			cacheInvalidationWaitTimeout: time.Microsecond,
			wantDirtyKeys:                []string{"test-key"},
		},
		"upload fails": {
			in:          newInput(),
			uploadFails: true,
			wantErr:     true,
		},
		"invalidation fails": {
			in:                       newInput(),
			invalidationFails:        true,
			wantCacheInvalidationErr: true,
		},
		"input is nil": {
			wantErr: true,
		},
		"key is nil": {
			in:      &s3.PutObjectInput{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cdnClient := &fakeCDNClient{}
			uploadClient := &stubUploadClient{}
			if tc.invalidationFails {
				cdnClient.createInvalidationErr = errors.New("invalidation failed")
			}
			if tc.uploadFails {
				uploadClient.uploadErr = errors.New("upload failed")
			}
			if tc.in != nil {
				tc.in.Body = bytes.NewReader([]byte("test-data"))
			}

			client := &Client{
				cdnClient:                    cdnClient,
				uploadClient:                 uploadClient,
				distributionID:               "test-distribution-id",
				cacheInvalidationStrategy:    tc.cacheInvalidationStrategy,
				cacheInvalidationWaitTimeout: tc.cacheInvalidationWaitTimeout,
        logger:                       slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
			}
			_, err := client.Upload(context.Background(), tc.in)

			var invalidationErr *InvalidationError
			if tc.wantCacheInvalidationErr {
				assert.ErrorAs(err, &invalidationErr)
				return
			}
			if tc.wantErr {
				assert.False(errors.As(err, &invalidationErr))
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.wantDirtyKeys, client.dirtyKeys)
			assert.Equal(tc.wantInvalidationIDs, client.invalidationIDs)
			assert.Equal("test-data", string(uploadClient.uploadedData))
			assert.Equal(tc.wantInvalidations, cdnClient.createInvalidationCounter)
		})
	}
}

func TestDeleteObject(t *testing.T) {
	newObjectInput := func(nilInput, nilKey bool) *s3.DeleteObjectInput {
		if nilInput {
			return nil
		}
		if nilKey {
			return &s3.DeleteObjectInput{}
		}
		return &s3.DeleteObjectInput{
			Bucket: ptr("test-bucket"),
			Key:    ptr("test-key"),
		}
	}
	newObjectsInput := func(nilInput, nilKey bool) *s3.DeleteObjectsInput {
		if nilInput {
			return nil
		}
		if nilKey {
			return &s3.DeleteObjectsInput{
				Delete: &s3types.Delete{
					Objects: []s3types.ObjectIdentifier{{}},
				},
			}
		}
		return &s3.DeleteObjectsInput{
			Bucket: ptr("test-bucket"),
			Delete: &s3types.Delete{
				Objects: []s3types.ObjectIdentifier{
					{Key: ptr("test-key")},
				},
			},
		}
	}

	testCases := map[string]struct {
		nilInput                     bool
		nilKey                       bool
		cacheInvalidationStrategy    CacheInvalidationStrategy
		cacheInvalidationWaitTimeout time.Duration
		deleteFails                  bool
		invalidationFails            bool
		wantInvalidations            int
		wantCacheInvalidationErr     bool
		wantErr                      bool
		wantDirtyKeys                []string
		wantInvalidationIDs          []string
	}{
		"eager invalidation": {
			cacheInvalidationStrategy:    CacheInvalidateEager,
			cacheInvalidationWaitTimeout: time.Microsecond,
			wantInvalidations:            1,
			wantInvalidationIDs:          []string{"test-invalidation-id-1"},
		},
		"lazy invalidation": {
			cacheInvalidationStrategy:    CacheInvalidateBatchOnFlush,
			cacheInvalidationWaitTimeout: time.Microsecond,
			wantDirtyKeys:                []string{"test-key"},
		},
		"delete fails": {
			deleteFails: true,
			wantErr:     true,
		},
		"invalidation fails": {
			invalidationFails:        true,
			wantCacheInvalidationErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cdnClient := &fakeCDNClient{}
			s3Client := &stubObjectStorageClient{}
			if tc.invalidationFails {
				cdnClient.createInvalidationErr = errors.New("invalidation failed")
			}
			if tc.deleteFails {
				s3Client.err = errors.New("delete failed")
			}

			client := &Client{
				cdnClient:                    cdnClient,
				s3Client:                     s3Client,
				distributionID:               "test-distribution-id",
				cacheInvalidationStrategy:    tc.cacheInvalidationStrategy,
				cacheInvalidationWaitTimeout: tc.cacheInvalidationWaitTimeout,
        logger:                       slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
			}
			_, err := client.DeleteObject(context.Background(), newObjectInput(tc.nilInput, tc.nilKey))

			var invalidationErr *InvalidationError
			if tc.wantCacheInvalidationErr {
				assert.ErrorAs(err, &invalidationErr)
				return
			}
			if tc.wantErr {
				assert.False(errors.As(err, &invalidationErr))
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.wantDirtyKeys, client.dirtyKeys)
			assert.Equal(tc.wantInvalidationIDs, client.invalidationIDs)
			assert.Equal(tc.wantInvalidations, cdnClient.createInvalidationCounter)
		})
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cdnClient := &fakeCDNClient{}
			s3Client := &stubObjectStorageClient{}
			if tc.invalidationFails {
				cdnClient.createInvalidationErr = errors.New("invalidation failed")
			}
			if tc.deleteFails {
				s3Client.err = errors.New("delete failed")
			}

			client := &Client{
				cdnClient:                    cdnClient,
				s3Client:                     s3Client,
				distributionID:               "test-distribution-id",
				cacheInvalidationStrategy:    tc.cacheInvalidationStrategy,
				cacheInvalidationWaitTimeout: tc.cacheInvalidationWaitTimeout,
        logger:                       slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
			}
			_, err := client.DeleteObjects(context.Background(), newObjectsInput(tc.nilInput, tc.nilKey))

			var invalidationErr *InvalidationError
			if tc.wantCacheInvalidationErr {
				assert.ErrorAs(err, &invalidationErr)
				return
			}
			if tc.wantErr {
				assert.False(errors.As(err, &invalidationErr))
				assert.Error(err)
				return
			}

			require.NoError(err)
			assert.Equal(tc.wantDirtyKeys, client.dirtyKeys)
			assert.Equal(tc.wantInvalidationIDs, client.invalidationIDs)
			assert.Equal(tc.wantInvalidations, cdnClient.createInvalidationCounter)
		})
	}
}

func TestFlush(t *testing.T) {
	testCases := map[string]struct {
		dirtyKeys                    []string
		invalidationIDs              []string
		cacheInvalidationWaitTimeout time.Duration
		invalidationFails            bool
		invalidationStatus           map[string]*string
		wantInvalidations            int
		wantDanglingInvalidationIDs  []string
		wantStatusChecks             map[string]int
		wantCacheInvalidationErr     bool
	}{
		"mixed invalidation": {
			dirtyKeys: []string{"test-key-1", "test-key-2"},
			invalidationIDs: []string{
				"test-invalidation-id-2",
				"test-invalidation-id-3",
			},
			cacheInvalidationWaitTimeout: time.Microsecond,
			invalidationStatus: map[string]*string{
				"test-invalidation-id-1": ptr("Completed"),
				"test-invalidation-id-2": ptr("Completed"),
				"test-invalidation-id-3": ptr("Completed"),
			},
			wantInvalidations: 1, // keys are batched
			wantStatusChecks: map[string]int{
				"test-invalidation-id-1": 1,
				"test-invalidation-id-2": 1,
				"test-invalidation-id-3": 1,
			},
		},
		"dirty key invalidation": {
			dirtyKeys:                    []string{"test-key-1", "test-key-2"},
			cacheInvalidationWaitTimeout: time.Microsecond,
			invalidationStatus: map[string]*string{
				"test-invalidation-id-1": ptr("Completed"),
			},
			wantInvalidations: 1, // keys are batched
			wantStatusChecks: map[string]int{
				"test-invalidation-id-1": 1,
			},
		},
		"not waiting for invalidation": {
			dirtyKeys: []string{"test-key-1", "test-key-2"},
			invalidationIDs: []string{
				"test-invalidation-id-2",
				"test-invalidation-id-3",
			},
			wantInvalidations: 1, // keys are batched
			wantDanglingInvalidationIDs: []string{
				"test-invalidation-id-2",
				"test-invalidation-id-3",
				"test-invalidation-id-1",
			},
		},
		"invalidation fails": {
			dirtyKeys:                []string{"test-key-1", "test-key-2"},
			invalidationFails:        true,
			wantCacheInvalidationErr: true,
		},
		"many keys": {
			dirtyKeys: func() []string {
				keys := make([]string, 3000)
				for i := range keys {
					keys[i] = fmt.Sprintf("test-key-%d", i)
				}
				return keys
			}(),
			wantInvalidations: 1, // keys are batched
			wantDanglingInvalidationIDs: []string{
				"test-invalidation-id-1",
			},
		},
		"too many keys": {
			dirtyKeys: func() []string {
				keys := make([]string, 3001)
				for i := range keys {
					keys[i] = fmt.Sprintf("test-key-%d", i)
				}
				return keys
			}(),
			wantCacheInvalidationErr: true,
		},
		"waiting for invalidation times out": {
			dirtyKeys: []string{"test-key-1"},
			invalidationIDs: []string{
				"test-invalidation-id-2",
				"test-invalidation-id-3",
			},
			cacheInvalidationWaitTimeout: time.Microsecond,
			wantCacheInvalidationErr:     true,
		},
		"no invalidations": {
			dirtyKeys: []string{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			cdnClient := &fakeCDNClient{
				status: tc.invalidationStatus,
			}
			uploadClient := &stubUploadClient{
				uploadErr: errors.New("Upload should not be called"),
			}
			if tc.invalidationFails {
				cdnClient.createInvalidationErr = errors.New("invalidation failed")
			}

			client := &Client{
				cdnClient:                    cdnClient,
				uploadClient:                 uploadClient,
				distributionID:               "test-distribution-id",
				cacheInvalidationWaitTimeout: tc.cacheInvalidationWaitTimeout,
				dirtyKeys:                    tc.dirtyKeys,
				invalidationIDs:              tc.invalidationIDs,
        logger:                       slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
			}
			err := client.Flush(context.Background())

			if tc.wantCacheInvalidationErr {
				var invalidationErr *InvalidationError
				assert.ErrorAs(err, &invalidationErr)
				return
			}

			require.NoError(err)
			assert.Empty(client.dirtyKeys)
			assert.Equal(tc.wantDanglingInvalidationIDs, client.invalidationIDs)
			assert.Equal(tc.wantInvalidations, cdnClient.createInvalidationCounter)
			assert.Equal(tc.wantStatusChecks, cdnClient.statusCheckCounter)
		})
	}
}

func TestConcurrency(t *testing.T) {
	newInput := func() *s3.PutObjectInput {
		return &s3.PutObjectInput{
			Bucket: ptr("test-bucket"),
			Key:    ptr("test-key"),
			Body:   bytes.NewReader([]byte("test-data")),
		}
	}

	cdnClient := &fakeCDNClient{}
	s3Client := &stubObjectStorageClient{}
	uploadClient := &stubUploadClient{}

	client := &Client{
		cdnClient:                    cdnClient,
		s3Client:                     s3Client,
		uploadClient:                 uploadClient,
		distributionID:               "test-distribution-id",
		cacheInvalidationWaitTimeout: 50 * time.Millisecond,
    logger:                       slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
	}

	var wg sync.WaitGroup

	upload := func() {
		defer wg.Done()
		_, _ = client.Upload(context.Background(), newInput())
	}
	deleteObject := func() {
		defer wg.Done()
		_, _ = client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: ptr("test-bucket"),
			Key:    ptr("test-key"),
		})
	}
	deleteObjects := func() {
		defer wg.Done()
		_, _ = client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
			Bucket: ptr("test-bucket"),
			Delete: &s3types.Delete{
				Objects: []s3types.ObjectIdentifier{
					{Key: ptr("test-key")},
				},
			},
		})
	}
	flushClient := func() {
		defer wg.Done()
		_ = client.Flush(context.Background())
	}

	for i := 0; i < 100; i++ {
		wg.Add(4)
		go upload()
		go deleteObject()
		go deleteObjects()
		go flushClient()
	}

	wg.Wait()
}

type fakeCDNClient struct {
	mux                       sync.Mutex
	createInvalidationCounter int
	statusCheckCounter        map[string]int
	status                    map[string]*string

	createInvalidationErr error
	getInvalidationErr    error
}

func (c *fakeCDNClient) CreateInvalidation(
	_ context.Context, _ *cloudfront.CreateInvalidationInput, _ ...func(*cloudfront.Options),
) (*cloudfront.CreateInvalidationOutput, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.createInvalidationCounter++
	ctr := c.createInvalidationCounter
	return &cloudfront.CreateInvalidationOutput{
		Invalidation: &cftypes.Invalidation{
			Id: ptr(fmt.Sprintf("test-invalidation-id-%d", ctr)),
		},
	}, c.createInvalidationErr
}

func (c *fakeCDNClient) GetInvalidation(
	_ context.Context, input *cloudfront.GetInvalidationInput, _ ...func(*cloudfront.Options),
) (*cloudfront.GetInvalidationOutput, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.statusCheckCounter == nil {
		c.statusCheckCounter = make(map[string]int)
	}
	c.statusCheckCounter[*input.Id]++
	status := "Unknown"
	if s, ok := c.status[*input.Id]; ok {
		status = *s
	}

	return &cloudfront.GetInvalidationOutput{
		Invalidation: &cftypes.Invalidation{
			CreateTime: ptr(time.Now()),
			Id:         input.Id,
			Status:     ptr(status),
		},
	}, c.getInvalidationErr
}

type stubUploadClient struct {
	mux          sync.Mutex
	uploadErr    error
	uploadedData []byte
}

func (s *stubUploadClient) Upload(
	_ context.Context, input *s3.PutObjectInput,
	_ ...func(*s3manager.Uploader),
) (*s3manager.UploadOutput, error) {
	var err error
	s.mux.Lock()
	defer s.mux.Unlock()
	s.uploadedData, err = io.ReadAll(input.Body)
	if err != nil {
		panic(err)
	}
	return nil, s.uploadErr
}

type stubObjectStorageClient struct {
	deleteObjectOut  *s3.DeleteObjectOutput
	deleteObjectsOut *s3.DeleteObjectsOutput
	err              error
}

func (s *stubObjectStorageClient) DeleteObject(_ context.Context, _ *s3.DeleteObjectInput,
	_ ...func(*s3.Options),
) (*s3.DeleteObjectOutput, error) {
	return s.deleteObjectOut, s.err
}

func (s *stubObjectStorageClient) DeleteObjects(
	_ context.Context, _ *s3.DeleteObjectsInput,
	_ ...func(*s3.Options),
) (*s3.DeleteObjectsOutput, error) {
	return s.deleteObjectsOut, s.err
}

// currently not needed so no-Op.
func (s *stubObjectStorageClient) GetObject(
	_ context.Context, _ *s3.GetObjectInput,
	_ ...func(*s3.Options),
) (*s3.GetObjectOutput, error) {
	return nil, nil
}

// currently not needed so no-Op.
func (s *stubObjectStorageClient) ListObjectsV2(
	_ context.Context, _ *s3.ListObjectsV2Input,
	_ ...func(*s3.Options),
) (*s3.ListObjectsV2Output, error) {
	return nil, nil
}
