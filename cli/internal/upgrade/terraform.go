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

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
)

// NewTerraformUpgrader returns a new TerraformUpgrader.
func NewTerraformUpgrader(tfClient tfClient, outWriter io.Writer) (*TerraformUpgrader, error) {
	return &TerraformUpgrader{
		tf:        tfClient,
		outWriter: outWriter,
	}, nil
}

// TerraformUpgrader is responsible for performing Terraform migrations on cluster upgrades.
type TerraformUpgrader struct {
	tf        tfClient
	outWriter io.Writer
}

// TerraformUpgradeOptions are the options used for the Terraform upgrade.
type TerraformUpgradeOptions struct {
	// LogLevel is the log level used for Terraform.
	LogLevel terraform.LogLevel
	// CSP is the cloud provider to perform the upgrade on.
	CSP cloudprovider.Provider
	// Vars are the Terraform variables used for the upgrade.
	Vars terraform.Variables
	// Targets are the Terraform targets used for the upgrade.
	Targets []string
	// OutputFile is the file to write the Terraform output to.
	OutputFile string
}

// PlanTerraformMigrations prepares the upgrade workspace and plans the Terraform migrations for the Constellation upgrade.
// If a diff exists, it's being written to the upgrader's output writer. It also returns
// a bool indicating whether a diff exists.
func (u *TerraformUpgrader) PlanTerraformMigrations(ctx context.Context, opts TerraformUpgradeOptions) (bool, error) {
	err := u.tf.PrepareUpgradeWorkspace(
		filepath.Join("terraform", strings.ToLower(opts.CSP.String())),
		constants.TerraformWorkingDir,
		filepath.Join(constants.UpgradeDir, constants.TerraformUpgradeWorkingDir),
		opts.Vars,
	)
	if err != nil {
		return false, fmt.Errorf("preparing terraform workspace: %w", err)
	}

	hasDiff, err := u.tf.Plan(ctx, opts.LogLevel, constants.TerraformUpgradePlanFile, opts.Targets...)
	if err != nil {
		return false, fmt.Errorf("terraform plan: %w", err)
	}

	if hasDiff {
		if err := u.tf.ShowPlan(ctx, opts.LogLevel, constants.TerraformUpgradePlanFile, u.outWriter); err != nil {
			return false, fmt.Errorf("terraform show plan: %w", err)
		}
	}

	return hasDiff, nil
}

// ApplyTerraformMigrations applies the migerations planned by PlanTerraformMigrations.
// If PlanTerraformMigrations has not been executed before, it will return an error.
// In case of a successful upgrade, the output will be written to the specified file and the old Terraform directory is replaced
// By the new one.
func (u *TerraformUpgrader) ApplyTerraformMigrations(ctx context.Context, file file.Handler, opts TerraformUpgradeOptions) error {
	tfOutput, err := u.tf.CreateCluster(ctx, opts.LogLevel, opts.Targets...)
	if err != nil {
		return fmt.Errorf("terraform apply: %w", err)
	}

	outputFileContents := clusterid.File{
		CloudProvider:  opts.CSP,
		InitSecret:     []byte(tfOutput.Secret),
		IP:             tfOutput.IP,
		UID:            tfOutput.UID,
		AttestationURL: tfOutput.AttestationURL,
	}

	if err := file.RemoveAll(constants.TerraformWorkingDir); err != nil {
		return fmt.Errorf("removing old terraform directory: %w", err)
	}

	if err := file.CopyDir(filepath.Join(constants.UpgradeDir, constants.TerraformUpgradeWorkingDir), constants.TerraformWorkingDir); err != nil {
		return fmt.Errorf("replacing old terraform directory with new one: %w", err)
	}

	if err := file.WriteJSON(opts.OutputFile, outputFileContents); err != nil {
		return fmt.Errorf("writing terraform output to file: %w", err)
	}

	return nil
}

// a tfClient performs the Terraform interactions in an upgrade.
type tfClient interface {
	PrepareUpgradeWorkspace(path, oldWorkingDir, newWorkingDir string, vars terraform.Variables) error
	ShowPlan(ctx context.Context, logLevel terraform.LogLevel, planFilePath string, output io.Writer) error
	Plan(ctx context.Context, logLevel terraform.LogLevel, planFile string, targets ...string) (bool, error)
	CreateCluster(ctx context.Context, logLevel terraform.LogLevel, targets ...string) (terraform.CreateOutput, error)
}
