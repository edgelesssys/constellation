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
	"net/url"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

// Uploader uploads image info to S3.
type Uploader struct {
	uploadClient uploadClient
	// bucket is the name of the S3 bucket to use.
	bucket string

	log *logger.Logger
}

// New creates a new Uploader.
func New(ctx context.Context, region, bucket string, log *logger.Logger) (*Uploader, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	s3client := s3.NewFromConfig(cfg)
	uploadClient := s3manager.NewUploader(s3client)

	return &Uploader{
		uploadClient: uploadClient,
		bucket:       bucket,
		log:          log,
	}, nil
}

// Upload marshals the image info to JSON and uploads it to S3.
func (a *Uploader) Upload(ctx context.Context, imageInfo versionsapi.ImageInfo) (string, error) {
	ver := versionsapi.Version{
		Ref:     imageInfo.Ref,
		Stream:  imageInfo.Stream,
		Version: imageInfo.Version,
		Kind:    versionsapi.VersionKindImage,
	}
	key, err := url.JoinPath(ver.ArtifactPath(versionsapi.APIV2), ver.Kind.String(), "info.json")
	if err != nil {
		return "", err
	}
	a.log.Debugf("Archiving image info to s3://%v/%v", a.bucket, key)
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
