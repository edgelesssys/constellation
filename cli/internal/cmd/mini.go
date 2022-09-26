/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import "github.com/spf13/cobra"

// NewMiniCmd creates a new cobra.Command for managing mini Constellation clusters.
func NewMiniCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mini",
		Short: "Manage mini Constellation clusters",
		Long:  "Manage mini Constellation clusters.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(newMiniUpCmd())
	cmd.AddCommand(newMiniDownCmd())

	return cmd
}
