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

	"github.com/edgelesssys/constellation/v2/api/attestationconfig"
	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

type iamUpgradeApplyFlags struct {
	rootFlags
	yes bool
}

func (f *iamUpgradeApplyFlags) parse(flags *pflag.FlagSet) error {
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

type iamUpgradeApplyCmd struct {
	fileHandler   file.Handler
	log           debugLog
	configFetcher attestationconfig.Fetcher
	flags         iamUpgradeApplyFlags
}

func runIAMUpgradeApply(cmd *cobra.Command, _ []string) error {
	fileHandler := file.NewHandler(afero.NewOsFs())
	upgradeID := generateUpgradeID(upgradeCmdKindIAM)
	upgradeDir := filepath.Join(constants.UpgradeDir, upgradeID)
	configFetcher := attestationconfig.NewFetcher()
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

	i := iamUpgradeApplyCmd{
		fileHandler:   fileHandler,
		log:           log,
		configFetcher: configFetcher,
	}
	if err := i.flags.parse(cmd.Flags()); err != nil {
		return err
	}

	return i.iamUpgradeApply(cmd, iamMigrateCmd, upgradeDir)
}

func (i iamUpgradeApplyCmd) iamUpgradeApply(cmd *cobra.Command, iamUpgrader iamUpgrader, upgradeDir string) error {
	conf, err := config.New(i.fileHandler, constants.ConfigFilename, i.configFetcher, i.flags.force)
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
	if !hasDiff && !i.flags.force {
		cmd.Println("No IAM migrations necessary.")
		return nil
	}

	// If there are any Terraform migrations to apply, ask for confirmation
	cmd.Println("The IAM upgrade requires a migration by applying an updated Terraform template. Please manually review the suggested changes.")
	if !i.flags.yes {
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
					filepath.Join(upgradeDir, constants.TerraformIAMUpgradeBackupDir),
				)
			}
			return errors.New("IAM upgrade aborted by user")
		}
	}
	i.log.Debug("Applying Terraform IAM migrations")
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
