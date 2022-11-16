/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"

	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// Terminator deletes cloud provider resources.
type Terminator struct {
	newTerraformClient func(ctx context.Context) (terraformClient, error)
	newLibvirtRunner   func() libvirtRunner
}

// NewTerminator create a new cloud terminator.
func NewTerminator() *Terminator {
	return &Terminator{
		newTerraformClient: func(ctx context.Context) (terraformClient, error) {
			return terraform.New(ctx, constants.TerraformWorkingDir)
		},
		newLibvirtRunner: func() libvirtRunner {
			return libvirt.New()
		},
	}
}

// Terminate deletes the could provider resources.
func (t *Terminator) Terminate(ctx context.Context) (retErr error) {
	defer func() {
		if retErr == nil {
			retErr = t.newLibvirtRunner().Stop(ctx)
		}
	}()

	cl, err := t.newTerraformClient(ctx)
	if err != nil {
		return err
	}
	defer cl.RemoveInstaller()

	return t.terminateTerraform(ctx, cl)
}

func (t *Terminator) terminateTerraform(ctx context.Context, cl terraformClient) error {
	if err := cl.DestroyCluster(ctx); err != nil {
		return err
	}
	return cl.CleanUpWorkspace()
}
