/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgrade

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// IAMMigrateCmd is a terraform migration command for IAM. Which is used for the tfMigrationClient.
type IAMMigrateCmd struct {
	tf               tfIAMClient
	upgradeID        string
	iamWorkspace     string
	upgradeWorkspace string
	csp              cloudprovider.Provider
	logLevel         terraform.LogLevel
}

// NewIAMMigrateCmd creates a new IAMMigrateCmd.
func NewIAMMigrateCmd(ctx context.Context, iamWorkspace, upgradeWorkspace, upgradeID string, csp cloudprovider.Provider, logLevel terraform.LogLevel) (*IAMMigrateCmd, error) {
	tfClient, err := terraform.New(ctx, filepath.Join(upgradeWorkspace, upgradeID, constants.TerraformIAMUpgradeWorkingDir))
	if err != nil {
		return nil, fmt.Errorf("setting up terraform client: %w", err)
	}
	return &IAMMigrateCmd{
		tf:               tfClient,
		upgradeID:        upgradeID,
		iamWorkspace:     iamWorkspace,
		upgradeWorkspace: upgradeWorkspace,
		csp:              csp,
		logLevel:         logLevel,
	}, nil
}

// String returns the name of the command.
func (c *IAMMigrateCmd) String() string {
	return "iam migration"
}

// UpgradeID returns the upgrade ID.
func (c *IAMMigrateCmd) UpgradeID() string {
	return c.upgradeID
}

// CheckTerraformMigrations checks whether Terraform migrations are possible in the current workspace.
func (c *IAMMigrateCmd) CheckTerraformMigrations(file file.Handler) error {
	return checkTerraformMigrations(file, c.upgradeWorkspace, c.upgradeID, constants.TerraformIAMUpgradeBackupDir)
}

// Plan prepares the upgrade workspace and plans the Terraform migrations for the Constellation upgrade, writing the plan to the outWriter.
func (c *IAMMigrateCmd) Plan(ctx context.Context, file file.Handler, outWriter io.Writer) (bool, error) {
	templateDir := filepath.Join("terraform", "iam", strings.ToLower(c.csp.String()))
	if err := terraform.PrepareIAMUpgradeWorkspace(file,
		templateDir,
		c.iamWorkspace,
		filepath.Join(c.upgradeWorkspace, c.upgradeID, constants.TerraformIAMUpgradeWorkingDir),
		filepath.Join(c.upgradeWorkspace, c.upgradeID, constants.TerraformIAMUpgradeBackupDir),
	); err != nil {
		return false, fmt.Errorf("preparing terraform workspace: %w", err)
	}

	hasDiff, err := c.tf.Plan(ctx, c.logLevel)
	if err != nil {
		return false, fmt.Errorf("terraform plan: %w", err)
	}

	if hasDiff {
		if err := c.tf.ShowPlan(ctx, c.logLevel, outWriter); err != nil {
			return false, fmt.Errorf("terraform show plan: %w", err)
		}
	}

	return hasDiff, nil
}

// Apply applies the Terraform IAM migrations for the Constellation upgrade.
func (c *IAMMigrateCmd) Apply(ctx context.Context, fileHandler file.Handler) error {
	if _, err := c.tf.ApplyIAM(ctx, c.csp, c.logLevel); err != nil {
		return fmt.Errorf("terraform apply: %w", err)
	}

	if err := fileHandler.RemoveAll(c.iamWorkspace); err != nil {
		return fmt.Errorf("removing old terraform directory: %w", err)
	}
	if err := fileHandler.CopyDir(
		filepath.Join(c.upgradeWorkspace, c.upgradeID, constants.TerraformIAMUpgradeWorkingDir),
		c.iamWorkspace,
	); err != nil {
		return fmt.Errorf("replacing old terraform directory with new one: %w", err)
	}

	if err := fileHandler.RemoveAll(filepath.Join(c.upgradeWorkspace, c.upgradeID, constants.TerraformIAMUpgradeWorkingDir)); err != nil {
		return fmt.Errorf("removing terraform upgrade directory: %w", err)
	}

	return nil
}
