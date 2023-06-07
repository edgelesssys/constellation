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
	awsRegion           = "eu-central-1"
	awsBucket           = "cdn-constellation-backend"
	invalidDefault      = 0
	envAwsKeyID         = "AWS_ACCESS_KEY_ID"
	envAwsKey           = "AWS_ACCESS_KEY"
	envCosignPwd        = "COSIGN_PASSWORD"
	envCosignPrivateKey = "COSIGN_PRIVATE_KEY"
)

var (
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
	rootCmd.Flags().BoolVar(&force, "force", false, "force to upload version regardless of comparison to latest API value.")
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

	latestAPIVersion, err := attestationconfigapi.NewFetcher().FetchAzureSEVSNPVersionLatest(ctx)
	if err != nil {
		return fmt.Errorf("fetching latest version: %w", err)
	}

	isNewer, err := isInputNewerThanLatestAPI(inputVersion, latestAPIVersion.AzureSEVSNPVersion)
	if err != nil {
		return fmt.Errorf("comparing versions: %w", err)
	}
	if isNewer || force {
		if force {
			cmd.Println("Forcing upload of new version")
		} else {
			cmd.Printf("Input version: %+v is newer than latest API version: %+v\n", inputVersion, latestAPIVersion)
		}
		sut, sutClose, err := attestationconfigapi.NewClient(ctx, cfg, []byte(cosignPwd), []byte(privateKey), false, log())
		defer func() {
			if err := sutClose(ctx); err != nil {
				cmd.Printf("closing repo: %v\n", err)
			}
		}()
		if err != nil {
			return fmt.Errorf("creating repo: %w", err)
		}

		if err := sut.UploadAzureSEVSNP(ctx, inputVersion, time.Now()); err != nil {
			return fmt.Errorf("uploading version: %w", err)
		}
		cmd.Printf("Successfully uploaded new Azure SEV-SNP version: %+v\n", inputVersion)
	} else {
		cmd.Printf("Input version: %+v is not newer than latest API version: %+v\n", inputVersion, latestAPIVersion)
	}
	return nil
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
