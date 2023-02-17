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
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/clusterid"
	"github.com/edgelesssys/constellation/v2/cli/internal/image"
	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
)

// Creator creates cloud resources.
type Creator struct {
	out                io.Writer
	image              imageFetcher
	newTerraformClient func(ctx context.Context) (terraformClient, error)
	newLibvirtRunner   func() libvirtRunner
	newRawDownloader   func() rawDownloader
}

// NewCreator creates a new creator.
func NewCreator(out io.Writer) *Creator {
	return &Creator{
		out:   out,
		image: image.New(),
		newTerraformClient: func(ctx context.Context) (terraformClient, error) {
			return terraform.New(ctx, constants.TerraformWorkingDir)
		},
		newLibvirtRunner: func() libvirtRunner {
			return libvirt.New()
		},
		newRawDownloader: func() rawDownloader {
			return image.NewDownloader()
		},
	}
}

// Create creates the handed amount of instances and all the needed resources.
func (c *Creator) Create(ctx context.Context, provider cloudprovider.Provider, config *config.Config, insType string, controlPlaneCount, workerCount int,
) (clusterid.File, error) {
	image, err := c.image.FetchReference(ctx, config)
	if err != nil {
		return clusterid.File{}, fmt.Errorf("fetching image reference: %w", err)
	}

	switch provider {
	case cloudprovider.AWS:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAWS(ctx, cl, config, insType, controlPlaneCount, workerCount, image)
	case cloudprovider.GCP:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createGCP(ctx, cl, config, insType, controlPlaneCount, workerCount, image)
	case cloudprovider.Azure:
		cl, err := c.newTerraformClient(ctx)
		if err != nil {
			return clusterid.File{}, err
		}
		defer cl.RemoveInstaller()
		return c.createAzure(ctx, cl, config, insType, controlPlaneCount, workerCount, image)
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
		return c.createQEMU(ctx, cl, lv, config, controlPlaneCount, workerCount, image)
	default:
		return clusterid.File{}, fmt.Errorf("unsupported cloud provider: %s", provider)
	}
}

func (c *Creator) createAWS(ctx context.Context, cl terraformClient, config *config.Config,
	insType string, controlPlaneCount, workerCount int, image string,
) (idFile clusterid.File, retErr error) {
	vars := terraform.AWSClusterVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               config.Name,
			CountControlPlanes: controlPlaneCount,
			CountWorkers:       workerCount,
			StateDiskSizeGB:    config.StateDiskSizeGB,
		},
		StateDiskType:          config.Provider.AWS.StateDiskType,
		Region:                 config.Provider.AWS.Region,
		Zone:                   config.Provider.AWS.Zone,
		InstanceType:           insType,
		AMIImageID:             image,
		IAMProfileControlPlane: config.Provider.AWS.IAMProfileControlPlane,
		IAMProfileWorkerNodes:  config.Provider.AWS.IAMProfileWorkerNodes,
		Debug:                  config.IsDebugCluster(),
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.AWS.String())), &vars); err != nil {
		return clusterid.File{}, err
	}

	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})
	tfOutput, err := cl.CreateCluster(ctx)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.AWS,
		InitSecret:    []byte(tfOutput.Secret),
		IP:            tfOutput.IP,
		UID:           tfOutput.UID,
	}, nil
}

func (c *Creator) createGCP(ctx context.Context, cl terraformClient, config *config.Config,
	insType string, controlPlaneCount, workerCount int, image string,
) (idFile clusterid.File, retErr error) {
	vars := terraform.GCPClusterVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               config.Name,
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
		ImageID:         image,
		Debug:           config.IsDebugCluster(),
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.GCP.String())), &vars); err != nil {
		return clusterid.File{}, err
	}

	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})
	tfOutput, err := cl.CreateCluster(ctx)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.GCP,
		InitSecret:    []byte(tfOutput.Secret),
		IP:            tfOutput.IP,
		UID:           tfOutput.UID,
	}, nil
}

