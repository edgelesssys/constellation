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

func planUpgrade(
	ctx context.Context, tfClient tfUpgradePlanner, fileHandler file.Handler,
	outWriter io.Writer, logLevel terraform.LogLevel, vars terraform.Variables,
	templateDir, existingWorkspace, backupDir string,
) (bool, error) {
	if err := ensureFileNotExist(fileHandler, backupDir); err != nil {
		return false, fmt.Errorf("workspace is not clean: %w", err)
	}

	// Prepare the new Terraform workspace and backup the old one
	err := tfClient.PrepareUpgradeWorkspace(
		templateDir,
		existingWorkspace,
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

func moveUpgradeToCurrent(fileHandler file.Handler, existingWorkspace, upgradeWorkingDir string) error {
	if err := fileHandler.RemoveAll(existingWorkspace); err != nil {
		return fmt.Errorf("removing old terraform directory: %w", err)
	}
	if err := fileHandler.CopyDir(
		upgradeWorkingDir,
		existingWorkspace,
	); err != nil {
		return fmt.Errorf("replacing old terraform directory with new one: %w", err)
	}

	if err := fileHandler.RemoveAll(upgradeWorkingDir); err != nil {
		return fmt.Errorf("removing terraform upgrade directory: %w", err)
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
