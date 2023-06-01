/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package staticupload provides a static file uploader/updater/remover for the CDN / static API.

This uploader uses AWS S3 as a backend and cloudfront as a CDN.
It understands how to upload files and invalidate the CDN cache accordingly.
*/
package staticupload

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/google/uuid"
)

// Client is a static file uploader/updater/remover for the CDN / static API.
// It has the same interface as the S3 uploader.
type Client struct {
	mux            sync.Mutex
	cdnClient      cdnClient
	uploadClient   uploadClient
	s3Client       objectStorageClient
	distributionID string
	BucketID       string

	cacheInvalidationStrategy    CacheInvalidationStrategy
	cacheInvalidationWaitTimeout time.Duration
	// dirtyKeys is a list of keys that still needs to be invalidated by us.
	dirtyKeys []string
	// invalidationIDs is a list of invalidation IDs that are currently in progress.
	invalidationIDs []string
}

// Config is the configuration for the Client.
type Config struct {
	// Region is the AWS region to use.
	Region string
	// Bucket is the name of the S3 bucket to use.
	Bucket string
	// DistributionID is the ID of the CloudFront distribution to use.
	DistributionID            string
	CacheInvalidationStrategy CacheInvalidationStrategy
	// CacheInvalidationWaitTimeout is the timeout to wait for the CDN cache to invalidate.
	// set to 0 to disable waiting for the CDN cache to invalidate.
	CacheInvalidationWaitTimeout time.Duration
}

// SetsDefault checks if all necessary values are set and sets default values otherwise.
func (c *Config) SetsDefault() {
	if c.DistributionID == "" {
		c.DistributionID = constants.CDNDefaultDistributionID
	}
}

// CacheInvalidationStrategy is the strategy to use for invalidating the CDN cache.
type CacheInvalidationStrategy int

const (
	// CacheInvalidateEager invalidates the CDN cache immediately for every key that is uploaded.
	CacheInvalidateEager CacheInvalidationStrategy = iota
	// CacheInvalidateBatchOnClose invalidates the CDN cache in batches when the client is closed.
	// This is useful when uploading many files at once but will fail if Close is not called.
	CacheInvalidateBatchOnClose
)

// InvalidationError is an error that occurs when invalidating the CDN cache.
type InvalidationError struct {
	inner error
}

// Error returns the error message.
func (e InvalidationError) Error() string {
	return fmt.Sprintf("invalidating CDN cache: %v", e.inner)
}

// Unwrap returns the inner error.
func (e InvalidationError) Unwrap() error {
	return e.inner
}

// New creates a new Client.
func New(ctx context.Context, config Config) (*Client, error) {
	config.SetsDefault()
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(config.Region))
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg)
	uploadClient := s3manager.NewUploader(s3Client)

	cdnClient := cloudfront.NewFromConfig(cfg)

	return &Client{
		cdnClient:                    cdnClient,
		s3Client:                     s3Client,
		uploadClient:                 uploadClient,
		distributionID:               config.DistributionID,
		cacheInvalidationStrategy:    config.CacheInvalidationStrategy,
		cacheInvalidationWaitTimeout: config.CacheInvalidationWaitTimeout,
		BucketID:                     config.Bucket,
	}, nil
}

// Close closes the client.
// It waits for all invalidations to finish.
// It returns nil on success or an error.
// The error will be of type InvalidationError if the CDN cache could not be invalidated.
func (c *Client) Close(ctx context.Context) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	// invalidate all dirty keys that have not been invalidated yet
	invalidationID, err := c.invalidateCacheForKeys(ctx, c.dirtyKeys)
	if err != nil {
		return err
	}
	c.invalidationIDs = append(c.invalidationIDs, invalidationID)
	c.dirtyKeys = nil

	return c.waitForInvalidations(ctx)
}

