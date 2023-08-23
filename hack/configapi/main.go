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
	rootCmd.Flags().StringP("maa-claims-path", "t", "", "File path to a json file containing the MAA claims.")
	rootCmd.Flags().StringP("upload-date", "d", "", "upload a version with this date as version name.")
	rootCmd.PersistentFlags().StringP("region", "r", awsRegion, "region of the targeted bucket.")
	rootCmd.PersistentFlags().StringP("bucket", "b", awsBucket, "bucket targeted by all operations.")
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

	flags, err := parseCliFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing cli flags: %w", err)
	}

	cfg := staticupload.Config{
		Bucket: flags.bucket,
		Region: flags.region,
	}

	log.Infof("Reading MAA claims from file: %s", flags.maaFilePath)
	maaClaimsBytes, err := os.ReadFile(flags.maaFilePath)
	if err != nil {
		return fmt.Errorf("reading MAA claims file: %w", err)
	}
	var maaTCB maaTokenTCBClaims
	if err = json.Unmarshal(maaClaimsBytes, &maaTCB); err != nil {
		return fmt.Errorf("unmarshalling MAA claims file: %w", err)
	}
	inputVersion := maaTCB.ToAzureSEVSNPVersion()
	log.Infof("Input version: %+v", inputVersion)

	latestAPIVersionAPI, err := attestationconfigapi.NewFetcher().FetchAzureSEVSNPVersionLatest(ctx, flags.uploadDate)
	if err != nil {
		return fmt.Errorf("fetching latest version: %w", err)
	}
	latestAPIVersion := latestAPIVersionAPI.AzureSEVSNPVersion

	isNewer, err := isInputNewerThanLatestAPI(inputVersion, latestAPIVersion)
	if err != nil {
		return fmt.Errorf("comparing versions: %w", err)
	}
	if !isNewer {
		log.Infof("Input version: %+v is not newer than latest API version: %+v", inputVersion, latestAPIVersion)
		return nil
	}
	log.Infof("Input version: %+v is newer than latest API version: %+v", inputVersion, latestAPIVersion)

	client, stop, err := attestationconfigapi.NewClient(ctx, cfg, []byte(cosignPwd), []byte(privateKey), false, log)
	defer func() {
		if err := stop(ctx); err != nil {
			cmd.Printf("stopping client: %v\n", err)
		}
	}()
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	if err := client.UploadAzureSEVSNP(ctx, inputVersion, flags.uploadDate); err != nil {
		return fmt.Errorf("uploading version: %w", err)
	}

	cmd.Printf("Successfully uploaded new Azure SEV-SNP version: %+v\n", inputVersion)
	return nil
}

type cliFlags struct {
	maaFilePath string
	uploadDate  time.Time
	region      string
	bucket      string
}

func parseCliFlags(cmd *cobra.Command) (cliFlags, error) {
	maaFilePath, err := cmd.Flags().GetString("maa-claims-path")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting maa claims path: %w", err)
	}

	dateStr, err := cmd.Flags().GetString("upload-date")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting upload date: %w", err)
	}
	uploadDate := time.Now()
	if dateStr != "" {
		uploadDate, err = time.Parse(attestationconfigapi.VersionFormat, dateStr)
		if err != nil {
			return cliFlags{}, fmt.Errorf("parsing date: %w", err)
		}
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting region: %w", err)
	}

	bucket, err := cmd.Flags().GetString("bucket")
	if err != nil {
		return cliFlags{}, fmt.Errorf("getting bucket: %w", err)
	}

	return cliFlags{
		maaFilePath: maaFilePath,
		uploadDate:  uploadDate,
		region:      region,
		bucket:      bucket,
	}, nil
}

// maaTokenTCBClaims describes the TCB information in a MAA token.
type maaTokenTCBClaims struct {
	IsolationTEE struct {
		TEESvn        uint8 `json:"x-ms-sevsnpvm-tee-svn"`
		SNPFwSvn      uint8 `json:"x-ms-sevsnpvm-snpfw-svn"`
		MicrocodeSvn  uint8 `json:"x-ms-sevsnpvm-microcode-svn"`
		BootloaderSvn uint8 `json:"x-ms-sevsnpvm-bootloader-svn"`
	} `json:"x-ms-isolation-tee"`
}

func (c maaTokenTCBClaims) ToAzureSEVSNPVersion() attestationconfigapi.AzureSEVSNPVersion {
	return attestationconfigapi.AzureSEVSNPVersion{
		TEE:        c.IsolationTEE.TEESvn,
		SNP:        c.IsolationTEE.SNPFwSvn,
		Microcode:  c.IsolationTEE.MicrocodeSvn,
		Bootloader: c.IsolationTEE.BootloaderSvn,
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
