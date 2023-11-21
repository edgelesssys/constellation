/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// NewIAMDestroyCmd returns a new cobra.Command for the iam destroy subcommand.
func newIAMDestroyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy an IAM configuration and delete local Terraform files",
		Long:  "Destroy an IAM configuration and delete local Terraform files.",
		Args:  cobra.ExactArgs(0),
		RunE:  runIAMDestroy,
	}

	cmd.Flags().BoolP("yes", "y", false, "destroy the IAM configuration without asking for confirmation")

	return cmd
}

type iamDestroyFlags struct {
	rootFlags
	yes bool
}

func (f *iamDestroyFlags) parse(flags *pflag.FlagSet) error {
	if err := f.rootFlags.parse(flags); err != nil {
		return err
	}

	yes, err := flags.GetBool("yes")
	if err != nil {
		return fmt.Errorf("getting 'yes' flag: %w", err)
	}
	f.yes = yes

	return nil
}

func runIAMDestroy(cmd *cobra.Command, _ []string) error {
	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("creating logger: %w", err)
	}
	defer log.Sync()
	spinner := newSpinner(cmd.ErrOrStderr())
	destroyer := cloudcmd.NewIAMDestroyer()
	fsHandler := file.NewHandler(afero.NewOsFs())

	c := &destroyCmd{log: log}
	if err := c.flags.parse(cmd.Flags()); err != nil {
		return err
	}

	return c.iamDestroy(cmd, spinner, destroyer, fsHandler)
}

type destroyCmd struct {
	log   debugLog
	flags iamDestroyFlags
}

func (c *destroyCmd) iamDestroy(cmd *cobra.Command, spinner spinnerInterf, destroyer iamDestroyer, fsHandler file.Handler) error {
	// check if there is a possibility that the cluster is still running by looking out for specific files
	c.log.Debugf("Checking if %q exists", c.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	if _, err := fsHandler.Stat(constants.AdminConfFilename); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %q still exists, please make sure to terminate your cluster before destroying your IAM configuration", c.flags.pathPrefixer.PrefixPrintablePath(constants.AdminConfFilename))
	}

	c.log.Debugf("Checking if %q exists", c.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))
	if _, err := fsHandler.Stat(constants.StateFilename); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %q still exists, please make sure to terminate your cluster before destroying your IAM configuration", c.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename))
	}

	gcpFileExists := false

	c.log.Debugf("Checking if %q exists", c.flags.pathPrefixer.PrefixPrintablePath(constants.GCPServiceAccountKeyFilename))
	if _, err := fsHandler.Stat(constants.GCPServiceAccountKeyFilename); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		c.log.Debugf("%q exists", c.flags.pathPrefixer.PrefixPrintablePath(constants.GCPServiceAccountKeyFilename))
		gcpFileExists = true
	}

	if !c.flags.yes {
		// Confirmation
		confirmString := "Do you really want to destroy your IAM configuration? Note that this will remove all resources in the resource group."
		if gcpFileExists {
			confirmString += fmt.Sprintf("\nThis will also delete %q", c.flags.pathPrefixer.PrefixPrintablePath(constants.GCPServiceAccountKeyFilename))
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
		c.log.Debugf("Starting to delete %q", c.flags.pathPrefixer.PrefixPrintablePath(constants.GCPServiceAccountKeyFilename))
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
	if err := destroyer.DestroyIAMConfiguration(cmd.Context(), constants.TerraformIAMWorkingDir, c.flags.tfLogLevel); err != nil {
		return fmt.Errorf("destroying IAM configuration: %w", err)
	}

	spinner.Stop() // stop the spinner to print a new line
	fmt.Println("Successfully destroyed IAM configuration")
	return nil
}

func (c *destroyCmd) deleteGCPServiceAccountKeyFile(cmd *cobra.Command, destroyer iamDestroyer, fsHandler file.Handler) (bool, error) {
	var fileSaKey gcpshared.ServiceAccountKey

	c.log.Debugf("Parsing %q", c.flags.pathPrefixer.PrefixPrintablePath(constants.GCPServiceAccountKeyFilename))
	if err := fsHandler.ReadJSON(constants.GCPServiceAccountKeyFilename, &fileSaKey); err != nil {
		return false, err
	}

	c.log.Debugf("Getting service account key from the tfstate")
	tfSaKey, err := destroyer.GetTfStateServiceAccountKey(cmd.Context(), constants.TerraformIAMWorkingDir)
	if err != nil {
		return false, err
	}

	c.log.Debugf("Checking if keys are the same")
	if tfSaKey != fileSaKey {
		cmd.Printf(
			"The key in %q don't match up with your Terraform state. %q will not be deleted.\n",
			c.flags.pathPrefixer.PrefixPrintablePath(constants.GCPServiceAccountKeyFilename),
			c.flags.pathPrefixer.PrefixPrintablePath(constants.GCPServiceAccountKeyFilename),
		)
		return true, nil
	}

	if err := fsHandler.Remove(constants.GCPServiceAccountKeyFilename); err != nil {
		return false, err
	}

	c.log.Debugf("Successfully deleted %q", c.flags.pathPrefixer.PrefixPrintablePath(constants.GCPServiceAccountKeyFilename))
	return true, nil
}
