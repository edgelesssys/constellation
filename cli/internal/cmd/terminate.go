/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// NewTerminateCmd returns a new cobra.Command for the terminate command.
func NewTerminateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Terminate a Constellation cluster",
		Long:  "Terminate a Constellation cluster. The cluster can't be started again, and all persistent storage will be lost.",
		Args:  cobra.NoArgs,
		RunE:  runTerminate,
	}
	cmd.Flags().BoolP("yes", "y", false, "terminate the cluster without further confirmation")
	return cmd
}

// runTerminate runs the terminate command.
func runTerminate(cmd *cobra.Command, args []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	spinner := newSpinner(cmd.ErrOrStderr())
	defer spinner.Stop()
	terminator := cloudcmd.NewTerminator()

	return terminate(cmd, terminator, fileHandler, spinner)
}

func terminate(cmd *cobra.Command, terminator cloudTerminator, fileHandler file.Handler, spinner spinnerInterf,
) error {
	yesFlag, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	if !yesFlag {
		cmd.Println("You are about to terminate a Constellation cluster.")
		cmd.Println("All of its associated resources will be DESTROYED.")
		cmd.Println("This action is irreversible and ALL DATA WILL BE LOST.")
		ok, err := askToConfirm(cmd, "Do you want to continue?")
		if err != nil {
			return err
		}
		if !ok {
			cmd.Println("The termination of the cluster was aborted.")
			return nil
		}
	}

	spinner.Start("Terminating", false)
	err = terminator.Terminate(cmd.Context())
	spinner.Stop()
	if err != nil {
		return fmt.Errorf("terminating Constellation cluster: %w", err)
	}

	cmd.Println("Your Constellation cluster was terminated successfully.")

	var retErr error
	if err := fileHandler.Remove(constants.AdminConfFilename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		retErr = multierr.Append(err, fmt.Errorf("failed to remove file: '%s', please remove it manually", constants.AdminConfFilename))
	}

	if err := fileHandler.Remove(constants.ClusterIDsFileName); err != nil && !errors.Is(err, fs.ErrNotExist) {
		retErr = multierr.Append(err, fmt.Errorf("failed to remove file: '%s', please remove it manually", constants.ClusterIDsFileName))
	}

	return retErr
}
