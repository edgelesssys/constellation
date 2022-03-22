package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cdbg",
	Short: "Constellation debugging client",
	Long: `cdbg is the constellation debugging client.
It connects to CoreOS instances running debugd and deploys a self-compiled version of the coordinator.`,
}

// Execute starts the CLI.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("dev-config", "", "debugd config file (required)")
	rootCmd.MarkPersistentFlagRequired("dev-config")
}
