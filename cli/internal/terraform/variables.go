/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Variables is a struct that holds all variables that are passed to Terraform.
type Variables interface {
	fmt.Stringer
}

// ClusterVariables should be used in places where a cluster is created.
type ClusterVariables interface {
	Variables
	// TODO(derpsteb): Rename this function once we have introduced an interface for config.Config.
	// GetCreateMAA does not follow Go's naming convention because we need to keep the CreateMAA property public for now.
	// There are functions creating Variables objects outside of this package.
	// These functions can only be moved into this package once we have introduced an interface for config.Config,
	// since we do not want to introduce a dependency on config.Config in this package.
	GetCreateMAA() bool
}

// VariablesFromBytes parses the given bytes into the given variables struct.
func VariablesFromBytes[T any](b []byte, vars *T) error {
	file, err := hclsyntax.ParseConfig(b, "", hcl.Pos{Line: 1, Column: 1})
	if err != nil {
		return fmt.Errorf("parsing variables: %w", err)
	}

	diags := gohcl.DecodeBody(file.Body, nil, vars)
	if diags.HasErrors() {
		return fmt.Errorf("decoding variables: %w", diags)
	}
	return nil
}

// AWSClusterVariables is user configuration for creating a cluster with Terraform on AWS.
type AWSClusterVariables struct {
	// Name of the cluster.
	Name string `hcl:"name" cty:"name"`
	// Region is the AWS region to use.
	Region string `hcl:"region" cty:"region"`
	// Zone is the AWS zone to use in the given region.
	Zone string `hcl:"zone" cty:"zone"`
	// ImageID is the ID of the AMI to use.
	ImageID string `hcl:"image_id" cty:"image_id"`
	// IAMProfileControlPlane is the IAM group to use for the control-plane nodes.
	IAMProfileControlPlane string `hcl:"iam_instance_profile_name_control_plane" cty:"iam_instance_profile_name_control_plane"`
	// IAMProfileWorkerNodes is the IAM group to use for the worker nodes.
	IAMProfileWorkerNodes string `hcl:"iam_instance_profile_name_worker_nodes" cty:"iam_instance_profile_name_worker_nodes"`
	// Debug is true if debug mode is enabled.
	Debug bool `hcl:"debug" cty:"debug"`
	// EnableSNP controls enablement of the EC2 cpu-option "AmdSevSnp".
	EnableSNP bool `hcl:"enable_snp" cty:"enable_snp"`
	// NodeGroups is a map of node groups to create.
	NodeGroups map[string]AWSNodeGroup `hcl:"node_groups" cty:"node_groups"`
	// CustomEndpoint is the (optional) custom dns hostname for the kubernetes api server.
	CustomEndpoint string `hcl:"custom_endpoint" cty:"custom_endpoint"`
	// InternalLoadBalancer is true if an internal load balancer should be created.
	InternalLoadBalancer bool `hcl:"internal_load_balancer" cty:"internal_load_balancer"`
	// AdditionalTags describes (optional) additional tags that should be applied to created resources.
	AdditionalTags cloudprovider.Tags `hcl:"additional_tags" cty:"additional_tags"`
}

