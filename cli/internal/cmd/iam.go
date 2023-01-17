/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"github.com/spf13/cobra"
)

// NewIAMCmd returns a new cobra.Command for the iam parent command. It needs another verb and does nothing on its own.
func NewIAMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Work with the IAM configuration on your cloud provider",
		Long:  "Work with the IAM configuration on your cloud provider.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(newIAMCreateCmd())

	return cmd
}

// NewIAMCreateCmd returns a new cobra.Command for the iam create parent command. It needs another verb, and does nothing on its own.
func newIAMCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create IAM configuration on a cloud platform for your Constellation cluster",
		Long:  "Create IAM configuration on a cloud platform for your Constellation cluster.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.PersistentFlags().Bool("generate-config", false, "automatically generate a configuration file and fill in the required fields")

	cmd.AddCommand(newIAMCreateAWSCmd())
	cmd.AddCommand(newIAMCreateAzureCmd())
	cmd.AddCommand(newIAMCreateGCPCmd())

	return cmd
}
