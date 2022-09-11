/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"github.com/spf13/cobra"
)

// NewUpgradeCmd returns a new cobra.Command for the upgrade command.
func NewUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Plan and perform an upgrade of a Constellation cluster",
		Long:  "Plan and perform an upgrade of a Constellation cluster.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(newUpgradePlanCmd())
	cmd.AddCommand(newUpgradeExecuteCmd())

	return cmd
}
