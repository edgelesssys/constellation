/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func runIAMDestroy(cmd *cobra.Command, _args []string) error {
	spinner := newSpinner(cmd.ErrOrStderr())
	destroyer := cloudcmd.NewIAMDestroyer(cmd.Context())
	fsHandler := file.NewHandler(afero.NewOsFs())

	return iamDestroy(cmd, spinner, destroyer, fsHandler)
}

func iamDestroy(cmd *cobra.Command, spinner spinnerInterf, destroyer iamDestroyer, fsHandler file.Handler) error {
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	gcpFileExists := false

	_, err = fsHandler.Stat(constants.GCPServiceAccountKeyFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		gcpFileExists = true
	}

	if !yes {
		// Confirmation
		confirmString := "Do you really want to destroy your IAM configuration?"
		if gcpFileExists {
			confirmString += " (This will also delete " + constants.GCPServiceAccountKeyFile + ")"
		}
		ok, err := askToConfirm(cmd, confirmString)
		if err != nil {
			return err
		}
		if !ok {
			cmd.Println("The destruction of the IAM configuration was aborted")
			return nil
		}
	}

	if gcpFileExists {
		proceed, err := deleteGCPServiceAccountKeyFile(cmd, destroyer, fsHandler)
		if err != nil {
			return err
		}
		if !proceed {
			cmd.Println("Destruction was aborted")
			return nil
		}
	}

	spinner.Start("Destroying IAM configuration", false)
	defer spinner.Stop()
	if err := destroyer.DestroyIAMConfiguration(cmd.Context()); err != nil {
		return fmt.Errorf("destroying IAM configuration: %w", err)
	}

	spinner.Stop() // stop the spinner to print a new line
	fmt.Println("Successfully destroyed IAM configuration")
	return nil
}

func deleteGCPServiceAccountKeyFile(cmd *cobra.Command, destroyer iamDestroyer, fsHandler file.Handler) (bool, error) {
	if _, err := fsHandler.Stat(constants.GCPServiceAccountKeyFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
		return true, err // file just doesn't exist
	}

	destroyed, err := destroyer.RunDeleteGCPKeyFile(cmd.Context())
	if err != nil {
		return false, err
	}
	if !destroyed {
		ok, err := askToConfirm(cmd, "The file gcpServiceAccountKey.json could not be deleted. Either it does not exist or the file belongs to another IAM configuration. Do you want to proceed anyway?")
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}
