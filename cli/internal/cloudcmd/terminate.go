/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/state"
)

// Terminator deletes cloud provider resources.
type Terminator struct {
	newTerraformClient func(ctx context.Context, provider cloudprovider.Provider) (terraformClient, error)
	newLibvirtRunner   func() libvirtRunner
}

// NewTerminator create a new cloud terminator.
func NewTerminator() *Terminator {
	return &Terminator{
		newTerraformClient: func(ctx context.Context, provider cloudprovider.Provider) (terraformClient, error) {
			return terraform.New(ctx, provider)
		},
		newLibvirtRunner: func() libvirtRunner {
			return libvirt.New()
		},
	}
}

// Terminate deletes the could provider resources defined in the constellation state.
func (t *Terminator) Terminate(ctx context.Context, state state.ConstellationState) error {
	provider := cloudprovider.FromString(state.CloudProvider)
	if provider == cloudprovider.Unknown {
		return fmt.Errorf("unknown cloud provider %s", state.CloudProvider)
	}

	cl, err := t.newTerraformClient(ctx, provider)
	if err != nil {
		return err
	}
	defer cl.RemoveInstaller()

	if provider == cloudprovider.QEMU {
		libvirt := t.newLibvirtRunner()
		return t.terminateQEMU(ctx, cl, libvirt)
	}

	return t.terminateTerraform(ctx, cl)
}

func (t *Terminator) terminateTerraform(ctx context.Context, cl terraformClient) error {
	if err := cl.DestroyCluster(ctx); err != nil {
		return err
	}
	return cl.CleanUpWorkspace()
}

func (t *Terminator) terminateQEMU(ctx context.Context, cl terraformClient, lv libvirtRunner) error {
	if err := cl.DestroyCluster(ctx); err != nil {
		return err
	}
	if err := lv.Stop(ctx); err != nil {
		return err
	}
	return cl.CleanUpWorkspace()
}
