/*
Copyright (c) Edgeless Systems GmbH
SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/cli/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const upgradeRequiresIAMMigration = true // TODO(elchead): needs to be set on every release. Can we automate this?

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
	upgradeID := "iam-" + time.Now().Format("20060102150405") + "-" + strings.Split(uuid.New().String(), "-")[0]
	iamMigrateCmd, err := upgrade.NewIAMMigrateCmd(cmd.Context(), upgradeID, cloudprovider.AWS, terraform.LogLevelDebug)
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
