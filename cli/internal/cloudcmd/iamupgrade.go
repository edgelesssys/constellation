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
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// IAMUpgrader handles upgrades to IAM resources required by Constellation.
type IAMUpgrader struct {
	tf                tfIAMUpgradeClient
	existingWorkspace string
	upgradeWorkspace  string
	fileHandler       file.Handler
	logLevel          terraform.LogLevel
}

// NewIAMUpgrader creates and initializes a new IAMUpgrader.
// The existingWorkspace is the directory holding the existing Terraform resources.
// The upgradeWorkspace is the directory used for the upgrade.
func NewIAMUpgrader(ctx context.Context, existingWorkspace, upgradeWorkspace string,
	logLevel terraform.LogLevel, fileHandler file.Handler,
) (*IAMUpgrader, error) {
	tfClient, err := terraform.New(ctx, filepath.Join(upgradeWorkspace, constants.TerraformIAMUpgradeWorkingDir))
	if err != nil {
		return nil, fmt.Errorf("setting up terraform client: %w", err)
	}

	return &IAMUpgrader{
		tf:                tfClient,
		existingWorkspace: existingWorkspace,
		upgradeWorkspace:  upgradeWorkspace,
		fileHandler:       fileHandler,
		logLevel:          logLevel,
	}, nil
}

// PlanIAMUpgrade prepares the upgrade workspace and plans the Terraform migrations for the Constellation upgrade, writing the plan to the outWriter.
func (u *IAMUpgrader) PlanIAMUpgrade(ctx context.Context, outWriter io.Writer, csp cloudprovider.Provider) (bool, error) {
	if err := ensureFileNotExist(u.fileHandler, filepath.Join(u.upgradeWorkspace, constants.TerraformIAMUpgradeBackupDir)); err != nil {
		return false, fmt.Errorf("workspace is not clean: %w", err)
	}

	templateDir := filepath.Join("terraform", "iam", strings.ToLower(csp.String()))
	if err := terraform.PrepareIAMUpgradeWorkspace(
		u.fileHandler,
		templateDir,
		u.existingWorkspace,
		filepath.Join(u.upgradeWorkspace, constants.TerraformIAMUpgradeWorkingDir),
		filepath.Join(u.upgradeWorkspace, constants.TerraformIAMUpgradeBackupDir),
	); err != nil {
		return false, fmt.Errorf("preparing terraform workspace: %w", err)
	}

	hasDiff, err := u.tf.Plan(ctx, u.logLevel)
	if err != nil {
		return false, fmt.Errorf("terraform plan: %w", err)
	}

	if hasDiff {
		if err := u.tf.ShowPlan(ctx, u.logLevel, outWriter); err != nil {
			return false, fmt.Errorf("terraform show plan: %w", err)
		}
	}

	return hasDiff, nil
}

// ApplyIAMUpgrade applies the Terraform IAM migrations for the IAM upgrade.
func (u *IAMUpgrader) ApplyIAMUpgrade(ctx context.Context, csp cloudprovider.Provider) error {
	if _, err := u.tf.ApplyIAM(ctx, csp, u.logLevel); err != nil {
		return fmt.Errorf("terraform apply: %w", err)
	}

	if err := u.fileHandler.RemoveAll(u.existingWorkspace); err != nil {
		return fmt.Errorf("removing old terraform directory: %w", err)
	}
	if err := u.fileHandler.CopyDir(
		filepath.Join(u.upgradeWorkspace, constants.TerraformIAMUpgradeWorkingDir),
		u.existingWorkspace,
	); err != nil {
		return fmt.Errorf("replacing old terraform directory with new one: %w", err)
	}

	if err := u.fileHandler.RemoveAll(filepath.Join(u.upgradeWorkspace, constants.TerraformIAMUpgradeWorkingDir)); err != nil {
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

type tfIAMUpgradeClient interface {
	tfPlanner
	ApplyIAM(ctx context.Context, csp cloudprovider.Provider, logLevel terraform.LogLevel) (terraform.IAMOutput, error)
}
