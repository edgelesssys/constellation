//go:build windows
// +build windows

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func newMiniUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Create and initialize a new MiniConstellation cluster",
		Long: "Create and initialize a new MiniConstellation cluster.\n" +
			"A mini cluster consists of a single control-plane and worker node, hosted using QEMU/KVM.\n",
		Args: cobra.ExactArgs(0),
		RunE: runUp,
	}

	// override global flag so we don't have a default value for the config
	cmd.Flags().String("config", "", "path to the config file to use for the cluster")

	return cmd
}

func runUp(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s", runtime.GOOS, runtime.GOARCH)
}
