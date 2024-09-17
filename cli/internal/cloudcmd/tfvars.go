/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/libvirt"
	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/mpimage"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

// The azurerm Terraform provider enforces its own convention of case sensitivity for Azure URIs which Azure's API itself does not enforce or, even worse, actually returns.
// These regular expression are used to make sure that the URIs we pass to Terraform are in the format that the provider expects.
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

// TerraformIAMUpgradeVars returns variables required to execute IAM upgrades with Terraform.
func TerraformIAMUpgradeVars(conf *config.Config, fileHandler file.Handler) (terraform.Variables, error) {
	// Load the tfvars of the existing IAM workspace.
	// Ideally we would only load values from the config file, but this currently does not hold all values required.
	// This should be refactored in the future.
	oldVarBytes, err := fileHandler.Read(filepath.Join(constants.TerraformIAMWorkingDir, "terraform.tfvars"))
	if err != nil {
		return nil, fmt.Errorf("reading existing IAM workspace: %w", err)
	}

	var vars terraform.Variables
	switch conf.GetProvider() {
	case cloudprovider.AWS:
		var oldVars terraform.AWSIAMVariables
		if err := terraform.VariablesFromBytes(oldVarBytes, &oldVars); err != nil {
			return nil, fmt.Errorf("parsing existing IAM workspace: %w", err)
		}
		vars = awsTerraformIAMVars(conf, oldVars)
	case cloudprovider.Azure:
		var oldVars terraform.AzureIAMVariables
		if err := terraform.VariablesFromBytes(oldVarBytes, &oldVars); err != nil {
			return nil, fmt.Errorf("parsing existing IAM workspace: %w", err)
		}
		vars = azureTerraformIAMVars(conf, oldVars)
	case cloudprovider.GCP:
		var oldVars terraform.GCPIAMVariables
		if err := terraform.VariablesFromBytes(oldVarBytes, &oldVars); err != nil {
			return nil, fmt.Errorf("parsing existing IAM workspace: %w", err)
		}
		vars = gcpTerraformIAMVars(conf, oldVars)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", conf.GetProvider())
	}
	return vars, nil
}

// awsTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the AWS variables.
func awsTerraformVars(conf *config.Config, imageRef string) *terraform.AWSClusterVariables {
	nodeGroups := make(map[string]terraform.AWSNodeGroup)
	for groupName, group := range conf.NodeGroups {
		nodeGroups[groupName] = terraform.AWSNodeGroup{
			Role:            role.FromString(group.Role).TFString(),
			StateDiskSizeGB: group.StateDiskSizeGB,
			InitialCount:    group.InitialCount,
			Zone:            group.Zone,
			InstanceType:    group.InstanceType,
			DiskType:        group.StateDiskType,
		}
	}

	return &terraform.AWSClusterVariables{
		Name:                   conf.Name,
		NodeGroups:             nodeGroups,
		Region:                 conf.Provider.AWS.Region,
		Zone:                   conf.Provider.AWS.Zone,
		ImageID:                imageRef,
		IAMProfileControlPlane: conf.Provider.AWS.IAMProfileControlPlane,
		IAMProfileWorkerNodes:  conf.Provider.AWS.IAMProfileWorkerNodes,
		Debug:                  conf.IsDebugCluster(),
		EnableSNP:              conf.GetAttestationConfig().GetVariant().Equal(variant.AWSSEVSNP{}),
		CustomEndpoint:         conf.CustomEndpoint,
		InternalLoadBalancer:   conf.InternalLoadBalancer,
		AdditionalTags:         conf.Tags,
	}
}

func awsTerraformIAMVars(conf *config.Config, oldVars terraform.AWSIAMVariables) *terraform.AWSIAMVariables {
	return &terraform.AWSIAMVariables{
		Region: conf.Provider.AWS.Region,
		Prefix: oldVars.Prefix,
	}
}

