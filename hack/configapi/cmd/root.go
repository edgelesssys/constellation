/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"go.uber.org/zap"

	"github.com/edgelesssys/constellation/v2/internal/staticupload"
	"github.com/spf13/cobra"
)

const (
	invalidDefault      = 0
	envAwsKeyID         = "AWS_ACCESS_KEY_ID"
	envAwsKey           = "AWS_ACCESS_KEY"
	envCosignPwd        = "COSIGN_PASSWORD"
	envCosignPrivateKey = "COSIGN_PRIVATE_KEY"
)

var (
	awsBucket       string
	awsRegion       string
	versionFilePath string
	force           bool
	// Cosign credentials.
	cosignPwd  string
	privateKey string
)

// Execute executes the root command.
func Execute() error {
	return newRootCmd().Execute()
}

// newRootCmd creates the root command.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "COSIGN_PASSWORD=$CPWD COSIGN_PRIVATE_KEY=$CKEY AWS_ACCESS_KEY_ID=$ID AWS_ACCESS_KEY=$KEY upload --version-file $FILE",
		Short: "Upload a set of versions specific to the azure-sev-snp attestation variant to the config api.",

		Long:    fmt.Sprintf("Upload a set of versions specific to the azure-sev-snp attestation variant to the config api. Please authenticate with AWS through your preferred method (e.g. environment variables, CLI) to be able to upload to S3. Set the %s and %s environment variables to authenticate with cosign.", envCosignPrivateKey, envCosignPwd),
		PreRunE: envCheck,
		RunE:    runCmd,
	}
	rootCmd.Flags().StringVarP(&versionFilePath, "version-file", "f", "", "File path to the version json file.")
	rootCmd.Flags().StringVar(&awsBucket, "bucket", "cdn-constellation-backend", "Bucket to upload files to.")
	rootCmd.Flags().StringVar(&awsRegion, "region", "eu-central-1", "Bucket to upload files to.")
	rootCmd.Flags().BoolVar(&force, "force", false, "force to upload version regardless of comparison to latest API value.")
	rootCmd.Flags().StringP("upload-date", "d", "", "upload a version with this date as version name. Setting it implies --force.")
	must(enforceRequiredFlags(rootCmd, "version-file"))
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
	cfg := staticupload.Config{
		Bucket: awsBucket,
		Region: awsRegion,
	}
	versionBytes, err := os.ReadFile(versionFilePath)
	if err != nil {
		return fmt.Errorf("reading version file: %w", err)
	}
	var inputVersion attestationconfigapi.AzureSEVSNPVersion
	if err = json.Unmarshal(versionBytes, &inputVersion); err != nil {
		return fmt.Errorf("unmarshalling version file: %w", err)
	}

	dateStr, err := cmd.Flags().GetString("upload-date")
	if err != nil {
		return fmt.Errorf("getting upload date: %w", err)
	}
	var uploadDate time.Time
	if dateStr != "" {
		uploadDate, err = time.Parse(attestationconfigapi.VersionFormat, dateStr)
		if err != nil {
			return fmt.Errorf("parsing date: %w", err)
		}
	} else {
		uploadDate = time.Now()
		force = true
	}

	doUpload := false
	if !force {
		latestAPIVersion, err := attestationconfigapi.NewFetcher().FetchAzureSEVSNPVersionLatest(ctx, time.Now())
		if err != nil {
			return fmt.Errorf("fetching latest version: %w", err)
		}

		isNewer, err := isInputNewerThanLatestAPI(inputVersion, latestAPIVersion.AzureSEVSNPVersion)
		if err != nil {
			return fmt.Errorf("comparing versions: %w", err)
		}
		cmd.Print(versionComparisonInformation(isNewer, inputVersion, latestAPIVersion.AzureSEVSNPVersion))
		doUpload = isNewer
	} else {
		doUpload = true
		cmd.Println("Forcing upload of new version")
	}

	if doUpload {
		sut, sutClose, err := attestationconfigapi.NewClient(ctx, cfg, []byte(cosignPwd), []byte(privateKey), false, log())
		defer func() {
			if err := sutClose(ctx); err != nil {
				cmd.Printf("closing repo: %v\n", err)
			}
		}()
		if err != nil {
			return fmt.Errorf("creating repo: %w", err)
		}

		if err := sut.UploadAzureSEVSNP(ctx, inputVersion, uploadDate); err != nil {
			return fmt.Errorf("uploading version: %w", err)
		}
		cmd.Printf("Successfully uploaded new Azure SEV-SNP version: %+v\n", inputVersion)
	}
	return nil
}

func versionComparisonInformation(isNewer bool, inputVersion attestationconfigapi.AzureSEVSNPVersion, latestAPIVersion attestationconfigapi.AzureSEVSNPVersion) string {
	if isNewer {
		return fmt.Sprintf("Input version: %+v is newer than latest API version: %+v\n", inputVersion, latestAPIVersion)
	}
	return fmt.Sprintf("Input version: %+v is not newer than latest API version: %+v\n", inputVersion, latestAPIVersion)
}

// isInputNewerThanLatestAPI compares all version fields with the latest API version and returns true if any input field is newer.
func isInputNewerThanLatestAPI(input, latest attestationconfigapi.AzureSEVSNPVersion) (bool, error) {
	inputValues := reflect.ValueOf(input)
	latestValues := reflect.ValueOf(latest)
	fields := reflect.TypeOf(input)
	num := fields.NumField()
	// validate that no input field is smaller than latest
	for i := 0; i < num; i++ {
		field := fields.Field(i)
		inputValue := inputValues.Field(i).Uint()
		latestValue := latestValues.Field(i).Uint()
		if inputValue < latestValue {
			return false, fmt.Errorf("input %s version: %d is older than latest API version: %d", field.Name, inputValue, latestValue)
		} else if inputValue > latestValue {
			return true, nil
		}
	}
	// check if any input field is greater than latest
	for i := 0; i < num; i++ {
		inputValue := inputValues.Field(i).Uint()
		latestValue := latestValues.Field(i).Uint()
		if inputValue > latestValue {
			return true, nil
		}
	}
	return false, nil
}

func enforceRequiredFlags(cmd *cobra.Command, flags ...string) error {
	for _, flag := range flags {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			return err
		}
	}
	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func log() *logger.Logger {
	return logger.New(logger.PlainLog, zap.DebugLevel).Named("attestationconfigapi")
}
