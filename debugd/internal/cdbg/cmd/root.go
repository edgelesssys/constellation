/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package cmd contains the cdbg CLI.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cdbg",
		Short: "Constellation debugging client",
		Long: `cdbg is the constellation debugging client.
	It connects to Constellation instances running debugd and deploys a self-compiled version of the bootstrapper.`,
		PersistentPreRunE: preRunRoot,
	}
	cmd.PersistentFlags().StringP("workspace", "C", "", "path to the Constellation workspace")
	cmd.PersistentFlags().Bool("force", false, "disables version validation errors - might result in corrupted clusters")

	must(cmd.MarkPersistentFlagDirname("workspace"))

	cmd.AddCommand(newDeployCmd())
	return cmd
}

// Execute starts the CLI.
func Execute() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func preRunRoot(cmd *cobra.Command, _ []string) error {
	cmd.SilenceUsage = true

	workspace, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return fmt.Errorf("getting workspace flag: %w", err)
	}

	// Change to workspace directory if set.
	if workspace != "" {
		if err := os.Chdir(workspace); err != nil {
			return fmt.Errorf("changing from current directory to workspace %q: %w", workspace, err)
		}
	}

	return nil
}

func must(err error) {
	if err == nil {
		return
	}
	panic(err)
}
