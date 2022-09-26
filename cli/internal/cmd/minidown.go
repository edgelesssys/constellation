/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"os"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/spf13/cobra"
)

func newMiniDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Destroy a mini Constellation cluster",
		Long:  "Destroy a mini Constellation cluster.",
		Args:  cobra.ExactArgs(0),
		RunE:  runDown,
	}

	return cmd
}

func runDown(cmd *cobra.Command, args []string) error {
	if err := runTerminate(cmd, args); err != nil {
		return err
	}
	return os.Remove(constants.MasterSecretFilename)
}
