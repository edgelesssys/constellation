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
	"github.com/spf13/pflag"

	"github.com/edgelesssys/constellation/v2/internal/cloudcmd"
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

type terminateFlags struct {
	rootFlags
	yes bool
}

func (f *terminateFlags) parse(flags *pflag.FlagSet) error {
	if err := f.rootFlags.parse(flags); err != nil {
		return err
	}

	yes, err := flags.GetBool("yes")
	if err != nil {
		return fmt.Errorf("getting 'yes' flag: %w", err)
	}
	f.yes = yes
	return nil
}

// runTerminate runs the terminate command.
func runTerminate(cmd *cobra.Command, _ []string) error {
	spinner, err := newSpinnerOrStderr(cmd)
	if err != nil {
		return fmt.Errorf("creating spinner: %w", err)
	}
	defer spinner.Stop()
	terminator := cloudcmd.NewTerminator()

	logger, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}

	t := &terminateCmd{log: logger, fileHandler: file.NewHandler(afero.NewOsFs())}
	if err := t.flags.parse(cmd.Flags()); err != nil {
		return err
	}

	return t.terminate(cmd, terminator, spinner)
}

type terminateCmd struct {
	log         debugLog
	fileHandler file.Handler
	flags       terminateFlags
}

func (t *terminateCmd) terminate(cmd *cobra.Command, terminator cloudTerminator, spinner spinnerInterf) error {
	if !t.flags.yes {
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
	err := terminator.Terminate(cmd.Context(), constants.TerraformWorkingDir, t.flags.tfLogLevel)
	spinner.Stop()
	if err != nil {
		return fmt.Errorf("terminating Constellation cluster: %w", err)
	}

	cmd.Println("Your Constellation cluster was terminated successfully.")

	var removeErr error
	if err := t.fileHandler.Remove(constants.AdminConfFilename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		removeErr = errors.Join(err, fmt.Errorf("failed to remove file: '%s', please remove it manually", t.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename)))
	}

	if err := t.fileHandler.Remove(constants.StateFilename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		removeErr = errors.Join(err, fmt.Errorf("failed to remove file: '%s', please remove it manually", t.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename)))
	}

	return removeErr
}
