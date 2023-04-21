/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/spf13/cobra"
)

// NewQEMUCmd returns the command that uploads an OS image to QEMU.
func NewQEMUCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "qemu",
		Short: "Upload OS image to QEMU",
		Long:  "Upload OS image to QEMU.",
		Args:  cobra.ExactArgs(0),
		RunE:  runQEMU,
	}

	return cmd
}

func runQEMU(cmd *cobra.Command, args []string) error {
	return runNOP(cmd, cloudprovider.QEMU, args)
}
