/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"

	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/spf13/cobra"
)

const (
	awsRegion           = "eu-central-1"
	awsBucket           = "cdn-constellation-backend"
	envCosignPwd        = "COSIGN_PASSWORD"
	envCosignPrivateKey = "COSIGN_PRIVATE_KEY"
)

var (
	maaFilePath string
	// Cosign credentials.
	cosignPwd  string
	privateKey string
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// newRootCmd creates the root command.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "COSIGN_PASSWORD=$CPWD COSIGN_PRIVATE_KEY=$CKEY upload --version-file $FILE",
		Short: "Upload a set of versions specific to the azure-sev-snp attestation variant to the config api.",

		Long: fmt.Sprintf("Upload a set of versions specific to the azure-sev-snp attestation variant to the config api."+
			"Please authenticate with AWS through your preferred method (e.g. environment variables, CLI)"+
			"to be able to upload to S3. Set the %s and %s environment variables to authenticate with cosign.",
			envCosignPrivateKey, envCosignPwd,
		),
		PreRunE: envCheck,
		RunE:    runCmd,
	}
	rootCmd.Flags().StringVarP(&maaFilePath, "maa-claims-path", "t", "", "File path to a json file containing the MAA claims.")
	rootCmd.Flags().StringP("upload-date", "d", "", "upload a version with this date as version name.")
	must(rootCmd.MarkFlagRequired("maa-claims-path"))
	rootCmd.AddCommand(newDeleteCmd())
	return rootCmd
}

func envCheck(_ *cobra.Command, _ []string) error {
	if os.Getenv(envCosignPrivateKey) == "" || os.Getenv(envCosignPwd) == "" {
		return fmt.Errorf("please set both %s and %s environment variables", envCosignPrivateKey, envCosignPwd)
	}
	cosignPwd = os.Getenv(envCosignPwd)
	privateKey = os.Getenv(envCosignPrivateKey)
	return nil
}

func runCmd(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	log := logger.New(logger.PlainLog, zap.DebugLevel).Named("attestationconfigapi")
	cfg := staticupload.Config{
		Bucket: awsBucket,
		Region: awsRegion,
	}
	maaClaimsBytes, err := os.ReadFile(maaFilePath)
	if err != nil {
		return fmt.Errorf("reading MAA claims file: %w", err)
	}
	var maaTCB maaTokenTCBClaims
	if err = json.Unmarshal(maaClaimsBytes, &maaTCB); err != nil {
		return fmt.Errorf("unmarshalling MAA claims file: %w", err)
	}
	inputVersion := maaTCB.ToAzureSEVSNPVersion()

	dateStr, err := cmd.Flags().GetString("upload-date")
	if err != nil {
		return fmt.Errorf("getting upload date: %w", err)
	}
	uploadDate := time.Now()
	if dateStr != "" {
		uploadDate, err = time.Parse(attestationconfigapi.VersionFormat, dateStr)
		if err != nil {
			return fmt.Errorf("parsing date: %w", err)
		}
	}

	latestAPIVersion, err := attestationconfigapi.NewFetcher().FetchAzureSEVSNPVersionLatest(ctx, uploadDate)
	if err != nil {
		return fmt.Errorf("fetching latest version: %w", err)
	}

	isNewer, err := isInputNewerThanLatestAPI(inputVersion, latestAPIVersion.AzureSEVSNPVersion)
	if err != nil {
		return fmt.Errorf("comparing versions: %w", err)
	}
	if !isNewer {
		fmt.Printf("Input version: %+v is not newer than latest API version: %+v\n", inputVersion, latestAPIVersion)
		return nil
	}
	fmt.Printf("Input version: %+v is newer than latest API version: %+v\n", inputVersion, latestAPIVersion)

	client, stop, err := attestationconfigapi.NewClient(ctx, cfg, []byte(cosignPwd), []byte(privateKey), false, log)
	defer func() {
		if err := stop(ctx); err != nil {
			cmd.Printf("stopping client: %v\n", err)
		}
	}()
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	if err := client.UploadAzureSEVSNP(ctx, inputVersion, uploadDate); err != nil {
		return fmt.Errorf("uploading version: %w", err)
	}

	cmd.Printf("Successfully uploaded new Azure SEV-SNP version: %+v\n", inputVersion)
	return nil
}

// maaTokenTCBClaims describes the TCB information in a MAA token.
type maaTokenTCBClaims struct {
	TEESvn        uint8 `json:"x-ms-sevsnpvm-tee-svn"`
	SNPFwSvn      uint8 `json:"x-ms-sevsnpvm-snpfw-svn"`
	MicrocodeSvn  uint8 `json:"x-ms-sevsnpvm-microcode-svn"`
	BootloaderSvn uint8 `json:"x-ms-sevsnpvm-bootloader-svn"`
}

func (c maaTokenTCBClaims) ToAzureSEVSNPVersion() attestationconfigapi.AzureSEVSNPVersion {
	return attestationconfigapi.AzureSEVSNPVersion{
		TEE:        c.TEESvn,
		SNP:        c.SNPFwSvn,
		Microcode:  c.MicrocodeSvn,
		Bootloader: c.BootloaderSvn,
	}
}

// isInputNewerThanLatestAPI compares all version fields with the latest API version and returns true if any input field is newer.
func isInputNewerThanLatestAPI(input, latest attestationconfigapi.AzureSEVSNPVersion) (bool, error) {
	if input == latest {
		return false, nil
	}

	if input.TEE < latest.TEE {
		return false, fmt.Errorf("input TEE version: %d is older than latest API version: %d", input.TEE, latest.TEE)
	}
	if input.SNP < latest.SNP {
		return false, fmt.Errorf("input SNP version: %d is older than latest API version: %d", input.SNP, latest.SNP)
	}
	if input.Microcode < latest.Microcode {
		return false, fmt.Errorf("input Microcode version: %d is older than latest API version: %d", input.Microcode, latest.Microcode)
	}
	if input.Bootloader < latest.Bootloader {
		return false, fmt.Errorf("input Bootloader version: %d is older than latest API version: %d", input.Bootloader, latest.Bootloader)
	}

	return true, nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
