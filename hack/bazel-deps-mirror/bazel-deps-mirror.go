/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// bazel-deps-mirror adds external dependencies to edgeless systems' mirror.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

func main() {
	if err := execute(); err != nil {
		os.Exit(1)
	}
}

func execute() error {
	rootCmd := newRootCmd()
	ctx, cancel := signalContext(context.Background(), os.Interrupt)
	defer cancel()
	return rootCmd.ExecuteContext(ctx)
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:              "bazel-deps-mirror",
		Short:            "Add external Bazel dependencies to edgeless systems' mirror.",
		Long:             "Add external Bazel dependencies to edgeless systems' mirror.",
		PersistentPreRun: preRunRoot,
	}

	rootCmd.SetOut(os.Stdout)

	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("region", "eu-central-1", "AWS region of the API S3 bucket")
	rootCmd.PersistentFlags().String("bucket", "cdn-constellation-backend", "S3 bucket name of the API")
	rootCmd.PersistentFlags().String("mirror-base-url", "https://cdn.confidential.cloud", "Base URL of the public mirror endpoint")

	rootCmd.AddCommand(newCheckCmd())
	rootCmd.AddCommand(newFixCmd())

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
