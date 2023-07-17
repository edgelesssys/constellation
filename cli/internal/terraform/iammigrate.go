/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

type tfClient interface {
	PrepareIAMUpgradeWorkspace(rootDir, workingDir, newWorkingDir, backupDir string) error
	Plan(ctx context.Context, logLevel LogLevel, planFile string) (bool, error)
	ShowPlan(ctx context.Context, logLevel LogLevel, planFile string, outWriter io.Writer) error
	CreateIAMConfig(ctx context.Context, csp cloudprovider.Provider, logLevel LogLevel) (IAMOutput, error)
}

type MigrationCmd interface {
	Plan(ctx context.Context) (bool, error)
	Apply(ctx context.Context, fileHandler file.Handler) error
	String() string
}

type IAMMigrateCmd struct {
	tf        tfClient
	upgradeID string
	csp       cloudprovider.Provider
	logLevel  LogLevel
	outWriter io.Writer
}

func NewIAMMigrateCmd(tf tfClient, upgradeID string, csp cloudprovider.Provider, logLevel LogLevel, outWriter io.Writer) *IAMMigrateCmd {
	return &IAMMigrateCmd{
		tf:        tf,
		upgradeID: upgradeID,
		csp:       csp,
		logLevel:  logLevel,
		outWriter: outWriter,
	}
}

func (c *IAMMigrateCmd) String() string {
	return "iam migration"
}

func (c *IAMMigrateCmd) Plan(ctx context.Context) (bool, error) {
	templateDir := filepath.Join("terraform", "iam", strings.ToLower(c.csp.String()))
	err := c.tf.PrepareIAMUpgradeWorkspace(
		templateDir,
		constants.TerraformIAMWorkingDir,
		filepath.Join(constants.UpgradeDir, c.upgradeID, constants.TerraformIAMUpgradeWorkingDir),
		filepath.Join(constants.UpgradeDir, c.upgradeID, constants.TerraformIAMUpgradeBackupDir), // TODO: use IAM backup dir
	)
	if err != nil {
		return false, fmt.Errorf("preparing terraform workspace: %w", err)
	}

	hasDiff, err := c.tf.Plan(ctx, c.logLevel, constants.TerraformUpgradePlanFile)
	if err != nil {
		return false, fmt.Errorf("terraform plan 1: %w", err)
	}

	if hasDiff {
		if err := c.tf.ShowPlan(ctx, c.logLevel, constants.TerraformUpgradePlanFile, c.outWriter); err != nil {
			return false, fmt.Errorf("terraform show plan: %w", err)
		}
	}

	return hasDiff, nil
}

func (c *IAMMigrateCmd) Apply(ctx context.Context, fileHandler file.Handler) error {
	_, err := c.tf.CreateIAMConfig(ctx, c.csp, c.logLevel)

	// TODO: put in template?

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
