/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Variables is a struct that holds all variables that are passed to Terraform.
type Variables interface {
	fmt.Stringer
}

// CommonVariables is user configuration for creating a cluster with Terraform.
type CommonVariables struct {
	// Name of the cluster.
	Name string
	// CountControlPlanes is the number of control-plane nodes to create.
	CountControlPlanes int
	// CountWorkers is the number of worker nodes to create.
	CountWorkers int
	// StateDiskSizeGB is the size of the state disk to allocate to each node, in GB.
	StateDiskSizeGB int
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *CommonVariables) String() string {
	b := &strings.Builder{}
	writeLinef(b, "name = %q", v.Name)
	writeLinef(b, "control_plane_count = %d", v.CountControlPlanes)
	writeLinef(b, "worker_count = %d", v.CountWorkers)
	writeLinef(b, "state_disk_size = %d", v.StateDiskSizeGB)

	return b.String()
}

// AWSClusterVariables is user configuration for creating a cluster with Terraform on AWS.
type AWSClusterVariables struct {
	// Name of the cluster.
	Name string `hcl:"name" cty:"name"`
	// Region is the AWS region to use.
	Region string `hcl:"region" cty:"region"`
	// Zone is the AWS zone to use in the given region.
	Zone string `hcl:"zone" cty:"zone"`
	// AMIImageID is the ID of the AMI image to use.
	AMIImageID string `hcl:"ami" cty:"ami"`
	// IAMGroupControlPlane is the IAM group to use for the control-plane nodes.
	IAMProfileControlPlane string `hcl:"iam_instance_profile_control_plane" cty:"iam_instance_profile_control_plane"`
	// IAMGroupWorkerNodes is the IAM group to use for the worker nodes.
	IAMProfileWorkerNodes string `hcl:"iam_instance_profile_worker_nodes" cty:"iam_instance_profile_worker_nodes"`
	// Debug is true if debug mode is enabled.
	Debug bool `hcl:"debug" cty:"debug"`
	// EnableSNP controls enablement of the EC2 cpu-option "AmdSevSnp".
	EnableSNP bool `hcl:"enable_snp" cty:"enable_snp"`
	// NodeGroups is a map of node groups to create.
	NodeGroups map[string]AWSNodeGroup `hcl:"node_groups" cty:"node_groups"`
}

// AWSNodeGroup is a node group to create on AWS.
type AWSNodeGroup struct {
	// Role is the role of the node group.
	Role string `hcl:"role" cty:"role"`
	// StateDiskSizeGB is the size of the state disk to allocate to each node, in GB.
	StateDiskSizeGB int `hcl:"disk_size" cty:"disk_size"`
	// InitialCount is the initial number of nodes to create in the node group.
	InitialCount int `hcl:"initial_count" cty:"initial_count"`
	// Zone is the AWS availability-zone to use in the given region.
	Zone string `hcl:"zone" cty:"zone"`
	// InstanceType is the type of the EC2 instance to use.
	InstanceType string `hcl:"instance_type" cty:"instance_type"`
	// DiskType is the EBS disk type to use for the state disk.
	DiskType string `hcl:"disk_type" cty:"disk_type"`
}

func (v *AWSClusterVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(v, f.Body())
	return string(f.Bytes())
}

// AWSIAMVariables is user configuration for creating the IAM configuration with Terraform on Microsoft Azure.
type AWSIAMVariables struct {
	// Region is the AWS location to use. (e.g. us-east-2)
	Region string
	// Prefix is the name prefix of the resources to use.
	Prefix string
}

// String returns a string representation of the IAM-specific variables, formatted as Terraform variables.
func (v *AWSIAMVariables) String() string {
	b := &strings.Builder{}
	writeLinef(b, "name_prefix = %q", v.Prefix)
	writeLinef(b, "region = %q", v.Region)

	return b.String()
}

