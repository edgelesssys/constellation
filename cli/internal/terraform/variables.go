/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

import (
	"fmt"
	"strings"
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

// AWSClusterVariables is user configuration for creating a cluster with Terraform on GCP.
type AWSClusterVariables struct {
	// CommonVariables contains common variables.
	CommonVariables
	// Region is the AWS region to use.
	Region string
	// Zone is the AWS zone to use in the given region.
	Zone string
	// AMIImageID is the ID of the AMI image to use.
	AMIImageID string
	// InstanceType is the type of the EC2 instance to use.
	InstanceType string
	// StateDiskType is the EBS disk type to use for the state disk.
	StateDiskType string
	// IAMGroupControlPlane is the IAM group to use for the control-plane nodes.
	IAMProfileControlPlane string
	// IAMGroupWorkerNodes is the IAM group to use for the worker nodes.
	IAMProfileWorkerNodes string
	// Debug is true if debug mode is enabled.
	Debug bool
}

func (v *AWSClusterVariables) String() string {
	b := &strings.Builder{}
	b.WriteString(v.CommonVariables.String())
	writeLinef(b, "region = %q", v.Region)
	writeLinef(b, "zone = %q", v.Zone)
	writeLinef(b, "ami = %q", v.AMIImageID)
	writeLinef(b, "instance_type = %q", v.InstanceType)
	writeLinef(b, "state_disk_type = %q", v.StateDiskType)
	writeLinef(b, "iam_instance_profile_control_plane = %q", v.IAMProfileControlPlane)
	writeLinef(b, "iam_instance_profile_worker_nodes = %q", v.IAMProfileWorkerNodes)
	writeLinef(b, "debug = %t", v.Debug)

	return b.String()
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
	// CommonVariables contains common variables.
	CommonVariables

	// Project is the ID of the GCP project to use.
	Project string
	// Region is the GCP region to use.
	Region string
	// Zone is the GCP zone to use.
	Zone string
	// CredentialsFile is the path to the GCP credentials file.
	CredentialsFile string
	// InstanceType is the GCP instance type to use.
	InstanceType string
	// StateDiskType is the GCP disk type to use for the state disk.
	StateDiskType string
	// ImageID is the ID of the GCP image to use.
	ImageID string
	// Debug is true if debug mode is enabled.
	Debug bool
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *GCPClusterVariables) String() string {
	b := &strings.Builder{}
	b.WriteString(v.CommonVariables.String())
	writeLinef(b, "project = %q", v.Project)
	writeLinef(b, "region = %q", v.Region)
	writeLinef(b, "zone = %q", v.Zone)
	writeLinef(b, "instance_type = %q", v.InstanceType)
	writeLinef(b, "state_disk_type = %q", v.StateDiskType)
	writeLinef(b, "image_id = %q", v.ImageID)
	writeLinef(b, "debug = %t", v.Debug)

	return b.String()
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
	// CommonVariables contains common variables.
	CommonVariables

	// ResourceGroup is the name of the Azure resource group to use.
	ResourceGroup string
	// Location is the Azure location to use.
	Location string
	// UserAssignedIdentity is the name of the Azure user-assigned identity to use.
	UserAssignedIdentity string
	// InstanceType is the Azure instance type to use.
	InstanceType string
	// StateDiskType is the Azure disk type to use for the state disk.
	StateDiskType string
	// ImageID is the ID of the Azure image to use.
	ImageID string
	// ConfidentialVM sets the VM to be confidential.
	ConfidentialVM bool
	// SecureBoot sets the VM to use secure boot.
	SecureBoot bool
	// CreateMAA sets whether a Microsoft Azure attestation provider should be created.
	CreateMAA bool
	// Debug is true if debug mode is enabled.
	Debug bool
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *AzureClusterVariables) String() string {
	b := &strings.Builder{}
	b.WriteString(v.CommonVariables.String())
	writeLinef(b, "resource_group = %q", v.ResourceGroup)
	writeLinef(b, "location = %q", v.Location)
	writeLinef(b, "user_assigned_identity = %q", v.UserAssignedIdentity)
	writeLinef(b, "instance_type = %q", v.InstanceType)
	writeLinef(b, "state_disk_type = %q", v.StateDiskType)
	writeLinef(b, "image_id = %q", v.ImageID)
	writeLinef(b, "confidential_vm = %t", v.ConfidentialVM)
	writeLinef(b, "secure_boot = %t", v.SecureBoot)
	writeLinef(b, "create_maa = %t", v.CreateMAA)
	writeLinef(b, "debug = %t", v.Debug)

	return b.String()
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
	// CommonVariables contains common variables.
	CommonVariables

	// Cloud is the (optional) name of the OpenStack cloud to use when reading the "clouds.yaml" configuration file. If empty, environment variables are used.
	Cloud string
	// AvailabilityZone is the OpenStack availability zone to use.
	AvailabilityZone string
	// Flavor is the ID of the OpenStack flavor (machine type) to use.
	FlavorID string
	// FloatingIPPoolID is the ID of the OpenStack floating IP pool to use for public IPs.
	FloatingIPPoolID string
	// StateDiskType is the OpenStack disk type to use for the state disk.
	StateDiskType string
	// ImageURL is the URL of the OpenStack image to use.
	ImageURL string
	// DirectDownload decides whether to download the image directly from the URL to OpenStack or to upload it from the local machine.
	DirectDownload bool
	// OpenstackUserDomainName is the OpenStack user domain name to use.
	OpenstackUserDomainName string
	// OpenstackUsername is the OpenStack user name to use.
	OpenstackUsername string
	// OpenstackPassword is the OpenStack password to use.
	OpenstackPassword string
	// Debug is true if debug mode is enabled.
	Debug bool
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *OpenStackClusterVariables) String() string {
	b := &strings.Builder{}
	b.WriteString(v.CommonVariables.String())
	if v.Cloud != "" {
		writeLinef(b, "cloud = %q", v.Cloud)
	}
	writeLinef(b, "availability_zone = %q", v.AvailabilityZone)
	writeLinef(b, "flavor_id = %q", v.FlavorID)
	writeLinef(b, "floating_ip_pool_id = %q", v.FloatingIPPoolID)
	writeLinef(b, "image_url = %q", v.ImageURL)
	writeLinef(b, "direct_download = %t", v.DirectDownload)
	writeLinef(b, "state_disk_type = %q", v.StateDiskType)
	writeLinef(b, "openstack_user_domain_name = %q", v.OpenstackUserDomainName)
	writeLinef(b, "openstack_username = %q", v.OpenstackUsername)
	writeLinef(b, "openstack_password = %q", v.OpenstackPassword)
	writeLinef(b, "debug = %t", v.Debug)

	return b.String()
}

// TODO: Add support for OpenStack IAM variables.

// QEMUVariables is user configuration for creating a QEMU cluster with Terraform.
type QEMUVariables struct {
	// CommonVariables contains common variables.
	CommonVariables

	// LibvirtURI is the libvirt connection URI.
	LibvirtURI string
	// LibvirtSocketPath is the path to the libvirt socket in case of unix socket.
	LibvirtSocketPath string
	// CPUCount is the number of CPUs to allocate to each node.
	CPUCount int
	// MemorySizeMiB is the amount of memory to allocate to each node, in MiB.
	MemorySizeMiB int
	// IPRangeStart is the first IP address in the IP range to allocate to the cluster.
	ImagePath string
	// ImageFormat is the format of the image from ImagePath.
	ImageFormat string
	// MetadataAPIImage is the container image to use for the metadata API.
	MetadataAPIImage string
	// MetadataLibvirtURI is the libvirt connection URI used by the metadata container.
	// In case of unix socket, this should be "qemu:///system".
	// Other wise it should be the same as LibvirtURI.
	MetadataLibvirtURI string
	// NVRAM is the path to the NVRAM template.
	NVRAM string
	// Firmware is the path to the firmware.
	Firmware string
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *QEMUVariables) String() string {
	b := &strings.Builder{}
	b.WriteString(v.CommonVariables.String())
	writeLinef(b, "libvirt_uri = %q", v.LibvirtURI)
	writeLinef(b, "libvirt_socket_path = %q", v.LibvirtSocketPath)
	writeLinef(b, "constellation_os_image = %q", v.ImagePath)
	writeLinef(b, "image_format = %q", v.ImageFormat)
	writeLinef(b, "vcpus = %d", v.CPUCount)
	writeLinef(b, "memory = %d", v.MemorySizeMiB)
	writeLinef(b, "metadata_api_image = %q", v.MetadataAPIImage)
	writeLinef(b, "metadata_libvirt_uri = %q", v.MetadataLibvirtURI)
	switch v.NVRAM {
	case "production":
		b.WriteString("nvram = \"/usr/share/OVMF/constellation_vars.production.fd\"\n")
	case "testing":
		b.WriteString("nvram = \"/usr/share/OVMF/constellation_vars.testing.fd\"\n")
	default:
		writeLinef(b, "nvram = %q", v.NVRAM)
	}
	if v.Firmware != "" {
		writeLinef(b, "firmware = %q", v.Firmware)
	}

	return b.String()
}

func writeLinef(builder *strings.Builder, format string, a ...any) {
	builder.WriteString(fmt.Sprintf(format, a...))
	builder.WriteByte('\n')
}
