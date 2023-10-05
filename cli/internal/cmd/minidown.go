/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newMiniDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Destroy a MiniConstellation cluster",
		Long:  "Destroy a MiniConstellation cluster.",
		Args:  cobra.ExactArgs(0),
		RunE:  runDown,
	}
	cmd.Flags().BoolP("yes", "y", false, "terminate the cluster without further confirmation")
	return cmd
}

func runDown(cmd *cobra.Command, args []string) error {
	if err := checkForMiniCluster(file.NewHandler(afero.NewOsFs())); err != nil {
		return fmt.Errorf("failed to destroy cluster: %w. Are you in the correct working directory?", err)
	}

	err := runTerminate(cmd, args)
	if removeErr := os.Remove(constants.MasterSecretFilename); removeErr != nil && !os.IsNotExist(removeErr) {
		err = errors.Join(err, removeErr)
	}
	return err
}

func checkForMiniCluster(fileHandler file.Handler) error {
	stateFile, err := state.ReadFromFile(fileHandler, constants.StateFilename)
	if err != nil {
		return fmt.Errorf("reading state file: %w", err)
	}

	if stateFile.Infrastructure.UID != constants.MiniConstellationUID {
		return errors.New("cluster is not a MiniConstellation cluster")
	}

	return nil
}