func (c *Creator) createAzure(ctx context.Context, cl terraformClient, config *config.Config,
	insType string, controlPlaneCount, workerCount int, image string,
) (idFile clusterid.File, retErr error) {
	vars := terraform.AzureClusterVariables{
		CommonVariables: terraform.CommonVariables{
			Name:               config.Name,
			CountControlPlanes: controlPlaneCount,
			CountWorkers:       workerCount,
			StateDiskSizeGB:    config.StateDiskSizeGB,
		},
		Location:             config.Provider.Azure.Location,
		ResourceGroup:        config.Provider.Azure.ResourceGroup,
		UserAssignedIdentity: config.Provider.Azure.UserAssignedIdentity,
		InstanceType:         insType,
		StateDiskType:        config.Provider.Azure.StateDiskType,
		ImageID:              image,
		ConfidentialVM:       *config.Provider.Azure.ConfidentialVM,
		SecureBoot:           *config.Provider.Azure.SecureBoot,
		Debug:                config.IsDebugCluster(),
	}

	vars = normalizeAzureURIs(vars)

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.Azure.String())), &vars); err != nil {
		return clusterid.File{}, err
	}

	defer rollbackOnError(context.Background(), c.out, &retErr, &rollbackerTerraform{client: cl})
	tfOutput, err := cl.CreateCluster(ctx)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.Azure,
		IP:            tfOutput.IP,
		InitSecret:    []byte(tfOutput.Secret),
		UID:           tfOutput.UID,
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

func normalizeAzureURIs(vars terraform.AzureClusterVariables) terraform.AzureClusterVariables {
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

func (c *Creator) createQEMU(ctx context.Context, cl terraformClient, lv libvirtRunner, config *config.Config,
	controlPlaneCount, workerCount int, source string,
) (idFile clusterid.File, retErr error) {
	qemuRollbacker := &rollbackerQEMU{client: cl, libvirt: lv, createdWorkspace: false}
	defer rollbackOnError(context.Background(), c.out, &retErr, qemuRollbacker)

	// TODO: render progress bar
	downloader := c.newRawDownloader()
	imagePath, err := downloader.Download(ctx, c.out, false, source, config.Image)
	if err != nil {
		return clusterid.File{}, fmt.Errorf("download raw image: %w", err)
	}

	libvirtURI := config.Provider.QEMU.LibvirtURI
	libvirtSocketPath := "."

	switch {
	// if no libvirt URI is specified, start a libvirt container
	case libvirtURI == "":
		if err := lv.Start(ctx, config.Name, config.Provider.QEMU.LibvirtContainerImage); err != nil {
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
			Name:               config.Name,
			CountControlPlanes: controlPlaneCount,
			CountWorkers:       workerCount,
			StateDiskSizeGB:    config.StateDiskSizeGB,
		},
		LibvirtURI:         libvirtURI,
		LibvirtSocketPath:  libvirtSocketPath,
		ImagePath:          imagePath,
		ImageFormat:        config.Provider.QEMU.ImageFormat,
		BzImagePath:        config.Provider.QEMU.KernelBzImage,
		InitrdPath:         config.Provider.QEMU.Initrd,
		KernelCmdline:      config.Provider.QEMU.KernelCmdline,
		CPUCount:           config.Provider.QEMU.VCPUs,
		MemorySizeMiB:      config.Provider.QEMU.Memory,
		MetadataAPIImage:   config.Provider.QEMU.MetadataAPIImage,
		MetadataLibvirtURI: metadataLibvirtURI,
		NVRAM:              config.Provider.QEMU.NVRAM,
		Firmware:           config.Provider.QEMU.Firmware,
	}

	if err := cl.PrepareWorkspace(path.Join("terraform", strings.ToLower(cloudprovider.QEMU.String())), &vars); err != nil {
		return clusterid.File{}, err
	}

	// Allow rollback of QEMU Terraform workspace from this point on
	qemuRollbacker.createdWorkspace = true

	tfOutput, err := cl.CreateCluster(ctx)
	if err != nil {
		return clusterid.File{}, err
	}

	return clusterid.File{
		CloudProvider: cloudprovider.QEMU,
		InitSecret:    []byte(tfOutput.Secret),
		IP:            tfOutput.IP,
		UID:           tfOutput.UID,
	}, nil
}
