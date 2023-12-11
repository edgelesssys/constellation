/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
)

func TestAWSClusterVariables(t *testing.T) {
	vars := AWSClusterVariables{
		Name: "cluster-name",
		NodeGroups: map[string]AWSNodeGroup{
			constants.ControlPlaneDefault: {
				Role:            role.ControlPlane.TFString(),
				StateDiskSizeGB: 30,
				InitialCount:    1,
				Zone:            "eu-central-1b",
				InstanceType:    "x1.foo",
				DiskType:        "foodisk",
			},
			constants.WorkerDefault: {
				Role:            role.Worker.TFString(),
				StateDiskSizeGB: 30,
				InitialCount:    2,
				Zone:            "eu-central-1c",
				InstanceType:    "x1.bar",
				DiskType:        "bardisk",
			},
		},
		Region:                 "eu-central-1",
		Zone:                   "eu-central-1a",
		AMIImageID:             "ami-0123456789abcdef",
		IAMProfileControlPlane: "arn:aws:iam::123456789012:instance-profile/cluster-name-controlplane",
		IAMProfileWorkerNodes:  "arn:aws:iam::123456789012:instance-profile/cluster-name-worker",
		Debug:                  true,
		EnableSNP:              true,
		CustomEndpoint:         "example.com",
	}

	// test that the variables are correctly rendered
	want := `name                               = "cluster-name"
region                             = "eu-central-1"
zone                               = "eu-central-1a"
ami                                = "ami-0123456789abcdef"
iam_instance_profile_control_plane = "arn:aws:iam::123456789012:instance-profile/cluster-name-controlplane"
iam_instance_profile_worker_nodes  = "arn:aws:iam::123456789012:instance-profile/cluster-name-worker"
debug                              = true
enable_snp                         = true
node_groups = {
  control_plane_default = {
    disk_size     = 30
    disk_type     = "foodisk"
    initial_count = 1
    instance_type = "x1.foo"
    role          = "control-plane"
    zone          = "eu-central-1b"
  }
  worker_default = {
    disk_size     = 30
    disk_type     = "bardisk"
    initial_count = 2
    instance_type = "x1.bar"
    role          = "worker"
    zone          = "eu-central-1c"
  }
}
custom_endpoint        = "example.com"
internal_load_balancer = false
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestAWSIAMVariables(t *testing.T) {
	vars := AWSIAMVariables{
		Region: "eu-central-1",
		Prefix: "my-prefix",
	}

	// test that the variables are correctly rendered
	want := `region      = "eu-central-1"
name_prefix = "my-prefix"
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestGCPClusterVariables(t *testing.T) {
	vars := GCPClusterVariables{
		Name:    "cluster-name",
		Project: "my-project",
		Region:  "eu-central-1",
		Zone:    "eu-central-1a",
		ImageID: "image-0123456789abcdef",
		Debug:   true,
		NodeGroups: map[string]GCPNodeGroup{
			constants.ControlPlaneDefault: {
				Role:            "control-plane",
				StateDiskSizeGB: 30,
				InitialCount:    1,
				Zone:            "eu-central-1a",
				InstanceType:    "n2d-standard-4",
				DiskType:        "pd-ssd",
			},
			constants.WorkerDefault: {
				Role:            "worker",
				StateDiskSizeGB: 10,
				InitialCount:    1,
				Zone:            "eu-central-1b",
				InstanceType:    "n2d-standard-8",
				DiskType:        "pd-ssd",
			},
		},
		CustomEndpoint: "example.com",
	}

	// test that the variables are correctly rendered
	want := `name     = "cluster-name"
project  = "my-project"
region   = "eu-central-1"
zone     = "eu-central-1a"
image_id = "image-0123456789abcdef"
debug    = true
node_groups = {
  control_plane_default = {
    disk_size     = 30
    disk_type     = "pd-ssd"
    initial_count = 1
    instance_type = "n2d-standard-4"
    role          = "control-plane"
    zone          = "eu-central-1a"
  }
  worker_default = {
    disk_size     = 10
    disk_type     = "pd-ssd"
    initial_count = 1
    instance_type = "n2d-standard-8"
    role          = "worker"
    zone          = "eu-central-1b"
  }
}
custom_endpoint        = "example.com"
internal_load_balancer = false
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestGCPIAMVariables(t *testing.T) {
	vars := GCPIAMVariables{
		Project:          "my-project",
		Region:           "eu-central-1",
		Zone:             "eu-central-1a",
		ServiceAccountID: "my-service-account",
	}

	// test that the variables are correctly rendered
	want := `project_id         = "my-project"
region             = "eu-central-1"
zone               = "eu-central-1a"
service_account_id = "my-service-account"
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestAzureClusterVariables(t *testing.T) {
	vars := AzureClusterVariables{
		Name: "cluster-name",
		NodeGroups: map[string]AzureNodeGroup{
			constants.ControlPlaneDefault: {
				Role:         "ControlPlane",
				InitialCount: 1,
				InstanceType: "Standard_D2s_v3",
				DiskType:     "StandardSSD_LRS",
				DiskSizeGB:   100,
			},
		},
		ConfidentialVM:       to.Ptr(true),
		ResourceGroup:        "my-resource-group",
		UserAssignedIdentity: "my-user-assigned-identity",
		ImageID:              "image-0123456789abcdef",
		CreateMAA:            to.Ptr(true),
		Debug:                to.Ptr(true),
		Location:             "eu-central-1",
		CustomEndpoint:       "example.com",
		MarketplaceImage: &AzureMarketplaceImageVariables{
			Publisher: "edgelesssys",
			Product:   "constellation",
			Name:      "constellation",
			Version:   "2.13.0",
		},
	}

	// test that the variables are correctly rendered
	want := `name                   = "cluster-name"
image_id               = "image-0123456789abcdef"
create_maa             = true
debug                  = true
resource_group         = "my-resource-group"
location               = "eu-central-1"
user_assigned_identity = "my-user-assigned-identity"
confidential_vm        = true
node_groups = {
  control_plane_default = {
    disk_size     = 100
    disk_type     = "StandardSSD_LRS"
    initial_count = 1
    instance_type = "Standard_D2s_v3"
    role          = "ControlPlane"
    zones         = null
  }
}
custom_endpoint        = "example.com"
internal_load_balancer = false
marketplace_image = {
  name      = "constellation"
  product   = "constellation"
  publisher = "edgelesssys"
  version   = "2.13.0"
}
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestAzureIAMVariables(t *testing.T) {
	vars := AzureIAMVariables{
		Region:           "eu-central-1",
		ServicePrincipal: "my-service-principal",
		ResourceGroup:    "my-resource-group",
	}

	// test that the variables are correctly rendered
	want := `region                 = "eu-central-1"
service_principal_name = "my-service-principal"
resource_group_name    = "my-resource-group"
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestOpenStackClusterVariables(t *testing.T) {
	vars := OpenStackClusterVariables{
		Name:                    "cluster-name",
		Cloud:                   toPtr("my-cloud"),
		FloatingIPPoolID:        "fip-pool-0123456789abcdef",
		ImageURL:                "https://example.com/image.raw",
		DirectDownload:          true,
		OpenstackUserDomainName: "my-user-domain",
		OpenstackUsername:       "my-username",
		OpenstackPassword:       "my-password",
		Debug:                   true,
		NodeGroups: map[string]OpenStackNodeGroup{
			constants.ControlPlaneDefault: {
				Role:            "control-plane",
				InitialCount:    1,
				FlavorID:        "flavor-0123456789abcdef",
				Zone:            "az-01",
				StateDiskType:   "performance-8",
				StateDiskSizeGB: 30,
			},
		},
		CustomEndpoint: "example.com",
	}

	// test that the variables are correctly rendered
	want := `name = "cluster-name"
node_groups = {
  control_plane_default = {
    flavor_id       = "flavor-0123456789abcdef"
    initial_count   = 1
    role            = "control-plane"
    state_disk_size = 30
    state_disk_type = "performance-8"
    zone            = "az-01"
  }
}
cloud                      = "my-cloud"
floating_ip_pool_id        = "fip-pool-0123456789abcdef"
image_url                  = "https://example.com/image.raw"
direct_download            = true
openstack_user_domain_name = "my-user-domain"
openstack_username         = "my-username"
openstack_password         = "my-password"
debug                      = true
custom_endpoint            = "example.com"
internal_load_balancer     = false
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestQEMUClusterVariables(t *testing.T) {
	vars := &QEMUVariables{
		Name: "cluster-name",
		NodeGroups: map[string]QEMUNodeGroup{
			"control-plane": {
				Role:         role.ControlPlane.TFString(),
				InitialCount: 1,
				DiskSize:     30,
				CPUCount:     4,
				MemorySize:   8192,
			},
		},
		Machine:            "q35",
		LibvirtURI:         "qemu:///system",
		LibvirtSocketPath:  "/var/run/libvirt/libvirt-sock",
		BootMode:           "uefi",
		ImagePath:          "/var/lib/libvirt/images/cluster-name.qcow2",
		ImageFormat:        "raw",
		MetadataAPIImage:   "example.com/metadata-api:latest",
		MetadataLibvirtURI: "qemu:///system",
		NVRAM:              "production",
		InitrdPath:         toPtr("/var/lib/libvirt/images/cluster-name-initrd"),
		KernelCmdline:      toPtr("console=ttyS0,115200n8"),
		CustomEndpoint:     "example.com",
	}

	// test that the variables are correctly rendered
	want := `name = "cluster-name"
node_groups = {
  control-plane = {
    disk_size     = 30
    initial_count = 1
    memory        = 8192
    role          = "control-plane"
    vcpus         = 4
  }
}
machine                 = "q35"
libvirt_uri             = "qemu:///system"
libvirt_socket_path     = "/var/run/libvirt/libvirt-sock"
constellation_boot_mode = "uefi"
constellation_os_image  = "/var/lib/libvirt/images/cluster-name.qcow2"
image_format            = "raw"
metadata_api_image      = "example.com/metadata-api:latest"
metadata_libvirt_uri    = "qemu:///system"
nvram                   = "/usr/share/OVMF/OVMF_VARS.fd"
constellation_initrd    = "/var/lib/libvirt/images/cluster-name-initrd"
constellation_cmdline   = "console=ttyS0,115200n8"
custom_endpoint         = "example.com"
internal_load_balancer  = false
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestVariablesFromBytes(t *testing.T) {
	assert := assert.New(t)

	awsVars := AWSIAMVariables{
		Region: "test",
	}
	var loadedAWSVars AWSIAMVariables
	err := VariablesFromBytes([]byte(awsVars.String()), &loadedAWSVars)
	assert.NoError(err)
	assert.Equal(awsVars, loadedAWSVars)

	azureVars := AzureIAMVariables{
		Region: "test",
	}
	var loadedAzureVars AzureIAMVariables
	err = VariablesFromBytes([]byte(azureVars.String()), &loadedAzureVars)
	assert.NoError(err)
	assert.Equal(azureVars, loadedAzureVars)

	gcpVars := GCPIAMVariables{
		Region: "test",
	}
	var loadedGCPVars GCPIAMVariables
	err = VariablesFromBytes([]byte(gcpVars.String()), &loadedGCPVars)
	assert.NoError(err)
	assert.Equal(gcpVars, loadedGCPVars)

	err = VariablesFromBytes([]byte("invalid"), &loadedGCPVars)
	assert.Error(err)
}
