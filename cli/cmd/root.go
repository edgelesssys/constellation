package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/spf13/cobra"
)

// Execute starts the CLI.
func Execute() error {
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
	must(rootCmd.MarkPersistentFlagFilename("config", "json"))

	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newVerifyCmd())
	rootCmd.AddCommand(newRecoverCmd())
	rootCmd.AddCommand(newTerminateCmd())
	rootCmd.AddCommand(newVersionCmd())

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

func preRunRoot(cmd *cobra.Command, args []string) {
	cmd.SilenceUsage = true
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
