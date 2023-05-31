/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	configapi "github.com/edgelesssys/constellation/v2/internal/api/attestationconfig"
	attestationconfigclient "github.com/edgelesssys/constellation/v2/internal/api/attestationconfig/client"
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
		Use:   "AWS_ACCESS_KEY_ID=$ID AWS_ACCESS_KEY=$KEY upload -b 2 -t 0 -s 6 -m 93 --cosign-pwd $PWD --private-key $FILE_PATH",
		Short: "Upload a set of versions specific to the azure-sev-snp attestation variant to the config api.",

		Long: "Upload a set of versions specific to the azure-sev-snp attestation variant to the config api. Please authenticate with AWS through your preferred method (e.g. environment variables, CLI) to be able to upload to S3.",
		RunE: runCmd,
	}
	rootCmd.PersistentFlags().StringVarP(&versionFilePath, "version-file", "f", "", "File path to the version json file.")
	rootCmd.PersistentFlags().StringVar(&cosignPwd, "cosign-pwd", "", "Cosign password used to decrpyt the private key.")
	rootCmd.PersistentFlags().StringVar(&privateKeyPath, "private-key", "", "File path of private key used to sign the payload.")
	must(enforceRequiredFlags(rootCmd, "version-file", "cosign-pwd", "private-key"))

	return rootCmd
}

func runCmd(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	cfg := staticupload.Config{
		Bucket: awsBucket,
		Region: awsRegion,
	}
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("reading private key: %w", err)
	}

	versionBytes, err := os.ReadFile(versionFilePath)
	if err != nil {
		return fmt.Errorf("reading version file: %w", err)
	}
	var versions configapi.AzureSEVSNPVersion
	err = json.Unmarshal(versionBytes, &versions)
	if err != nil {
		return fmt.Errorf("unmarshalling version file: %w", err)
	}

	sut, sutClose, err := attestationconfigclient.New(ctx, cfg, []byte(cosignPwd), privateKey)
	if err != nil {
		return fmt.Errorf("creating repo: %w", err)
	}
	defer func() {
		if err := sutClose(ctx); err != nil {
			fmt.Printf("closing repo: %v\n", err)
		}
	}()

	if err := sut.UploadAzureSEVSNP(ctx, versions, time.Now()); err != nil {
		return fmt.Errorf("uploading version: %w", err)
	}
	cmd.Printf("Successfully uploaded new Azure SEV-SNP version: %+v\n", versions)
	return nil
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
