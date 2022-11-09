/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
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
		err = multierr.Append(err, removeErr)
	}
	return err
}

func checkForMiniCluster(fileHandler file.Handler) error {
	var idFile clusterid.File
	if err := fileHandler.ReadJSON(constants.ClusterIDsFileName, &idFile); err != nil {
		return err
	}
	if idFile.CloudProvider != cloudprovider.QEMU {
		return errors.New("cluster is not a QEMU based Constellation")
	}
	if idFile.UID != "mini" {
		return errors.New("cluster is not a MiniConstellation cluster")
	}

	return nil
}
