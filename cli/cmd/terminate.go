package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"

	"github.com/edgelesssys/constellation/cli/cloud/cloudcmd"
	"github.com/edgelesssys/constellation/cli/file"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/state"
)

func newTerminateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Terminate an existing Constellation cluster",
		Long:  "Terminate an existing Constellation cluster. The cluster can't be started again, and all persistent storage will be lost.",
		Args:  cobra.NoArgs,
		RunE:  runTerminate,
	}
	return cmd
}

// runTerminate runs the terminate command.
func runTerminate(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	terminator := cloudcmd.NewTerminator()

	return terminate(cmd, terminator, fileHandler)
}

func terminate(cmd *cobra.Command, terminator cloudTerminator, fileHandler file.Handler) error {
	var stat state.ConstellationState
	if err := fileHandler.ReadJSON(constants.StateFilename, &stat); err != nil {
		return err
	}

	cmd.Println("Terminating ...")

	if err := terminator.Terminate(cmd.Context(), stat); err != nil {
		return err
	}

	cmd.Println("Your Constellation cluster was terminated successfully.")

	var retErr error
	if err := fileHandler.Remove(constants.StateFilename); err != nil {
		retErr = multierr.Append(err, fmt.Errorf("failed to remove file '%s', please remove manually", constants.StateFilename))
	}

	if err := fileHandler.Remove(constants.AdminConfFilename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		retErr = multierr.Append(err, fmt.Errorf("failed to remove file '%s', please remove manually", constants.AdminConfFilename))
	}

	if err := fileHandler.Remove(constants.WGQuickConfigFilename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		retErr = multierr.Append(err, fmt.Errorf("failed to remove file '%s', please remove manually", constants.WGQuickConfigFilename))
	}

	return retErr
}
