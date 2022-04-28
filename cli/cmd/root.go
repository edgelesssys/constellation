package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "constellation",
	Short:        "Set up your Constellation cluster.",
	Long:         "Set up your Constellation cluster.",
	SilenceUsage: true,
}

// Execute starts the CLI.
func Execute() error {
	ctx, cancel := signalContext(context.Background(), os.Interrupt)
	defer cancel()
	return rootCmd.ExecuteContext(ctx)
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

func init() {
	cobra.EnableCommandSorting = false
	// Set output of cmd.Print to stdout.
	rootCmd.SetOut(os.Stdout)
	// Disable --no-description flag for completion command.
	rootCmd.CompletionOptions.DisableNoDescFlag = true
	rootCmd.PersistentFlags().String("dev-config", "", "Set this flag to create the Constellation using settings from a development config.")
	rootCmd.AddCommand(newCreateCmd())
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newVerifyCmd())
	rootCmd.AddCommand(newTerminateCmd())
	rootCmd.AddCommand(newVersionCmd())
}
