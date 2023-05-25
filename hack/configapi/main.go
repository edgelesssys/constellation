/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/spf13/cobra"
)

const (
	awsRegion      = "eu-central-1"
	awsBucket      = "cdn-constellation-backend"
	invalidDefault = 0
)

var (
	// AWS S3 credentials.
	awsAccessKeyID string
	awsAccessKey   string

	// Azure SEV-SNP version numbers.
	bootloaderVersion uint8
	teeVersion        uint8
	snpVersion        uint8
	microcodeVersion  uint8
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	myCmd := &cobra.Command{
		Use:   "upload a set of versions specific to the azure-sev-snp attestation variant to the config api",
		Short: "upload a set of versions specific to the azure-sev-snp attestation variant to the config api",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			cfg := uri.AWSS3Config{
				Bucket:      awsBucket,
				AccessKeyID: awsAccessKeyID,
				AccessKey:   awsAccessKey,
				Region:      awsRegion,
			}
			sut, err := configapi.NewAttestationVersionRepo(ctx, cfg)
			if err != nil {
				panic(err)
			}
			versions := configapi.AzureSEVSNPVersion{
				Bootloader: bootloaderVersion,
				TEE:        teeVersion,
				SNP:        snpVersion,
				Microcode:  microcodeVersion,
			}

			if err := sut.UploadAzureSEVSNP(ctx, versions, time.Now()); err != nil {
				panic(err)
			} else {
				fmt.Println("Successfully uploaded version numbers", versions)
			}
		},
	}
	myCmd.PersistentFlags().Uint8VarP(&bootloaderVersion, "bootloader-version", "b", invalidDefault, "Bootloader version number")
	handleError(myCmd.MarkPersistentFlagRequired("bootloader-version"))

	myCmd.PersistentFlags().Uint8VarP(&teeVersion, "tee-version", "t", invalidDefault, "TEE version number")
	handleError(myCmd.MarkPersistentFlagRequired("tee-version"))

	myCmd.PersistentFlags().Uint8VarP(&snpVersion, "snp-version", "s", invalidDefault, "SNP version number")
	handleError(myCmd.MarkPersistentFlagRequired("snp-version"))

	myCmd.PersistentFlags().Uint8VarP(&microcodeVersion, "microcode-version", "m", invalidDefault, "Microcode version number")
	handleError(myCmd.MarkPersistentFlagRequired("microcode-version"))

	myCmd.PersistentFlags().StringVar(&awsAccessKeyID, "key-id", "", "ID of the Access key to use for AWS tests. Required for AWS KMS and storage test.")
	handleError(myCmd.MarkPersistentFlagRequired("key-id"))

	myCmd.PersistentFlags().StringVar(&awsAccessKey, "key", "", "Access key to use for AWS tests. Required for AWS KMS and storage test.")
	handleError(myCmd.MarkPersistentFlagRequired("key"))
	handleError(myCmd.Execute())
}
