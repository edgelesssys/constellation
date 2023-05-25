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
*/
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cftypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/mod/semver"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
)

// Client is the client for the versions API.
type Client struct {
	config                       aws.Config
	cloudfrontClient             *cloudfront.Client
	s3Client                     *s3.Client
	uploadClient                 *s3manager.Uploader
	bucket                       string
	distributionID               string
	cacheInvalidationWaitTimeout time.Duration

	dirtyPaths []string // written paths to be invalidated
	dryRun     bool     // no write operations are performed

	log *logger.Logger
}

// NewReadOnlyClient creates a new read-only client.
// This client can be used to fetch objects but cannot write updates.
func NewReadOnlyClient(ctx context.Context, region, bucket, distributionID string,
	log *logger.Logger,
) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	s3c := s3.NewFromConfig(cfg)

	return &Client{
		config:         cfg,
		s3Client:       s3c,
		bucket:         bucket,
		distributionID: distributionID,
		dryRun:         true,
		log:            log,
	}, nil
}

// NewClient creates a new client for the versions API.
func NewClient(ctx context.Context, region, bucket, distributionID string, dryRun bool,
	log *logger.Logger,
) (*Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}

	cloudfrontC := cloudfront.NewFromConfig(cfg)
	s3C := s3.NewFromConfig(cfg)
	uploadC := s3manager.NewUploader(s3C)

	return &Client{
		config:                       cfg,
		cloudfrontClient:             cloudfrontC,
		s3Client:                     s3C,
		uploadClient:                 uploadC,
		bucket:                       bucket,
		distributionID:               distributionID,
		dryRun:                       dryRun,
		log:                          log,
		cacheInvalidationWaitTimeout: 10 * time.Minute,
	}, nil
}

// FetchVersionList fetches the given version list from the versions API.
func (c *Client) FetchVersionList(ctx context.Context, list versionsapi.List) (versionsapi.List, error) {
	return fetch(ctx, c, list)
}

// UpdateVersionList updates the given version list in the versions API.
func (c *Client) UpdateVersionList(ctx context.Context, list versionsapi.List) error {
	semver.Sort(list.Versions)
	return update(ctx, c, list)
}

// FetchVersionLatest fetches the latest version from the versions API.
func (c *Client) FetchVersionLatest(ctx context.Context, latest versionsapi.Latest) (versionsapi.Latest, error) {
	return fetch(ctx, c, latest)
}

// UpdateVersionLatest updates the latest version in the versions API.
func (c *Client) UpdateVersionLatest(ctx context.Context, latest versionsapi.Latest) error {
	return update(ctx, c, latest)
}

// FetchImageInfo fetches the given image info from the versions API.
func (c *Client) FetchImageInfo(ctx context.Context, imageInfo versionsapi.ImageInfo) (versionsapi.ImageInfo, error) {
	return fetch(ctx, c, imageInfo)
}

// UpdateImageInfo updates the given image info in the versions API.
func (c *Client) UpdateImageInfo(ctx context.Context, imageInfo versionsapi.ImageInfo) error {
	return update(ctx, c, imageInfo)
}

// FetchCLIInfo fetches the given CLI info from the versions API.
func (c *Client) FetchCLIInfo(ctx context.Context, cliInfo versionsapi.CLIInfo) (versionsapi.CLIInfo, error) {
	return fetch(ctx, c, cliInfo)
}

// UpdateCLIInfo updates the given CLI info in the versions API.
func (c *Client) UpdateCLIInfo(ctx context.Context, cliInfo versionsapi.CLIInfo) error {
	return update(ctx, c, cliInfo)
}

// DeleteRef deletes the given ref from the versions API.
func (c *Client) DeleteRef(ctx context.Context, ref string) error {
	if err := versionsapi.ValidateRef(ref); err != nil {
		return fmt.Errorf("validating ref: %w", err)
	}

	refPath := path.Join(constants.CDNAPIPrefix, "ref", ref)
	if err := c.deletePath(ctx, refPath); err != nil {
		return fmt.Errorf("deleting ref path: %w", err)
	}

	return nil
}

