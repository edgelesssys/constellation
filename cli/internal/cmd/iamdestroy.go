/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/spf13/cobra"
)

func runDestroyIAMUser(cmd *cobra.Command, _args []string) error {
	spinner := newSpinner(cmd.ErrOrStderr())
	destroyer := cloudcmd.NewIAMDestroyer(cmd.Context())

	return destroyIAMUser(cmd, spinner, destroyer)
}

func destroyIAMUser(cmd *cobra.Command, spinner spinnerInterf, destroyer iamDestroyer) error {
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	if !yes {
		// Confirmation
		ok, err := askToConfirm(cmd, "Do you really want to destroy your IAM user?")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("The destruction of the IAM user was aborted")
			return nil
		}
	}

	spinner.Start("Destroying IAM User", false)

	if err := destroyer.DestroyIAMUser(cmd.Context()); err != nil {
		return fmt.Errorf("Couldn't destroy IAM User: %w", err)
	}
	spinner.Stop()
	fmt.Println("Successfully destroyed IAM User")
	return nil
}