// invalidate invalidates the CDN cache for the given keys.
// It either performs the invalidation immediately or adds them to the list of dirty keys.
func (c *Client) invalidate(ctx context.Context, keys []string) error {
	if c.cacheInvalidationStrategy == CacheInvalidateBatchOnClose {
		// save as dirty key for batch invalidation on Close
		c.mux.Lock()
		defer c.mux.Unlock()
		c.dirtyKeys = append(c.dirtyKeys, keys...)
		return nil
	}
	// eagerly invalidate the CDN cache
	invalidationID, err := c.invalidateCacheForKeys(ctx, keys)
	if err != nil {
		return err
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.invalidationIDs = append(c.invalidationIDs, invalidationID)
	return nil
}

// invalidateCacheForKeys invalidates the CDN cache for the given list of keys.
// It returns the invalidation ID without waiting for the invalidation to finish.
// The list of keys must not be longer than 3000 as specified by AWS:
// https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/Invalidation.html#InvalidationLimits
func (c *Client) invalidateCacheForKeys(ctx context.Context, keys []string) (string, error) {
	if len(keys) > 3000 {
		return "", InvalidationError{inner: fmt.Errorf("too many keys to invalidate: %d", len(keys))}
	}

	for i, key := range keys {
		if !strings.HasPrefix(key, "/") {
			keys[i] = "/" + key
		}
	}

	in := &cloudfront.CreateInvalidationInput{
		DistributionId: &c.distributionID,
		InvalidationBatch: &cftypes.InvalidationBatch{
			CallerReference: ptr(uuid.New().String()),
			Paths: &cftypes.Paths{
				Items:    keys,
				Quantity: ptr(int32(len(keys))),
			},
		},
	}
	invalidation, err := c.cdnClient.CreateInvalidation(ctx, in)
	if err != nil {
		return "", InvalidationError{inner: fmt.Errorf("creating invalidation: %w", err)}
	}
	if invalidation.Invalidation == nil || invalidation.Invalidation.Id == nil {
		return "", InvalidationError{inner: fmt.Errorf("invalidation ID is not set")}
	}
	return *invalidation.Invalidation.Id, nil
}

// waitForInvalidations waits for all invalidations to finish.
func (c *Client) waitForInvalidations(ctx context.Context) error {
	if c.cacheInvalidationWaitTimeout == 0 {
		return nil
	}

	waiter := cloudfront.NewInvalidationCompletedWaiter(c.cdnClient)
	for _, invalidationID := range c.invalidationIDs {
		waitIn := &cloudfront.GetInvalidationInput{
			DistributionId: &c.distributionID,
			Id:             &invalidationID,
		}
		if err := waiter.Wait(ctx, waitIn, c.cacheInvalidationWaitTimeout); err != nil {
			return InvalidationError{inner: fmt.Errorf("waiting for invalidation to complete: %w", err)}
		}
	}
	c.invalidationIDs = nil
	return nil
}

type uploadClient interface {
	Upload(
		ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader),
	) (*s3manager.UploadOutput, error)
}

type getClient interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type deleteClient interface {
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.DeleteObjectOutput, error)
	DeleteObjects(
		ctx context.Context, params *s3.DeleteObjectsInput,
		optFns ...func(*s3.Options),
	) (*s3.DeleteObjectsOutput, error)
}

type cdnClient interface {
	CreateInvalidation(
		ctx context.Context, params *cloudfront.CreateInvalidationInput, optFns ...func(*cloudfront.Options),
	) (*cloudfront.CreateInvalidationOutput, error)
	GetInvalidation(
		context.Context, *cloudfront.GetInvalidationInput, ...func(*cloudfront.Options),
	) (*cloudfront.GetInvalidationOutput, error)
}

type objectStorageClient interface {
	deleteClient
	getClient
}

// statically assert that Client implements the uploadClient interface.
var _ uploadClient = (*Client)(nil)

// statically assert that Client implements the deleteClient interface.
var _ objectStorageClient = (*Client)(nil)

func ptr[T any](t T) *T {
	return &t
}
