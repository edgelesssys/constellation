/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func newIAMUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Find and apply upgrades to your IAM profile",
		Long:  "Find and apply upgrades to your IAM profile.",
		Args:  cobra.ExactArgs(0),
	}
	cmd.AddCommand(newIAMUpgradeApplyCmd())
	return cmd
}

func newIAMUpgradeApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply an upgrade to an IAM profile",
		Long:  "Apply an upgrade to an IAM profile.",
		Args:  cobra.NoArgs,
		RunE:  runIAMUpgradeApply,
	}
	cmd.Flags().BoolP("yes", "y", false, "run upgrades without further confirmation")
	return cmd
}

type iamUpgradeApplyCmd struct {
	fileHandler   file.Handler
	log           debugLog
	configFetcher attestationconfigapi.Fetcher
}

func runIAMUpgradeApply(cmd *cobra.Command, _ []string) error {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("parsing force argument: %w", err)
	}
	fileHandler := file.NewHandler(afero.NewOsFs())
	upgradeID := generateUpgradeID(upgradeCmdKindIAM)
	upgradeDir := filepath.Join(constants.UpgradeDir, upgradeID)
	configFetcher := attestationconfigapi.NewFetcher()
	iamMigrateCmd, err := cloudcmd.NewIAMUpgrader(
		cmd.Context(),
		constants.TerraformIAMWorkingDir,
		upgradeDir,
		terraform.LogLevelDebug,
		fileHandler,
	)
	if err != nil {
		return fmt.Errorf("setting up IAM migration command: %w", err)
	}

	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("setting up logger: %w", err)
	}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	i := iamUpgradeApplyCmd{
		fileHandler:   fileHandler,
		log:           log,
		configFetcher: configFetcher,
	}

	return i.iamUpgradeApply(cmd, iamMigrateCmd, force, yes)
}

func (i iamUpgradeApplyCmd) iamUpgradeApply(cmd *cobra.Command, iamUpgrader iamUpgrader, force, yes bool) error {
	conf, err := config.New(i.fileHandler, constants.ConfigFilename, i.configFetcher, force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}

	vars, err := cloudcmd.TerraformIAMUpgradeVars(conf, i.fileHandler)
	if err != nil {
		return fmt.Errorf("getting terraform variables: %w", err)
	}
	hasDiff, err := iamUpgrader.PlanIAMUpgrade(cmd.Context(), cmd.OutOrStderr(), vars, conf.GetProvider())
	if err != nil {
		return fmt.Errorf("planning terraform migrations: %w", err)
	}
	if !hasDiff && !force {
		cmd.Println("No IAM migrations necessary.")
		return nil
	}

	// If there are any Terraform migrations to apply, ask for confirmation
	cmd.Println("The IAM upgrade requires a migration by applying an updated Terraform template. Please manually review the suggested changes.")
	if !yes {
		ok, err := askToConfirm(cmd, "Do you want to apply the IAM upgrade?")
		if err != nil {
			return fmt.Errorf("asking for confirmation: %w", err)
		}
		if !ok {
			cmd.Println("Aborting upgrade.")
			// User doesn't expect to see any changes in his workspace after aborting an "upgrade apply",
			// therefore, roll back to the backed up state.
			if err := iamUpgrader.RestoreIAMWorkspace(); err != nil {
				return fmt.Errorf(
					"restoring Terraform workspace: %w, restore the Terraform workspace manually from %s ",
					err,
					filepath.Join(constants.UpgradeDir, "<upgrade-id>", constants.TerraformIAMUpgradeBackupDir),
				)
			}
			return errors.New("IAM upgrade aborted by user")
		}
	}
	i.log.Debugf("Applying Terraform IAM migrations")
	if err := iamUpgrader.ApplyIAMUpgrade(cmd.Context(), conf.GetProvider()); err != nil {
		return fmt.Errorf("applying terraform migrations: %w", err)
	}

	cmd.Println("IAM profile successfully applied.")

	return nil
}

type iamUpgrader interface {
	PlanIAMUpgrade(ctx context.Context, outWriter io.Writer, vars terraform.Variables, csp cloudprovider.Provider) (bool, error)
	ApplyIAMUpgrade(ctx context.Context, csp cloudprovider.Provider) error
	RestoreIAMWorkspace() error
}
