/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
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
func (t *Terminator) Terminate(ctx context.Context, provider cloudprovider.Provider) (retErr error) {
	if provider == cloudprovider.Unknown {
		return errors.New("unknown cloud provider")
	}

	if provider == cloudprovider.QEMU {
		if runtime.GOARCH != "amd64" || runtime.GOOS != "linux" {
			return fmt.Errorf("termination of a QEMU based Constellation is not supported for %s/%s", runtime.GOOS, runtime.GOARCH)
		}

		libvirt := t.newLibvirtRunner()
		defer func() {
			if retErr == nil {
				retErr = libvirt.Stop(ctx)
			}
		}()
	}

	cl, err := t.newTerraformClient(ctx, provider)
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
