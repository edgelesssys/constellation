/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgrade

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/cloudcmd"
	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// NewTerraformUpgrader returns a new TerraformUpgrader.
func NewTerraformUpgrader(tfClient tfClient, outWriter io.Writer, fileHandler file.Handler) (*TerraformUpgrader, error) {
	return &TerraformUpgrader{
		tf:            tfClient,
		policyPatcher: cloudcmd.NewAzurePolicyPatcher(),
		outWriter:     outWriter,
		fileHandler:   fileHandler,
	}, nil
}

// TerraformUpgrader is responsible for performing Terraform migrations on cluster upgrades.
type TerraformUpgrader struct {
	tf            tfClient
	policyPatcher policyPatcher
	outWriter     io.Writer
	fileHandler   file.Handler
}

// TerraformUpgradeOptions are the options used for the Terraform upgrade.
type TerraformUpgradeOptions struct {
	// LogLevel is the log level used for Terraform.
	LogLevel terraform.LogLevel
	// CSP is the cloud provider to perform the upgrade on.
	CSP cloudprovider.Provider
	// Vars are the Terraform variables used for the upgrade.
	Vars terraform.Variables
}

// CheckTerraformMigrations checks whether Terraform migrations are possible in the current workspace.
// If the files that will be written during the upgrade already exist, it returns an error.
func (u *TerraformUpgrader) CheckTerraformMigrations(upgradeID string) error {
	var existingFiles []string
	filesToCheck := []string{
		filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformUpgradeBackupDir),
	}

	for _, f := range filesToCheck {
		if err := checkFileExists(u.fileHandler, &existingFiles, f); err != nil {
			return fmt.Errorf("checking terraform migrations: %w", err)
		}
	}

	if len(existingFiles) > 0 {
		return fmt.Errorf("file(s) %s already exist", strings.Join(existingFiles, ", "))
	}
	return nil
}

// checkFileExists checks whether a file exists and adds it to the existingFiles slice if it does.
func checkFileExists(fileHandler file.Handler, existingFiles *[]string, filename string) error {
	_, err := fileHandler.Stat(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("checking %s: %w", filename, err)
		}
		return nil
	}

	*existingFiles = append(*existingFiles, filename)
	return nil
}

// PlanTerraformMigrations prepares the upgrade workspace and plans the Terraform migrations for the Constellation upgrade.
// If a diff exists, it's being written to the upgrader's output writer. It also returns
// a bool indicating whether a diff exists.
func (u *TerraformUpgrader) PlanTerraformMigrations(ctx context.Context, opts TerraformUpgradeOptions, upgradeID string) (bool, error) {
	// Prepare the new Terraform workspace and backup the old one
	err := u.tf.PrepareUpgradeWorkspace(
		filepath.Join("terraform", strings.ToLower(opts.CSP.String())),
		constants.TerraformWorkingDir,
		filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformUpgradeWorkingDir),
		filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformUpgradeBackupDir),
		opts.Vars,
	)
	if err != nil {
		return false, fmt.Errorf("preparing terraform workspace: %w", err)
	}

	// Backup the old constellation-id.json file
	if err := u.fileHandler.CopyFile(
		constants.ClusterIDsFileName,
		filepath.Join(constants.UpgradeDir, upgradeID, constants.ClusterIDsFileName+".old"),
	); err != nil {
		return false, fmt.Errorf("backing up %s: %w", constants.ClusterIDsFileName, err)
	}

	hasDiff, err := u.tf.Plan(ctx, opts.LogLevel, constants.TerraformUpgradePlanFile)
	if err != nil {
		return false, fmt.Errorf("terraform plan: %w", err)
	}

	if hasDiff {
		if err := u.tf.ShowPlan(ctx, opts.LogLevel, constants.TerraformUpgradePlanFile, u.outWriter); err != nil {
			return false, fmt.Errorf("terraform show plan: %w", err)
		}
	}

	return hasDiff, nil
}

