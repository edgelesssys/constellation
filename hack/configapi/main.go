/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/configapi"
	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/spf13/cobra"
)

const (
	awsRegion      = "eu-central-1"
	awsBucket      = "cdn-constellation-backend"
	invalidDefault = 0
	envAwsKeyID    = "AWS_ACCESS_KEY_ID"
	envAwsKey      = "AWS_ACCESS_KEY"
)

var (
	// Azure SEV-SNP version numbers.
	bootloaderVersion uint8
	teeVersion        uint8
	snpVersion        uint8
	microcodeVersion  uint8

	// Cosign credentials.
	cosignPwd      string
	privateKeyPath string
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	myCmd := &cobra.Command{
		Use:   fmt.Sprintf("please set the %s and %s environment variables and all flags", envAwsKeyID, envAwsKey),
		Short: "upload a set of versions specific to the azure-sev-snp attestation variant to the config api",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			cfg := staticupload.Config{
				Bucket: awsBucket,
				Region: awsRegion,
			}
			privateKey := getBytesFromFilePath(privateKeyPath)

			sut, err := configapi.NewAttestationVersionRepo(ctx, cfg, []byte(cosignPwd), privateKey)
			handleError(err)
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
	myCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if _, present := os.LookupEnv(envAwsKey); !present {
			return fmt.Errorf("%s not set", envAwsKey)
		}
		if _, present := os.LookupEnv(envAwsKeyID); !present {
			return fmt.Errorf("%s not set", envAwsKeyID)
		}
		return nil
	}
	myCmd.PersistentFlags().Uint8VarP(&bootloaderVersion, "bootloader-version", "b", invalidDefault, "Bootloader version number")
	handleError(myCmd.MarkPersistentFlagRequired("bootloader-version"))

	myCmd.PersistentFlags().Uint8VarP(&teeVersion, "tee-version", "t", invalidDefault, "TEE version number")
	handleError(myCmd.MarkPersistentFlagRequired("tee-version"))

	myCmd.PersistentFlags().Uint8VarP(&snpVersion, "snp-version", "s", invalidDefault, "SNP version number")
	handleError(myCmd.MarkPersistentFlagRequired("snp-version"))

	myCmd.PersistentFlags().Uint8VarP(&microcodeVersion, "microcode-version", "m", invalidDefault, "Microcode version number")
	handleError(myCmd.MarkPersistentFlagRequired("microcode-version"))

	myCmd.PersistentFlags().StringVar(&cosignPwd, "cosign-pwd", "", "Cosign password used to decrpyt the private key.")
	handleError(myCmd.MarkPersistentFlagRequired("cosign-pwd"))

	myCmd.PersistentFlags().StringVar(&privateKeyPath, "private-key", "", "File path of private key used to sign the payload.")
	handleError(myCmd.MarkPersistentFlagRequired("private-key"))

	handleError(myCmd.Execute())
}

func getBytesFromFilePath(path string) []byte {
	file, err := os.Open(path)
	handleError(err)

	content, err := io.ReadAll(file)
	handleError(err)
	return content
}
