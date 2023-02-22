/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"github.com/spf13/cobra"
)

// NewConfigCmd creates a new config parent command. Config needs another
// verb, and does nothing on its own.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Work with the Constellation configuration file",
		Long:  "Work with the Constellation configuration file.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(newConfigGenerateCmd())
	cmd.AddCommand(newConfigFetchMeasurementsCmd())
	cmd.AddCommand(newConfigInstanceTypesCmd())
	cmd.AddCommand(newConfigKubernetesVersionsCmd())

	return cmd
}
