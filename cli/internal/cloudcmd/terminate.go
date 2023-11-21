/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"

	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
)

// Terminator deletes cloud provider resources.
type Terminator struct {
	newTerraformClient func(ctx context.Context, tfWorkspace string) (tfDestroyer, error)
	newLibvirtRunner   func() libvirtRunner
}

// NewTerminator create a new cloud terminator.
func NewTerminator() *Terminator {
	return &Terminator{
		newTerraformClient: func(ctx context.Context, tfWorkspace string) (tfDestroyer, error) {
			return terraform.New(ctx, tfWorkspace)
		},
		newLibvirtRunner: func() libvirtRunner {
			return libvirt.New()
		},
	}
}

// Terminate deletes the could provider resources.
func (t *Terminator) Terminate(ctx context.Context, tfWorkspace string, logLevel terraform.LogLevel) (retErr error) {
	defer func() {
		if retErr == nil {
			retErr = t.newLibvirtRunner().Stop(ctx)
		}
	}()

	cl, err := t.newTerraformClient(ctx, tfWorkspace)
	if err != nil {
		return err
	}
	defer cl.RemoveInstaller()

	return t.terminateTerraform(ctx, cl, logLevel)
}

func (t *Terminator) terminateTerraform(ctx context.Context, cl tfDestroyer, logLevel terraform.LogLevel) error {
	if err := cl.Destroy(ctx, logLevel); err != nil {
		return err
	}
	return cl.CleanUpWorkspace()
}
