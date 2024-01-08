/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
)

func deleteAzure(ctx context.Context, client *attestationconfigapi.Client, cfg deleteConfig) error {
	if cfg.provider != cloudprovider.Azure && cfg.kind != snpReport {
		return fmt.Errorf("provider %s and kind %s not supported", cfg.provider, cfg.kind)
	}

	return client.DeleteSEVSNPVersion(ctx, variant.AzureSEVSNP{}, cfg.version)
}

func ptr[A any](a A) *A { return &a }

func deleteRecursive(ctx context.Context, path string, client *staticupload.Client, cfg deleteConfig) error {
	resp, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(cfg.bucket),
		Prefix: aws.String(path),
	})
	if err != nil {
		return err
	}

	// Delete all objects in the path.
	objIDs := make([]s3types.ObjectIdentifier, len(resp.Contents))
	for i, obj := range resp.Contents {
		objIDs[i] = s3types.ObjectIdentifier{Key: obj.Key}
	}
	if len(objIDs) > 0 {
		_, err = client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(cfg.bucket),
			Delete: &s3types.Delete{
				Objects: objIDs,
				Quiet:   ptr(true),
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
