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
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// TerraformUpgradeOptions are the options used for the Terraform upgrade.
type TerraformUpgradeOptions struct {
	// LogLevel is the log level used for Terraform.
	LogLevel terraform.LogLevel
	// CSP is the cloud provider to perform the upgrade on.
	CSP cloudprovider.Provider
	// Vars are the Terraform variables used for the upgrade.
	Vars             terraform.Variables
	TFWorkspace      string
	UpgradeWorkspace string
}

// TerraformUpgrader is responsible for performing Terraform migrations on cluster upgrades.
type TerraformUpgrader struct {
	tf            tfResourceClient
	policyPatcher policyPatcher
	outWriter     io.Writer
	fileHandler   file.Handler
	upgradeID     string
}

// NewTerraformUpgrader returns a new TerraformUpgrader.
func NewTerraformUpgrader(tfClient tfResourceClient, outWriter io.Writer, fileHandler file.Handler, upgradeID string,
) *TerraformUpgrader {
	return &TerraformUpgrader{
		tf:            tfClient,
		policyPatcher: cloudcmd.NewAzurePolicyPatcher(),
		outWriter:     outWriter,
		fileHandler:   fileHandler,
		upgradeID:     upgradeID,
	}
}

// CheckTerraformMigrations checks whether Terraform migrations are possible in the current workspace.
// If the files that will be written during the upgrade already exist, it returns an error.
func (u *TerraformUpgrader) CheckTerraformMigrations(upgradeWorkspace string) error {
	return checkTerraformMigrations(u.fileHandler, upgradeWorkspace, u.upgradeID, constants.TerraformUpgradeBackupDir)
}

// PlanTerraformMigrations prepares the upgrade workspace and plans the Terraform migrations for the Constellation upgrade.
// If a diff exists, it's being written to the upgrader's output writer. It also returns
// a bool indicating whether a diff exists.
func (u *TerraformUpgrader) PlanTerraformMigrations(ctx context.Context, opts TerraformUpgradeOptions) (bool, error) {
	// Prepare the new Terraform workspace and backup the old one
	err := u.tf.PrepareUpgradeWorkspace(
		filepath.Join("terraform", strings.ToLower(opts.CSP.String())),
		opts.TFWorkspace,
		filepath.Join(opts.UpgradeWorkspace, u.upgradeID, constants.TerraformUpgradeWorkingDir),
		filepath.Join(opts.UpgradeWorkspace, u.upgradeID, constants.TerraformUpgradeBackupDir),
		opts.Vars,
	)
	if err != nil {
		return false, fmt.Errorf("preparing terraform workspace: %w", err)
	}

	hasDiff, err := u.tf.Plan(ctx, opts.LogLevel)
	if err != nil {
		return false, fmt.Errorf("terraform plan: %w", err)
	}

	if hasDiff {
		if err := u.tf.ShowPlan(ctx, opts.LogLevel, u.outWriter); err != nil {
			return false, fmt.Errorf("terraform show plan: %w", err)
		}
	}

	return hasDiff, nil
}

// CleanUpTerraformMigrations cleans up the Terraform migration workspace, for example when an upgrade is
// aborted by the user.
func (u *TerraformUpgrader) CleanUpTerraformMigrations(upgradeWorkspace string) error {
	return CleanUpTerraformMigrations(upgradeWorkspace, u.upgradeID, u.fileHandler)
}

// ApplyTerraformMigrations applies the migrations planned by PlanTerraformMigrations.
// If PlanTerraformMigrations has not been executed before, it will return an error.
// In case of a successful upgrade, the output will be written to the specified file and the old Terraform directory is replaced
// By the new one.
func (u *TerraformUpgrader) ApplyTerraformMigrations(ctx context.Context, opts TerraformUpgradeOptions) (terraform.ApplyOutput, error) {
	tfOutput, err := u.tf.CreateCluster(ctx, opts.CSP, opts.LogLevel)
	if err != nil {
		return tfOutput, fmt.Errorf("terraform apply: %w", err)
	}
	if tfOutput.Azure != nil {
		if err := u.policyPatcher.Patch(ctx, tfOutput.Azure.AttestationURL); err != nil {
			return tfOutput, fmt.Errorf("patching policies: %w", err)
		}
	}
	if err := u.fileHandler.RemoveAll(opts.TFWorkspace); err != nil {
		return tfOutput, fmt.Errorf("removing old terraform directory: %w", err)
	}

	if err := u.fileHandler.CopyDir(
		filepath.Join(opts.UpgradeWorkspace, u.upgradeID, constants.TerraformUpgradeWorkingDir),
		opts.TFWorkspace,
	); err != nil {
		return tfOutput, fmt.Errorf("replacing old terraform directory with new one: %w", err)
	}
	if err := u.fileHandler.RemoveAll(filepath.Join(opts.UpgradeWorkspace, u.upgradeID, constants.TerraformUpgradeWorkingDir)); err != nil {
		return tfOutput, fmt.Errorf("removing terraform upgrade directory: %w", err)
	}
	return tfOutput, nil
}

// UpgradeID returns the ID of the upgrade.
func (u *TerraformUpgrader) UpgradeID() string {
	return u.upgradeID
}

// CleanUpTerraformMigrations cleans up the Terraform upgrade directory.
func CleanUpTerraformMigrations(upgradeWorkspace, upgradeID string, fileHandler file.Handler) error {
	upgradeDir := filepath.Join(upgradeWorkspace, upgradeID)
	if err := fileHandler.RemoveAll(upgradeDir); err != nil {
		return fmt.Errorf("cleaning up file %s: %w", upgradeDir, err)
	}
	return nil
}

// CheckTerraformMigrations checks whether Terraform migrations are possible in the current workspace.
func checkTerraformMigrations(file file.Handler, upgradeWorkspace, upgradeID, upgradeSubDir string) error {
	var existingFiles []string
	filesToCheck := []string{
		filepath.Join(upgradeWorkspace, upgradeID, upgradeSubDir),
	}

	for _, f := range filesToCheck {
		if err := checkFileExists(file, &existingFiles, f); err != nil {
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

type tfClientCommon interface {
	ShowPlan(ctx context.Context, logLevel terraform.LogLevel, output io.Writer) error
	Plan(ctx context.Context, logLevel terraform.LogLevel) (bool, error)
}

// tfResourceClient is a Terraform client for managing cluster resources.
type tfResourceClient interface {
	PrepareUpgradeWorkspace(embeddedPath, oldWorkingDir, newWorkingDir, backupDir string, vars terraform.Variables) error
	CreateCluster(ctx context.Context, provider cloudprovider.Provider, logLevel terraform.LogLevel) (terraform.ApplyOutput, error)
	tfClientCommon
}

// tfIAMClient is a Terraform client for managing IAM resources.
type tfIAMClient interface {
	ApplyIAMConfig(ctx context.Context, csp cloudprovider.Provider, logLevel terraform.LogLevel) (terraform.IAMOutput, error)
	tfClientCommon
}

// policyPatcher interacts with the CSP (currently only applies for Azure) to update the attestation policy.
type policyPatcher interface {
	Patch(ctx context.Context, attestationURL string) error
}
