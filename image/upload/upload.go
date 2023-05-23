/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// upload uploads os images.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/edgelesssys/constellation/v2/image/upload/internal/cmd"
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
		Use:              "upload",
		Short:            "Uploads OS images to supported CSPs",
		Long:             "Uploads OS images to supported CSPs.",
		PersistentPreRun: preRunRoot,
	}

	rootCmd.SetOut(os.Stdout)

	rootCmd.PersistentFlags().String("raw-image", "", "Path to os image in CSP specific format that should be uploaded.")
	rootCmd.PersistentFlags().String("pki", "", "Base path to the PKI (secure boot signing) files.")
	rootCmd.PersistentFlags().String("attestation-variant", "", "Attestation variant of the image being uploaded.")
	rootCmd.PersistentFlags().String("version", "", "Shortname of the os image version.")
	rootCmd.PersistentFlags().String("timestamp", "", "Optional timestamp to use for resource names. Uses format 2006-01-02T15:04:05Z07:00.")
	rootCmd.PersistentFlags().String("region", "eu-central-1", "AWS region of the archive S3 bucket")
	rootCmd.PersistentFlags().String("bucket", "cdn-constellation-backend", "S3 bucket name of the archive")
	rootCmd.PersistentFlags().String("out", "", "Optional path to write the upload result to. If not set, the result is written to stdout.")
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	must(rootCmd.MarkPersistentFlagRequired("raw-image"))
	must(rootCmd.MarkPersistentFlagRequired("attestation-variant"))
	must(rootCmd.MarkPersistentFlagRequired("version"))

	rootCmd.AddCommand(cmd.NewAWSCmd())
	rootCmd.AddCommand(cmd.NewAzureCmd())
	rootCmd.AddCommand(cmd.NewGCPCommand())
	rootCmd.AddCommand(cmd.NewOpenStackCmd())
	rootCmd.AddCommand(cmd.NewQEMUCmd())

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
