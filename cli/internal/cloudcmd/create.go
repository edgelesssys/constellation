/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime"
	"strings"

	azurecl "github.com/edgelesssys/constellation/v2/cli/internal/azure/client"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
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
	newLibvirtRunner   func() libvirtRunner
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
		newLibvirtRunner: func() libvirtRunner {
			return libvirt.New()
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
	case cloudprovider.AWS:
		// TODO: Remove this once AWS is supported.
		if os.Getenv("CONSTELLATION_AWS_DEV") != "1" {
			return state.ConstellationState{}, fmt.Errorf("AWS is not supported yet")
		}
		cl, err := c.newTerraformClient(ctx, provider)
		if err != nil {
			return state.ConstellationState{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAWS(ctx, cl, config, name, insType, controlPlaneCount, workerCount)
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
		lv := c.newLibvirtRunner()
		return c.createQEMU(ctx, cl, lv, name, config, controlPlaneCount, workerCount)
	default:
		return state.ConstellationState{}, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

func (c *Creator) createAWS(ctx context.Context, cl terraformClient, config *config.Config,
	name, insType string, controlPlaneCount, workerCount int,
) (stat state.ConstellationState, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})

	vars := &terraform.AWSVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               name,
			CountControlPlanes: controlPlaneCount,
			CountWorkers:       workerCount,
			StateDiskSizeGB:    config.StateDiskSizeGB,
		},
		Region:                 config.Provider.AWS.Region,
		Zone:                   config.Provider.AWS.Zone,
		InstanceType:           insType,
		AMIImageID:             config.Provider.AWS.Image,
		IAMProfileControlPlane: config.Provider.AWS.IAMProfileControlPlane,
		IAMProfileWorkerNodes:  config.Provider.AWS.IAMProfileWorkerNodes,
	}

	if err := cl.CreateCluster(ctx, name, vars); err != nil {
		return state.ConstellationState{}, err
	}

	return cl.GetState(), nil
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

func (c *Creator) createQEMU(ctx context.Context, cl terraformClient, lv libvirtRunner, name string, config *config.Config,
	controlPlaneCount, workerCount int,
) (stat state.ConstellationState, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerQEMU{client: cl, libvirt: lv})

	libvirtURI := config.Provider.QEMU.LibvirtURI
	libvirtSocketPath := "."

	switch {
	// if no libvirt URI is specified, start a libvirt container
	case libvirtURI == "":
		if err := lv.Start(ctx, name, config.Provider.QEMU.LibvirtContainerImage); err != nil {
			return state.ConstellationState{}, err
		}
		libvirtURI = libvirt.LibvirtTCPConnectURI

	// socket for system URI should be in /var/run/libvirt/libvirt-sock
	case libvirtURI == "qemu:///system":
		libvirtSocketPath = "/var/run/libvirt/libvirt-sock"

	// socket for session URI should be in /run/user/<uid>/libvirt/libvirt-sock
	case libvirtURI == "qemu:///session":
		libvirtSocketPath = fmt.Sprintf("/run/user/%d/libvirt/libvirt-sock", os.Getuid())

	// if a unix socket is specified we need to parse the URI to get the socket path
	case strings.HasPrefix(libvirtURI, "qemu+unix://"):
		unixURI, err := url.Parse(strings.TrimPrefix(libvirtURI, "qemu+unix://"))
		if err != nil {
			return state.ConstellationState{}, err
		}
		libvirtSocketPath = unixURI.Query().Get("socket")
		if libvirtSocketPath == "" {
			return state.ConstellationState{}, fmt.Errorf("socket path not specified in qemu+unix URI: %s", libvirtURI)
		}
	}

	metadataLibvirtURI := libvirtURI
	if libvirtSocketPath != "." {
		metadataLibvirtURI = "qemu:///system"
	}

	vars := &terraform.QEMUVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               name,
			CountControlPlanes: controlPlaneCount,
			CountWorkers:       workerCount,
			StateDiskSizeGB:    config.StateDiskSizeGB,
		},
		LibvirtURI:         libvirtURI,
		LibvirtSocketPath:  libvirtSocketPath,
		ImagePath:          config.Provider.QEMU.Image,
		ImageFormat:        config.Provider.QEMU.ImageFormat,
		CPUCount:           config.Provider.QEMU.VCPUs,
		MemorySizeMiB:      config.Provider.QEMU.Memory,
		MetadataAPIImage:   config.Provider.QEMU.MetadataAPIImage,
		MetadataLibvirtURI: metadataLibvirtURI,
	}

	if err := cl.CreateCluster(ctx, name, vars); err != nil {
		return state.ConstellationState{}, err
	}

	return cl.GetState(), nil
}
