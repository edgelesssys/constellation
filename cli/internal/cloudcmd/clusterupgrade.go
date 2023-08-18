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
	return planUpgrade(
		ctx, u.tf, u.fileHandler, outWriter, u.logLevel, vars,
		filepath.Join("terraform", strings.ToLower(csp.String())),
		u.existingWorkspace,
		filepath.Join(u.upgradeWorkspace, constants.TerraformUpgradeBackupDir),
	)
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

	if err := moveUpgradeToCurrent(
		u.fileHandler,
		u.existingWorkspace,
		filepath.Join(u.upgradeWorkspace, constants.TerraformUpgradeWorkingDir),
	); err != nil {
		return tfOutput, fmt.Errorf("promoting upgrade workspace to current workspace: %w", err)
	}

	return tfOutput, nil
}
