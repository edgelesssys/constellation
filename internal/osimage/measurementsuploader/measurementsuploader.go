/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// package measurementsuploader is used to upload measurements (v2) JSON files (and signatures) to S3.
package measurementsuploader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	versionsapi "github.com/edgelesssys/constellation/v2/internal/api/versions"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
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

// Upload uploads the measurements v2 JSON file and its signature to S3.
func (a *Uploader) Upload(ctx context.Context, rawMeasurement, signature io.ReadSeeker) (string, string, error) {
	// parse the measurements to get the ref, stream, and version
	var measurements measurements.ImageMeasurementsV2
	if err := json.NewDecoder(rawMeasurement).Decode(&measurements); err != nil {
		return "", "", err
	}
	if _, err := rawMeasurement.Seek(0, io.SeekStart); err != nil {
		return "", "", err
	}

	ver := versionsapi.Version{
		Ref:     measurements.Ref,
		Stream:  measurements.Stream,
		Version: measurements.Version,
		Kind:    versionsapi.VersionKindImage,
	}
	key, err := url.JoinPath(ver.ArtifactPath(versionsapi.APIV2), ver.Kind.String(), "measurements.json")
	if err != nil {
		return "", "", err
	}
	sigKey, err := url.JoinPath(ver.ArtifactPath(versionsapi.APIV2), ver.Kind.String(), "measurements.json.sig")
	if err != nil {
		return "", "", err
	}
	a.log.Debugf("Archiving image measurements to s3://%v/%v and s3://%v/%v", a.bucket, key, a.bucket, sigKey)
	if _, err = a.uploadClient.Upload(ctx, &s3.PutObjectInput{
		Bucket:            &a.bucket,
		Key:               &key,
		Body:              rawMeasurement,
		ChecksumAlgorithm: s3types.ChecksumAlgorithmSha256,
	}); err != nil {
		return "", "", fmt.Errorf("uploading measurements: %w", err)
	}
	if _, err = a.uploadClient.Upload(ctx, &s3.PutObjectInput{
		Bucket:            &a.bucket,
		Key:               &sigKey,
		Body:              signature,
		ChecksumAlgorithm: s3types.ChecksumAlgorithmSha256,
	}); err != nil {
		return "", "", fmt.Errorf("uploading measurements signature: %w", err)
	}
	return constants.CDNRepositoryURL + "/" + key, constants.CDNRepositoryURL + "/" + sigKey, nil
}

type uploadClient interface {
	Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}
