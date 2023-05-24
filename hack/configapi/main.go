/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/spf13/pflag"
)

var (
	// AWS S3 credentials.
	awsAccessKeyID = pflag.String("key-id", "", "ID of the Access key to use for AWS tests. Required for AWS KMS and storage test.")
	awsAccessKey   = pflag.String("key", "", "Access key to use for AWS tests. Required for AWS KMS and storage test.")
	awsRegion      = "eu-central-1"
	awsBucket      = "cdn-constellation-backend"

	// Azure SEV-SNP version numbers.
	bootloaderVersion = pflag.Uint8P("bootloader-version", "b", 0, "Bootloader version number")
	teeVersion        = pflag.Uint8P("tee-version", "t", 0, "TEE version number")
	snpVersion        = pflag.Uint8P("snp-version", "s", 0, "SNP version number")
	microcodeVersion  = pflag.Uint8P("microcode-version", "m", 0, "Microcode version number")
)

func main() {
	pflag.Parse()
	if *awsAccessKey == "" || *awsAccessKeyID == "" || awsBucket == "" || awsRegion == "" {
		pflag.Usage()
		fmt.Println("Required flags not set: --aws-access-key, --aws-access-key-id, --aws-bucket, --aws-region. Skipping tests.")
		os.Exit(0)
	}
	ctx := context.Background()
	cfg := uri.AWSS3Config{
		Bucket:      awsBucket,
		AccessKeyID: *awsAccessKeyID,
		AccessKey:   *awsAccessKey,
		Region:      awsRegion,
	}
	sut, err := configapi.NewAttestationVersionRepo(ctx, cfg)
	if err != nil {
		panic(err)
	}
	versions := configapi.AzureSEVSNPVersion{
		Bootloader: *bootloaderVersion,
		TEE:        *teeVersion,
		SNP:        *snpVersion,
		Microcode:  *microcodeVersion,
	}

	if err := sut.UploadAzureSEVSNP(ctx, versions, time.Now()); err != nil {
		panic(err)
	} else {
		fmt.Println("Successfully uploaded version numbers", versions)
	}
}
