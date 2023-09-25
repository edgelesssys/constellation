/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package metadata

import (
	"context"

	"github.com/edgelesssys/constellation/v2/internal/role"
)

// InstanceMetadata describes metadata of a peer.
type InstanceMetadata struct {
	Name       string
	ProviderID string
	Role       role.Role
	// VPCIP is the primary IP address of the instance in the VPC.
	VPCIP string

	// SecondaryIPRange is the VPC wide CIDR from which subnets are attached to VMs as AliasIPRanges.
	// May be empty on certain CSPs.
	SecondaryIPRange string
	// AliasIPRanges is a list of IP ranges that are attached.
	// May be empty on certain CSPs.
	AliasIPRanges []string
}

// InstanceSelfer provide instance metadata about themselves.
type InstanceSelfer interface {
	// Self retrieves the current instance.
	Self(ctx context.Context) (InstanceMetadata, error)
}

// InstanceLister list information about instance metadata.
type InstanceLister interface {
	// List retrieves all instances belonging to the current constellation.
	List(ctx context.Context) ([]InstanceMetadata, error)
}
