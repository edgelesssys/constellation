/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/osimage/measurementsuploader"
	"github.com/spf13/cobra"
)

// newMeasurementsUploadCmd creates a new upload command.
func newMeasurementsUploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Uploads OS image measurements to S3",
		Long:  "Uploads OS image measurements to S3.",
		Args:  cobra.ExactArgs(0),
		RunE:  runMeasurementsUpload,
	}

	cmd.SetOut(os.Stdout)

	cmd.Flags().String("measurements", "", "Path to measurements file to upload")
	cmd.Flags().String("signature", "", "Path to signature file to upload")
	cmd.Flags().String("region", "eu-central-1", "AWS region of the archive S3 bucket")
	cmd.Flags().String("bucket", "cdn-constellation-backend", "S3 bucket name of the archive")
	cmd.Flags().String("distribution-id", constants.CDNDefaultDistributionID, "CloudFront distribution ID of the API")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")

	must(cmd.MarkFlagRequired("measurements"))
	must(cmd.MarkFlagRequired("signature"))

	return cmd
}

func runMeasurementsUpload(cmd *cobra.Command, _ []string) error {
	workdir := os.Getenv("BUILD_WORKING_DIRECTORY")
	if len(workdir) > 0 {
		must(os.Chdir(workdir))
	}

	flags, err := parseUploadMeasurementsFlags(cmd)
	if err != nil {
		return err
	}

	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	uploadC, uploadCClose, err := measurementsuploader.New(cmd.Context(), flags.region, flags.bucket, flags.distributionID, log)
	if err != nil {
		return fmt.Errorf("uploading image info: %w", err)
	}
	defer func() {
		if err := uploadCClose(cmd.Context()); err != nil {
			log.Errorf("closing upload client: %v", err)
		}
	}()

	measurements, err := os.Open(flags.measurementsPath)
	if err != nil {
		return fmt.Errorf("uploading image measurements: opening measurements file: %w", err)
	}
	defer measurements.Close()
	signature, err := os.Open(flags.signaturePath)
	if err != nil {
		return fmt.Errorf("uploading image measurements: opening signature file: %w", err)
	}
	defer signature.Close()

	measurementsURL, signatureURL, err := uploadC.Upload(cmd.Context(), measurements, signature)
	if err != nil {
		return fmt.Errorf("uploading image info: %w", err)
	}
	log.Infof("Uploaded image measurements to %s (and signature to %s)", measurementsURL, signatureURL)
	return nil
}
