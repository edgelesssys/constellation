/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
	"github.com/edgelesssys/constellation/v2/internal/osimage/archive"
	nopupload "github.com/edgelesssys/constellation/v2/internal/osimage/nop"
	uplosiupload "github.com/edgelesssys/constellation/v2/internal/osimage/uplosi"
	"github.com/spf13/cobra"
)

// NewUplosiCmd returns the command that uses uplosi for uploading os images.
func NewUplosiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uplosi",
		Short: "Templates uplosi configuration files",
		Long:  "Templates uplosi configuration files.",
		Args:  cobra.ExactArgs(0),
		RunE:  runUplosi,
	}

	cmd.SetOut(os.Stdout)

	cmd.Flags().String("raw-image", "", "Path to os image in CSP specific format that should be uploaded.")
	cmd.Flags().String("attestation-variant", "", "Attestation variant of the image being uploaded.")
	cmd.Flags().String("csp", "", "Cloud service provider that we want to upload this image to. If not set, the csp is guessed from the raw-image file name.")
	cmd.Flags().String("ref", "", "Ref of the OS image (part of image shortname).")
	cmd.Flags().String("stream", "", "Stream of the OS image (part of the image shortname).")
	cmd.Flags().String("version", "", "Semantic version of the os image (part of the image shortname).")
	cmd.Flags().String("region", "eu-central-1", "AWS region of the archive S3 bucket")
	cmd.Flags().String("bucket", "cdn-constellation-backend", "S3 bucket name of the archive")
	cmd.Flags().String("distribution-id", "E1H77EZTHC3NE4", "CloudFront distribution ID of the API")
	cmd.Flags().String("out", "", "Optional path to write the upload result to. If not set, the result is written to stdout.")
	cmd.Flags().String("uplosi-path", "uplosi", "Path to the uplosi binary.")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")
	must(cmd.MarkFlagRequired("raw-image"))
	must(cmd.MarkFlagRequired("version"))
	must(cmd.MarkFlagRequired("ref"))

	return cmd
}

func runUplosi(cmd *cobra.Command, _ []string) error {
	flags, err := parseUplosiFlags(cmd)
	if err != nil {
		return err
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: flags.logLevel}))
	log.Debug(fmt.Sprintf("Parsed flags: %+v", flags))

	archiveC, archiveCClose, err := archive.New(cmd.Context(), flags.region, flags.bucket, flags.distributionID, log)
	if err != nil {
		return err
	}
	defer func() {
		if err := archiveCClose(cmd.Context()); err != nil {
			log.Error(fmt.Sprintf("closing archive client: %v", err))
		}
	}()

	var uploadC uploader
	switch flags.provider {
	case cloudprovider.AWS, cloudprovider.Azure, cloudprovider.GCP:
		uploadC = uplosiupload.New(flags.uplosiPath, log)
	default:
		uploadC = nopupload.New(log)
	}

	imageOpener := func() (io.ReadSeekCloser, error) {
		return os.Open(flags.rawImage)
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
		Timestamp:          getTimestamp(),
		ImageReader:        imageOpener,
		ImagePath:          flags.rawImage,
	}

	return uploadImage(cmd.Context(), archiveC, uploadC, uploadReq, out)
}

func getTimestamp() time.Time {
	epoch := os.Getenv("SOURCE_DATE_EPOCH")
	epochSecs, err := strconv.ParseInt(epoch, 10, 64)
	if epoch == "" || err != nil {
		return time.Now().UTC()
	}
	return time.Unix(epochSecs, 0).UTC()
}
