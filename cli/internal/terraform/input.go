/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package terraform

// CreateClusterInput is user configuration for creating a cluster with Terraform.
type CreateClusterInput struct {
	// CountControlPlanes is the number of control-plane nodes to create.
	CountControlPlanes int
	// CountWorkers is the number of worker nodes to create.
	CountWorkers int
	// QEMU is the configuration for QEMU clusters.
	QEMU QEMUInput
}

// QEMUInput is user configuration for creating a QEMU cluster with Terraform.
type QEMUInput struct {
	// CPUCount is the number of CPUs to allocate to each node.
	CPUCount int
	// MemorySizeMiB is the amount of memory to allocate to each node, in MiB.
	MemorySizeMiB int
	// StateDiskSizeGB is the size of the state disk to allocate to each node, in GB.
	StateDiskSizeGB int
	// IPRangeStart is the first IP address in the IP range to allocate to the cluster.
	IPRangeStart int
	// ImagePath is the path to the image to use for the nodes.
	ImagePath string
	// ImageFormat is the format of the image from ImagePath.
	ImageFormat string
	// MetadataAPIImage is the container image to use for the metadata API.
	MetadataAPIImage string
}