func normalizeAzureURIs(vars *terraform.AzureClusterVariables) *terraform.AzureClusterVariables {
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

// azureTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the Azure variables.
func azureTerraformVars(conf *config.Config, imageRef string) (*terraform.AzureClusterVariables, error) {
	nodeGroups := make(map[string]terraform.AzureNodeGroup)
	for groupName, group := range conf.NodeGroups {
		zones := strings.Split(group.Zone, ",")
		if len(zones) == 0 || (len(zones) == 1 && zones[0] == "") {
			zones = nil
		}
		nodeGroups[groupName] = terraform.AzureNodeGroup{
			Role:         role.FromString(group.Role).TFString(),
			InitialCount: group.InitialCount,
			InstanceType: group.InstanceType,
			DiskSizeGB:   group.StateDiskSizeGB,
			DiskType:     group.StateDiskType,
			Zones:        zones,
		}
	}
	vars := &terraform.AzureClusterVariables{
		SubscriptionID:       conf.Provider.Azure.SubscriptionID,
		Name:                 conf.Name,
		NodeGroups:           nodeGroups,
		Location:             conf.Provider.Azure.Location,
		CreateMAA:            toPtr(conf.GetAttestationConfig().GetVariant().Equal(variant.AzureSEVSNP{})),
		Debug:                toPtr(conf.IsDebugCluster()),
		ConfidentialVM:       toPtr(!conf.GetAttestationConfig().GetVariant().Equal(variant.AzureTrustedLaunch{})),
		SecureBoot:           conf.Provider.Azure.SecureBoot,
		UserAssignedIdentity: conf.Provider.Azure.UserAssignedIdentity,
		ResourceGroup:        conf.Provider.Azure.ResourceGroup,
		CustomEndpoint:       conf.CustomEndpoint,
		InternalLoadBalancer: conf.InternalLoadBalancer,
		MarketplaceImage:     nil,
		AdditionalTags:       conf.Tags,
	}

	if conf.UseMarketplaceImage() {
		image, err := mpimage.NewFromURI(imageRef)
		if err != nil {
			return nil, fmt.Errorf("parsing marketplace image URI: %w", err)
		}

		azureImage, ok := image.(mpimage.AzureMarketplaceImage)
		if !ok {
			return nil, fmt.Errorf("expected Azure marketplace image, got %T", image)
		}

		// If a marketplace image is used, only the marketplace reference is required.
		vars.MarketplaceImage = &terraform.AzureMarketplaceImageVariables{
			Publisher: azureImage.Publisher,
			Product:   azureImage.Offer,
			Name:      azureImage.SKU,
			Version:   azureImage.Version,
		}
	} else {
		// If not, we need to specify the exact CommunityGalleries/.. image reference.
		vars.ImageID = imageRef
	}

	vars = normalizeAzureURIs(vars)
	return vars, nil
}

func azureTerraformIAMVars(conf *config.Config, oldVars terraform.AzureIAMVariables) *terraform.AzureIAMVariables {
	return &terraform.AzureIAMVariables{
		SubscriptionID:   conf.Provider.Azure.SubscriptionID,
		Location:         conf.Provider.Azure.Location,
		ServicePrincipal: oldVars.ServicePrincipal,
		ResourceGroup:    conf.Provider.Azure.ResourceGroup,
	}
}

// gcpTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the GCP variables.
func gcpTerraformVars(conf *config.Config, imageRef string) *terraform.GCPClusterVariables {
	nodeGroups := make(map[string]terraform.GCPNodeGroup)
	for groupName, group := range conf.NodeGroups {
		nodeGroups[groupName] = terraform.GCPNodeGroup{
			Role:            role.FromString(group.Role).TFString(),
			StateDiskSizeGB: group.StateDiskSizeGB,
			InitialCount:    group.InitialCount,
			Zone:            group.Zone,
			InstanceType:    group.InstanceType,
			DiskType:        group.StateDiskType,
		}
	}

	ccTech := "SEV"
	if conf.GetAttestationConfig().GetVariant().Equal(variant.GCPSEVSNP{}) {
		ccTech = "SEV_SNP"
	}

	return &terraform.GCPClusterVariables{
		Name:                 conf.Name,
		NodeGroups:           nodeGroups,
		Project:              conf.Provider.GCP.Project,
		Region:               conf.Provider.GCP.Region,
		Zone:                 conf.Provider.GCP.Zone,
		ImageID:              imageRef,
		Debug:                conf.IsDebugCluster(),
		CustomEndpoint:       conf.CustomEndpoint,
		InternalLoadBalancer: conf.InternalLoadBalancer,
		CCTechnology:         ccTech,
		AdditionalLabels:     conf.Tags,
	}
}

func gcpTerraformIAMVars(conf *config.Config, oldVars terraform.GCPIAMVariables) *terraform.GCPIAMVariables {
	return &terraform.GCPIAMVariables{
		Project:          conf.Provider.GCP.Project,
		Region:           conf.Provider.GCP.Region,
		Zone:             conf.Provider.GCP.Zone,
		ServiceAccountID: oldVars.ServiceAccountID,
	}
}

// openStackTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the OpenStack variables.
func openStackTerraformVars(conf *config.Config, imageRef string) (*terraform.OpenStackClusterVariables, error) {
	if _, hasOSAuthURL := os.LookupEnv("OS_AUTH_URL"); !hasOSAuthURL && conf.Provider.OpenStack.Cloud == "" {
		return nil, errors.New(
			"neither environment variable OS_AUTH_URL nor cloud name for \"clouds.yaml\" is set. OpenStack authentication requires a set of " +
				"OS_* environment variables that are typically sourced into the current shell with an openrc file " +
				"or a cloud name for \"clouds.yaml\". " +
				"See https://docs.openstack.org/openstacksdk/latest/user/config/configuration.html for more information",
		)
	}

	nodeGroups := make(map[string]terraform.OpenStackNodeGroup)
	for groupName, group := range conf.NodeGroups {
		nodeGroups[groupName] = terraform.OpenStackNodeGroup{
			Role:            role.FromString(group.Role).TFString(),
			StateDiskSizeGB: group.StateDiskSizeGB,
			InitialCount:    group.InitialCount,
			FlavorID:        group.InstanceType,
			Zone:            group.Zone,
			StateDiskType:   group.StateDiskType,
		}
	}

	// since openstack does not support tags in the form of key = value, the tags will be converted
	// to an array of "key=value" strings
	tags := []string{}
	for key, value := range conf.Tags {
		tags = append(tags, fmt.Sprintf("%s=%s", key, value))
	}

	return &terraform.OpenStackClusterVariables{
		Name:                    conf.Name,
		Cloud:                   toPtr(conf.Provider.OpenStack.Cloud),
		OpenStackCloudsYAMLPath: conf.Provider.OpenStack.CloudsYAMLPath,
		FloatingIPPoolID:        conf.Provider.OpenStack.FloatingIPPoolID,
		ImageID:                 imageRef,
		Debug:                   conf.IsDebugCluster(),
		NodeGroups:              nodeGroups,
		CustomEndpoint:          conf.CustomEndpoint,
		InternalLoadBalancer:    conf.InternalLoadBalancer,
		STACKITProjectID:        conf.Provider.OpenStack.STACKITProjectID,
		AdditionalTags:          tags,
	}, nil
}

// qemuTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the QEMU variables.
func qemuTerraformVars(
	ctx context.Context, conf *config.Config, imageRef string,
	lv libvirtRunner, downloader rawDownloader,
) (*terraform.QEMUVariables, error) {
	if runtime.GOARCH != "amd64" || runtime.GOOS != "linux" {
		return nil, fmt.Errorf("creation of a QEMU based Constellation is not supported for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	imagePath, err := downloader.Download(ctx, nil, false, imageRef, conf.Image)
	if err != nil {
		return nil, fmt.Errorf("download raw image: %w", err)
	}

	libvirtURI := conf.Provider.QEMU.LibvirtURI
	libvirtSocketPath := "."

	switch {
	// if no libvirt URI is specified, start a libvirt container
	case libvirtURI == "":
		if err := lv.Start(ctx, conf.Name, conf.Provider.QEMU.LibvirtContainerImage); err != nil {
			return nil, fmt.Errorf("start libvirt container: %w", err)
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
			return nil, err
		}
		libvirtSocketPath = unixURI.Query().Get("socket")
		if libvirtSocketPath == "" {
			return nil, fmt.Errorf("socket path not specified in qemu+unix URI: %s", libvirtURI)
		}
	}

	metadataLibvirtURI := libvirtURI
	if libvirtSocketPath != "." {
		metadataLibvirtURI = "qemu:///system"
	}

	var firmware *string
	if conf.Provider.QEMU.Firmware != "" {
		firmware = &conf.Provider.QEMU.Firmware
	}

	nodeGroups := make(map[string]terraform.QEMUNodeGroup)
	for groupName, group := range conf.NodeGroups {
		nodeGroups[groupName] = terraform.QEMUNodeGroup{
			Role:         role.FromString(group.Role).TFString(),
			InitialCount: group.InitialCount,
			DiskSize:     group.StateDiskSizeGB,
			CPUCount:     conf.Provider.QEMU.VCPUs,
			MemorySize:   conf.Provider.QEMU.Memory,
		}
	}
	return &terraform.QEMUVariables{
		Name:              conf.Name,
		LibvirtURI:        libvirtURI,
		LibvirtSocketPath: libvirtSocketPath,
		// TODO(malt3): auto select boot mode based on attestation variant.
		// requires image info v2.
		BootMode:           "uefi",
		ImagePath:          imagePath,
		ImageFormat:        conf.Provider.QEMU.ImageFormat,
		NodeGroups:         nodeGroups,
		Machine:            "q35", // TODO(elchead): make configurable AB#3225
		MetadataAPIImage:   conf.Provider.QEMU.MetadataAPIImage,
		MetadataLibvirtURI: metadataLibvirtURI,
		NVRAM:              conf.Provider.QEMU.NVRAM,
		Firmware:           firmware,
		// TODO(malt3) enable once we have a way to auto-select values for these
		// requires image info v2.
		// BzImagePath:        placeholder,
		// InitrdPath:         placeholder,
		// KernelCmdline:      placeholder,
	}, nil
}

func toPtr[T any](v T) *T {
	return &v
}
