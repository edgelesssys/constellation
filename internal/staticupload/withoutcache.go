/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package staticupload

import (
	"context"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	s3manager "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/edgelesssys/constellation/v2/internal/logger"
)

// ClientWithoutCache is a static file uploader/updater/remover for the CDN / static API without CDN cache invalidation.
type ClientWithoutCache struct {
	s3Client     *s3.Client
	uploadClient *s3manager.Uploader
	bucket       string
	DryRun       bool
	Logger       *logger.Logger
}

// NewClientWithoutCDNCache creates a new client for the versions API without CDN cache invalidation.
func NewClientWithoutCDNCache(ctx context.Context, region, bucket string, dryRun bool, log *logger.Logger) (*ClientWithoutCache, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg)
	uploadClient := s3manager.NewUploader(s3Client)
	client := &ClientWithoutCache{
		s3Client:     s3Client,
		uploadClient: uploadClient,
		bucket:       bucket,
		DryRun:       dryRun,
		Logger:       log,
	}
	return client, nil
}
