/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: BUSL-1.1
*/

package cmd

import "github.com/spf13/cobra"

// NewIAMCmd returns a new cobra.Command for the iam parent command. It needs another verb and does nothing on its own.
func NewIAMCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Work with the IAM configuration on your cloud provider",
		Long:  "Work with the IAM configuration on your cloud provider.",
		Args:  cobra.ExactArgs(0),
	}

	cmd.AddCommand(newIAMCreateCmd())
	cmd.AddCommand(newIAMDestroyCmd())
	cmd.AddCommand(newIAMUpgradeCmd())
	return cmd
}
