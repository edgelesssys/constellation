/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/edgelesssys/constellation/v2/internal/maa"
	"github.com/spf13/cobra"
)

// NewMaaPatchCmd returns a new cobra.Command for the maa-patch command.
func NewMaaPatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maa-patch <attestation-url>",
		Short: "Patch the MAA's attestation policy",
		Long:  "Patch the MAA's attestation policy.",
		Args: cobra.MatchAll(
			cobra.ExactArgs(1),
			func(_ *cobra.Command, args []string) error {
				if _, err := url.Parse(args[0]); err != nil {
					return fmt.Errorf("argument %s is not a valid URL: %w", args[0], err)
				}
				return nil
			},
		),
		RunE:   runPatchMAA,
		Hidden: true, // we don't want to show this command to the user directly.
	}

	return cmd
}

type maaPatchCmd struct {
	log     debugLog
	patcher patcher
}

func runPatchMAA(cmd *cobra.Command, args []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	p := maa.NewAzurePolicyPatcher()

	c := &maaPatchCmd{log: log, patcher: p}

	return c.patchMAA(cmd, args[0])
}

func (c *maaPatchCmd) patchMAA(cmd *cobra.Command, attestationURL string) error {
	c.log.Debug(fmt.Sprintf("Using attestation URL %q", attestationURL))

	if err := c.patcher.Patch(cmd.Context(), attestationURL); err != nil {
		return fmt.Errorf("patching MAA attestation policy: %w", err)
	}

	return nil
}

type patcher interface {
	Patch(ctx context.Context, attestationURL string) error
}