// GCPClusterVariables is user configuration for creating resources with Terraform on GCP.
type GCPClusterVariables struct {
	// Name of the cluster.
	Name string `hcl:"name" cty:"name"`
	// Project is the ID of the GCP project to use.
	Project string `hcl:"project" cty:"project"`
	// Region is the GCP region to use.
	Region string `hcl:"region" cty:"region"`
	// Zone is the GCP zone to use.
	Zone string `hcl:"zone" cty:"zone"`
	// ImageID is the ID of the GCP image to use.
	ImageID string `hcl:"image_id" cty:"image_id"`
	// Debug is true if debug mode is enabled.
	Debug bool `hcl:"debug" cty:"debug"`
	// NodeGroups is a map of node groups to create.
	NodeGroups map[string]GCPNodeGroup `hcl:"node_groups" cty:"node_groups"`
}

// GCPNodeGroup is a node group to create on GCP.
type GCPNodeGroup struct {
	// Role is the role of the node group.
	Role string `hcl:"role" cty:"role"`
	// StateDiskSizeGB is the size of the state disk to allocate to each node, in GB.
	StateDiskSizeGB int `hcl:"disk_size" cty:"disk_size"`
	// InitialCount is the initial number of nodes to create in the node group.
	InitialCount int    `hcl:"initial_count" cty:"initial_count"`
	Zone         string `hcl:"zone" cty:"zone"`
	InstanceType string `hcl:"instance_type" cty:"instance_type"`
	DiskType     string `hcl:"disk_type" cty:"disk_type"`
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *GCPClusterVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(v, f.Body())
	return string(f.Bytes())
}

// GCPIAMVariables is user configuration for creating the IAM confioguration with Terraform on GCP.
type GCPIAMVariables struct {
	// Project is the ID of the GCP project to use.
	Project string
	// Region is the GCP region to use.
	Region string
	// Zone is the GCP zone to use.
	Zone string
	// ServiceAccountID is the ID of the service account to use.
	ServiceAccountID string
}

// String returns a string representation of the IAM-specific variables, formatted as Terraform variables.
func (v *GCPIAMVariables) String() string {
	b := &strings.Builder{}
	writeLinef(b, "project_id = %q", v.Project)
	writeLinef(b, "region = %q", v.Region)
	writeLinef(b, "zone = %q", v.Zone)
	writeLinef(b, "service_account_id = %q", v.ServiceAccountID)

	return b.String()
}

// AzureClusterVariables is user configuration for creating a cluster with Terraform on Azure.
type AzureClusterVariables struct {
	// Name of the cluster.
	Name string `hcl:"name" cty:"name"`
	// ImageID is the ID of the Azure image to use.
	ImageID string `hcl:"image_id" cty:"image_id"`
	// CreateMAA sets whether a Microsoft Azure attestation provider should be created.
	CreateMAA *bool `hcl:"create_maa" cty:"create_maa"`
	// Debug is true if debug mode is enabled.
	Debug *bool `hcl:"debug" cty:"debug"`
	// ResourceGroup is the name of the Azure resource group to use.
	ResourceGroup string `hcl:"resource_group" cty:"resource_group"`
	// Location is the Azure location to use.
	Location string `hcl:"location" cty:"location"`
	// UserAssignedIdentity is the name of the Azure user-assigned identity to use.
	UserAssignedIdentity string `hcl:"user_assigned_identity" cty:"user_assigned_identity"`
	// ConfidentialVM sets the VM to be confidential.
	ConfidentialVM *bool `hcl:"confidential_vm" cty:"confidential_vm"`
	// SecureBoot sets the VM to use secure boot.
	SecureBoot *bool `hcl:"secure_boot" cty:"secure_boot"`
	// NodeGroups is a map of node groups to create.
	NodeGroups map[string]AzureNodeGroup `hcl:"node_groups" cty:"node_groups"`
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *AzureClusterVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(v, f.Body())
	return string(f.Bytes())
}

