/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package client provides a client for the versions API.

The client needs to be authenticated with AWS. It should be used in internal
development and CI tools for administrative tasks. For just fetching information
from the API, use the fetcher package instead.

Needed IAM permissions for read mode:
- "s3:GetObject"
- "s3:ListBucket"

Additional needed IAM permissions for write mode:
- "s3:PutObject"
- "s3:DeleteObject"
- "cloudfront:CreateInvalidation"

Thread-safety of the bucket is not guaranteed. The client is not thread-safe.

Each sub-API included in the Constellation Resource API should define it's resources by implementing types that implement apiObject.
The new package can then call this package's Fetch and Update functions to get/modify resources from the API.
*/
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"go.uber.org/zap"
)

// Client is the a general client for all APIs.
type Client struct {
	s3Client
	s3ClientClose                func(ctx context.Context) error
	bucket                       string
	cacheInvalidationWaitTimeout time.Duration

	dirtyPaths []string // written paths to be invalidated
	DryRun     bool     // no write operations are performed

	Log *logger.Logger
}

// NewReadOnlyClient creates a new read-only client.
// This client can be used to fetch objects but cannot write updates.
func NewReadOnlyClient(ctx context.Context, region, bucket, distributionID string,
	log *logger.Logger,
) (*Client, CloseFunc, error) {
	staticUploadClient, staticUploadClientClose, err := staticupload.New(ctx, staticupload.Config{
		Region:                    region,
		Bucket:                    bucket,
		DistributionID:            distributionID,
		CacheInvalidationStrategy: staticupload.CacheInvalidateBatchOnFlush,
	})
	if err != nil {
		return nil, nil, err
	}

	client := &Client{
		s3Client:      staticUploadClient,
		s3ClientClose: staticUploadClientClose,
		bucket:        bucket,
		DryRun:        true,
		Log:           log,
	}
	clientClose := func(ctx context.Context) error {
		return client.Close(ctx)
	}

	return client, clientClose, nil
}

// NewClient creates a new client for the versions API.
func NewClient(ctx context.Context, region, bucket, distributionID string, dryRun bool,
	log *logger.Logger,
) (*Client, CloseFunc, error) {
	staticUploadClient, staticUploadClientClose, err := staticupload.New(ctx, staticupload.Config{
		Region:                    region,
		Bucket:                    bucket,
		DistributionID:            distributionID,
		CacheInvalidationStrategy: staticupload.CacheInvalidateBatchOnFlush,
	})
	if err != nil {
		return nil, nil, err
	}

	client := &Client{
		s3Client:                     staticUploadClient,
		s3ClientClose:                staticUploadClientClose,
		bucket:                       bucket,
		DryRun:                       dryRun,
		Log:                          log,
		cacheInvalidationWaitTimeout: 10 * time.Minute,
	}
	clientClose := func(ctx context.Context) error {
		return client.Close(ctx)
	}

	return client, clientClose, nil
}

// Close closes the client.
// It invalidates the CDN cache for all uploaded files.
func (c *Client) Close(ctx context.Context) error {
	if c.s3ClientClose == nil {
		return nil
	}
	return c.s3ClientClose(ctx)
}

// DeletePath deletes all objects at a given path from a s3 bucket.
func (c *Client) DeletePath(ctx context.Context, path string) error {
	listIn := &s3.ListObjectsV2Input{
		Bucket: &c.bucket,
		Prefix: &path,
	}
	c.Log.Debugf("Listing objects in %s", path)
	objs := []s3types.Object{}
	out := &s3.ListObjectsV2Output{IsTruncated: true}
	for out.IsTruncated {
		var err error
		out, err = c.s3Client.ListObjectsV2(ctx, listIn)
		if err != nil {
			return fmt.Errorf("listing objects in %s: %w", path, err)
		}
		objs = append(objs, out.Contents...)
	}
	c.Log.Debugf("Found %d objects in %s", len(objs), path)

	if len(objs) == 0 {
		c.Log.Warnf("Path %s is already empty", path)
		return nil
	}

	objIDs := make([]s3types.ObjectIdentifier, len(objs))
	for i, obj := range objs {
		objIDs[i] = s3types.ObjectIdentifier{Key: obj.Key}
	}

	if c.DryRun {
		c.Log.Debugf("DryRun: Deleting %d objects with IDs %v", len(objs), objIDs)
		return nil
	}

	c.dirtyPaths = append(c.dirtyPaths, "/"+path)

	deleteIn := &s3.DeleteObjectsInput{
		Bucket: &c.bucket,
		Delete: &s3types.Delete{
			Objects: objIDs,
		},
	}
	c.Log.Debugf("Deleting %d objects in %s", len(objs), path)
	if _, err := c.s3Client.DeleteObjects(ctx, deleteIn); err != nil {
		return fmt.Errorf("deleting objects in %s: %w", path, err)
	}

	return nil
}

