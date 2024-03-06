/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	infoupload "github.com/edgelesssys/constellation/v2/internal/osimage/imageinfo"
	"github.com/spf13/cobra"
)

// NewInfoCmd creates a new info parent command.
func NewInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [flags] <image-info.json>...",
		Short: "Uploads OS image info to S3",
		Long:  "Uploads OS image info to S3.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runInfo,
	}

	cmd.SetOut(os.Stdout)

	cmd.Flags().String("region", "eu-central-1", "AWS region of the archive S3 bucket")
	cmd.Flags().String("bucket", "cdn-constellation-backend", "S3 bucket name of the archive")
	cmd.Flags().String("distribution-id", constants.CDNDefaultDistributionID, "CloudFront distribution ID of the API")
	cmd.Flags().Bool("verbose", false, "Enable verbose output")

	return cmd
}

func runInfo(cmd *cobra.Command, args []string) error {
	workdir := os.Getenv("BUILD_WORKING_DIRECTORY")
	if len(workdir) > 0 {
		must(os.Chdir(workdir))
	}

	flags, err := parseS3Flags(cmd)
	if err != nil {
		return err
	}

	log := logger.NewTextLogger(flags.logLevel)
	log.Debug(fmt.Sprintf(
`Parsed flags:
  region: %q
  bucket: %q
  distribution-id: %q
`, flags.region, flags.bucket, flags.distributionID))
	info, err := readInfoArgs(args)
	if err != nil {
		return err
	}

	uploadC, uploadCClose, err := infoupload.New(cmd.Context(), flags.region, flags.bucket, flags.distributionID, log)
	if err != nil {
		return fmt.Errorf("uploading image info: %w", err)
	}
	defer func() {
		if err := uploadCClose(cmd.Context()); err != nil {
			log.Error(fmt.Sprintf("closing upload client: %v", err))
		}
	}()

	url, err := uploadC.Upload(cmd.Context(), info)
	if err != nil {
		return fmt.Errorf("uploading image info: %w", err)
	}
	log.Info(fmt.Sprintf("Uploaded image info to %s", url))
	return nil
}

func readInfoArgs(paths []string) (versionsapi.ImageInfo, error) {
	infos := make([]versionsapi.ImageInfo, len(paths))
	for i, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return versionsapi.ImageInfo{}, err
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&infos[i]); err != nil {
			return versionsapi.ImageInfo{}, err
		}
	}
	return versionsapi.MergeImageInfos(infos...)
}