// AzureNodeGroup is a node group to create on Azure.
type AzureNodeGroup struct {
	// Role is the role of the node group.
	Role string `hcl:"role" cty:"role"`
	// InitialCount is optional for upgrades.
	InitialCount *int      `hcl:"initial_count" cty:"initial_count"`
	InstanceType string    `hcl:"instance_type" cty:"instance_type"`
	DiskSizeGB   int       `hcl:"disk_size" cty:"disk_size"`
	DiskType     string    `hcl:"disk_type" cty:"disk_type"`
	Zones        *[]string `hcl:"zones" cty:"zones"`
}

// AzureIAMVariables is user configuration for creating the IAM configuration with Terraform on Microsoft Azure.
type AzureIAMVariables struct {
	// Region is the Azure region to use. (e.g. westus)
	Region string
	// ServicePrincipal is the name of the service principal to use.
	ServicePrincipal string
	// ResourceGroup is the name of the resource group to use.
	ResourceGroup string
}

// String returns a string representation of the IAM-specific variables, formatted as Terraform variables.
func (v *AzureIAMVariables) String() string {
	b := &strings.Builder{}
	writeLinef(b, "service_principal_name = %q", v.ServicePrincipal)
	writeLinef(b, "region = %q", v.Region)
	writeLinef(b, "resource_group_name = %q", v.ResourceGroup)

	return b.String()
}

// OpenStackClusterVariables is user configuration for creating a cluster with Terraform on OpenStack.
type OpenStackClusterVariables struct {
	// Name of the cluster.
	Name string `hcl:"name" cty:"name"`
	// NodeGroups is a map of node groups to create.
	NodeGroups map[string]OpenStackNodeGroup `hcl:"node_groups" cty:"node_groups"`
	// Cloud is the (optional) name of the OpenStack cloud to use when reading the "clouds.yaml" configuration file. If empty, environment variables are used.
	Cloud *string `hcl:"cloud" cty:"cloud"`
	// Flavor is the ID of the OpenStack flavor (machine type) to use.
	FlavorID string `hcl:"flavor_id" cty:"flavor_id"`
	// FloatingIPPoolID is the ID of the OpenStack floating IP pool to use for public IPs.
	FloatingIPPoolID string `hcl:"floating_ip_pool_id" cty:"floating_ip_pool_id"`
	// ImageURL is the URL of the OpenStack image to use.
	ImageURL string `hcl:"image_url" cty:"image_url"`
	// DirectDownload decides whether to download the image directly from the URL to OpenStack or to upload it from the local machine.
	DirectDownload bool `hcl:"direct_download" cty:"direct_download"`
	// OpenstackUserDomainName is the OpenStack user domain name to use.
	OpenstackUserDomainName string `hcl:"openstack_user_domain_name" cty:"openstack_user_domain_name"`
	// OpenstackUsername is the OpenStack user name to use.
	OpenstackUsername string `hcl:"openstack_username" cty:"openstack_username"`
	// OpenstackPassword is the OpenStack password to use.
	OpenstackPassword string `hcl:"openstack_password" cty:"openstack_password"`
	// Debug is true if debug mode is enabled.
	Debug bool `hcl:"debug" cty:"debug"`
}

// OpenStackNodeGroup is a node group to create on OpenStack.
type OpenStackNodeGroup struct {
	// Role is the role of the node group.
	Role string `hcl:"role" cty:"role"`
	// InitialCount is the number of instances to create.
	InitialCount int `hcl:"initial_count" cty:"initial_count"`
	// Zone is the OpenStack availability zone to use.
	Zone string `hcl:"zone" cty:"zone"`
	// StateDiskType is the OpenStack disk type to use for the state disk.
	StateDiskType string `hcl:"state_disk_type" cty:"state_disk_type"`
	// StateDiskSizeGB is the size of the state disk to allocate to each node, in GB.
	StateDiskSizeGB int `hcl:"state_disk_size" cty:"state_disk_size"`
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *OpenStackClusterVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(v, f.Body())
	return string(f.Bytes())
}

