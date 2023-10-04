/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

// TerraformUpgradeVars returns variables required to execute the Terraform scripts.
func TerraformUpgradeVars(conf *config.Config) (terraform.Variables, error) {
	// Note that we don't pass any real image as imageRef, as we ignore changes to the image in the terraform.
	// The image is updates via our operator.
	// Still, the terraform variable verification must accept the values.
	// For AWS, we enforce some basic constraints on the image variable.
	// For Azure, the provider enforces the format below.
	// For GCP, any placeholder works.
	var vars terraform.Variables
	switch conf.GetProvider() {
	case cloudprovider.AWS:
		vars = awsTerraformVars(conf, "ami-placeholder")
	case cloudprovider.Azure:
		vars = azureTerraformVars(conf, "/communityGalleries/myGalleryName/images/myImageName/versions/latest")
	case cloudprovider.GCP:
		vars = gcpTerraformVars(conf, "placeholder")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", conf.GetProvider())
	}
	return vars, nil
}

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
		AMIImageID:             imageRef,
		IAMProfileControlPlane: conf.Provider.AWS.IAMProfileControlPlane,
		IAMProfileWorkerNodes:  conf.Provider.AWS.IAMProfileWorkerNodes,
		Debug:                  conf.IsDebugCluster(),
		EnableSNP:              conf.GetAttestationConfig().GetVariant().Equal(variant.AWSSEVSNP{}),
		CustomEndpoint:         conf.CustomEndpoint,
		InternalLoadBalancer:   conf.InternalLoadBalancer,
	}
}

func awsTerraformIAMVars(conf *config.Config, oldVars terraform.AWSIAMVariables) *terraform.AWSIAMVariables {
	return &terraform.AWSIAMVariables{
		Region: conf.Provider.AWS.Region,
		Prefix: oldVars.Prefix,
	}
}

// azureTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the Azure variables.
func azureTerraformVars(conf *config.Config, imageRef string) *terraform.AzureClusterVariables {
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
		Name:                 conf.Name,
		NodeGroups:           nodeGroups,
		Location:             conf.Provider.Azure.Location,
		ImageID:              imageRef,
		CreateMAA:            toPtr(conf.GetAttestationConfig().GetVariant().Equal(variant.AzureSEVSNP{})),
		Debug:                toPtr(conf.IsDebugCluster()),
		ConfidentialVM:       toPtr(conf.GetAttestationConfig().GetVariant().Equal(variant.AzureSEVSNP{})),
		SecureBoot:           conf.Provider.Azure.SecureBoot,
		UserAssignedIdentity: conf.Provider.Azure.UserAssignedIdentity,
		ResourceGroup:        conf.Provider.Azure.ResourceGroup,
		CustomEndpoint:       conf.CustomEndpoint,
		InternalLoadBalancer: conf.InternalLoadBalancer,
	}

	vars = normalizeAzureURIs(vars)
	return vars
}

func azureTerraformIAMVars(conf *config.Config, oldVars terraform.AzureIAMVariables) *terraform.AzureIAMVariables {
	return &terraform.AzureIAMVariables{
		Region:           conf.Provider.Azure.Location,
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
func openStackTerraformVars(conf *config.Config, imageRef string) *terraform.OpenStackClusterVariables {
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
	return &terraform.OpenStackClusterVariables{
		Name:                    conf.Name,
		Cloud:                   toPtr(conf.Provider.OpenStack.Cloud),
		FloatingIPPoolID:        conf.Provider.OpenStack.FloatingIPPoolID,
		ImageURL:                imageRef,
		DirectDownload:          *conf.Provider.OpenStack.DirectDownload,
		OpenstackUserDomainName: conf.Provider.OpenStack.UserDomainName,
		OpenstackUsername:       conf.Provider.OpenStack.Username,
		OpenstackPassword:       conf.Provider.OpenStack.Password,
		Debug:                   conf.IsDebugCluster(),
		NodeGroups:              nodeGroups,
		CustomEndpoint:          conf.CustomEndpoint,
		InternalLoadBalancer:    conf.InternalLoadBalancer,
	}
}

// qemuTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the QEMU variables.
func qemuTerraformVars(conf *config.Config, imageRef string, libvirtURI, libvirtSocketPath, metadataLibvirtURI string) *terraform.QEMUVariables {
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
		ImagePath:          imageRef,
		ImageFormat:        conf.Provider.QEMU.ImageFormat,
		NodeGroups:         nodeGroups,
		Machine:            "q35", // TODO(elchead): make configurable AB#3225
		MetadataAPIImage:   conf.Provider.QEMU.MetadataAPIImage,
		MetadataLibvirtURI: metadataLibvirtURI,
		NVRAM:              conf.Provider.QEMU.NVRAM,
		// TODO(malt3) enable once we have a way to auto-select values for these
		// requires image info v2.
		// BzImagePath:        placeholder,
		// InitrdPath:         placeholder,
		// KernelCmdline:      placeholder,
	}
}
