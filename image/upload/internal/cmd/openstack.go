/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/spf13/cobra"
)

// NewOpenStackCmd returns the command that uploads an OS image to OpenStack.
func NewOpenStackCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openstack",
		Short: "Upload OS image to OpenStack",
		Long:  "Upload OS image to OpenStack.",
		Args:  cobra.ExactArgs(0),
		RunE:  runOpenStack,
	}

	return cmd
}

func runOpenStack(cmd *cobra.Command, args []string) error {
	return runNOP(cmd, cloudprovider.OpenStack, args)
}
