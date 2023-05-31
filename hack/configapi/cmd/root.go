/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"encoding/json"
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
	versionFilePath string
	// Cosign credentials.
	cosignPwd      string
	privateKeyPath string
)

// Execute executes the root command.
func Execute() error {
	return NewRootCmd().Execute()
}

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	uploadCmd := &cobra.Command{
		Use:   "AWS_ACCESS_KEY_ID=$ID AWS_ACCESS_KEY=$KEY upload -b 2 -t 0 -s 6 -m 93 --cosign-pwd $PWD --private-key $FILE_PATH",
		Short: "Upload a set of versions specific to the azure-sev-snp attestation variant to the config api.",

		Long: "Upload a set of versions specific to the azure-sev-snp attestation variant to the config api. Please authenticate with AWS through your preferred method (e.g. environment variables, CLI) to be able to upload to S3.",
		RunE: runCmd,
	}
	uploadCmd.PersistentFlags().StringVarP(&versionFilePath, "version-file", "f", "", "File path to the version json file.")

	uploadCmd.PersistentFlags().StringVar(&cosignPwd, "cosign-pwd", "", "Cosign password used to decrpyt the private key.")

	uploadCmd.PersistentFlags().StringVar(&privateKeyPath, "private-key", "", "File path of private key used to sign the payload.")
	return uploadCmd
}

func runCmd(cmd *cobra.Command, _ []string) error {
	if err := enforceRequiredFlags(cmd, "version-file", "cosign-pwd", "private-key"); err != nil {
		return err
	}
	ctx := context.Background()
	cfg := staticupload.Config{
		Bucket: awsBucket,
		Region: awsRegion,
	}
	privateKey, err := getBytesFromFilePath(privateKeyPath)
	if err != nil {
		return fmt.Errorf("reading private key: %w", err)
	}

	versionBytes, err := getBytesFromFilePath(versionFilePath)
	if err != nil {
		return fmt.Errorf("reading version file: %w", err)
	}
	var versions configapi.AzureSEVSNPVersion
	err = json.Unmarshal(versionBytes, &versions)
	if err != nil {
		return fmt.Errorf("unmarshalling version file: %w", err)
	}

	sut, err := configapi.NewAttestationVersionRepo(ctx, cfg, []byte(cosignPwd), privateKey)
	if err != nil {
		return fmt.Errorf("creating repo: %w", err)
	}

	if err := sut.UploadAzureSEVSNP(ctx, versions, time.Now()); err != nil {
		return fmt.Errorf("uploading version: %w", err)
	}
	fmt.Printf("Successfully uploaded new Azure SEV-SNP version: %+v\n", versions)
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

func getBytesFromFilePath(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	return content, nil
}