// CleanUpTerraformMigrations cleans up the Terraform migration workspace, for example when an upgrade is
// aborted by the user.
func (u *TerraformUpgrader) CleanUpTerraformMigrations(upgradeID string) error {
	cleanupFiles := []string{
		filepath.Join(constants.UpgradeDir, upgradeID),
	}

	for _, f := range cleanupFiles {
		if err := u.fileHandler.RemoveAll(f); err != nil {
			return fmt.Errorf("cleaning up file %s: %w", f, err)
		}
	}

	return nil
}

// ApplyTerraformMigrations applies the migerations planned by PlanTerraformMigrations.
// If PlanTerraformMigrations has not been executed before, it will return an error.
// In case of a successful upgrade, the output will be written to the specified file and the old Terraform directory is replaced
// By the new one.
func (u *TerraformUpgrader) ApplyTerraformMigrations(ctx context.Context, opts TerraformUpgradeOptions, upgradeID string) error {
	tfOutput, err := u.tf.CreateCluster(ctx, opts.LogLevel)
	if err != nil {
		return fmt.Errorf("terraform apply: %w", err)
	}

	// AttestationURL is only set for Azure.
	if tfOutput.AttestationURL != "" {
		if err := u.policyPatcher.Patch(ctx, tfOutput.AttestationURL); err != nil {
			return fmt.Errorf("patching policies: %w", err)
		}
	}

	outputFileContents := clusterid.File{
		CloudProvider:  opts.CSP,
		InitSecret:     []byte(tfOutput.Secret),
		IP:             tfOutput.IP,
		UID:            tfOutput.UID,
		AttestationURL: tfOutput.AttestationURL,
	}

	if err := u.fileHandler.RemoveAll(constants.TerraformWorkingDir); err != nil {
		return fmt.Errorf("removing old terraform directory: %w", err)
	}

	if err := u.fileHandler.CopyDir(filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformUpgradeWorkingDir), constants.TerraformWorkingDir); err != nil {
		return fmt.Errorf("replacing old terraform directory with new one: %w", err)
	}

	if err := u.fileHandler.RemoveAll(filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformUpgradeWorkingDir)); err != nil {
		return fmt.Errorf("removing terraform upgrade directory: %w", err)
	}

	if err := u.mergeClusterIDFile(outputFileContents); err != nil {
		return fmt.Errorf("merging migration output into %s: %w", constants.ClusterIDsFileName, err)
	}

	return nil
}

// mergeClusterIDFile merges the output of the migration into the constellation-id.json file.
func (u *TerraformUpgrader) mergeClusterIDFile(migrationOutput clusterid.File) error {
	idFile := &clusterid.File{}
	if err := u.fileHandler.ReadJSON(constants.ClusterIDsFileName, idFile); err != nil {
		return fmt.Errorf("reading %s: %w", constants.ClusterIDsFileName, err)
	}

	if err := u.fileHandler.WriteJSON(constants.ClusterIDsFileName, idFile.Merge(migrationOutput), file.OptOverwrite); err != nil {
		return fmt.Errorf("writing %s: %w", constants.ClusterIDsFileName, err)
	}

	return nil
}

// a tfClient performs the Terraform interactions in an upgrade.
type tfClient interface {
	PrepareUpgradeWorkspace(path, oldWorkingDir, newWorkingDir, upgradeID string, vars terraform.Variables) error
	ShowPlan(ctx context.Context, logLevel terraform.LogLevel, planFilePath string, output io.Writer) error
	Plan(ctx context.Context, logLevel terraform.LogLevel, planFile string) (bool, error)
	CreateCluster(ctx context.Context, logLevel terraform.LogLevel) (terraform.CreateOutput, error)
}

// policyPatcher interacts with the CSP (currently only applies for Azure) to update the attestation policy.
type policyPatcher interface {
	Patch(ctx context.Context, attestationURL string) error
}
