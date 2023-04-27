/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

// oci-pin generates Go code and shasum files for OCI images.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

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
		Use:              "oci-pin",
		Short:            "Generate pinning artifacts for OCI images.",
		Long:             "Generate pinning artifacts (Go code, shasum files) for OCI images.",
		PersistentPreRun: preRunRoot,
	}

	rootCmd.SetOut(os.Stdout)

	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output.")

	rootCmd.AddCommand(newCodegenCmd())
	rootCmd.AddCommand(newSumCmd())
	rootCmd.AddCommand(newMergeCmd())

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

func splitRepoTag(ref string) (registry, prefix, name, tag string, err error) {
	// last colon is separator between name and tag
	tagSep := strings.LastIndexByte(ref, ':')
	if tagSep == -1 {
		return "", "", "", "", fmt.Errorf("invalid OCI image reference %q: missing tag", ref)
	}
	tag = ref[tagSep+1:]
	base := ref[:tagSep]

	// first slash is separator between registry and full name
	registrySep := strings.IndexByte(base, '/')
	if registrySep == -1 {
		return "", "", "", "", fmt.Errorf("invalid OCI image reference %q: missing registry", ref)
	}

	registry = base[:registrySep]
	fullName := base[registrySep+1:]

	// last slash is separator between prefix and short name
	nameSep := strings.LastIndexByte(fullName, '/')
	if nameSep == -1 {
		name = fullName
	} else {
		prefix = fullName[:nameSep]
		name = fullName[nameSep+1:]
	}
	return
}
