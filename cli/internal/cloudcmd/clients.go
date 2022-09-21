/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"

	azurecl "github.com/edgelesssys/constellation/v2/cli/internal/azure/client"
	gcpcl "github.com/edgelesssys/constellation/v2/cli/internal/gcp/client"
	"github.com/edgelesssys/constellation/v2/internal/state"
)

type gcpclient interface {
	GetState() state.ConstellationState
	SetState(state.ConstellationState)
	CreateVPCs(ctx context.Context) error
	CreateFirewall(ctx context.Context, input gcpcl.FirewallInput) error
	CreateInstances(ctx context.Context, input gcpcl.CreateInstancesInput) error
	CreateLoadBalancers(ctx context.Context, isDebugCluster bool) error
	TerminateFirewall(ctx context.Context) error
	TerminateVPCs(context.Context) error
	TerminateLoadBalancers(context.Context) error
	TerminateInstances(context.Context) error
	Close() error
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
