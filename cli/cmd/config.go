package cmd

import (
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Work with the Constellation configuration file",
		Long:  "Generate & manipulate configuration file for Constellation.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(newConfigGenerateCmd())

	return cmd
}
