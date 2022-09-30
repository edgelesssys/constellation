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

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
)

// Creator creates cloud resources.
type Creator struct {
	out                io.Writer
	newTerraformClient func(ctx context.Context, provider cloudprovider.Provider) (terraformClient, error)
	newLibvirtRunner   func() libvirtRunner
}

// NewCreator creates a new creator.
func NewCreator(out io.Writer) *Creator {
	return &Creator{
		out: out,
		newTerraformClient: func(ctx context.Context, provider cloudprovider.Provider) (terraformClient, error) {
			return terraform.New(ctx, provider)
		},
		newLibvirtRunner: func() libvirtRunner {
			return libvirt.New()
		},
	}
}

// Create creates the handed amount of instances and all the needed resources.
func (c *Creator) Create(ctx context.Context, provider cloudprovider.Provider, config *config.Config, name, insType string, controlPlaneCount, workerCount int,
) (clusterid.File, error) {
	switch provider {
	case cloudprovider.AWS:
		// TODO: Remove this once AWS is supported.
		if os.Getenv("CONSTELLATION_AWS_DEV") != "1" {
			return state.ConstellationState{}, fmt.Errorf("AWS isn't supported yet")
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
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createGCP(ctx, cl, config, name, insType, controlPlaneCount, workerCount)
	case cloudprovider.Azure:
		cl, err := c.newTerraformClient(ctx, provider)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAzure(ctx, cl, config, name, insType, controlPlaneCount, workerCount)
	case cloudprovider.QEMU:
		if runtime.GOARCH != "amd64" || runtime.GOOS != "linux" {
			return clusterid.File{}, fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		cl, err := c.newTerraformClient(ctx, provider)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		lv := c.newLibvirtRunner()
		return c.createQEMU(ctx, cl, lv, name, config, controlPlaneCount, workerCount)
	default:
		return clusterid.File{}, fmt.Errorf("unsupported cloud provider: %s", provider)
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
) (idFile clusterid.File, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})

	vars := terraform.GCPVariables{
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

	ip, err := cl.CreateCluster(ctx, name, &vars)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.GCP,
		IP:            ip,
	}, nil
}

func (c *Creator) createAzure(ctx context.Context, cl terraformClient, config *config.Config,
	name, insType string, controlPlaneCount, workerCount int,
) (idFile clusterid.File, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})

	vars := terraform.AzureVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               name,
			CountControlPlanes: controlPlaneCount,
			CountWorkers:       workerCount,
			StateDiskSizeGB:    config.StateDiskSizeGB,
		},
		Location:             config.Provider.Azure.Location,
		ResourceGroup:        config.Provider.Azure.ResourceGroup,
		UserAssignedIdentity: config.Provider.Azure.UserAssignedIdentity,
		InstanceType:         insType,
		StateDiskType:        config.Provider.Azure.StateDiskType,
		ImageID:              config.Provider.Azure.Image,
		ConfidentialVM:       *config.Provider.Azure.ConfidentialVM,
		Debug:                config.IsDebugCluster(),
	}

	vars = normalizeAzureURIs(vars)

	ip, err := cl.CreateCluster(ctx, name, &vars)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.Azure,
		IP:            ip,
	}, nil
}

func normalizeAzureURIs(vars terraform.AzureVariables) terraform.AzureVariables {
	vars.UserAssignedIdentity = strings.ReplaceAll(vars.UserAssignedIdentity, "resourcegroup", "resourceGroup")

	vars.ImageID = strings.ReplaceAll(vars.ImageID, "CommunityGalleries", "communityGalleries")
	vars.ImageID = strings.ReplaceAll(vars.ImageID, "Images", "images")
	vars.ImageID = strings.ReplaceAll(vars.ImageID, "Versions", "versions")

	return vars
}

func (c *Creator) createQEMU(ctx context.Context, cl terraformClient, lv libvirtRunner, name string, config *config.Config,
	controlPlaneCount, workerCount int,
) (idFile clusterid.File, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerQEMU{client: cl, libvirt: lv})

	libvirtURI := config.Provider.QEMU.LibvirtURI
	libvirtSocketPath := "."

	switch {
	// if no libvirt URI is specified, start a libvirt container
	case libvirtURI == "":
		if err := lv.Start(ctx, name, config.Provider.QEMU.LibvirtContainerImage); err != nil {
			return clusterid.File{}, err
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
			return clusterid.File{}, err
		}
		libvirtSocketPath = unixURI.Query().Get("socket")
		if libvirtSocketPath == "" {
			return clusterid.File{}, fmt.Errorf("socket path not specified in qemu+unix URI: %s", libvirtURI)
		}
	}

	metadataLibvirtURI := libvirtURI
	if libvirtSocketPath != "." {
		metadataLibvirtURI = "qemu:///system"
	}

	vars := terraform.QEMUVariables{
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

	ip, err := cl.CreateCluster(ctx, name, &vars)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.QEMU,
		IP:            ip,
	}, nil
}
