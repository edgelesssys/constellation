/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"fmt"
	"io"
	"runtime"

	azurecl "github.com/edgelesssys/constellation/v2/cli/internal/azure/client"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudtypes"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/state"
)

// Creator creates cloud resources.
type Creator struct {
	out                io.Writer
	newTerraformClient func(ctx context.Context, provider cloudprovider.Provider) (terraformClient, error)
	newAzureClient     func(subscriptionID, tenantID, name, location, resourceGroup string) (azureclient, error)
}

// NewCreator creates a new creator.
func NewCreator(out io.Writer) *Creator {
	return &Creator{
		out: out,
		newTerraformClient: func(ctx context.Context, provider cloudprovider.Provider) (terraformClient, error) {
			return terraform.New(ctx, provider)
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
		cl, err := c.newTerraformClient(ctx, provider)
		if err != nil {
			return state.ConstellationState{}, err
		}
		defer cl.RemoveInstaller()
		return c.createGCP(ctx, cl, config, name, insType, controlPlaneCount, workerCount)
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
	case cloudprovider.QEMU:
		if runtime.GOARCH != "amd64" || runtime.GOOS != "linux" {
			return state.ConstellationState{}, fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		cl, err := c.newTerraformClient(ctx, provider)
		if err != nil {
			return state.ConstellationState{}, err
		}
		defer cl.RemoveInstaller()
		return c.createQEMU(ctx, cl, name, config, controlPlaneCount, workerCount)
	default:
		return state.ConstellationState{}, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

func (c *Creator) createGCP(ctx context.Context, cl terraformClient, config *config.Config,
	name, insType string, controlPlaneCount, workerCount int,
) (stat state.ConstellationState, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})

	vars := &terraform.GCPVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               name,
			CountControlPlanes: controlPlaneCount,
			CountWorkers:       workerCount,
			StateDiskSizeGB:    config.StateDiskSizeGB,
		},
		Project:         config.Provider.GCP.Project,
		Region:          config.Provider.GCP.Region,
		Zone:            config.Provider.GCP.Zone,
		CredentialsFile: config.Provider.GCP.ServiceAccountKeyPath,
		InstanceType:    insType,
		StateDiskType:   config.Provider.GCP.StateDiskType,
		ImageID:         config.Provider.GCP.Image,
		Debug:           config.IsDebugCluster(),
	}

	if err := cl.CreateCluster(ctx, name, vars); err != nil {
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

func (c *Creator) createQEMU(ctx context.Context, cl terraformClient, name string, config *config.Config,
	controlPlaneCount, workerCount int,
) (stat state.ConstellationState, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})

	vars := &terraform.QEMUVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               name,
			CountControlPlanes: controlPlaneCount,
			CountWorkers:       workerCount,
			StateDiskSizeGB:    config.StateDiskSizeGB,
		},
		ImagePath:        config.Provider.QEMU.Image,
		ImageFormat:      config.Provider.QEMU.ImageFormat,
		CPUCount:         config.Provider.QEMU.VCPUs,
		MemorySizeMiB:    config.Provider.QEMU.Memory,
		IPRangeStart:     config.Provider.QEMU.IPRangeStart,
		MetadataAPIImage: config.Provider.QEMU.MetadataAPIImage,
	}

	if err := cl.CreateCluster(ctx, name, vars); err != nil {
		return state.ConstellationState{}, err
	}

	return cl.GetState(), nil
}
