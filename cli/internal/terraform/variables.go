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

// GCPVariables is user configuration for creating a cluster with Terraform on GCP.
type AWSVariables struct {
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
	// IAMGroupControlPlane is the IAM group to use for the control-plane nodes.
	IAMProfileControlPlane string
	// IAMGroupWorkerNodes is the IAM group to use for the worker nodes.
	IAMProfileWorkerNodes string
	// Debug is true if debug mode is enabled.
	Debug bool
}

// GCPVariables is user configuration for creating a cluster with Terraform on GCP.
type GCPVariables struct {
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

func (v *AWSVariables) String() string {
	b := &strings.Builder{}
	b.WriteString(v.CommonVariables.String())
	writeLinef(b, "region = %q", v.Region)
	writeLinef(b, "zone = %q", v.Zone)
	writeLinef(b, "ami = %q", v.AMIImageID)
	writeLinef(b, "instance_type = %q", v.InstanceType)
	writeLinef(b, "iam_instance_profile_control_plane = %q", v.IAMProfileControlPlane)
	writeLinef(b, "iam_instance_profile_worker_nodes = %q", v.IAMProfileWorkerNodes)
	writeLinef(b, "debug = %t", v.Debug)

	return b.String()
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *GCPVariables) String() string {
	b := &strings.Builder{}
	b.WriteString(v.CommonVariables.String())
	writeLinef(b, "project = %q", v.Project)
	writeLinef(b, "region = %q", v.Region)
	writeLinef(b, "zone = %q", v.Zone)
	writeLinef(b, "credentials_file = %q", v.CredentialsFile)
	writeLinef(b, "instance_type = %q", v.InstanceType)
	writeLinef(b, "state_disk_type = %q", v.StateDiskType)
	writeLinef(b, "image_id = %q", v.ImageID)
	writeLinef(b, "debug = %t", v.Debug)

	return b.String()
}

// AzureVariables is user configuration for creating a cluster with Terraform on Azure.
type AzureVariables struct {
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
	// Debug is true if debug mode is enabled.
	Debug bool
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *AzureVariables) String() string {
	b := &strings.Builder{}
	b.WriteString(v.CommonVariables.String())
	writeLinef(b, "resource_group = %q", v.ResourceGroup)
	writeLinef(b, "location = %q", v.Location)
	writeLinef(b, "user_assigned_identity = %q", v.UserAssignedIdentity)
	writeLinef(b, "instance_type = %q", v.InstanceType)
	writeLinef(b, "state_disk_type = %q", v.StateDiskType)
	writeLinef(b, "image_id = %q", v.ImageID)
	writeLinef(b, "confidential_vm = %t", v.ConfidentialVM)
	writeLinef(b, "debug = %t", v.Debug)

	return b.String()
}

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
}

// String returns a string representation of the variables, formatted as Terraform variables.
func (v *QEMUVariables) String() string {
	b := &strings.Builder{}
	b.WriteString(v.CommonVariables.String())
	writeLinef(b, "libvirt_uri = %q", v.LibvirtURI)
	writeLinef(b, "libvirt_socket_path = %q", v.LibvirtSocketPath)
	writeLinef(b, "constellation_coreos_image = %q", v.ImagePath)
	writeLinef(b, "image_format = %q", v.ImageFormat)
	writeLinef(b, "vcpus = %d", v.CPUCount)
	writeLinef(b, "memory = %d", v.MemorySizeMiB)
	writeLinef(b, "metadata_api_image = %q", v.MetadataAPIImage)
	writeLinef(b, "metadata_libvirt_uri = %q", v.MetadataLibvirtURI)

	return b.String()
}

func writeLinef(builder *strings.Builder, format string, a ...interface{}) {
	builder.WriteString(fmt.Sprintf(format, a...))
	builder.WriteByte('\n')
}
