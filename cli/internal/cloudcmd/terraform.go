/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudcmd

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/role"
)

// TerraformUpgradeVars returns variables required to execute the Terraform scripts.
func TerraformUpgradeVars(conf *config.Config, imageRef string) (terraform.Variables, error) {
	switch conf.GetProvider() {
	case cloudprovider.AWS:
		vars := awsTerraformVars(conf, imageRef, nil, nil)
		return vars, nil
	case cloudprovider.Azure:
		vars := azureTerraformVars(conf, imageRef, nil, nil)
		return vars, nil
	case cloudprovider.GCP:
		vars := gcpTerraformVars(conf, imageRef, nil, nil)
		return vars, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", conf.GetProvider())
	}
}

// awsTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the AWS variables.
func awsTerraformVars(conf *config.Config, imageRef string, controlPlaneCount, workerCount *int) *terraform.AWSClusterVariables {
	return &terraform.AWSClusterVariables{
		Name: conf.Name,
		NodeGroups: map[string]terraform.AWSNodeGroup{
			constants.ControlPlaneDefault: {
				Role:            role.ControlPlane.TFString(),
				StateDiskSizeGB: conf.StateDiskSizeGB,
				InitialCount:    controlPlaneCount,
				Zone:            conf.Provider.AWS.Zone,
				InstanceType:    conf.Provider.AWS.InstanceType,
				DiskType:        conf.Provider.AWS.StateDiskType,
			},
			constants.WorkerDefault: {
				Role:            role.Worker.TFString(),
				StateDiskSizeGB: conf.StateDiskSizeGB,
				InitialCount:    workerCount,
				Zone:            conf.Provider.AWS.Zone,
				InstanceType:    conf.Provider.AWS.InstanceType,
				DiskType:        conf.Provider.AWS.StateDiskType,
			},
		},
		Region:                 conf.Provider.AWS.Region,
		Zone:                   conf.Provider.AWS.Zone,
		AMIImageID:             imageRef,
		IAMProfileControlPlane: conf.Provider.AWS.IAMProfileControlPlane,
		IAMProfileWorkerNodes:  conf.Provider.AWS.IAMProfileWorkerNodes,
		Debug:                  conf.IsDebugCluster(),
		EnableSNP:              conf.GetAttestationConfig().GetVariant().Equal(variant.AWSSEVSNP{}),
	}
}

// azureTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the Azure variables.
func azureTerraformVars(conf *config.Config, imageRef string, controlPlaneCount, workerCount *int) *terraform.AzureClusterVariables {
	vars := &terraform.AzureClusterVariables{
		Name: conf.Name,
		NodeGroups: map[string]terraform.AzureNodeGroup{
			constants.ControlPlaneDefault: {
				Role:         "control-plane",
				InitialCount: controlPlaneCount,
				InstanceType: conf.Provider.Azure.InstanceType,
				DiskSizeGB:   conf.StateDiskSizeGB,
				DiskType:     conf.Provider.Azure.StateDiskType,
				Zones:        nil, // TODO(elchead): support zones AB#3225. check if lifecycle arg is required.
			},
			constants.WorkerDefault: {
				Role:         "worker",
				InitialCount: workerCount,
				InstanceType: conf.Provider.Azure.InstanceType,
				DiskSizeGB:   conf.StateDiskSizeGB,
				DiskType:     conf.Provider.Azure.StateDiskType,
				Zones:        nil,
			},
		},
		Location:             conf.Provider.Azure.Location,
		ImageID:              imageRef,
		CreateMAA:            toPtr(conf.GetAttestationConfig().GetVariant().Equal(variant.AzureSEVSNP{})),
		Debug:                toPtr(conf.IsDebugCluster()),
		ConfidentialVM:       toPtr(conf.GetAttestationConfig().GetVariant().Equal(variant.AzureSEVSNP{})),
		SecureBoot:           conf.Provider.Azure.SecureBoot,
		UserAssignedIdentity: conf.Provider.Azure.UserAssignedIdentity,
		ResourceGroup:        conf.Provider.Azure.ResourceGroup,
	}

	vars = normalizeAzureURIs(vars)
	return vars
}

// gcpTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the GCP variables.
func gcpTerraformVars(conf *config.Config, imageRef string, controlPlaneCount, workerCount *int) *terraform.GCPClusterVariables {
	return &terraform.GCPClusterVariables{
		Name: conf.Name,
		NodeGroups: map[string]terraform.GCPNodeGroup{
			constants.ControlPlaneDefault: {
				Role:            role.ControlPlane.TFString(),
				StateDiskSizeGB: conf.StateDiskSizeGB,
				InitialCount:    controlPlaneCount,
				Zone:            conf.Provider.GCP.Zone,
				InstanceType:    conf.Provider.GCP.InstanceType,
				DiskType:        conf.Provider.GCP.StateDiskType,
			},
			constants.WorkerDefault: {
				Role:            role.Worker.TFString(),
				StateDiskSizeGB: conf.StateDiskSizeGB,
				InitialCount:    workerCount,
				Zone:            conf.Provider.GCP.Zone,
				InstanceType:    conf.Provider.GCP.InstanceType,
				DiskType:        conf.Provider.GCP.StateDiskType,
			},
		},
		Project: conf.Provider.GCP.Project,
		Region:  conf.Provider.GCP.Region,
		Zone:    conf.Provider.GCP.Zone,
		ImageID: imageRef,
		Debug:   conf.IsDebugCluster(),
	}
}

// openStackTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the OpenStack variables.
func openStackTerraformVars(conf *config.Config, imageRef string, controlPlaneCount, workerCount *int) *terraform.OpenStackClusterVariables {
	return &terraform.OpenStackClusterVariables{
		Name:                    conf.Name,
		Cloud:                   toPtr(conf.Provider.OpenStack.Cloud),
		FlavorID:                conf.Provider.OpenStack.FlavorID,
		FloatingIPPoolID:        conf.Provider.OpenStack.FloatingIPPoolID,
		ImageURL:                imageRef,
		DirectDownload:          *conf.Provider.OpenStack.DirectDownload,
		OpenstackUserDomainName: conf.Provider.OpenStack.UserDomainName,
		OpenstackUsername:       conf.Provider.OpenStack.Username,
		OpenstackPassword:       conf.Provider.OpenStack.Password,
		Debug:                   conf.IsDebugCluster(),
		NodeGroups: map[string]terraform.OpenStackNodeGroup{
			constants.ControlPlaneDefault: {
				Role:            role.ControlPlane.TFString(),
				InitialCount:    controlPlaneCount,
				Zone:            conf.Provider.OpenStack.AvailabilityZone, // TODO(elchead): make configurable AB#3225
				StateDiskType:   conf.Provider.OpenStack.StateDiskType,
				StateDiskSizeGB: conf.StateDiskSizeGB,
			},
			constants.WorkerDefault: {
				Role:            role.Worker.TFString(),
				InitialCount:    workerCount,
				Zone:            conf.Provider.OpenStack.AvailabilityZone, // TODO(elchead): make configurable AB#3225
				StateDiskType:   conf.Provider.OpenStack.StateDiskType,
				StateDiskSizeGB: conf.StateDiskSizeGB,
			},
		},
	}
}

// qemuTerraformVars provides variables required to execute the Terraform scripts.
// It should be the only place to declare the QEMU variables.
func qemuTerraformVars(conf *config.Config, imageRef string, controlPlaneCount, workerCount *int, libvirtURI, libvirtSocketPath, metadataLibvirtURI string) *terraform.QEMUVariables {
	return &terraform.QEMUVariables{
		Name:              conf.Name,
		LibvirtURI:        libvirtURI,
		LibvirtSocketPath: libvirtSocketPath,
		// TODO(malt3): auto select boot mode based on attestation variant.
		// requires image info v2.
		BootMode:    "uefi",
		ImagePath:   imageRef,
		ImageFormat: conf.Provider.QEMU.ImageFormat,
		NodeGroups: map[string]terraform.QEMUNodeGroup{
			constants.ControlPlaneDefault: {
				Role:         role.ControlPlane.TFString(),
				InitialCount: controlPlaneCount,
				DiskSize:     conf.StateDiskSizeGB,
				CPUCount:     conf.Provider.QEMU.VCPUs,
				MemorySize:   conf.Provider.QEMU.Memory,
			},
			constants.WorkerDefault: {
				Role:         role.Worker.TFString(),
				InitialCount: workerCount,
				DiskSize:     conf.StateDiskSizeGB,
				CPUCount:     conf.Provider.QEMU.VCPUs,
				MemorySize:   conf.Provider.QEMU.Memory,
			},
		},
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
