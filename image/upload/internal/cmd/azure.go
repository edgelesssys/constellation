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
	azureupload "github.com/edgelesssys/constellation/v2/internal/osimage/azure"
	"github.com/spf13/cobra"
)

// newAzureCmd returns the command that uploads an OS image to Azure.
func newAzureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure",
		Short: "Upload OS image to Azure",
		Long:  "Upload OS image to Azure.",
		Args:  cobra.ExactArgs(0),
		RunE:  runAzure,
	}

	cmd.Flags().String("az-subscription", "0d202bbb-4fa7-4af8-8125-58c269a05435", "Azure subscription to use")
	cmd.Flags().String("az-location", "northeurope", "Azure location to use")
	cmd.Flags().String("az-resource-group", "constellation-images", "Azure resource group to use")
	return cmd
}

func runAzure(cmd *cobra.Command, _ []string) error {
	workdir := os.Getenv("BUILD_WORKING_DIRECTORY")
	if len(workdir) > 0 {
		must(os.Chdir(workdir))
	}

	flags, err := parseAzureFlags(cmd)
	if err != nil {
		return err
	}
	log := logger.New(logger.PlainLog, flags.logLevel)
	log.Debugf("Parsed flags: %+v", flags)

	archiveC, err := archive.New(cmd.Context(), flags.region, flags.bucket, log)
	if err != nil {
		return err
	}

	uploadC, err := azureupload.New(flags.azSubscription, flags.azLocation, flags.azResourceGroup, log)
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
		Provider:           flags.provider,
		Version:            flags.version,
		AttestationVariant: flags.attestationVariant,
		SBDatabase:         sbDatabase,
		UEFIVarStore:       uefiVarStore,
		Size:               size,
		Timestamp:          flags.timestamp,
		Image:              file,
	}
	return uploadImage(cmd.Context(), archiveC, uploadC, uploadReq, out)
}
