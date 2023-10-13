/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/spf13/cobra"
)

// runTerraformApply checks if changes to Terraform are required and applies them.
func (a *applyCmd) runTerraformApply(cmd *cobra.Command, conf *config.Config, stateFile *state.State, upgradeDir string) error {
	a.log.Debugf("Checking if Terraform migrations are required")
	migrationRequired, err := a.planTerraformMigration(cmd, conf)
	if err != nil {
		return fmt.Errorf("planning Terraform migrations: %w", err)
	}

	if !migrationRequired {
		a.log.Debugf("No changes to infrastructure required, skipping Terraform migrations")
		return nil
	}

	a.log.Debugf("Migrating terraform resources for infrastructure changes")
	postMigrationInfraState, err := a.migrateTerraform(cmd, conf, upgradeDir)
	if err != nil {
		return fmt.Errorf("performing Terraform migrations: %w", err)
	}

	// Merge the pre-upgrade state with the post-migration infrastructure values
	a.log.Debugf("Updating state file with new infrastructure state")
	if _, err := stateFile.Merge(
		// temporary state with post-migration infrastructure values
		state.New().SetInfrastructure(postMigrationInfraState),
	); err != nil {
		return fmt.Errorf("merging pre-upgrade state with post-migration infrastructure values: %w", err)
	}

	// Write the post-migration state to disk
	if err := stateFile.WriteToFile(a.fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

// planTerraformMigration checks if the Constellation version the cluster is being upgraded to requires a migration.
func (a *applyCmd) planTerraformMigration(cmd *cobra.Command, conf *config.Config) (bool, error) {
	a.log.Debugf("Planning Terraform migrations")
	vars, err := cloudcmd.TerraformUpgradeVars(conf)
	if err != nil {
		return false, fmt.Errorf("parsing upgrade variables: %w", err)
	}
	a.log.Debugf("Using Terraform variables:\n%+v", vars)

	// Check if there are any Terraform migrations to apply

	// Add manual migrations here if required
	//
	// var manualMigrations []terraform.StateMigration
	// for _, migration := range manualMigrations {
	// 	  u.log.Debugf("Adding manual Terraform migration: %s", migration.DisplayName)
	// 	  u.upgrader.AddManualStateMigration(migration)
	// }

	return a.clusterUpgrader.PlanClusterUpgrade(cmd.Context(), cmd.OutOrStdout(), vars, conf.GetProvider())
}

// migrateTerraform migrates an existing Terraform state and the post-migration infrastructure state is returned.
func (a *applyCmd) migrateTerraform(cmd *cobra.Command, conf *config.Config, upgradeDir string) (state.Infrastructure, error) {
	// Ask for confirmation first
	fmt.Fprintln(cmd.OutOrStdout(), "The upgrade requires a migration of Constellation cloud resources by applying an updated Terraform template. Please manually review the suggested changes below.")
	if !a.flags.yes {
		ok, err := askToConfirm(cmd, "Do you want to apply the Terraform migrations?")
		if err != nil {
			return state.Infrastructure{}, fmt.Errorf("asking for confirmation: %w", err)
		}
		if !ok {
			cmd.Println("Aborting upgrade.")
			// User doesn't expect to see any changes in his workspace after aborting an "upgrade apply",
			// therefore, roll back to the backed up state.
			if err := a.clusterUpgrader.RestoreClusterWorkspace(); err != nil {
				return state.Infrastructure{}, fmt.Errorf(
					"restoring Terraform workspace: %w, restore the Terraform workspace manually from %s ",
					err,
					filepath.Join(upgradeDir, constants.TerraformUpgradeBackupDir),
				)
			}
			return state.Infrastructure{}, fmt.Errorf("cluster upgrade aborted by user")
		}
	}
	a.log.Debugf("Applying Terraform migrations")

	a.spinner.Start("Migrating Terraform resources", false)
	infraState, err := a.clusterUpgrader.ApplyClusterUpgrade(cmd.Context(), conf.GetProvider())
	a.spinner.Stop()
	if err != nil {
		return state.Infrastructure{}, fmt.Errorf("applying terraform migrations: %w", err)
	}

	cmd.Printf("Infrastructure migrations applied successfully and output written to: %s\n"+
		"A backup of the pre-upgrade state has been written to: %s\n",
		a.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename),
		a.flags.pathPrefixer.PrefixPrintablePath(filepath.Join(upgradeDir, constants.TerraformUpgradeBackupDir)),
	)
	return infraState, nil
}
