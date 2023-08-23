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

func upgradeRequiresIAMMigration(provider cloudprovider.Provider) bool {
	switch provider {
	default:
		return false
	}
}

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
	configFetcher attestationconfigapi.Fetcher
	log           debugLog
}

func runIAMUpgradeApply(cmd *cobra.Command, _ []string) error {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("parsing force argument: %w", err)
	}
	fileHandler := file.NewHandler(afero.NewOsFs())
	configFetcher := attestationconfigapi.NewFetcher()

	upgradeID := generateUpgradeID(upgradeCmdKindIAM)
	upgradeDir := filepath.Join(constants.UpgradeDir, upgradeID)
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
		configFetcher: configFetcher,
		log:           log,
	}

	return i.iamUpgradeApply(cmd, iamMigrateCmd, upgradeDir, force, yes)
}

func (i iamUpgradeApplyCmd) iamUpgradeApply(cmd *cobra.Command, iamUpgrader iamUpgrader, upgradeDir string, force, yes bool) error {
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
		return err
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
			// Remove the upgrade directory
			if err := i.fileHandler.RemoveAll(upgradeDir); err != nil {
				return fmt.Errorf("cleaning up upgrade directory %s: %w", upgradeDir, err)
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
}
