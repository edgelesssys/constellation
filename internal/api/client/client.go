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
	"log/slog"
	"strings"
	"time"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/sigstore"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
)

// Client is the a general client for all APIs.
type Client struct {
	s3Client
	s3ClientClose func(ctx context.Context) error
	bucket        string

	dirtyPaths []string // written paths to be invalidated
	DryRun     bool     // no write operations are performed

	log *slog.Logger
}

// NewReadOnlyClient creates a new read-only client.
// This client can be used to fetch objects but cannot write updates.
func NewReadOnlyClient(ctx context.Context, region, bucket, distributionID string,
	log *slog.Logger,
) (*Client, CloseFunc, error) {
	staticUploadClient, staticUploadClientClose, err := staticupload.New(ctx, staticupload.Config{
		Region:                       region,
		Bucket:                       bucket,
		DistributionID:               distributionID,
		CacheInvalidationStrategy:    staticupload.CacheInvalidateBatchOnFlush,
		CacheInvalidationWaitTimeout: 10 * time.Minute,
	}, log)
	if err != nil {
		return nil, nil, err
	}

	client := &Client{
		s3Client:      staticUploadClient,
		s3ClientClose: staticUploadClientClose,
		bucket:        bucket,
		DryRun:        true,
		log:           log,
	}
	clientClose := func(ctx context.Context) error {
		return client.Close(ctx)
	}

	return client, clientClose, nil
}

// NewClient creates a new client for the versions API.
func NewClient(ctx context.Context, region, bucket, distributionID string, dryRun bool,
	log *slog.Logger,
) (*Client, CloseFunc, error) {
	staticUploadClient, staticUploadClientClose, err := staticupload.New(ctx, staticupload.Config{
		Region:                       region,
		Bucket:                       bucket,
		DistributionID:               distributionID,
		CacheInvalidationStrategy:    staticupload.CacheInvalidateBatchOnFlush,
		CacheInvalidationWaitTimeout: 10 * time.Minute,
	}, log)
	if err != nil {
		return nil, nil, err
	}

	client := &Client{
		s3Client:      staticUploadClient,
		s3ClientClose: staticUploadClientClose,
		bucket:        bucket,
		DryRun:        dryRun,
		log:           log,
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
		c.log.Debug("Client has no s3ClientClose")
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
	c.log.Debug(fmt.Sprintf("Listing objects in %q", path))
	objs := []s3types.Object{}
	out := &s3.ListObjectsV2Output{IsTruncated: ptr(true)}
	for out.IsTruncated != nil && *out.IsTruncated {
		var err error
		out, err = c.s3Client.ListObjectsV2(ctx, listIn)
		if err != nil {
			return fmt.Errorf("listing objects in %s: %w", path, err)
		}
		objs = append(objs, out.Contents...)
	}
	c.log.Debug(fmt.Sprintf("Found %d objects in %q", len(objs), path))

	if len(objs) == 0 {
		c.log.Warn(fmt.Sprintf("Path %s is already empty", path))
		return nil
	}

	objIDs := make([]s3types.ObjectIdentifier, len(objs))
	for i, obj := range objs {
		objIDs[i] = s3types.ObjectIdentifier{Key: obj.Key}
	}

	if c.DryRun {
		c.log.Debug(fmt.Sprintf("DryRun: Deleting %d objects with IDs %v", len(objs), objIDs))
		return nil
	}

	c.dirtyPaths = append(c.dirtyPaths, "/"+path)

	deleteIn := &s3.DeleteObjectsInput{
		Bucket: &c.bucket,
		Delete: &s3types.Delete{
			Objects: objIDs,
		},
	}
	c.log.Debug(fmt.Sprintf("Deleting %d objects in %q", len(objs), path))
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

// Fetch fetches the given apiObject from the public Constellation API.
func Fetch[T APIObject](ctx context.Context, c *Client, obj T) (T, error) {
	if err := obj.ValidateRequest(); err != nil {
		return *new(T), fmt.Errorf("validating request for %T: %w", obj, err)
	}

	in := &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    ptr(obj.JSONPath()),
	}

	c.log.Debug(fmt.Sprintf("Fetching %T from s3: %q", obj, obj.JSONPath()))
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
		c.log.With(slog.String("bucket", c.bucket), slog.String("key", obj.JSONPath()), slog.String("body", string(rawJSON))).Debug("DryRun: s3 put object")
		return nil
	}

	in := &s3.PutObjectInput{
		Bucket: &c.bucket,
		Key:    ptr(obj.JSONPath()),
		Body:   bytes.NewBuffer(rawJSON),
	}

	c.dirtyPaths = append(c.dirtyPaths, "/"+obj.JSONPath())

	c.log.Debug(fmt.Sprintf("Uploading %T to s3: %q", obj, obj.JSONPath()))
	if _, err := c.Upload(ctx, in); err != nil {
		return fmt.Errorf("uploading %T: %w", obj, err)
	}

	return nil
}

// SignAndUpdate signs the given apiObject and updates the object and it's signature in the API.
// This function should be used in favor of manually managing signatures.
// The signing is specified as part of the signer argument.
func SignAndUpdate(ctx context.Context, c *Client, obj APIObject, signer sigstore.Signer) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshaling %T: %w", obj, err)
	}

	dataSignature, err := signer.Sign(data)
	if err != nil {
		return fmt.Errorf("sign version file: %w", err)
	}

	signature := signature{
		Signed:    obj.JSONPath(),
		Signature: dataSignature,
	}

	if err := Update(ctx, c, obj); err != nil {
		return fmt.Errorf("updating %T: %w", obj, err)
	}

	if err := Update(ctx, c, signature); err != nil {
		return fmt.Errorf("updating %T: %w", obj, err)
	}

	return nil
}

// DeleteWithSignature deletes the given apiObject and it's signature from the public Constellation API.
func DeleteWithSignature(ctx context.Context, c *Client, obj APIObject) error {
	if err := Delete(ctx, c, obj); err != nil {
		return fmt.Errorf("deleting %T: %w", obj, err)
	}

	sig := signature{Signed: obj.JSONPath()}
	if err := Delete(ctx, c, sig); err != nil {
		return fmt.Errorf("deleting %T: %w", sig, err)
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

	c.log.Debug(fmt.Sprintf("Deleting %T from s3: %q", obj, obj.JSONPath()))
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

// signature manages the signature of a object saved at location 'Signed'.
type signature struct {
	// Signed is the object that is signed.
	Signed string `json:"signed"`
	// Signature is the signature of `Signed`.
	Signature []byte `json:"signature"`
}

// JSONPath returns the path to the JSON file for the request to the config api.
func (s signature) JSONPath() string {
	return s.Signed + ".sig"
}

// ValidateRequest validates the request.
func (s signature) ValidateRequest() error {
	if !strings.HasSuffix(s.Signed, ".json") {
		return errors.New("signed object missing .json suffix")
	}

	return nil
}

// Validate checks that the signature is base64 encoded.
func (s signature) Validate() error {
	return sigstore.IsBase64(s.Signature)
}
