/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"

	azurecl "github.com/edgelesssys/constellation/v2/cli/internal/azure/client"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/state"
)

type terraformClient interface {
	GetState() state.ConstellationState
	CreateCluster(ctx context.Context, name string, input terraform.Variables) error
	DestroyCluster(ctx context.Context) error
	CleanUpWorkspace() error
	RemoveInstaller()
}

type azureclient interface {
	GetState() state.ConstellationState
	SetState(state.ConstellationState)
	CreateApplicationInsight(ctx context.Context) error
	CreateExternalLoadBalancer(ctx context.Context, isDebugCluster bool) error
	CreateVirtualNetwork(ctx context.Context) error
	CreateSecurityGroup(ctx context.Context, input azurecl.NetworkSecurityGroupInput) error
	CreateInstances(ctx context.Context, input azurecl.CreateInstancesInput) error
	TerminateResourceGroupResources(ctx context.Context) error
}
