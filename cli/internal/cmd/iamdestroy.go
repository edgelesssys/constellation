/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
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

	fsHandler := file.NewHandler(afero.NewOsFs())
	if _, err := fsHandler.Stat("gcpServiceAccountKey.json"); err == nil {
		if !yes {
			ok, err := askToConfirm(cmd, "There seems to be a gcpServiceAccountKey.json file. Do you want to delete it? (Note that you may not be able to save generated private keys for GCP if the file exists)")
			if err != nil {
				return err
			}
			if ok {
				destroyed, err := destroyer.DeleteGCPServiceAccountKeyFile(cmd.Context())
				if err != nil {
					return err
				}
				if !destroyed {
					ok, err := askToConfirm(cmd, "The file gcpServiceAccountKey.json could not be deleted. Either it does not exist or the file belongs to another IAM user. Do you want to proceed anyway?")
					if err != nil {
						return err
					}
					if !ok {
						return nil
					}
				}
			}
		} else {
			destroyed, err := destroyer.DeleteGCPServiceAccountKeyFile(cmd.Context())
			if err != nil {
				return err
			}
			if !destroyed {
				ok, err := askToConfirm(cmd, "The file gcpServiceAccountKey.json could not be deleted. Either it does not exist or the file belongs to another IAM user. Do you want to proceed anyway?")
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}
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
