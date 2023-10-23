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

	"github.com/edgelesssys/constellation/v2/cli/internal/state"
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
// existingWorkspace is the directory holding the existing Terraform resources.
// upgradeWorkspace is the directory to use for holding temporary files and resources required to apply the upgrade.
func NewClusterUpgrader(ctx context.Context, existingWorkspace, upgradeWorkspace string,
	logLevel terraform.LogLevel, fileHandler file.Handler,
) (*ClusterUpgrader, error) {
	tfClient, err := terraform.New(ctx, constants.TerraformWorkingDir)
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

// PlanClusterUpgrade prepares the upgrade workspace and plans the possible Terraform migrations for Constellation's cluster resources (Loadbalancers, VMs, networks etc.).
// In case of possible migrations, the diff is written to outWriter and this function returns true.
func (u *ClusterUpgrader) PlanClusterUpgrade(ctx context.Context, outWriter io.Writer, vars terraform.Variables, csp cloudprovider.Provider,
) (bool, error) {
	return planUpgrade(
		ctx, u.tf, u.fileHandler, outWriter, u.logLevel, vars,
		filepath.Join(constants.TerraformEmbeddedDir, strings.ToLower(csp.String())),
		filepath.Join(u.upgradeWorkspace, constants.TerraformUpgradeBackupDir),
	)
}

// RestoreClusterWorkspace rolls back the existing workspace to the backup directory created when planning an upgrade,
// when the user decides to not apply an upgrade after planning it.
// Note that this will not apply the restored state from the backup.
func (u *ClusterUpgrader) RestoreClusterWorkspace() error {
	return restoreBackup(u.fileHandler, u.existingWorkspace, filepath.Join(u.upgradeWorkspace, constants.TerraformUpgradeBackupDir))
}

// ApplyClusterUpgrade applies the Terraform migrations planned by PlanClusterUpgrade.
// On success, the workspace of the Upgrader replaces the existing Terraform workspace.
func (u *ClusterUpgrader) ApplyClusterUpgrade(ctx context.Context, csp cloudprovider.Provider) (state.Infrastructure, error) {
	infraState, err := u.tf.ApplyCluster(ctx, csp, u.logLevel)
	if err != nil {
		return infraState, fmt.Errorf("terraform apply: %w", err)
	}
	if infraState.Azure != nil {
		if err := u.policyPatcher.Patch(ctx, infraState.Azure.AttestationURL); err != nil {
			return infraState, fmt.Errorf("patching policies: %w", err)
		}
	}

	return infraState, nil
}