// DeleteVersion deletes the given version from the versions API.
// The version will be removed from version lists and latest versions, and the versioned
// objects are deleted.
// Notice that the versions API can get into an inconsistent state if the version is the latest
// version but there is no older version of the same minor version available.
// Manual update of latest versions is required in this case.
func (c *Client) DeleteVersion(ctx context.Context, ver versionsapi.Version) error {
	var retErr error

	c.log.Debugf("Deleting version %s from minor version list", ver.Version)
	possibleNewLatest, err := c.deleteVersionFromMinorVersionList(ctx, ver)
	if err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("removing from minor version list: %w", err))
	}

	c.log.Debugf("Checking latest version for %s", ver.Version)
	if err := c.deleteVersionFromLatest(ctx, ver, possibleNewLatest); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("updating latest version: %w", err))
	}

	c.log.Debugf("Deleting artifact path %s for %s", ver.ArtifactPath(versionsapi.APIV1), ver.Version)
	if err := c.deletePath(ctx, ver.ArtifactPath(versionsapi.APIV1)); err != nil {
		retErr = errors.Join(retErr, fmt.Errorf("deleting artifact path: %w", err))
	}

	return retErr
}

// InvalidateCache invalidates the CDN cache for the paths that have been written.
// The function should be deferred after the client has been created.
func (c *Client) InvalidateCache(ctx context.Context) error {
	if len(c.dirtyPaths) == 0 {
		c.log.Debugf("No dirty paths, skipping cache invalidation")
		return nil
	}

	if c.dryRun {
		c.log.Debugf("DryRun: cloudfront create invalidation {DistributionID: %v, Paths: %v}", c.distributionID, c.dirtyPaths)
		return nil
	}

	c.log.Debugf("Paths to invalidate: %v", c.dirtyPaths)

	in := &cloudfront.CreateInvalidationInput{
		DistributionId: &c.distributionID,
		InvalidationBatch: &cftypes.InvalidationBatch{
			CallerReference: ptr(fmt.Sprintf("%d", time.Now().Unix())),
			Paths: &cftypes.Paths{
				Items:    c.dirtyPaths,
				Quantity: ptr(int32(len(c.dirtyPaths))),
			},
		},
	}
	invalidation, err := c.cloudfrontClient.CreateInvalidation(ctx, in)
	if err != nil {
		return fmt.Errorf("creating invalidation: %w", err)
	}

	c.log.Debugf("Waiting for invalidation %s to complete", *invalidation.Invalidation.Id)
	waiter := cloudfront.NewInvalidationCompletedWaiter(c.cloudfrontClient)
	waitIn := &cloudfront.GetInvalidationInput{
		DistributionId: &c.distributionID,
		Id:             invalidation.Invalidation.Id,
	}
	if err := waiter.Wait(ctx, waitIn, c.cacheInvalidationWaitTimeout); err != nil {
		return fmt.Errorf("waiting for invalidation to complete: %w", err)
	}

	return nil
}

type apiObject interface {
	ValidateRequest() error
	Validate() error
	JSONPath() string
}

func fetch[T apiObject](ctx context.Context, c *Client, obj T) (T, error) {
	if err := obj.ValidateRequest(); err != nil {
		return *new(T), fmt.Errorf("validating request for %T: %w", obj, err)
	}

	in := &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    ptr(obj.JSONPath()),
	}

	c.log.Debugf("Fetching %T from s3: %s", obj, obj.JSONPath())
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

func update[T apiObject](ctx context.Context, c *Client, obj T) error {
	if err := obj.Validate(); err != nil {
		return fmt.Errorf("validating %T struct: %w", obj, err)
	}

	rawJSON, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshaling %T struct: %w", obj, err)
	}

	if c.dryRun {
		c.log.Debugf("DryRun: s3 put object {Bucket: %v, Key: %v, Body: %v", c.bucket, obj.JSONPath(), string(rawJSON))
		return nil
	}

	in := &s3.PutObjectInput{
		Bucket: &c.bucket,
		Key:    ptr(obj.JSONPath()),
		Body:   bytes.NewBuffer(rawJSON),
	}

	c.dirtyPaths = append(c.dirtyPaths, "/"+obj.JSONPath())

	c.log.Debugf("Uploading %T to s3: %v", obj, obj.JSONPath())
	if _, err := c.uploadClient.Upload(ctx, in); err != nil {
		return fmt.Errorf("uploading %T: %w", obj, err)
	}

	return nil
}

