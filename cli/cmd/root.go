/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package cmd is the entrypoint of the Constellation CLI.

Business logic of the CLI shall be implemented in the internal/cmd package.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/edgelesssys/constellation/v2/cli/internal/cmd"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/spf13/cobra"
)

// Execute starts the CLI.
func Execute() error {
	cobra.EnableCommandSorting = false
	rootCmd := NewRootCmd()
	ctx, cancel := signalContext(context.Background(), os.Interrupt)
	defer cancel()
	return rootCmd.ExecuteContext(ctx)
}

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "constellation",
		Short:            "Manage your Constellation cluster",
		Long:             "Manage your Constellation cluster.",
		PersistentPreRun: preRunRoot,
	}

	// Set output of cmd.Print to stdout. (By default, it's stderr.)
	rootCmd.SetOut(os.Stdout)

	rootCmd.PersistentFlags().String("config", constants.ConfigFilename, "path to the configuration file")
	must(rootCmd.MarkPersistentFlagFilename("config", "yaml"))

	rootCmd.PersistentFlags().Bool("debug", false, "enable debug logging")
	rootCmd.PersistentFlags().Bool("force", false, "disable version compatibility checks - might result in corrupted clusters")

	rootCmd.AddCommand(cmd.NewConfigCmd())
	rootCmd.AddCommand(cmd.NewCreateCmd())
	rootCmd.AddCommand(cmd.NewInitCmd())
	rootCmd.AddCommand(cmd.NewMiniCmd())
	rootCmd.AddCommand(cmd.NewVerifyCmd())
	rootCmd.AddCommand(cmd.NewUpgradeCmd())
	rootCmd.AddCommand(cmd.NewRecoverCmd())
	rootCmd.AddCommand(cmd.NewTerminateCmd())
	rootCmd.AddCommand(cmd.NewVersionCmd())
	rootCmd.AddCommand(cmd.NewIAMCmd())

	return rootCmd
}

// signalContext returns a context that is canceled on the handed signal.
// The signal isn't watched after its first occurrence. Call the cancel
// function to ensure the internal goroutine is stopped and the signal isn't
// watched any longer.
func signalContext(ctx context.Context, sig os.Signal) (context.Context, context.CancelFunc) {
	sigCtx, stop := signal.NotifyContext(ctx, sig)
	done := make(chan struct{}, 1)
	stopDone := make(chan struct{}, 1)

	go func() {
		defer func() { stopDone <- struct{}{} }()
		defer stop()
		select {
		case <-sigCtx.Done():
			fmt.Println(" Signal caught. Press ctrl+c again to terminate the program immediately.")
		case <-done:
		}
	}()

	cancelFunc := func() {
		done <- struct{}{}
		<-stopDone
	}

	return sigCtx, cancelFunc
}

func preRunRoot(cmd *cobra.Command, _ []string) {
	cmd.SilenceUsage = true
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
