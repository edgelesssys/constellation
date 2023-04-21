/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package archive is used to archive OS images in S3.
package archive

import (
	"context"
	"io"
	"net/url"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
)

// Archivist uploads OS images to S3.
type Archivist struct {
	uploadClient uploadClient
	// bucket is the name of the S3 bucket to use.
	bucket string

	log *logger.Logger
}

// New creates a new Archivist.
func New(ctx context.Context, region, bucket string, log *logger.Logger) (*Archivist, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	s3client := s3.NewFromConfig(cfg)
	uploadClient := s3manager.NewUploader(s3client)

	return &Archivist{
		uploadClient: uploadClient,
		bucket:       bucket,
		log:          log,
	}, nil
}

// Archive reads the OS image in img and uploads it as key.
func (a *Archivist) Archive(ctx context.Context, version versionsapi.Version, csp, variant string, img io.Reader) (string, error) {
	key, err := url.JoinPath(version.ArtifactPath(), version.Kind.String(), "csp", csp, variant, "image.raw")
	if err != nil {
		return "", err
	}
	a.log.Debugf("Archiving OS image %s %s %v to s3://%v/%v", csp, variant, version.ShortPath(), a.bucket, key)
	_, err = a.uploadClient.Upload(ctx, &s3.PutObjectInput{
		Bucket:            &a.bucket,
		Key:               &key,
		Body:              img,
		ChecksumAlgorithm: s3types.ChecksumAlgorithmSha256,
	})
	return baseURL + key, err
}

type uploadClient interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

const baseURL = "https://cdn.confidential.cloud/"
