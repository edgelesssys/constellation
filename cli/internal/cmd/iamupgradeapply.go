/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/cli/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

func upgradeRequiresIAMMigration(provider cloudprovider.Provider) bool {
	switch provider {
	case cloudprovider.AWS:
		return true // needs to be set on every release. Can we automate this?
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
	cmd.Flags().BoolP("yes", "y", false, "run upgrades without further confirmation\n")
	return cmd
}

func runIAMUpgradeApply(cmd *cobra.Command, _ []string) error {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("parsing force argument: %w", err)
	}
	fileHandler := file.NewHandler(afero.NewOsFs())
	configFetcher := attestationconfigapi.NewFetcher()
	conf, err := config.New(fileHandler, configPath, configFetcher, force)
	var configValidationErr *config.ValidationError
	if errors.As(err, &configValidationErr) {
		cmd.PrintErrln(configValidationErr.LongMessage())
	}
	if err != nil {
		return err
	}
	upgradeID := "iam-" + time.Now().Format("20060102150405") + "-" + strings.Split(uuid.New().String(), "-")[0]
	iamMigrateCmd, err := upgrade.NewIAMMigrateCmd(cmd.Context(), upgradeID, conf.GetProvider(), terraform.LogLevelDebug)
	if err != nil {
		return fmt.Errorf("setting up IAM migration command: %w", err)
	}

	log, err := newCLILogger(cmd)
	if err != nil {
		return fmt.Errorf("setting up logger: %w", err)
	}
	migrator := &tfMigrationClient{log}

	yes, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}
	err = migrator.applyMigration(cmd, file.NewHandler(afero.NewOsFs()), iamMigrateCmd, yes)
	if err != nil {
		return fmt.Errorf("applying IAM migration: %w", err)
	}
	cmd.Println("IAM profile successfully applied.")
	return nil
}
