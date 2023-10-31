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

	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/edgelesssys/constellation/v2/internal/maa"
)

const (
	// WithRollbackOnError indicates a rollback should be performed on error.
	WithRollbackOnError RollbackBehavior = true
	// WithoutRollbackOnError indicates a rollback should not be performed on error.
	WithoutRollbackOnError RollbackBehavior = false
)

// RollbackBehavior is a boolean flag that indicates whether a rollback should be performed.
type RollbackBehavior bool

// Applier creates or updates cloud resources.
type Applier struct {
	fileHandler     file.Handler
	imageFetcher    imageFetcher
	libvirtRunner   libvirtRunner
	rawDownloader   rawDownloader
	policyPatcher   policyPatcher
	terraformClient tfResourceClient
	logLevel        terraform.LogLevel

	workingDir string
	backupDir  string
	out        io.Writer
}

// NewApplier creates a new Applier.
func NewApplier(
	ctx context.Context, out io.Writer, workingDir, backupDir string,
	logLevel terraform.LogLevel, fileHandler file.Handler,
) (*Applier, func(), error) {
	tfClient, err := terraform.New(ctx, workingDir)
	if err != nil {
		return nil, nil, fmt.Errorf("setting up terraform client: %w", err)
	}

	return &Applier{
		fileHandler:     fileHandler,
		imageFetcher:    imagefetcher.New(),
		libvirtRunner:   libvirt.New(),
		rawDownloader:   imagefetcher.NewDownloader(),
		policyPatcher:   maa.NewAzurePolicyPatcher(),
		terraformClient: tfClient,
		logLevel:        logLevel,
		workingDir:      workingDir,
		backupDir:       backupDir,
		out:             out,
	}, tfClient.RemoveInstaller, nil
}

// Plan plans the given configuration and prepares the Terraform workspace.
func (a *Applier) Plan(ctx context.Context, conf *config.Config) (bool, error) {
	vars, err := a.terraformApplyVars(ctx, conf)
	if err != nil {
		return false, fmt.Errorf("creating terraform variables: %w", err)
	}

	return plan(
		ctx, a.terraformClient, a.fileHandler, a.out, a.logLevel, vars,
		filepath.Join(constants.TerraformEmbeddedDir, strings.ToLower(conf.GetProvider().String())),
		a.workingDir,
		filepath.Join(a.backupDir, constants.TerraformUpgradeBackupDir),
	)
}

// Apply applies the prepared configuration by creating or updating cloud resources.
func (a *Applier) Apply(ctx context.Context, csp cloudprovider.Provider, withRollback RollbackBehavior) (infra state.Infrastructure, retErr error) {
	if withRollback {
		var rollbacker rollbacker
		switch csp {
		case cloudprovider.QEMU:
			rollbacker = &rollbackerQEMU{client: a.terraformClient, libvirt: a.libvirtRunner}
		default:
			rollbacker = &rollbackerTerraform{client: a.terraformClient}
		}
		defer rollbackOnError(a.out, &retErr, rollbacker, a.logLevel)
	}

	infraState, err := a.terraformClient.ApplyCluster(ctx, csp, a.logLevel)
	if err != nil {
		return infraState, fmt.Errorf("terraform apply: %w", err)
	}
	if csp == cloudprovider.Azure && infraState.Azure != nil {
		if err := a.policyPatcher.Patch(ctx, infraState.Azure.AttestationURL); err != nil {
			return infraState, fmt.Errorf("patching policies: %w", err)
		}
	}

	return infraState, nil
}

// RestoreWorkspace rolls back the existing workspace to the backup directory created when planning an action,
// and the user decides to not apply it.
// Note that this will not apply the restored state from the backup.
func (a *Applier) RestoreWorkspace() error {
	return restoreBackup(a.fileHandler, a.workingDir, filepath.Join(a.backupDir, constants.TerraformUpgradeBackupDir))
}

func (a *Applier) terraformApplyVars(ctx context.Context, conf *config.Config) (terraform.Variables, error) {
	imageRef, err := a.imageFetcher.FetchReference(
		ctx,
		conf.GetProvider(),
		conf.GetAttestationConfig().GetVariant(),
		conf.Image, conf.GetRegion(),
	)
	if err != nil {
		return nil, fmt.Errorf("fetching image reference: %w", err)
	}

	switch conf.GetProvider() {
	case cloudprovider.AWS:
		return awsTerraformVars(conf, imageRef), nil
	case cloudprovider.Azure:
		return azureTerraformVars(conf, imageRef), nil
	case cloudprovider.GCP:
		return gcpTerraformVars(conf, imageRef), nil
	case cloudprovider.OpenStack:
		return openStackTerraformVars(conf, imageRef)
	case cloudprovider.QEMU:
		return qemuTerraformVars(ctx, conf, imageRef, a.libvirtRunner, a.rawDownloader)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", conf.GetProvider())
	}
}

// policyPatcher interacts with the CSP (currently only applies for Azure) to update the attestation policy.
type policyPatcher interface {
	Patch(ctx context.Context, attestationURL string) error
}
