/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

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

// ClusterUpgrader is responsible for performing Terraform migrations on cluster upgrades.
type ClusterUpgrader struct {
	tf                tfClusterUpgradeClient
	policyPatcher     policyPatcher
	fileHandler       file.Handler
	existingWorkspace string
	upgradeWorkspace  string
	logLevel          terraform.LogLevel
}

// NewClusterUpgrader initializes and returns a new ClusterUpgrader.
func NewClusterUpgrader(ctx context.Context, existingWorkspace, upgradeWorkspace string,
	logLevel terraform.LogLevel, fileHandler file.Handler,
) (*ClusterUpgrader, error) {
	tfClient, err := terraform.New(ctx, filepath.Join(upgradeWorkspace, constants.TerraformUpgradeWorkingDir))
	if err != nil {
		return nil, fmt.Errorf("setting up terraform client: %w", err)
	}

	return &ClusterUpgrader{
		tf:                tfClient,
		policyPatcher:     NewAzurePolicyPatcher(),
		fileHandler:       fileHandler,
		existingWorkspace: existingWorkspace,
		upgradeWorkspace:  upgradeWorkspace,
		logLevel:          logLevel,
	}, nil
}

// PlanClusterUpgrade prepares the upgrade workspace and plans the possible migrations for Constellation's cluster resources.
// If a diff exists, it's being written to the upgrader's output writer. It also returns
// a bool indicating whether a diff exists.
func (u *ClusterUpgrader) PlanClusterUpgrade(ctx context.Context, outWriter io.Writer, vars terraform.Variables, csp cloudprovider.Provider,
) (bool, error) {
	if err := ensureFileNotExist(u.fileHandler, filepath.Join(u.upgradeWorkspace, constants.TerraformUpgradeBackupDir)); err != nil {
		return false, fmt.Errorf("workspace is not clean: %w", err)
	}

	// Prepare the new Terraform workspace and backup the old one
	err := u.tf.PrepareUpgradeWorkspace(
		filepath.Join("terraform", strings.ToLower(csp.String())),
		u.existingWorkspace,
		filepath.Join(u.upgradeWorkspace, constants.TerraformUpgradeWorkingDir),
		filepath.Join(u.upgradeWorkspace, constants.TerraformUpgradeBackupDir),
		vars,
	)
	if err != nil {
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

// ApplyClusterUpgrade applies the migrations planned by PlanClusterUpgrade.
// If PlanTerraformMigrations has not been executed before, it will return an error.
// In case of a successful upgrade, the output will be written to the specified file and the old Terraform directory is replaced
// By the new one.
func (u *ClusterUpgrader) ApplyClusterUpgrade(ctx context.Context, csp cloudprovider.Provider) (terraform.ApplyOutput, error) {
	tfOutput, err := u.tf.ApplyCluster(ctx, csp, u.logLevel)
	if err != nil {
		return tfOutput, fmt.Errorf("terraform apply: %w", err)
	}
	if tfOutput.Azure != nil {
		if err := u.policyPatcher.Patch(ctx, tfOutput.Azure.AttestationURL); err != nil {
			return tfOutput, fmt.Errorf("patching policies: %w", err)
		}
	}
	if err := u.fileHandler.RemoveAll(u.existingWorkspace); err != nil {
		return tfOutput, fmt.Errorf("removing old terraform directory: %w", err)
	}
	if err := u.fileHandler.CopyDir(
		filepath.Join(u.upgradeWorkspace, constants.TerraformUpgradeWorkingDir),
		u.existingWorkspace,
	); err != nil {
		return tfOutput, fmt.Errorf("replacing old terraform directory with new one: %w", err)
	}

	if err := u.fileHandler.RemoveAll(filepath.Join(u.upgradeWorkspace, constants.TerraformUpgradeWorkingDir)); err != nil {
		return tfOutput, fmt.Errorf("removing terraform upgrade directory: %w", err)
	}
	return tfOutput, nil
}

type tfClusterUpgradeClient interface {
	tfPlanner
	ApplyCluster(ctx context.Context, provider cloudprovider.Provider, logLevel terraform.LogLevel) (terraform.ApplyOutput, error)
	PrepareUpgradeWorkspace(embeddedPath, oldWorkingDir, newWorkingDir, backupDir string, vars terraform.Variables) error
}
