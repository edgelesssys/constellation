/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// planUpgrade prepares a workspace and plans the possible Terraform migrations.
// In case of possible migrations, the diff is written to outWriter and this function returns true.
func planUpgrade(
	ctx context.Context, tfClient tfUpgradePlanner, fileHandler file.Handler,
	outWriter io.Writer, logLevel terraform.LogLevel, vars terraform.Variables,
	templateDir, backupDir string,
) (bool, error) {
	if err := ensureFileNotExist(fileHandler, backupDir); err != nil {
		return false, fmt.Errorf("backup directory %s already exists: %w", backupDir, err)
	}

	// Backup the old Terraform workspace and move the embedded Terraform files into the workspace.
	err := tfClient.PrepareUpgradeWorkspace(
		templateDir,
		backupDir,
		vars,
	)
	if err != nil {
		return false, fmt.Errorf("preparing terraform workspace: %w", err)
	}

	hasDiff, err := tfClient.Plan(ctx, logLevel)
	if err != nil {
		return false, fmt.Errorf("terraform plan: %w", err)
	}

	if hasDiff {
		if err := tfClient.ShowPlan(ctx, logLevel, outWriter); err != nil {
			return false, fmt.Errorf("terraform show plan: %w", err)
		}
	}

	return hasDiff, nil
}

// restoreBackup replaces the existing Terraform workspace with the backup.
func restoreBackup(fileHandler file.Handler, workingDir, backupDir string) error {
	if err := fileHandler.RemoveAll(workingDir); err != nil {
		return fmt.Errorf("removing existing workspace: %w", err)
	}
	if err := fileHandler.CopyDir(
		backupDir,
		workingDir,
	); err != nil {
		return fmt.Errorf("replacing terraform workspace with backup: %w", err)
	}

	if err := fileHandler.RemoveAll(backupDir); err != nil {
		return fmt.Errorf("removing backup directory: %w", err)
	}
	return nil
}

// ensureFileNotExist checks if a single file or directory does not exist, returning an error if it does.
func ensureFileNotExist(fileHandler file.Handler, fileName string) error {
	if _, err := fileHandler.Stat(fileName); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("checking %q: %w", fileName, err)
		}
		return nil
	}
	return fmt.Errorf("%q already exists", fileName)
}
