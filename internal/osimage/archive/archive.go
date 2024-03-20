/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package archive is used to archive OS images in S3.
package archive

import (
	"context"
	"fmt"
	"io"
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

// Archivist uploads OS images to S3.
type Archivist struct {
	uploadClient      uploadClient
	uploadClientClose func(ctx context.Context) error
	// bucket is the name of the S3 bucket to use.
	bucket string

	log *slog.Logger
}

// New creates a new Archivist.
func New(ctx context.Context, region, bucket, distributionID string, log *slog.Logger) (*Archivist, CloseFunc, error) {
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

	archivist := &Archivist{
		uploadClient:      staticUploadClient,
		uploadClientClose: staticUploadClientClose,
		bucket:            bucket,
		log:               log,
	}
	archivistClose := func(ctx context.Context) error {
		return archivist.Close(ctx)
	}

	return archivist, archivistClose, nil
}

// Close closes the uploader.
// It invalidates the CDN cache for all uploaded files.
func (a *Archivist) Close(ctx context.Context) error {
	if a.uploadClientClose == nil {
		return nil
	}
	return a.uploadClientClose(ctx)
}

// Archive reads the OS image in img and uploads it as key.
func (a *Archivist) Archive(ctx context.Context, version versionsapi.Version, csp, attestationVariant string, img io.Reader) (string, error) {
	key, err := url.JoinPath(version.ArtifactPath(versionsapi.APIV1), version.Kind().String(), "csp", csp, attestationVariant, "image.raw")
	if err != nil {
		return "", err
	}
	a.log.Debug(fmt.Sprintf("Archiving OS image %q to s3://%s/%s", fmt.Sprintf("%s %s %v", csp, attestationVariant, version.ShortPath()), a.bucket, key))
	_, err = a.uploadClient.Upload(ctx, &s3.PutObjectInput{
		Bucket:            &a.bucket,
		Key:               &key,
		Body:              img,
		ChecksumAlgorithm: s3types.ChecksumAlgorithmSha256,
	})
	return constants.CDNRepositoryURL + "/" + key, err
}

type uploadClient interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

// CloseFunc is a function that closes the client.
type CloseFunc func(ctx context.Context) error
