/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"fmt"
	"io"

	azurecl "github.com/edgelesssys/constellation/v2/cli/internal/azure/client"
	"github.com/edgelesssys/constellation/v2/cli/internal/gcp"
	gcpcl "github.com/edgelesssys/constellation/v2/cli/internal/gcp/client"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/state"
)

// Creator creates cloud resources.
type Creator struct {
	out            io.Writer
	newGCPClient   func(ctx context.Context, project, zone, region, name string) (gcpclient, error)
	newAzureClient func(subscriptionID, tenantID, name, location, resourceGroup string) (azureclient, error)
}

// NewCreator creates a new creator.
func NewCreator(out io.Writer) *Creator {
	return &Creator{
		out: out,
		newGCPClient: func(ctx context.Context, project, zone, region, name string) (gcpclient, error) {
			return gcpcl.NewInitialized(ctx, project, zone, region, name)
		},
		newAzureClient: func(subscriptionID, tenantID, name, location, resourceGroup string) (azureclient, error) {
			return azurecl.NewInitialized(subscriptionID, tenantID, name, location, resourceGroup)
		},
	}
}

// Create creates the handed amount of instances and all the needed resources.
func (c *Creator) Create(ctx context.Context, provider cloudprovider.Provider, config *config.Config, name, insType string, controlPlaneCount, workerCount int,
) (state.ConstellationState, error) {
	// Use debug ingress firewall rules when debug mode / image is enabled
	var ingressRules cloudtypes.Firewall
	if config.IsDebugCluster() {
		ingressRules = constants.IngressRulesDebug
	} else {
		ingressRules = constants.IngressRulesNoDebug
	}

	switch provider {
	case cloudprovider.GCP:
		cl, err := c.newGCPClient(
			ctx,
			config.Provider.GCP.Project,
			config.Provider.GCP.Zone,
			config.Provider.GCP.Region,
			name,
		)
		if err != nil {
			return state.ConstellationState{}, err
		}
		defer cl.Close()
		return c.createGCP(ctx, cl, config, insType, controlPlaneCount, workerCount, ingressRules)
	case cloudprovider.Azure:
		cl, err := c.newAzureClient(
			config.Provider.Azure.SubscriptionID,
			config.Provider.Azure.TenantID,
			name,
			config.Provider.Azure.Location,
			config.Provider.Azure.ResourceGroup,
		)
		if err != nil {
			return state.ConstellationState{}, err
		}
		return c.createAzure(ctx, cl, config, insType, controlPlaneCount, workerCount, ingressRules)
	default:
		return state.ConstellationState{}, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

func (c *Creator) createGCP(ctx context.Context, cl gcpclient, config *config.Config, insType string, controlPlaneCount, workerCount int, ingressRules cloudtypes.Firewall,
) (stat state.ConstellationState, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerGCP{client: cl})

	if err := cl.CreateVPCs(ctx); err != nil {
		return state.ConstellationState{}, err
	}

	if err := cl.CreateFirewall(ctx, gcpcl.FirewallInput{
		Ingress: ingressRules,
		Egress:  constants.EgressRules,
	}); err != nil {
		return state.ConstellationState{}, err
	}

	// additionally create allow-internal rules
	internalFirewallInput := gcpcl.FirewallInput{
		Ingress: cloudtypes.Firewall{
			{
				Name:     "allow-cluster-internal-tcp",
				Protocol: "tcp",
				IPRange:  gcpcl.SubnetExtCIDR,
			},
			{
				Name:     "allow-cluster-internal-udp",
				Protocol: "udp",
				IPRange:  gcpcl.SubnetExtCIDR,
			},
			{
				Name:     "allow-cluster-internal-icmp",
				Protocol: "icmp",
				IPRange:  gcpcl.SubnetExtCIDR,
			},
			{
				Name:     "allow-node-internal-tcp",
				Protocol: "tcp",
				IPRange:  gcpcl.SubnetCIDR,
			},
			{
				Name:     "allow-node-internal-udp",
				Protocol: "udp",
				IPRange:  gcpcl.SubnetCIDR,
			},
			{
				Name:     "allow-node-internal-icmp",
				Protocol: "icmp",
				IPRange:  gcpcl.SubnetCIDR,
			},
		},
	}
	if err := cl.CreateFirewall(ctx, internalFirewallInput); err != nil {
		return state.ConstellationState{}, err
	}

	createInput := gcpcl.CreateInstancesInput{
		EnableSerialConsole: config.IsDebugCluster(),
		CountControlPlanes:  controlPlaneCount,
		CountWorkers:        workerCount,
		ImageID:             config.Provider.GCP.Image,
		InstanceType:        insType,
		StateDiskSizeGB:     config.StateDiskSizeGB,
		StateDiskType:       config.Provider.GCP.StateDiskType,
		KubeEnv:             gcp.KubeEnv,
	}
	if err := cl.CreateInstances(ctx, createInput); err != nil {
		return state.ConstellationState{}, err
	}

	if err := cl.CreateLoadBalancers(ctx, config.IsDebugCluster()); err != nil {
		return state.ConstellationState{}, err
	}

	return cl.GetState(), nil
}

func (c *Creator) createAzure(ctx context.Context, cl azureclient, config *config.Config, insType string, controlPlaneCount, workerCount int, ingressRules cloudtypes.Firewall,
) (stat state.ConstellationState, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerAzure{client: cl})

	if err := cl.CreateApplicationInsight(ctx); err != nil {
		return state.ConstellationState{}, err
	}
	if err := cl.CreateExternalLoadBalancer(ctx, config.IsDebugCluster()); err != nil {
		return state.ConstellationState{}, err
	}
	if err := cl.CreateVirtualNetwork(ctx); err != nil {
		return state.ConstellationState{}, err
	}

	if err := cl.CreateSecurityGroup(ctx, azurecl.NetworkSecurityGroupInput{
		Ingress: ingressRules,
		Egress:  constants.EgressRules,
	}); err != nil {
		return state.ConstellationState{}, err
	}
	createInput := azurecl.CreateInstancesInput{
		CountControlPlanes:   controlPlaneCount,
		CountWorkers:         workerCount,
		InstanceType:         insType,
		StateDiskSizeGB:      config.StateDiskSizeGB,
		StateDiskType:        config.Provider.Azure.StateDiskType,
		Image:                config.Provider.Azure.Image,
		UserAssingedIdentity: config.Provider.Azure.UserAssignedIdentity,
		ConfidentialVM:       *config.Provider.Azure.ConfidentialVM,
	}
	if err := cl.CreateInstances(ctx, createInput); err != nil {
		return state.ConstellationState{}, err
	}

	return cl.GetState(), nil
}
