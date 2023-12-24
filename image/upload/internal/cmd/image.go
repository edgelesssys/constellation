/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"os"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/spf13/cobra"
)

// NewImageCmd creates a new image parent command. Image needs another
// verb, and does nothing on its own.
func NewImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Uploads OS images to supported CSPs",
		Long:  "Uploads OS images to supported CSPs.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.SetOut(os.Stdout)

	cmd.PersistentFlags().String("raw-image", "", "Path to os image in CSP specific format that should be uploaded.")
	cmd.PersistentFlags().Bool("secure-boot", false, "Enables secure boot support.")
	cmd.PersistentFlags().String("pki", "", "Base path to the PKI (secure boot signing) files.")
	cmd.PersistentFlags().String("attestation-variant", "", "Attestation variant of the image being uploaded.")
	cmd.PersistentFlags().String("version", "", "Shortname of the os image version.")
	cmd.PersistentFlags().String("timestamp", "", "Optional timestamp to use for resource names. Uses format 2006-01-02T15:04:05Z07:00.")
	cmd.PersistentFlags().String("region", "eu-central-1", "AWS region of the archive S3 bucket")
	cmd.PersistentFlags().String("bucket", "cdn-constellation-backend", "S3 bucket name of the archive")
	cmd.PersistentFlags().String("distribution-id", constants.CDNDefaultDistributionID, "CloudFront distribution ID of the API")
	cmd.PersistentFlags().String("out", "", "Optional path to write the upload result to. If not set, the result is written to stdout.")
	cmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	must(cmd.MarkPersistentFlagRequired("raw-image"))
	must(cmd.MarkPersistentFlagRequired("attestation-variant"))
	must(cmd.MarkPersistentFlagRequired("version"))

	cmd.AddCommand(newAWSCmd())
	cmd.AddCommand(newAzureCmd())
	cmd.AddCommand(newGCPCommand())
	cmd.AddCommand(newOpenStackCmd())
	cmd.AddCommand(newQEMUCmd())

	return cmd
}
