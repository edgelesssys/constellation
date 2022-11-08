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
	"regexp"
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
	newTerraformClient func(ctx context.Context) (terraformClient, error)
	newLibvirtRunner   func() libvirtRunner
}

// NewCreator creates a new creator.
func NewCreator(out io.Writer) *Creator {
	return &Creator{
		out: out,
		newTerraformClient: func(ctx context.Context) (terraformClient, error) {
			return terraform.New(ctx)
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
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAWS(ctx, cl, config, name, insType, controlPlaneCount, workerCount)
	case cloudprovider.GCP:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createGCP(ctx, cl, config, name, insType, controlPlaneCount, workerCount)
	case cloudprovider.Azure:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAzure(ctx, cl, config, name, insType, controlPlaneCount, workerCount)
	case cloudprovider.QEMU:
		if runtime.GOARCH != "amd64" || runtime.GOOS != "linux" {
			return clusterid.File{}, fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		cl, err := c.newTerraformClient(ctx)
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
) (idFile clusterid.File, retErr error) {
	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})

	vars := &terraform.AWSVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               name,
			CountControlPlanes: controlPlaneCount,
			CountWorkers:       workerCount,
			StateDiskSizeGB:    config.StateDiskSizeGB,
		},
		StateDiskType:          config.Provider.AWS.StateDiskType,
		Region:                 config.Provider.AWS.Region,
		Zone:                   config.Provider.AWS.Zone,
		InstanceType:           insType,
		AMIImageID:             config.Provider.AWS.Image,
		IAMProfileControlPlane: config.Provider.AWS.IAMProfileControlPlane,
		IAMProfileWorkerNodes:  config.Provider.AWS.IAMProfileWorkerNodes,
		Debug:                  config.IsDebugCluster(),
	}

	ip, err := cl.CreateCluster(ctx, cloudprovider.AWS, name, vars)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.AWS,
		IP:            ip,
	}, nil
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

	ip, err := cl.CreateCluster(ctx, cloudprovider.GCP, name, &vars)
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
		SecureBoot:           *config.Provider.Azure.SecureBoot,
		Debug:                config.IsDebugCluster(),
	}

	vars = normalizeAzureURIs(vars)

	ip, err := cl.CreateCluster(ctx, cloudprovider.Azure, name, &vars)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.Azure,
		IP:            ip,
	}, nil
}

// The azurerm Terraform provider enforces its own convention of case sensitivity for Azure URIs which Azure's API itself does not enforce or, even worse, actually returns.
// Let's go loco with case insensitive Regexp here and fix the user input here to be compliant with this arbitrary design decision.
var (
	caseInsensitiveSubscriptionsRegexp          = regexp.MustCompile(`(?i)\/subscriptions\/`)
	caseInsensitiveResourceGroupRegexp          = regexp.MustCompile(`(?i)\/resourcegroups\/`)
	caseInsensitiveProvidersRegexp              = regexp.MustCompile(`(?i)\/providers\/`)
	caseInsensitiveUserAssignedIdentitiesRegexp = regexp.MustCompile(`(?i)\/userassignedidentities\/`)
	caseInsensitiveMicrosoftManagedIdentity     = regexp.MustCompile(`(?i)\/microsoft.managedidentity\/`)
	caseInsensitiveCommunityGalleriesRegexp     = regexp.MustCompile(`(?i)\/communitygalleries\/`)
	caseInsensitiveImagesRegExp                 = regexp.MustCompile(`(?i)\/images\/`)
	caseInsensitiveVersionsRegExp               = regexp.MustCompile(`(?i)\/versions\/`)
)

func normalizeAzureURIs(vars terraform.AzureVariables) terraform.AzureVariables {
	vars.UserAssignedIdentity = caseInsensitiveSubscriptionsRegexp.ReplaceAllString(vars.UserAssignedIdentity, "/subscriptions/")
	vars.UserAssignedIdentity = caseInsensitiveResourceGroupRegexp.ReplaceAllString(vars.UserAssignedIdentity, "/resourceGroups/")
	vars.UserAssignedIdentity = caseInsensitiveProvidersRegexp.ReplaceAllString(vars.UserAssignedIdentity, "/providers/")
	vars.UserAssignedIdentity = caseInsensitiveUserAssignedIdentitiesRegexp.ReplaceAllString(vars.UserAssignedIdentity, "/userAssignedIdentities/")
	vars.UserAssignedIdentity = caseInsensitiveMicrosoftManagedIdentity.ReplaceAllString(vars.UserAssignedIdentity, "/Microsoft.ManagedIdentity/")
	vars.ImageID = caseInsensitiveCommunityGalleriesRegexp.ReplaceAllString(vars.ImageID, "/communityGalleries/")
	vars.ImageID = caseInsensitiveImagesRegExp.ReplaceAllString(vars.ImageID, "/images/")
	vars.ImageID = caseInsensitiveVersionsRegExp.ReplaceAllString(vars.ImageID, "/versions/")

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
		NVRAM:              config.Provider.QEMU.NVRAM,
		Firmware:           config.Provider.QEMU.Firmware,
	}

	ip, err := cl.CreateCluster(ctx, cloudprovider.QEMU, name, &vars)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.QEMU,
		IP:            ip,
	}, nil
}
