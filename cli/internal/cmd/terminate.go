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

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/cmd/pathprefix"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// NewTerminateCmd returns a new cobra.Command for the terminate command.
func NewTerminateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Terminate a Constellation cluster",
		Long: "Terminate a Constellation cluster.\n\n" +
			"The cluster can't be started again, and all persistent storage will be lost.",
		Args: cobra.NoArgs,
		RunE: runTerminate,
	}
	cmd.Flags().BoolP("yes", "y", false, "terminate the cluster without further confirmation")
	return cmd
}

// runTerminate runs the terminate command.
func runTerminate(cmd *cobra.Command, _ []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	spinner, err := newSpinnerOrStderr(cmd)
	if err != nil {
		return fmt.Errorf("creating spinner: %w", err)
	}
	defer spinner.Stop()
	terminator := cloudcmd.NewTerminator()

	return terminate(cmd, terminator, fileHandler, spinner)
}

func terminate(cmd *cobra.Command, terminator cloudTerminator, fileHandler file.Handler, spinner spinnerInterf,
) error {
	flags, err := parseTerminateFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}
	pf := pathprefix.New(flags.workspace)

	if !flags.yes {
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
	err = terminator.Terminate(cmd.Context(), constants.TerraformWorkingDir, flags.logLevel)
	spinner.Stop()
	if err != nil {
		return fmt.Errorf("terminating Constellation cluster: %w", err)
	}

	cmd.Println("Your Constellation cluster was terminated successfully.")

	var removeErr error
	if err := fileHandler.Remove(constants.AdminConfFilename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		removeErr = errors.Join(err, fmt.Errorf("failed to remove file: '%s', please remove it manually", pf.PrefixPrintablePath(constants.AdminConfFilename)))
	}

	if err := fileHandler.Remove(constants.StateFilename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		removeErr = errors.Join(err, fmt.Errorf("failed to remove file: '%s', please remove it manually", pf.PrefixPrintablePath(constants.StateFilename)))
	}

	return removeErr
}

type terminateFlags struct {
	yes       bool
	workspace string
	logLevel  terraform.LogLevel
}

func parseTerminateFlags(cmd *cobra.Command) (terminateFlags, error) {
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return terminateFlags{}, fmt.Errorf("parsing yes bool: %w", err)
	}
	logLevelString, err := cmd.Flags().GetString("tf-log")
	if err != nil {
		return terminateFlags{}, fmt.Errorf("parsing tf-log string: %w", err)
	}
	logLevel, err := terraform.ParseLogLevel(logLevelString)
	if err != nil {
		return terminateFlags{}, fmt.Errorf("parsing Terraform log level %s: %w", logLevelString, err)
	}
	workspace, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return terminateFlags{}, fmt.Errorf("parsing workspace string: %w", err)
	}

	return terminateFlags{
		yes:       yes,
		workspace: workspace,
		logLevel:  logLevel,
	}, nil
}
