//go:build !windows
// +build !windows

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
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

	return cmd
}

func runDown(cmd *cobra.Command, args []string) error {
	return cloudcmd.ErrQEMUTerminationNotSupportedOnPlatform
}
