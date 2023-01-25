/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/
package cmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func runDestroyIAMUser(cmd *cobra.Command, _args []string) error {
	spinner := newSpinner(cmd.ErrOrStderr())
	destroyer := cloudcmd.NewIAMDestroyer(cmd.Context())
	fsHandler := file.NewHandler(afero.NewOsFs())

	return destroyIAMUser(cmd, spinner, destroyer, fsHandler)
}

func destroyIAMUser(cmd *cobra.Command, spinner spinnerInterf, destroyer iamDestroyer, fsHandler file.Handler) error {
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

	if proceed, err := deleteGCPServiceAccountKeyFile(cmd, destroyer, fsHandler); err != nil && !proceed {
		return err
	}

	spinner.Start("Destroying IAM User", false)

	if err := destroyer.DestroyIAMUser(cmd.Context()); err != nil {
		return fmt.Errorf("couldn't destroy IAM User: %w", err)
	}
	spinner.Stop()
	fmt.Println("Successfully destroyed IAM User")
	return nil
}

func deleteGCPServiceAccountKeyFile(cmd *cobra.Command, destroyer iamDestroyer, fsHandler file.Handler) (bool, error) {
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return false, err
	}

	if _, err := fsHandler.Stat(constants.GCPServiceAccountKeyFile); err != nil {
		return true, err // file doesn't exist
	}

	if !yes {
		ok, err := askToConfirm(cmd, "There seems to be a gcpServiceAccountKey.json file. Do you want to delete it? (Note that you may not be able to save generated private keys for GCP if the file exists)")
		if err != nil {
			return false, err
		}
		if !ok {
			return true, nil
		}
	}

	destroyed, err := destroyer.RunDeleteGCPKeyFile(cmd.Context())
	if err != nil {
		return false, err
	}
	if !destroyed {
		ok, err := askToConfirm(cmd, "The file gcpServiceAccountKey.json could not be deleted. Either it does not exist or the file belongs to another IAM user. Do you want to proceed anyway?")
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
