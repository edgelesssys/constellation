/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAWSClusterVariables(t *testing.T) {
	vars := AWSClusterVariables{
		CommonVariables: CommonVariables{
			Name:               "cluster-name",
			CountControlPlanes: 1,
			CountWorkers:       2,
			StateDiskSizeGB:    30,
		},
		Region:                 "eu-central-1",
		Zone:                   "eu-central-1a",
		AMIImageID:             "ami-0123456789abcdef",
		InstanceType:           "x1.foo",
		StateDiskType:          "bardisk",
		IAMProfileControlPlane: "arn:aws:iam::123456789012:instance-profile/cluster-name-controlplane",
		IAMProfileWorkerNodes:  "arn:aws:iam::123456789012:instance-profile/cluster-name-worker",
		Debug:                  true,
		EnableSNP:              true,
	}

	// test that the variables are correctly rendered
	want := `name = "cluster-name"
control_plane_count = 1
worker_count = 2
state_disk_size = 30
region = "eu-central-1"
zone = "eu-central-1a"
ami = "ami-0123456789abcdef"
instance_type = "x1.foo"
state_disk_type = "bardisk"
iam_instance_profile_control_plane = "arn:aws:iam::123456789012:instance-profile/cluster-name-controlplane"
iam_instance_profile_worker_nodes = "arn:aws:iam::123456789012:instance-profile/cluster-name-worker"
debug = true
enable_snp = true
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
	want := `name_prefix = "my-prefix"
region = "eu-central-1"
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
			"control_plane_default": {
				Role:            "ControlPlane",
				StateDiskSizeGB: 30,
				InitialCount:    1,
				Zone:            "eu-central-1a",
				InstanceType:    "n2d-standard-4",
				DiskType:        "pd-ssd",
			},
			"worker_default": {
				Role:            "Worker",
				StateDiskSizeGB: 10,
				InitialCount:    1,
				Zone:            "eu-central-1b",
				InstanceType:    "n2d-standard-8",
				DiskType:        "pd-ssd",
			},
		},
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
    role          = "ControlPlane"
    zone          = "eu-central-1a"
  }
  worker_default = {
    disk_size     = 10
    disk_type     = "pd-ssd"
    initial_count = 1
    instance_type = "n2d-standard-8"
    role          = "Worker"
    zone          = "eu-central-1b"
  }
}
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
	want := `project_id = "my-project"
region = "eu-central-1"
zone = "eu-central-1a"
service_account_id = "my-service-account"
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestAzureClusterVariables(t *testing.T) {
	vars := AzureClusterVariables{
		Name: "cluster-name",
		NodeGroups: map[string]AzureNodeGroup{
			"control_plane_default": {
				Role:          "ControlPlane",
				InstanceCount: 1,
				InstanceType:  "Standard_D2s_v3",
				DiskType:      "StandardSSD_LRS",
				DiskSizeGB:    100,
				Location:      "eu-central-1",
			},
		},
		ConfidentialVM:       true,
		SecureBoot:           false,
		ResourceGroup:        "my-resource-group",
		UserAssignedIdentity: "my-user-assigned-identity",
		ImageID:              "image-0123456789abcdef",
		CreateMAA:            true,
		Debug:                true,
	}

	// test that the variables are correctly rendered
	want := `name                   = "cluster-name"
image_id               = "image-0123456789abcdef"
create_maa             = true
debug                  = true
resource_group         = "my-resource-group"
user_assigned_identity = "my-user-assigned-identity"
confidential_vm        = true
secure_boot            = false
node_groups = {
  control_plane_default = {
    disk_size      = 100
    disk_type      = "StandardSSD_LRS"
    instance_count = 1
    instance_type  = "Standard_D2s_v3"
    location       = "eu-central-1"
    role           = "ControlPlane"
  }
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
	want := `service_principal_name = "my-service-principal"
region = "eu-central-1"
resource_group_name = "my-resource-group"
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestOpenStackClusterVariables(t *testing.T) {
	vars := OpenStackClusterVariables{
		CommonVariables: CommonVariables{
			Name:               "cluster-name",
			CountControlPlanes: 1,
			CountWorkers:       2,
			StateDiskSizeGB:    30,
		},
		Cloud:                   "my-cloud",
		AvailabilityZone:        "az-01",
		FlavorID:                "flavor-0123456789abcdef",
		FloatingIPPoolID:        "fip-pool-0123456789abcdef",
		StateDiskType:           "performance-8",
		ImageURL:                "https://example.com/image.raw",
		DirectDownload:          true,
		OpenstackUserDomainName: "my-user-domain",
		OpenstackUsername:       "my-username",
		OpenstackPassword:       "my-password",
		Debug:                   true,
	}

	// test that the variables are correctly rendered
	want := `name = "cluster-name"
control_plane_count = 1
worker_count = 2
state_disk_size = 30
cloud = "my-cloud"
availability_zone = "az-01"
flavor_id = "flavor-0123456789abcdef"
floating_ip_pool_id = "fip-pool-0123456789abcdef"
image_url = "https://example.com/image.raw"
direct_download = true
state_disk_type = "performance-8"
openstack_user_domain_name = "my-user-domain"
openstack_username = "my-username"
openstack_password = "my-password"
debug = true
`
	got := vars.String()
	assert.Equal(t, want, got)
}

func TestQEMUClusterVariables(t *testing.T) {
	vars := QEMUVariables{
		CommonVariables: CommonVariables{
			Name:               "cluster-name",
			CountControlPlanes: 1,
			CountWorkers:       2,
			StateDiskSizeGB:    30,
		},
		LibvirtURI:         "qemu:///system",
		LibvirtSocketPath:  "/var/run/libvirt/libvirt-sock",
		BootMode:           "uefi",
		CPUCount:           4,
		MemorySizeMiB:      8192,
		ImagePath:          "/var/lib/libvirt/images/cluster-name.qcow2",
		ImageFormat:        "raw",
		MetadataAPIImage:   "example.com/metadata-api:latest",
		MetadataLibvirtURI: "qemu:///system",
		NVRAM:              "production",
		Firmware:           "/usr/share/OVMF/OVMF_CODE.fd",
		BzImagePath:        "/var/lib/libvirt/images/cluster-name-bzimage",
		InitrdPath:         "/var/lib/libvirt/images/cluster-name-initrd",
		KernelCmdline:      "console=ttyS0,115200n8",
	}

	// test that the variables are correctly rendered
	want := `name = "cluster-name"
control_plane_count = 1
worker_count = 2
state_disk_size = 30
libvirt_uri = "qemu:///system"
libvirt_socket_path = "/var/run/libvirt/libvirt-sock"
constellation_os_image = "/var/lib/libvirt/images/cluster-name.qcow2"
image_format = "raw"
constellation_boot_mode = "uefi"
constellation_kernel = "/var/lib/libvirt/images/cluster-name-bzimage"
constellation_initrd = "/var/lib/libvirt/images/cluster-name-initrd"
constellation_cmdline = "console=ttyS0,115200n8"
vcpus = 4
memory = 8192
metadata_api_image = "example.com/metadata-api:latest"
metadata_libvirt_uri = "qemu:///system"
nvram = "/usr/share/OVMF/constellation_vars.production.fd"
firmware = "/usr/share/OVMF/OVMF_CODE.fd"
`
	got := vars.String()
	assert.Equal(t, want, got)
}