// TODO(malt3): Add support for OpenStack IAM variables.

// QEMUVariables is user configuration for creating a QEMU cluster with Terraform.
type QEMUVariables struct {
	// Name is the name to use for the cluster.
	Name string `hcl:"name" cty:"name"`
	// NodeGroups is a map of node groups to create.
	NodeGroups map[string]QEMUNodeGroup `hcl:"node_groups" cty:"node_groups"`
	// Machine is the type of machine to use.  use 'q35' for secure boot and 'pc' for non secure boot. See 'qemu-system-x86_64 -machine help'
	Machine string `hcl:"machine" cty:"machine"`
	// LibvirtURI is the libvirt connection URI.
	LibvirtURI string `hcl:"libvirt_uri" cty:"libvirt_uri"`
	// LibvirtSocketPath is the path to the libvirt socket in case of unix socket.
	LibvirtSocketPath string `hcl:"libvirt_socket_path" cty:"libvirt_socket_path"`
	// BootMode is the boot mode to use.
	// Can be either "uefi" or "direct-linux-boot".
	BootMode string `hcl:"constellation_boot_mode" cty:"constellation_boot_mode"`
	// ImagePath is the path to the image to use for the nodes.
	ImagePath string `hcl:"constellation_os_image" cty:"constellation_os_image"`
	// ImageFormat is the format of the image from ImagePath.
	ImageFormat string `hcl:"image_format" cty:"image_format"`
	// MetadataAPIImage is the container image to use for the metadata API.
	MetadataAPIImage string `hcl:"metadata_api_image" cty:"metadata_api_image"`
	// MetadataLibvirtURI is the libvirt connection URI used by the metadata container.
	// In case of unix socket, this should be "qemu:///system".
	// Other wise it should be the same as LibvirtURI.
	MetadataLibvirtURI string `hcl:"metadata_libvirt_uri" cty:"metadata_libvirt_uri"`
	// NVRAM is the path to the NVRAM template.
	NVRAM string `hcl:"nvram" cty:"nvram"`
	// Firmware is the path to the firmware.
	Firmware *string `hcl:"firmware" cty:"firmware"`
	// BzImagePath is the path to the bzImage (kernel).
	BzImagePath *string `hcl:"constellation_kernel" cty:"constellation_kernel"`
	// InitrdPath is the path to the initrd.
	InitrdPath *string `hcl:"constellation_initrd" cty:"constellation_initrd"`
	// KernelCmdline is the kernel command line.
	KernelCmdline *string `hcl:"constellation_cmdline" cty:"constellation_cmdline"`
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *QEMUVariables) String() string {
	// copy v object
	vCopy := *v
	switch vCopy.NVRAM {
	case "production":
		vCopy.NVRAM = "/usr/share/OVMF/constellation_vars.production.fd"
	case "testing":
		vCopy.NVRAM = "/usr/share/OVMF/constellation_vars.testing.fd"
	}
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(vCopy, f.Body())
	return string(f.Bytes())
}

// QEMUNodeGroup is a node group for a QEMU cluster.
type QEMUNodeGroup struct {
	// Role is the role of the node group.
	Role string `hcl:"role" cty:"role"`
	// InitialCount is the number of instances to create.
	InitialCount int `hcl:"initial_count" cty:"initial_count"`
	// DiskSize is the size of the disk to allocate to each node, in GiB.
	DiskSize int `hcl:"disk_size" cty:"disk_size"`
	// CPUCount is the number of CPUs to allocate to each node.
	CPUCount int `hcl:"vcpus" cty:"vcpus"`
	// MemorySize is the amount of memory to allocate to each node, in MiB.
	MemorySize int `hcl:"memory" cty:"memory"`
}

func writeLinef(builder *strings.Builder, format string, a ...any) {
	builder.WriteString(fmt.Sprintf(format, a...))
	builder.WriteByte('\n')
}