func ptr[T any](t T) *T {
	return &t
}

// APIObject is an object that is used to perform CRUD operations on the API.
type APIObject interface {
	ValidateRequest() error
	Validate() error
	JSONPath() string
}

// Fetch fetches the given apiObject from the public Constellation CDN.
func Fetch[T APIObject](ctx context.Context, c *Client, obj T) (T, error) {
	if err := obj.ValidateRequest(); err != nil {
		return *new(T), fmt.Errorf("validating request for %T: %w", obj, err)
	}

	in := &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    ptr(obj.JSONPath()),
	}

	c.Log.Debugf("Fetching %T from s3: %s", obj, obj.JSONPath())
	out, err := c.s3Client.GetObject(ctx, in)
	var noSuchkey *s3types.NoSuchKey
	if errors.As(err, &noSuchkey) {
		return *new(T), &NotFoundError{err: err}
	} else if err != nil {
		return *new(T), fmt.Errorf("getting s3 object at %s: %w", obj.JSONPath(), err)
	}
	defer out.Body.Close()

	var newObj T
	if err := json.NewDecoder(out.Body).Decode(&newObj); err != nil {
		return *new(T), fmt.Errorf("decoding %T: %w", obj, err)
	}

	if newObj.Validate() != nil {
		return *new(T), fmt.Errorf("received invalid %T: %w", newObj, newObj.Validate())
	}

	return newObj, nil
}

// Update creates/updates the given apiObject in the public Constellation API.
func Update(ctx context.Context, c *Client, obj APIObject) error {
	if err := obj.Validate(); err != nil {
		return fmt.Errorf("validating %T struct: %w", obj, err)
	}

	rawJSON, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshaling %T struct: %w", obj, err)
	}

	if c.DryRun {
		c.Log.With(zap.String("bucket", c.bucket), zap.String("key", obj.JSONPath()), zap.String("body", string(rawJSON))).Debugf("DryRun: s3 put object")
		return nil
	}

	in := &s3.PutObjectInput{
		Bucket: &c.bucket,
		Key:    ptr(obj.JSONPath()),
		Body:   bytes.NewBuffer(rawJSON),
	}

	c.dirtyPaths = append(c.dirtyPaths, "/"+obj.JSONPath())

	c.Log.Debugf("Uploading %T to s3: %v", obj, obj.JSONPath())
	if _, err := c.Upload(ctx, in); err != nil {
		return fmt.Errorf("uploading %T: %w", obj, err)
	}

	return nil
}

// Delete deletes the given apiObject from the public Constellation API.
func Delete(ctx context.Context, c *Client, obj APIObject) error {
	if err := obj.ValidateRequest(); err != nil {
		return fmt.Errorf("validating request for %T: %w", obj, err)
	}

	in := &s3.DeleteObjectInput{
		Bucket: &c.bucket,
		Key:    ptr(obj.JSONPath()),
	}

	c.Log.Debugf("Deleting %T from s3: %s", obj, obj.JSONPath())
	if _, err := c.DeleteObject(ctx, in); err != nil {
		return fmt.Errorf("deleting s3 object at %s: %w", obj.JSONPath(), err)
	}

	return nil
}

// NotFoundError is an error that is returned when a resource is not found.
type NotFoundError struct {
	err error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("the requested resource was not found: %s", e.err.Error())
}

func (e *NotFoundError) Unwrap() error {
	return e.err
}

type s3Client interface {
	GetObject(
		ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options),
	) (*s3.GetObjectOutput, error)
	ListObjectsV2(
		ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options),
	) (*s3.ListObjectsV2Output, error)
	DeleteObjects(
		ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options),
	) (*s3.DeleteObjectsOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput,
		optFns ...func(*s3.Options),
	) (*s3.DeleteObjectOutput, error)
	uploadClient
}

type uploadClient interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// CloseFunc is a function that closes the client.
type CloseFunc func(ctx context.Context) error
