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
	gcpupload "github.com/edgelesssys/constellation/v2/internal/osimage/gcp"
	"github.com/spf13/cobra"
)

// NewGCPCommand returns the command that uploads an OS image to GCP.
func NewGCPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "Upload OS image to GCP",
		Long:  "Upload OS image to GCP.",
		Args:  cobra.ExactArgs(0),
		RunE:  runGCP,
	}

	cmd.Flags().String("gcp-project", "constellation-images", "GCP project to use")
	cmd.Flags().String("gcp-location", "europe-west3", "GCP location to use")
	cmd.Flags().String("gcp-bucket", "constellation-images", "GCP bucket to use")
	return cmd
}

func runGCP(cmd *cobra.Command, _ []string) error {
	workdir := os.Getenv("BUILD_WORKING_DIRECTORY")
	if len(workdir) > 0 {
		must(os.Chdir(workdir))
	}

	flags, err := parseGCPFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	archiveC, err := archive.New(cmd.Context(), flags.region, flags.bucket, log)
	if err != nil {
		return err
	}

	uploadC, err := gcpupload.New(cmd.Context(), flags.gcpProject, flags.gcpLocation, flags.gcpBucket, log)
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

	sbDatabase, uefiVarStore, err := loadSecureBootKeys(flags.pki)
	if err != nil {
		return err
	}

	uploadReq := &osimage.UploadRequest{
		Provider:     flags.provider,
		Version:      flags.version,
		Variant:      flags.variant,
		SBDatabase:   sbDatabase,
		UEFIVarStore: uefiVarStore,
		Size:         size,
		Timestamp:    flags.timestamp,
		Image:        file,
	}
	return uploadImage(cmd.Context(), archiveC, uploadC, uploadReq, out)
}
