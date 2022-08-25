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

	cmd.AddCommand(newUpgradeExecuteCmd())

	return cmd
}
