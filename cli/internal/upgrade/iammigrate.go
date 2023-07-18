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

// MigrationCmd is an interface for all terraform upgrade / migration commands.
type MigrationCmd interface {
	CheckTerraformMigrations(file file.Handler) error
	Plan(ctx context.Context, file file.Handler, outWriter io.Writer) (bool, error)
	Apply(ctx context.Context, fileHandler file.Handler) error
	String() string
}

// IAMMigrateCmd is a terraform migration command for IAM.
type IAMMigrateCmd struct {
	tf        tfIAMClient
	upgradeID string
	csp       cloudprovider.Provider
	logLevel  terraform.LogLevel
}

// NewIAMMigrateCmd creates a new IAMMigrateCmd.
func NewIAMMigrateCmd(ctx context.Context, upgradeID string, csp cloudprovider.Provider, logLevel terraform.LogLevel) (*IAMMigrateCmd, error) {
	tfClient, err := terraform.New(ctx, filepath.Join(constants.UpgradeDir, upgradeID, constants.TerraformIAMUpgradeWorkingDir))
	if err != nil {
		return nil, fmt.Errorf("setting up terraform client: %w", err)
	}
	return &IAMMigrateCmd{
		tf:        tfClient,
		upgradeID: upgradeID,
		csp:       csp,
		logLevel:  logLevel,
	}, nil
}

// String returns the name of the command.
func (c *IAMMigrateCmd) String() string {
	return "iam migration"
}

// CheckTerraformMigrations checks whether Terraform migrations are possible in the current workspace.
func (c *IAMMigrateCmd) CheckTerraformMigrations(file file.Handler) error {
	return CheckTerraformMigrations(file, c.upgradeID, constants.TerraformIAMUpgradeBackupDir)
}

// Plan prepares the upgrade workspace and plans the Terraform migrations for the Constellation upgrade, writing the plan to the outWriter.
// TODO put outWriter as argument.
func (c *IAMMigrateCmd) Plan(ctx context.Context, file file.Handler, outWriter io.Writer) (bool, error) {
	templateDir := filepath.Join("terraform", "iam", strings.ToLower(c.csp.String()))
	err := terraform.PrepareIAMUpgradeWorkspace(file,
		templateDir,
		constants.TerraformIAMWorkingDir,
		filepath.Join(constants.UpgradeDir, c.upgradeID, constants.TerraformIAMUpgradeWorkingDir),
		filepath.Join(constants.UpgradeDir, c.upgradeID, constants.TerraformIAMUpgradeBackupDir),
	)
	if err != nil {
		return false, fmt.Errorf("preparing terraform workspace: %w", err)
	}

	hasDiff, err := c.tf.Plan(ctx, c.logLevel, constants.TerraformUpgradePlanFile)
	if err != nil {
		return false, fmt.Errorf("terraform plan: %w", err)
	}

	if hasDiff {
		if err := c.tf.ShowPlan(ctx, c.logLevel, constants.TerraformUpgradePlanFile, outWriter); err != nil {
			return false, fmt.Errorf("terraform show plan: %w", err)
		}
	}

	return hasDiff, nil
}

// Apply applies the Terraform IAM migrations for the Constellation upgrade.
func (c *IAMMigrateCmd) Apply(ctx context.Context, fileHandler file.Handler) error {
	_, err := c.tf.CreateIAMConfig(ctx, c.csp, c.logLevel) // TODO rename CreateIAMConfig to ApplyIAMConfig to reflect usage for migration too

	// TODO: put in template, since moving files is also done in other TF migrations
	if err := fileHandler.RemoveAll(constants.TerraformIAMWorkingDir); err != nil {
		return fmt.Errorf("removing old terraform directory: %w", err)
	}
	if err := fileHandler.CopyDir(filepath.Join(constants.UpgradeDir, c.upgradeID, constants.TerraformIAMUpgradeWorkingDir), constants.TerraformIAMWorkingDir); err != nil {
		return fmt.Errorf("replacing old terraform directory with new one: %w", err)
	}

	if err := fileHandler.RemoveAll(filepath.Join(constants.UpgradeDir, c.upgradeID, constants.TerraformIAMUpgradeWorkingDir)); err != nil {
		return fmt.Errorf("removing terraform upgrade directory: %w", err)
	}

	return err
}
