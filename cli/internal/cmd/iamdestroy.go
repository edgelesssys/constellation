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
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
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

	return c.iamDestroy(cmd, spinner, destroyer, fsHandler)
}

type destroyCmd struct {
	log debugLog
}

func (c *destroyCmd) iamDestroy(cmd *cobra.Command, spinner spinnerInterf, destroyer iamDestroyer, fsHandler file.Handler) error {
	flags, err := c.parseDestroyFlags(cmd)
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	// check if there is a possibility that the cluster is still running by looking out for specific files
	c.log.Debugf("Checking if %q exists", adminConfPath(flags.workspace))
	_, err = fsHandler.Stat(constants.AdminConfFilename)
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %q still exists, please make sure to terminate your cluster before destroying your IAM configuration", adminConfPath(flags.workspace))
	}
	c.log.Debugf("Checking if %q exists", clusterIDsPath(flags.workspace))
	_, err = fsHandler.Stat(constants.ClusterIDsFilename)
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file %q still exists, please make sure to terminate your cluster before destroying your IAM configuration", clusterIDsPath(flags.workspace))
	}

	gcpFileExists := false

	c.log.Debugf("Checking if %q exists", gcpServiceAccountKeyPath(flags.workspace))
	_, err = fsHandler.Stat(gcpServiceAccountKeyFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		c.log.Debugf("%q exists", gcpServiceAccountKeyPath(flags.workspace))
		gcpFileExists = true
	}

	if !flags.yes {
		// Confirmation
		confirmString := "Do you really want to destroy your IAM configuration? Note that this will remove all resources in the resource group."
		if gcpFileExists {
			confirmString += fmt.Sprintf("\nThis will also delete %q", gcpServiceAccountKeyPath(flags.workspace))
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
		c.log.Debugf("Starting to delete %q", gcpServiceAccountKeyPath(flags.workspace))
		proceed, err := c.deleteGCPServiceAccountKeyFile(cmd, destroyer, flags.workspace, fsHandler)
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
	if err := destroyer.DestroyIAMConfiguration(cmd.Context(), constants.TerraformIAMWorkingDir, flags.tfLogLevel); err != nil {
		return fmt.Errorf("destroying IAM configuration: %w", err)
	}

	spinner.Stop() // stop the spinner to print a new line
	fmt.Println("Successfully destroyed IAM configuration")
	return nil
}

func (c *destroyCmd) deleteGCPServiceAccountKeyFile(cmd *cobra.Command, destroyer iamDestroyer, workspace string, fsHandler file.Handler) (bool, error) {
	var fileSaKey gcpshared.ServiceAccountKey

	c.log.Debugf("Parsing %q", gcpServiceAccountKeyPath(workspace))
	if err := fsHandler.ReadJSON(gcpServiceAccountKeyFile, &fileSaKey); err != nil {
		return false, err
	}

	c.log.Debugf("Getting service account key from the tfstate")
	tfSaKey, err := destroyer.GetTfStateServiceAccountKey(cmd.Context(), constants.TerraformIAMWorkingDir)
	if err != nil {
		return false, err
	}

	c.log.Debugf("Checking if keys are the same")
	if tfSaKey != fileSaKey {
		cmd.Printf("The key in %q don't match up with your Terraform state. %q will not be deleted.\n", gcpServiceAccountKeyPath(workspace), gcpServiceAccountKeyPath(workspace))
		return true, nil
	}

	if err := fsHandler.Remove(gcpServiceAccountKeyFile); err != nil {
		return false, err
	}

	c.log.Debugf("Successfully deleted %q", gcpServiceAccountKeyPath(workspace))
	return true, nil
}

type destroyFlags struct {
	yes        bool
	workspace  string
	tfLogLevel terraform.LogLevel
}

// parseDestroyFlags parses the flags of the create command.
func (c *destroyCmd) parseDestroyFlags(cmd *cobra.Command) (destroyFlags, error) {
	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return destroyFlags{}, fmt.Errorf("parsing yes bool: %w", err)
	}
	c.log.Debugf("Yes flag is %t", yes)

	workspace, err := cmd.Flags().GetString("workspace")
	if err != nil {
		return destroyFlags{}, fmt.Errorf("parsing workspace string: %w", err)
	}
	c.log.Debugf("Workspace set to %q", workspace)

	logLevelString, err := cmd.Flags().GetString("tf-log")
	if err != nil {
		return destroyFlags{}, fmt.Errorf("parsing tf-log string: %w", err)
	}
	logLevel, err := terraform.ParseLogLevel(logLevelString)
	if err != nil {
		return destroyFlags{}, fmt.Errorf("parsing Terraform log level %s: %w", logLevelString, err)
	}
	c.log.Debugf("Terraform logs will be written into %s at level %s", terraformLogPath(workspace), logLevel.String())

	return destroyFlags{
		tfLogLevel: logLevel,
		workspace:  workspace,
		yes:        yes,
	}, nil
}
