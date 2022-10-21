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

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cdbg",
		Short: "Constellation debugging client",
		Long: `cdbg is the constellation debugging client.
	It connects to Constellation instances running debugd and deploys a self-compiled version of the bootstrapper.`,
	}
	cmd.PersistentFlags().String("config", constants.ConfigFilename, "Constellation config file")
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
