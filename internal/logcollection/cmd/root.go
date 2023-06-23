/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"os"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logco",
		Short: "Constellation log collection client (logco)",
		Long:  "logco manages log collection via Logstash and Filebeat to the central OpenSearch log platform.",
	}

	cmd.PersistentFlags().String("config", constants.ConfigFilename, "(required) Constellation config file")
	must(cobra.MarkFlagRequired(cmd.PersistentFlags(), "config"))

	cmd.AddCommand(newPrepareCmd())

	return cmd
}

// Execute starts the CLI.
func Execute() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
