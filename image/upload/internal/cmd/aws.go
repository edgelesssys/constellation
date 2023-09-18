/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
	"github.com/edgelesssys/constellation/v2/internal/osimage/archive"
	awsupload "github.com/edgelesssys/constellation/v2/internal/osimage/aws"
	"github.com/spf13/cobra"
)

// newAWSCmd returns the command that uploads an OS image to AWS.
func newAWSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "Upload OS image to AWS",
		Long:  "Upload OS image to AWS.",
		Args:  cobra.ExactArgs(0),
		RunE:  runAWS,
	}

	cmd.Flags().String("aws-region", "eu-central-1", "AWS region used during AMI creation")
	cmd.Flags().String("aws-bucket", "constellation-images", "S3 bucket used during AMI creation")
	return cmd
}

func runAWS(cmd *cobra.Command, _ []string) error {
	workdir := os.Getenv("BUILD_WORKING_DIRECTORY")
	if len(workdir) > 0 {
		must(os.Chdir(workdir))
	}

	flags, err := parseAWSFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	archiveC, archiveCClose, err := archive.New(cmd.Context(), flags.region, flags.bucket, flags.distributionID, log)
	if err != nil {
		return err
	}
	defer func() {
		if err := archiveCClose(cmd.Context()); err != nil {
			log.Errorf("closing archive client: %v", err)
		}
	}()

	uploadC, err := awsupload.New(flags.awsRegion, flags.awsBucket, log)
	if err != nil {
		return fmt.Errorf("uploading image: %w", err)
	}

	file, err := os.Open(flags.rawImage)
	if err != nil {
		return fmt.Errorf("uploading image: opening image file %w", err)
	}
	defer file.Close()
	size, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	if len(flags.out) > 0 {
		outF, err := os.Create(flags.out)
		if err != nil {
			return fmt.Errorf("uploading image: opening output file %w", err)
		}
		defer outF.Close()
		out = outF
	}

	uploadReq := &osimage.UploadRequest{
		Provider:           flags.provider,
		Version:            flags.version,
		AttestationVariant: flags.attestationVariant,
		SecureBoot:         flags.secureBoot,
		Size:               size,
		Timestamp:          flags.timestamp,
		Image:              file,
	}

	if flags.secureBoot {
		sbDatabase, uefiVarStore, err := loadSecureBootKeys(flags.pki)
		if err != nil {
			return err
		}
		uploadReq.SBDatabase = sbDatabase
		uploadReq.UEFIVarStore = uefiVarStore
	}

	return uploadImage(cmd.Context(), archiveC, uploadC, uploadReq, out)
}
