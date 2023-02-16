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
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func runIAMDestroy(cmd *cobra.Command, _args []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	spinner := newSpinner(cmd.ErrOrStderr())
	destroyer, err := cloudcmd.NewIAMDestroyer(cmd.Context())
	if err != nil {
		return err
	}
	fsHandler := file.NewHandler(afero.NewOsFs())

	c := &destroyCmd{log: log}

	return c.iamDestroy(cmd, spinner, destroyer, fsHandler)
}

type destroyCmd struct {
	log debugLog
}

func (c *destroyCmd) iamDestroy(cmd *cobra.Command, spinner spinnerInterf, destroyer iamDestroyer, fsHandler file.Handler) error {
	// check if there is a possibility that the cluster is still running by looking out for specific files
	c.log.Debugf("Checking if %s exists", constants.AdminConfFilename)
	_, err := fsHandler.Stat(constants.AdminConfFilename)
	if !errors.Is(err, os.ErrNotExist) {
		return errors.New("file " + constants.AdminConfFilename + " still exists, please make sure to terminate your cluster before destroying your IAM configuration")
	}
	c.log.Debugf("Checking if %s exists", constants.ClusterIDsFileName)
	_, err = fsHandler.Stat(constants.ClusterIDsFileName)
	if !errors.Is(err, os.ErrNotExist) {
		return errors.New("file " + constants.ClusterIDsFileName + " still exists, please make sure to terminate your cluster before destroying your IAM configuration")
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}
	c.log.Debugf("\"yes\" flag is set to %t", yes)

	gcpFileExists := false

	c.log.Debugf("Checking if %s exists", constants.GCPServiceAccountKeyFile)
	_, err = fsHandler.Stat(constants.GCPServiceAccountKeyFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		c.log.Debugf("%s exists", constants.GCPServiceAccountKeyFile)
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
		c.log.Debugf("Starting to delete %s", constants.GCPServiceAccountKeyFile)
		proceed, err := c.deleteGCPServiceAccountKeyFile(cmd, destroyer, fsHandler)
		if err != nil {
			return err
		}
		if !proceed {
			cmd.Println("Destruction was aborted")
			return nil
		}
	}

	c.log.Debugf("Starting to destroy IAM configuration")

	spinner.Start("Destroying IAM configuration", false)
	defer spinner.Stop()
	if err := destroyer.DestroyIAMConfiguration(cmd.Context()); err != nil {
		return fmt.Errorf("destroying IAM configuration: %w", err)
	}

	spinner.Stop() // stop the spinner to print a new line
	fmt.Println("Successfully destroyed IAM configuration")
	return nil
}

func (c *destroyCmd) deleteGCPServiceAccountKeyFile(cmd *cobra.Command, destroyer iamDestroyer, fsHandler file.Handler) (bool, error) {
	c.log.Debugf("Checking if %s exists", constants.GCPServiceAccountKeyFile)
	if _, err := fsHandler.Stat(constants.GCPServiceAccountKeyFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
		c.log.Debugf("File %s doesn't exist", constants.GCPServiceAccountKeyFile)
		return true, nil // file just doesn't exist
	}

	var fileSaKey gcpshared.ServiceAccountKey

	c.log.Debugf("Parsing %s", constants.GCPServiceAccountKeyFile)
	if err := fsHandler.ReadJSON(constants.GCPServiceAccountKeyFile, &fileSaKey); err != nil {
		return false, err
	}

	c.log.Debugf("Getting service account key from the tfstate")
	tfSaKey, err := destroyer.GetTfstateServiceAccountKey(cmd.Context())
	if err != nil {
		return false, err
	}

	c.log.Debugf("Checking if keys are the same")
	if tfSaKey != fileSaKey {
		return false, nil
	}

	if err := fsHandler.Remove(constants.GCPServiceAccountKeyFile); err != nil {
		ok, err := askToConfirm(cmd, "The file gcpServiceAccountKey.json could not be deleted. Either it does not exist or the file belongs to another IAM configuration. Do you want to proceed anyway?")
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	c.log.Debugf("Successfully deleted %s", constants.GCPServiceAccountKeyFile)
	return true, nil
}
