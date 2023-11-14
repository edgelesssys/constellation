/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/edgelesssys/constellation/v2/internal/verify"
)

func uploadAzure(ctx context.Context, client *attestationconfigapi.Client, cfg uploadConfig, fs file.Handler, log *logger.Logger) error {
	if cfg.kind != snpReport {
		return fmt.Errorf("kind %s not supported", cfg.kind)
	}

	log.Infof("Reading SNP report from file: %s", cfg.path)
	var report verify.Report
	if err := fs.ReadJSON(cfg.path, &report); err != nil {
		return fmt.Errorf("reading snp report: %w", err)
	}

	inputVersion := convertTCBVersionToAzureVersion(report.SNPReport.LaunchTCB)
	log.Infof("Input report: %+v", inputVersion)

	latestAPIVersionAPI, err := attestationconfigapi.NewFetcherWithCustomCDNAndCosignKey(cfg.url, cfg.cosignPublicKey).FetchSEVSNPVersionLatest(ctx, variant.AzureSEVSNP{})
	if err != nil {
		if errors.Is(err, attestationconfigapi.ErrNoVersionsFound) {
			log.Infof("No versions found in API, but assuming that we are uploading the first version.")
		} else {
			return fmt.Errorf("fetching latest version: %w", err)
		}
	}
	latestAPIVersion := latestAPIVersionAPI.SEVSNPVersion
	if err := client.UploadSEVSNPVersionLatest(ctx, variant.AzureSEVSNP{}, inputVersion, latestAPIVersion, cfg.uploadDate, cfg.force); err != nil {
		if errors.Is(err, attestationconfigapi.ErrNoNewerVersion) {
			log.Infof("Input version: %+v is not newer than latest API version: %+v", inputVersion, latestAPIVersion)
			return nil
		}
		return fmt.Errorf("updating latest version: %w", err)
	}

	return nil
}

func convertTCBVersionToAzureVersion(tcb verify.TCBVersion) attestationconfigapi.SEVSNPVersion {
	return attestationconfigapi.SEVSNPVersion{
		Bootloader: tcb.Bootloader,
		TEE:        tcb.TEE,
		SNP:        tcb.SNP,
		Microcode:  tcb.Microcode,
	}
}

func deleteAzure(ctx context.Context, client *attestationconfigapi.Client, cfg deleteConfig) error {
	if cfg.provider == cloudprovider.Azure && cfg.kind == snpReport {
		return client.DeleteSEVSNPVersion(ctx, variant.AzureSEVSNP{}, cfg.version)
	}

	return fmt.Errorf("provider %s and kind %s not supported", cfg.provider, cfg.kind)
}

func deleteRecursiveAzure(ctx context.Context, client *staticupload.Client, cfg deleteConfig) error {
	path := "constellation/v1/attestation/azure-sev-snp"
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
				Quiet:   true,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
