/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/osimage"
	"github.com/edgelesssys/constellation/v2/internal/osimage/archive"
	nopupload "github.com/edgelesssys/constellation/v2/internal/osimage/nop"
	"github.com/spf13/cobra"
)

func runNOP(cmd *cobra.Command, provider cloudprovider.Provider, _ []string) error {
	workdir := os.Getenv("BUILD_WORKING_DIRECTORY")
	if len(workdir) > 0 {
		must(os.Chdir(workdir))
	}

	flags, err := parseCommonFlags(cmd)
	if err != nil {
		return err
	}
	flags.provider = provider
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

	uploadC := nopupload.New(log)

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
