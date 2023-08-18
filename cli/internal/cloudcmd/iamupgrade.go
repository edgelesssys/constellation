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
func (u *IAMUpgrader) PlanIAMUpgrade(ctx context.Context, outWriter io.Writer, vars terraform.Variables, csp cloudprovider.Provider) (bool, error) {
	return planUpgrade(
		ctx, u.tf, u.fileHandler, outWriter, u.logLevel, vars,
		filepath.Join("terraform", "iam", strings.ToLower(csp.String())),
		u.existingWorkspace,
		filepath.Join(u.upgradeWorkspace, constants.TerraformIAMUpgradeBackupDir),
	)
}

// ApplyIAMUpgrade applies the Terraform IAM migrations for the IAM upgrade.
func (u *IAMUpgrader) ApplyIAMUpgrade(ctx context.Context, csp cloudprovider.Provider) error {
	if _, err := u.tf.ApplyIAM(ctx, csp, u.logLevel); err != nil {
		return fmt.Errorf("terraform apply: %w", err)
	}

	if err := moveUpgradeToCurrent(
		u.fileHandler,
		u.existingWorkspace,
		filepath.Join(u.upgradeWorkspace, constants.TerraformIAMUpgradeWorkingDir),
	); err != nil {
		return fmt.Errorf("promoting upgrade workspace to current workspace: %w", err)
	}

	return nil
}
