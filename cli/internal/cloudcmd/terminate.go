/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"fmt"

	azurecl "github.com/edgelesssys/constellation/v2/cli/internal/azure/client"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/state"
)

// Terminator deletes cloud provider resources.
type Terminator struct {
	newTerraformClient func(ctx context.Context) (terraformClient, error)
	newAzureClient     func(subscriptionID, tenantID string) (azureclient, error)
	newLibvirtRunner   func() libvirtRunner
}

// NewTerminator create a new cloud terminator.
func NewTerminator() *Terminator {
	return &Terminator{
		newTerraformClient: func(ctx context.Context) (terraformClient, error) {
			return terraform.New(ctx, cloudprovider.GCP)
		},
		newAzureClient: func(subscriptionID, tenantID string) (azureclient, error) {
			return azurecl.NewFromDefault(subscriptionID, tenantID)
		},
		newLibvirtRunner: func() libvirtRunner {
			return libvirt.New()
		},
	}
}

// Terminate deletes the could provider resources defined in the constellation state.
func (t *Terminator) Terminate(ctx context.Context, state state.ConstellationState) error {
	provider := cloudprovider.FromString(state.CloudProvider)
	switch provider {
	case cloudprovider.Azure:
		cl, err := t.newAzureClient(state.AzureSubscription, state.AzureTenant)
		if err != nil {
			return err
		}
		return t.terminateAzure(ctx, cl, state)
	case cloudprovider.AWS:
		fallthrough
	case cloudprovider.GCP:
		cl, err := t.newTerraformClient(ctx)
		if err != nil {
			return err
		}
		defer cl.RemoveInstaller()
		return t.terminateTerraform(ctx, cl)
	case cloudprovider.QEMU:
		cl, err := t.newTerraformClient(ctx)
		if err != nil {
			return err
		}
		defer cl.RemoveInstaller()
		libvirt := t.newLibvirtRunner()
		return t.terminateQEMU(ctx, cl, libvirt)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (t *Terminator) terminateAzure(ctx context.Context, cl azureclient, state state.ConstellationState) error {
	cl.SetState(state)

	return cl.TerminateResourceGroupResources(ctx)
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