func (c *Client) deleteVersionFromMinorVersionList(ctx context.Context, ver versionsapi.Version,
) (*versionsapi.Latest, error) {
	minorList := versionsapi.List{
		Ref:         ver.Ref,
		Stream:      ver.Stream,
		Granularity: versionsapi.GranularityMinor,
		Base:        ver.WithGranularity(versionsapi.GranularityMinor),
		Kind:        versionsapi.VersionKindImage,
	}
	c.log.Debugf("Fetching minor version list for version %s", ver.Version)
	minorList, err := c.FetchVersionList(ctx, minorList)
	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
		c.log.Warnf("Minor version list for version %s not found", ver.Version)
		c.log.Warnf("Skipping update of minor version list")
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("fetching minor version list for version %s: %w", ver.Version, err)
	}

	if !minorList.Contains(ver.Version) {
		c.log.Warnf("Version %s is not in minor version list %s", ver.Version, minorList.JSONPath())
		c.log.Warnf("Skipping update of minor version list")
		return nil, nil
	}

	semver.Sort(minorList.Versions)
	for i, v := range minorList.Versions {
		if v == ver.Version {
			minorList.Versions = append(minorList.Versions[:i], minorList.Versions[i+1:]...)
			break
		}
	}

	var latest *versionsapi.Latest
	if len(minorList.Versions) != 0 {
		latest = &versionsapi.Latest{
			Ref:     ver.Ref,
			Stream:  ver.Stream,
			Kind:    versionsapi.VersionKindImage,
			Version: minorList.Versions[len(minorList.Versions)-1],
		}
		c.log.Debugf("Possible latest version replacement %q", latest.Version)
	}

	if c.dryRun {
		c.log.Debugf("DryRun: Updating minor version list %s to %v", minorList.JSONPath(), minorList)
		return latest, nil
	}

	c.log.Debugf("Updating minor version list %s", minorList.JSONPath())
	if err := c.UpdateVersionList(ctx, minorList); err != nil {
		return latest, fmt.Errorf("updating minor version list %s: %w", minorList.JSONPath(), err)
	}

	c.log.Debugf("Removed version %s from minor version list %s", ver.Version, minorList.JSONPath())
	return latest, nil
}

func (c *Client) deleteVersionFromLatest(ctx context.Context, ver versionsapi.Version, possibleNewLatest *versionsapi.Latest,
) error {
	latest := versionsapi.Latest{
		Ref:    ver.Ref,
		Stream: ver.Stream,
		Kind:   versionsapi.VersionKindImage,
	}
	c.log.Debugf("Fetching latest version from %s", latest.JSONPath())
	latest, err := c.FetchVersionLatest(ctx, latest)
	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
		c.log.Warnf("Latest version for %s not found", latest.JSONPath())
		return nil
	} else if err != nil {
		return fmt.Errorf("fetching latest version: %w", err)
	}

	if latest.Version != ver.Version {
		c.log.Debugf("Latest version is %s, not the deleted version %s", latest.Version, ver.Version)
		return nil
	}

	if possibleNewLatest == nil {
		c.log.Errorf("Latest version is %s, but no new latest version was found", latest.Version)
		c.log.Errorf("A manual update of latest at %s might be needed", latest.JSONPath())
		return fmt.Errorf("latest version is %s, but no new latest version was found", latest.Version)
	}

	if c.dryRun {
		c.log.Debugf("Would update latest version from %s to %s", latest.Version, possibleNewLatest.Version)
		return nil
	}

	c.log.Infof("Updating latest version from %s to %s", latest.Version, possibleNewLatest.Version)
	if err := c.UpdateVersionLatest(ctx, *possibleNewLatest); err != nil {
		return fmt.Errorf("updating latest version: %w", err)
	}

	return nil
}

func (c *Client) deletePath(ctx context.Context, path string) error {
	listIn := &s3.ListObjectsV2Input{
		Bucket: &c.bucket,
		Prefix: &path,
	}
	c.log.Debugf("Listing objects in %s", path)
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
	c.log.Debugf("Found %d objects in %s", len(objs), path)

	if len(objs) == 0 {
		c.log.Warnf("Path %s is already empty", path)
		return nil
	}

	objIDs := make([]s3types.ObjectIdentifier, len(objs))
	for i, obj := range objs {
		objIDs[i] = s3types.ObjectIdentifier{Key: obj.Key}
	}

	if c.dryRun {
		c.log.Debugf("DryRun: Deleting %d objects with IDs %v", len(objs), objIDs)
		return nil
	}

	c.dirtyPaths = append(c.dirtyPaths, "/"+path)

	deleteIn := &s3.DeleteObjectsInput{
		Bucket: &c.bucket,
		Delete: &s3types.Delete{
			Objects: objIDs,
		},
	}
	c.log.Debugf("Deleting %d objects in %s", len(objs), path)
	if _, err := c.s3Client.DeleteObjects(ctx, deleteIn); err != nil {
		return fmt.Errorf("deleting objects in %s: %w", path, err)
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

func ptr[T any](t T) *T {
	return &t
}
