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

// UpgradeRequiresIAMMigration returns true if the given cloud provider requires an IAM migration.
func UpgradeRequiresIAMMigration(provider cloudprovider.Provider) bool {
	switch provider {
	default:
		return false
	}
}

// IAMUpgrader handles upgrades to IAM resources required by Constellation.
type IAMUpgrader struct {
	tf                tfIAMUpgradeClient
	existingWorkspace string
	upgradeWorkspace  string
	fileHandler       file.Handler
	logLevel          terraform.LogLevel
}

// NewIAMUpgrader creates and initializes a new IAMUpgrader.
// existingWorkspace is the directory holding the existing Terraform resources.
// upgradeWorkspace is the directory to use for holding temporary files and resources required to apply the upgrade.
func NewIAMUpgrader(ctx context.Context, existingWorkspace, upgradeWorkspace string,
	logLevel terraform.LogLevel, fileHandler file.Handler,
) (*IAMUpgrader, error) {
	tfClient, err := terraform.New(ctx, existingWorkspace)
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

// PlanIAMUpgrade prepares the upgrade workspace and plans the possible Terraform migrations for Constellation's IAM resources (service accounts, permissions etc.).
// In case of possible migrations, the diff is written to outWriter and this function returns true.
func (u *IAMUpgrader) PlanIAMUpgrade(ctx context.Context, outWriter io.Writer, vars terraform.Variables, csp cloudprovider.Provider) (bool, error) {
	return planApply(
		ctx, u.tf, u.fileHandler, outWriter, u.logLevel, vars,
		filepath.Join(constants.TerraformEmbeddedDir, "iam", strings.ToLower(csp.String())),
		u.existingWorkspace,
		filepath.Join(u.upgradeWorkspace, constants.TerraformIAMUpgradeBackupDir),
	)
}

// RestoreIAMWorkspace rolls back the existing workspace to the backup directory created when planning an upgrade,
// when the user decides to not apply an upgrade after planning it.
// Note that this will not apply the restored state from the backup.
func (u *IAMUpgrader) RestoreIAMWorkspace() error {
	return restoreBackup(u.fileHandler, u.existingWorkspace, filepath.Join(u.upgradeWorkspace, constants.TerraformIAMUpgradeBackupDir))
}

// ApplyIAMUpgrade applies the Terraform IAM migrations planned by PlanIAMUpgrade.
// On success, the workspace of the Upgrader replaces the existing Terraform workspace.
func (u *IAMUpgrader) ApplyIAMUpgrade(ctx context.Context, csp cloudprovider.Provider) error {
	if _, err := u.tf.ApplyIAM(ctx, csp, u.logLevel); err != nil {
		return fmt.Errorf("terraform apply: %w", err)
	}

	return nil
}
