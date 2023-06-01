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

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	attestationconfigclient "github.com/edgelesssys/constellation/v2/internal/api/attestationconfig/client"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfig/fetcher"

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
	versionFilePath string
	// Cosign credentials.
	cosignPwd      string
	privateKeyPath string
)

// Execute executes the root command.
func Execute() error {
	return newRootCmd().Execute()
}

// newRootCmd creates the root command.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "AWS_ACCESS_KEY_ID=$ID AWS_ACCESS_KEY=$KEY upload --version-file $FILE --cosign-pwd $PWD --private-key $FILE_PATH",
		Short: "Upload a set of versions specific to the azure-sev-snp attestation variant to the config api.",

		Long: "Upload a set of versions specific to the azure-sev-snp attestation variant to the config api. Please authenticate with AWS through your preferred method (e.g. environment variables, CLI) to be able to upload to S3.",
		RunE: runCmd,
	}
	rootCmd.PersistentFlags().StringVarP(&versionFilePath, "version-file", "f", "", "File path to the version json file.")
	rootCmd.PersistentFlags().StringVar(&cosignPwd, "cosign-pwd", "", "Cosign password used to decrpyt the private key. We use the release key to sign versions.")
	rootCmd.PersistentFlags().StringVar(&privateKeyPath, "private-key", "", "File path of private key used to sign the payload. We use the release key to sign versions.")
	must(enforceRequiredFlags(rootCmd, "version-file", "cosign-pwd", "private-key"))

	return rootCmd
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
	var inputVersion attestationconfig.AzureSEVSNPVersion
	err = json.Unmarshal(versionBytes, &inputVersion)
	if err != nil {
		return fmt.Errorf("unmarshalling version file: %w", err)
	}

	fetcher := fetcher.New()
	latestAPIVersion, err := fetcher.FetchAzureSEVSNPVersionLatest(ctx)
	if err != nil {
		return fmt.Errorf("fetching latest version: %w", err)
	}

	isNewer, err := IsInputNewerThanLatestAPI(inputVersion, latestAPIVersion.AzureSEVSNPVersion)
	if err != nil {
		return fmt.Errorf("comparing versions: %w", err)
	}
	if isNewer {
		privateKey, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return fmt.Errorf("reading private key: %w", err)
		}
		sut, sutClose, err := attestationconfigclient.New(ctx, cfg, []byte(cosignPwd), privateKey)
		defer func() {
			if err := sutClose(ctx); err != nil {
				fmt.Printf("closing repo: %v\n", err)
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

// IsInputNewerThanLatestAPI compares all version fields with the latest API version and returns true if any input field is newer.
func IsInputNewerThanLatestAPI(input, latest attestationconfig.AzureSEVSNPVersion) (bool, error) {
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
		if err := cmd.MarkPersistentFlagRequired(flag); err != nil {
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
