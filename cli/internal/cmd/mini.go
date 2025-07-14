/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package cmd

import "github.com/spf13/cobra"

// NewMiniCmd creates a new cobra.Command for managing MiniConstellation clusters.
func NewMiniCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mini",
		Short: "Manage MiniConstellation clusters",
		Long:  "Manage MiniConstellation clusters.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(newMiniUpCmd())
	cmd.AddCommand(newMiniDownCmd())

	return cmd
}
