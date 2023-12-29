/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package imageinfo is used to upload image info JSON files to S3.
package imageinfo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
)

// Uploader uploads image info to S3.
type Uploader struct {
	uploadClient      uploadClient
	uploadClientClose func(ctx context.Context) error
	// bucket is the name of the S3 bucket to use.
	bucket string

	log *slog.Logger
}

// New creates a new Uploader.
func New(ctx context.Context, region, bucket, distributionID string, log *slog.Logger) (*Uploader, CloseFunc, error) {
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

	uploader := &Uploader{
		uploadClient:      staticUploadClient,
		uploadClientClose: staticUploadClientClose,
		bucket:            bucket,
		log:               log,
	}
	uploaderClose := func(ctx context.Context) error {
		return uploader.Close(ctx)
	}
	return uploader, uploaderClose, nil
}

// Close closes the uploader.
// It invalidates the CDN cache for all uploaded files.
func (a *Uploader) Close(ctx context.Context) error {
	if a.uploadClientClose == nil {
		return nil
	}
	return a.uploadClientClose(ctx)
}

// Upload marshals the image info to JSON and uploads it to S3.
func (a *Uploader) Upload(ctx context.Context, imageInfo versionsapi.ImageInfo) (string, error) {
	ver, err := versionsapi.NewVersion(imageInfo.Ref, imageInfo.Stream, imageInfo.Version, versionsapi.VersionKindImage)
	if err != nil {
		return "", fmt.Errorf("creating version: %w", err)
	}
	key, err := url.JoinPath(ver.ArtifactPath(versionsapi.APIV2), ver.Kind().String(), "info.json")
	if err != nil {
		return "", err
	}
	a.log.Debug("Archiving image info to s3://%v/%v", a.bucket, key)
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(imageInfo); err != nil {
		return "", err
	}
	_, err = a.uploadClient.Upload(ctx, &s3.PutObjectInput{
		Bucket:            &a.bucket,
		Key:               &key,
		Body:              buf,
		ChecksumAlgorithm: s3types.ChecksumAlgorithmSha256,
	})
	return constants.CDNRepositoryURL + "/" + key, err
}

type uploadClient interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// CloseFunc is a function that closes the client.
type CloseFunc func(ctx context.Context) error