// GetCreateMAA gets the CreateMAA variable.
// TODO(derpsteb): Rename this function once we have introduced an interface for config.Config.
func (a *AWSClusterVariables) GetCreateMAA() bool {
	return false
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (a *AWSClusterVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(a, f.Body())
	return string(f.Bytes())
}

// AWSNodeGroup is a node group to create on AWS.
type AWSNodeGroup struct {
	// Role is the role of the node group.
	Role string `hcl:"role" cty:"role"`
	// StateDiskSizeGB is the size of the state disk to allocate to each node, in GB.
	StateDiskSizeGB int `hcl:"disk_size" cty:"disk_size"`
	// InitialCount is the initial number of nodes to create in the node group.
	// During upgrades this value ignored.
	InitialCount int `hcl:"initial_count" cty:"initial_count"`
	// Zone is the AWS availability-zone to use in the given region.
	Zone string `hcl:"zone" cty:"zone"`
	// InstanceType is the type of the EC2 instance to use.
	InstanceType string `hcl:"instance_type" cty:"instance_type"`
	// DiskType is the EBS disk type to use for the state disk.
	DiskType string `hcl:"disk_type" cty:"disk_type"`
}

// AWSIAMVariables is user configuration for creating the IAM configuration with Terraform on Microsoft Azure.
type AWSIAMVariables struct {
	// Region is the AWS location to use. (e.g. us-east-2)
	Region string `hcl:"region" cty:"region"`
	// Prefix is the name prefix of the resources to use.
	Prefix string `hcl:"name_prefix" cty:"name_prefix"`
}

// String returns a string representation of the IAM-specific variables, formatted as Terraform variables.
func (v *AWSIAMVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(v, f.Body())
	return string(f.Bytes())
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
	// CustomEndpoint is the (optional) custom dns hostname for the kubernetes api server.
	CustomEndpoint string `hcl:"custom_endpoint" cty:"custom_endpoint"`
	// InternalLoadBalancer is true if an internal load balancer should be created.
	InternalLoadBalancer bool `hcl:"internal_load_balancer" cty:"internal_load_balancer"`
	// CCTechnology is the confidential computing technology to use on the VMs. (`SEV` or `SEV_SNP`)
	CCTechnology string `hcl:"cc_technology" cty:"cc_technology"`
	// AdditionalLables are (optional) additional labels that should be applied to created resources.
	AdditionalLabels cloudprovider.Tags `hcl:"additional_labels" cty:"additional_labels"`
}

// GetCreateMAA gets the CreateMAA variable.
// TODO(derpsteb): Rename this function once we have introduced an interface for config.Config.
func (g *GCPClusterVariables) GetCreateMAA() bool {
	return false
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (g *GCPClusterVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(g, f.Body())
	return string(f.Bytes())
}

// GCPNodeGroup is a node group to create on GCP.
type GCPNodeGroup struct {
	// Role is the role of the node group.
	Role string `hcl:"role" cty:"role"`
	// StateDiskSizeGB is the size of the state disk to allocate to each node, in GB.
	StateDiskSizeGB int `hcl:"disk_size" cty:"disk_size"`
	// InitialCount is the initial number of nodes to create in the node group.
	// During upgrades this value ignored.
	InitialCount int    `hcl:"initial_count" cty:"initial_count"`
	Zone         string `hcl:"zone" cty:"zone"`
	InstanceType string `hcl:"instance_type" cty:"instance_type"`
	DiskType     string `hcl:"disk_type" cty:"disk_type"`
}

// GCPIAMVariables is user configuration for creating the IAM configuration with Terraform on GCP.
type GCPIAMVariables struct {
	// Project is the ID of the GCP project to use.
	Project string `hcl:"project_id" cty:"project_id"`
	// Region is the GCP region to use.
	Region string `hcl:"region" cty:"region"`
	// Zone is the GCP zone to use.
	Zone string `hcl:"zone" cty:"zone"`
	// ServiceAccountID is the ID of the service account to use.
	ServiceAccountID string `hcl:"service_account_id" cty:"service_account_id"`
}

// String returns a string representation of the IAM-specific variables, formatted as Terraform variables.
func (v *GCPIAMVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(v, f.Body())
	return string(f.Bytes())
}

// AzureClusterVariables is user configuration for creating a cluster with Terraform on Azure.
type AzureClusterVariables struct {
	// SubscriptionID is the Azure subscription ID to use.
	SubscriptionID string `hcl:"subscription_id" cty:"subscription_id"`
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
	// CustomEndpoint is the (optional) custom dns hostname for the kubernetes api server.
	CustomEndpoint string `hcl:"custom_endpoint" cty:"custom_endpoint"`
	// InternalLoadBalancer is true if an internal load balancer should be created.
	InternalLoadBalancer bool `hcl:"internal_load_balancer" cty:"internal_load_balancer"`
	// MarketplaceImage is the (optional) Azure Marketplace image to use.
	MarketplaceImage *AzureMarketplaceImageVariables `hcl:"marketplace_image" cty:"marketplace_image"`
	// AdditionalTags are (optional) additional tags that get applied to created resources.
	AdditionalTags cloudprovider.Tags `hcl:"additional_tags" cty:"additional_tags"`
}

// GetCreateMAA gets the CreateMAA variable.
// TODO(derpsteb): Rename this function once we have introduced an interface for config.Config.
func (a *AzureClusterVariables) GetCreateMAA() bool {
	if a.CreateMAA == nil {
		return false
	}

	return *a.CreateMAA
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (a *AzureClusterVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(a, f.Body())
	return string(f.Bytes())
}

// AzureNodeGroup is a node group to create on Azure.
type AzureNodeGroup struct {
	// Role is the role of the node group.
	Role string `hcl:"role" cty:"role"`
	// InitialCount is optional for upgrades.
	InitialCount int      `hcl:"initial_count" cty:"initial_count"`
	InstanceType string   `hcl:"instance_type" cty:"instance_type"`
	DiskSizeGB   int      `hcl:"disk_size" cty:"disk_size"`
	DiskType     string   `hcl:"disk_type" cty:"disk_type"`
	Zones        []string `hcl:"zones" cty:"zones"`
}

// AzureIAMVariables is user configuration for creating the IAM configuration with Terraform on Microsoft Azure.
type AzureIAMVariables struct {
	// SubscriptionID is the Azure subscription ID to use.
	SubscriptionID string `hcl:"subscription_id,optional" cty:"subscription_id"` // TODO(v2.18): remove optional tag. This is only required for migration from var files that dont have the value yet.
	// Location is the Azure location to use. (e.g. westus)
	Location string `hcl:"location" cty:"location"`
	// ServicePrincipal is the name of the service principal to use.
	ServicePrincipal string `hcl:"service_principal_name" cty:"service_principal_name"`
	// ResourceGroup is the name of the resource group to use.
	ResourceGroup string `hcl:"resource_group_name" cty:"resource_group_name"`
}

// AzureMarketplaceImageVariables is a configuration for specifying an Azure Marketplace image.
type AzureMarketplaceImageVariables struct {
	// Publisher is the publisher ID of the image.
	Publisher string `hcl:"publisher" cty:"publisher"`
	// Product is the product ID of the image.
	Product string `hcl:"product" cty:"product"`
	// Name is the name of the image.
	Name string `hcl:"name" cty:"name"`
	// Version is the version of the image.
	Version string `hcl:"version" cty:"version"`
}

// String returns a string representation of the IAM-specific variables, formatted as Terraform variables.
func (v *AzureIAMVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(v, f.Body())
	return string(f.Bytes())
}

// OpenStackClusterVariables is user configuration for creating a cluster with Terraform on OpenStack.
type OpenStackClusterVariables struct {
	// Name of the cluster.
	Name string `hcl:"name" cty:"name"`
	// NodeGroups is a map of node groups to create.
	NodeGroups map[string]OpenStackNodeGroup `hcl:"node_groups" cty:"node_groups"`
	// Cloud is the name of the OpenStack cloud to use when reading the "clouds.yaml" configuration file. If empty, environment variables are used.
	Cloud *string `hcl:"cloud" cty:"cloud"`
	// OpenStackCloudsYAMLPath is the path to the OpenStack clouds.yaml file
	OpenStackCloudsYAMLPath string `hcl:"openstack_clouds_yaml_path" cty:"openstack_clouds_yaml_path"`
	// (STACKIT only) STACKITProjectID is the ID of the STACKIT project to use.
	STACKITProjectID string `hcl:"stackit_project_id" cty:"stackit_project_id"`
	// FloatingIPPoolID is the ID of the OpenStack floating IP pool to use for public IPs.
	FloatingIPPoolID string `hcl:"floating_ip_pool_id" cty:"floating_ip_pool_id"`
	// ImageID is the ID of the OpenStack image to use.
	ImageID string `hcl:"image_id" cty:"image_id"`
	// Debug is true if debug mode is enabled.
	Debug bool `hcl:"debug" cty:"debug"`
	// CustomEndpoint is the (optional) custom dns hostname for the kubernetes api server.
	CustomEndpoint string `hcl:"custom_endpoint" cty:"custom_endpoint"`
	// InternalLoadBalancer is true if an internal load balancer should be created.
	InternalLoadBalancer bool     `hcl:"internal_load_balancer" cty:"internal_load_balancer"`
	AdditionalTags       []string `hcl:"additional_tags" cty:"additional_tags"`
}

// GetCreateMAA gets the CreateMAA variable.
// TODO(derpsteb): Rename this function once we have introduced an interface for config.Config.
func (o *OpenStackClusterVariables) GetCreateMAA() bool {
	return false
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (o *OpenStackClusterVariables) String() string {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(o, f.Body())
	return string(f.Bytes())
}

// OpenStackNodeGroup is a node group to create on OpenStack.
type OpenStackNodeGroup struct {
	// Role is the role of the node group.
	Role string `hcl:"role" cty:"role"`
	// InitialCount is the number of instances to create.
	// InitialCount is optional for upgrades. OpenStack does not support upgrades yet but might in the future.
	InitialCount int `hcl:"initial_count" cty:"initial_count"`
	// Flavor is the ID of the OpenStack flavor (machine type) to use.
	FlavorID string `hcl:"flavor_id" cty:"flavor_id"`
	// Zone is the OpenStack availability zone to use.
	Zone string `hcl:"zone" cty:"zone"`
	// StateDiskType is the OpenStack disk type to use for the state disk.
	StateDiskType string `hcl:"state_disk_type" cty:"state_disk_type"`
	// StateDiskSizeGB is the size of the state disk to allocate to each node, in GB.
	StateDiskSizeGB int `hcl:"state_disk_size" cty:"state_disk_size"`
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
	ImagePath string `hcl:"image_id" cty:"image_id"`
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
	// CustomEndpoint is the (optional) custom dns hostname for the kubernetes api server.
	CustomEndpoint string `hcl:"custom_endpoint" cty:"custom_endpoint"`
	// InternalLoadBalancer is true if an internal load balancer should be created.
	InternalLoadBalancer bool `hcl:"internal_load_balancer" cty:"internal_load_balancer"`
}

// GetCreateMAA gets the CreateMAA variable.
// TODO(derpsteb): Rename this function once we have introduced an interface for config.Config.
func (q *QEMUVariables) GetCreateMAA() bool {
	return false
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (q *QEMUVariables) String() string {
	// copy v object
	vCopy := *q
	switch vCopy.NVRAM {
	case "production":
		vCopy.NVRAM = "/usr/share/OVMF/OVMF_VARS.fd"
	case "testing":
		vCopy.NVRAM = "/usr/share/OVMF/OVMF_VARS.fd"
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
	// InitialCount is optional for upgrades.
	// Upgrades are not implemented for QEMU. The type is similar to other NodeGroup types for consistency.
	InitialCount int `hcl:"initial_count" cty:"initial_count"`
	// DiskSize is the size of the disk to allocate to each node, in GiB.
	DiskSize int `hcl:"disk_size" cty:"disk_size"`
	// CPUCount is the number of CPUs to allocate to each node.
	CPUCount int `hcl:"vcpus" cty:"vcpus"`
	// MemorySize is the amount of memory to allocate to each node, in MiB.
	MemorySize int `hcl:"memory" cty:"memory"`
}

func toPtr[T any](v T) *T {
	return &v
}
