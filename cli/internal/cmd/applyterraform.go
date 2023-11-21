/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/state"
	"github.com/spf13/cobra"
)

// runTerraformApply checks if changes to Terraform are required and applies them.
func (a *applyCmd) runTerraformApply(cmd *cobra.Command, conf *config.Config, stateFile *state.State, upgradeDir string) error {
	a.log.Debugf("Checking if Terraform migrations are required")
	terraformClient, removeClient, err := a.newInfraApplier(cmd.Context())
	if err != nil {
		return fmt.Errorf("creating Terraform client: %w", err)
	}
	defer removeClient()

	// Check if we are creating a new cluster by checking if the Terraform workspace is empty
	isNewCluster, err := terraformClient.WorkingDirIsEmpty()
	if err != nil {
		return fmt.Errorf("checking if Terraform workspace is empty: %w", err)
	}

	if changesRequired, err := a.planTerraformChanges(cmd, conf, terraformClient); err != nil {
		return fmt.Errorf("planning Terraform migrations: %w", err)
	} else if !changesRequired {
		a.log.Debugf("No changes to infrastructure required, skipping Terraform migrations")
		return nil
	}

	a.log.Debugf("Apply new Terraform resources for infrastructure changes")
	newInfraState, err := a.applyTerraformChanges(cmd, conf, terraformClient, upgradeDir, isNewCluster)
	if err != nil {
		return err
	}

	// Merge the original state with the new infrastructure values
	a.log.Debugf("Updating state file with new infrastructure state")
	if _, err := stateFile.Merge(
		// temporary state with new infrastructure values
		state.New().SetInfrastructure(newInfraState),
	); err != nil {
		return fmt.Errorf("merging old state with new infrastructure values: %w", err)
	}

	// Write the new state to disk
	if err := stateFile.WriteToFile(a.fileHandler, constants.StateFilename); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

// planTerraformChanges checks if any changes to the Terraform state are required.
// If no state exists, this function will return true and the caller should create a new state.
func (a *applyCmd) planTerraformChanges(cmd *cobra.Command, conf *config.Config, terraformClient cloudApplier) (bool, error) {
	a.log.Debugf("Planning Terraform changes")

	// Check if there are any Terraform changes to apply

	// Add manual migrations here if required
	//
	// var manualMigrations []terraform.StateMigration
	// for _, migration := range manualMigrations {
	// 	  u.log.Debugf("Adding manual Terraform migration: %s", migration.DisplayName)
	// 	  u.infraApplier.AddManualStateMigration(migration)
	// }

	a.spinner.Start("Checking for infrastructure changes", false)
	defer a.spinner.Stop()
	return terraformClient.Plan(cmd.Context(), conf)
}

// applyTerraformChanges applies planned changes to a Terraform state and returns the resulting infrastructure state.
// If no state existed prior to this function call, a new cluster will be created.
func (a *applyCmd) applyTerraformChanges(
	cmd *cobra.Command, conf *config.Config, terraformClient cloudApplier, upgradeDir string, isNewCluster bool,
) (state.Infrastructure, error) {
	if isNewCluster {
		if err := printCreateInfo(cmd.OutOrStdout(), conf, a.log); err != nil {
			return state.Infrastructure{}, err
		}
		return a.applyTerraformChangesWithMessage(
			cmd, conf.GetProvider(), cloudcmd.WithRollbackOnError, terraformClient, upgradeDir,
			"Do you want to create this cluster?",
			"The creation of the cluster was aborted.",
			"cluster creation aborted by user",
			"Creating",
			"Cloud infrastructure created successfully.",
		)
	}

	cmd.Println("Changes of Constellation cloud resources are required by applying an updated Terraform template.")
	return a.applyTerraformChangesWithMessage(
		cmd, conf.GetProvider(), cloudcmd.WithoutRollbackOnError, terraformClient, upgradeDir,
		"Do you want to apply these Terraform changes?",
		"Aborting upgrade.",
		"cluster upgrade aborted by user",
		"Applying Terraform changes",
		fmt.Sprintf("Infrastructure migrations applied successfully and output written to: %s\n"+
			"A backup of the pre-upgrade state has been written to: %s",
			a.flags.pathPrefixer.PrefixPrintablePath(constants.StateFilename),
			a.flags.pathPrefixer.PrefixPrintablePath(filepath.Join(upgradeDir, constants.TerraformUpgradeBackupDir)),
		),
	)
}

func (a *applyCmd) applyTerraformChangesWithMessage(
	cmd *cobra.Command, csp cloudprovider.Provider, rollbackBehavior cloudcmd.RollbackBehavior,
	terraformClient cloudApplier, upgradeDir string,
	confirmationQst, abortMsg, abortErrorMsg, progressMsg, successMsg string,
) (state.Infrastructure, error) {
	// Ask for confirmation first
	if !a.flags.yes {
		ok, err := askToConfirm(cmd, confirmationQst)
		if err != nil {
			return state.Infrastructure{}, fmt.Errorf("asking for confirmation: %w", err)
		}
		if !ok {
			cmd.Println(abortMsg)
			// User doesn't expect to see any changes in their workspace after aborting an "apply",
			// therefore, restore the workspace to the previous state.
			if err := terraformClient.RestoreWorkspace(); err != nil {
				return state.Infrastructure{}, fmt.Errorf(
					"restoring Terraform workspace: %w, clean up or restore the Terraform workspace manually from %s ",
					err,
					filepath.Join(upgradeDir, constants.TerraformUpgradeBackupDir),
				)
			}
			return state.Infrastructure{}, errors.New(abortErrorMsg)
		}
	}
	a.log.Debugf("Applying Terraform changes")

	a.spinner.Start(progressMsg, false)
	infraState, err := terraformClient.Apply(cmd.Context(), csp, rollbackBehavior)
	a.spinner.Stop()
	if err != nil {
		return state.Infrastructure{}, fmt.Errorf("applying terraform changes: %w", err)
	}

	cmd.Println(successMsg)
	return infraState, nil
}

func printCreateInfo(out io.Writer, conf *config.Config, log debugLog) error {
	controlPlaneGroup, ok := conf.NodeGroups[constants.DefaultControlPlaneGroupName]
	if !ok {
		return fmt.Errorf("default control-plane node group %q not found in configuration", constants.DefaultControlPlaneGroupName)
	}
	controlPlaneType := controlPlaneGroup.InstanceType

	workerGroup, ok := conf.NodeGroups[constants.DefaultWorkerGroupName]
	if !ok {
		return fmt.Errorf("default worker node group %q not found in configuration", constants.DefaultWorkerGroupName)
	}
	workerGroupType := workerGroup.InstanceType

	var qemuInstanceType string
	if conf.GetProvider() == cloudprovider.QEMU {
		qemuInstanceType = fmt.Sprintf("%d-vCPUs", conf.Provider.QEMU.VCPUs)
		controlPlaneType = qemuInstanceType
		workerGroupType = qemuInstanceType
	}

	otherGroupNames := make([]string, 0, len(conf.NodeGroups)-2)
	for groupName := range conf.NodeGroups {
		if groupName != constants.DefaultControlPlaneGroupName && groupName != constants.DefaultWorkerGroupName {
			otherGroupNames = append(otherGroupNames, groupName)
		}
	}
	if len(otherGroupNames) > 0 {
		log.Debugf("Creating %d additional node groups: %v", len(otherGroupNames), otherGroupNames)
	}

	fmt.Fprintf(out, "The following Constellation cluster will be created:\n")
	fmt.Fprintf(out, "  %d control-plane node%s of type %s will be created.\n", controlPlaneGroup.InitialCount, isPlural(controlPlaneGroup.InitialCount), controlPlaneType)
	fmt.Fprintf(out, "  %d worker node%s of type %s will be created.\n", workerGroup.InitialCount, isPlural(workerGroup.InitialCount), workerGroupType)
	for _, groupName := range otherGroupNames {
		group := conf.NodeGroups[groupName]
		groupInstanceType := group.InstanceType
		if conf.GetProvider() == cloudprovider.QEMU {
			groupInstanceType = qemuInstanceType
		}
		fmt.Fprintf(out, "  group %s with %d node%s of type %s will be created.\n", groupName, group.InitialCount, isPlural(group.InitialCount), groupInstanceType)
	}

	return nil
}
