/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/edgelesssys/constellation/v2/cli/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/cobra"
)

// tfMigrationClient is a client for planning and applying Terraform migrations.
type tfMigrationClient struct {
	log DebugLog
}

// planMigration checks for Terraform migrations and asks for confirmation if there are any. The user input is returned as confirmedDiff.
// adapted from migrateTerraform().
func (u *tfMigrationClient) planMigration(cmd *cobra.Command, file file.Handler, migrateCmd tfMigrationCmd) (hasDiff bool, err error) {
	u.log.Debugf("Planning %s", migrateCmd.String())
	if err := migrateCmd.CheckTerraformMigrations(file); err != nil {
		return false, fmt.Errorf("checking workspace: %w", err)
	}
	hasDiff, err = migrateCmd.Plan(cmd.Context(), file, cmd.OutOrStdout())
	if err != nil {
		return hasDiff, fmt.Errorf("planning terraform migrations: %w", err)
	}
	return hasDiff, nil
}

// applyMigration plans and then applies the Terraform migration. The user is asked for confirmation if there are any changes.
// adapted from migrateTerraform().
func (u *tfMigrationClient) applyMigration(cmd *cobra.Command, upgradeWorkspace string, file file.Handler, migrateCmd tfMigrationCmd, yesFlag bool) error {
	hasDiff, err := u.planMigration(cmd, file, migrateCmd)
	if err != nil {
		return err
	}
	if hasDiff {
		// If there are any Terraform migrations to apply, ask for confirmation
		fmt.Fprintf(cmd.OutOrStdout(), "The %s upgrade requires a migration by applying an updated Terraform template. Please manually review the suggested changes below.\n", migrateCmd.String())
		if !yesFlag {
			ok, err := AskToConfirm(cmd, fmt.Sprintf("Do you want to apply the %s?", migrateCmd.String()))
			if err != nil {
				return fmt.Errorf("asking for confirmation: %w", err)
			}
			if !ok {
				cmd.Println("Aborting upgrade.")
				if err := upgrade.CleanUpTerraformMigrations(upgradeWorkspace, migrateCmd.UpgradeID(), file); err != nil {
					return fmt.Errorf("cleaning up workspace: %w", err)
				}
				return fmt.Errorf("aborted by user")
			}
		}
		u.log.Debugf("Applying Terraform %s migrations", migrateCmd.String())
		err := migrateCmd.Apply(cmd.Context(), file)
		if err != nil {
			return fmt.Errorf("applying terraform migrations: %w", err)
		}
	} else {
		u.log.Debugf("No Terraform diff detected")
	}
	return nil
}

// tfMigrationCmd is an interface for all terraform upgrade / migration commands.
type tfMigrationCmd interface {
	CheckTerraformMigrations(file file.Handler) error
	Plan(ctx context.Context, file file.Handler, outWriter io.Writer) (bool, error)
	Apply(ctx context.Context, fileHandler file.Handler) error
	String() string
	UpgradeID() string
}
