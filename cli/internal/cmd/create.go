/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/spf13/cobra"
)

// NewCreateCmd returns a new cobra.Command for the create command.
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create instances on a cloud platform for your Constellation cluster",
		Long:  "Create instances on a cloud platform for your Constellation cluster.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Bool("conformance", false, "")
			cmd.Flags().Bool("skip-helm-wait", false, "")
			cmd.Flags().Bool("merge-kubeconfig", false, "")
			cmd.Flags().Duration("timeout", 5*time.Minute, "")
			// Skip all phases but the infrastructure phase.
			cmd.Flags().StringSlice("skip-phases", allPhases(skipInfrastructurePhase), "")
			return runApply(cmd, args)
		},
		Deprecated: "use 'constellation apply' instead.",
	}
	cmd.Flags().BoolP("yes", "y", false, "create the cluster without further confirmation")
	return cmd
}

func isPlural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// validateCLIandConstellationVersionAreEqual checks if the image and microservice version are equal (down to patch level) to the CLI version.
func validateCLIandConstellationVersionAreEqual(cliVersion semver.Semver, imageVersion string, microserviceVersion semver.Semver) error {
	parsedImageVersion, err := versionsapi.NewVersionFromShortPath(imageVersion, versionsapi.VersionKindImage)
	if err != nil {
		return fmt.Errorf("parsing image version: %w", err)
	}

	semImage, err := semver.New(parsedImageVersion.Version())
	if err != nil {
		return fmt.Errorf("parsing image semantical version: %w", err)
	}

	if !cliVersion.MajorMinorEqual(semImage) {
		return fmt.Errorf("image version %q does not match the major and minor version of the cli version %q", semImage.String(), cliVersion.String())
	}
	if cliVersion.Compare(microserviceVersion) != 0 {
		return fmt.Errorf("cli version %q does not match microservice version %q", cliVersion.String(), microserviceVersion.String())
	}
	return nil
}
