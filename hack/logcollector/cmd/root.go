/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "loco",
		Short: "Constellation logcollection client",
		Long: `LoCo is the Constellation LOgCOllection client.
It prepares filebeat and logstash configurations for deployment.`,
	}

	cmd.AddCommand(newTemplateCmd())

	return cmd
}

// Execute starts the CLI.
func Execute() error {
	return newRootCmd().Execute()
}
