/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"

	"github.com/googleapis/gax-go/v2"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type instanceAPI interface {
	Get(ctx context.Context, req *computepb.GetInstanceRequest, opts ...gax.CallOption) (*computepb.Instance, error)
	List(ctx context.Context, req *computepb.ListInstancesRequest, opts ...gax.CallOption) InstanceIterator
	SetMetadata(ctx context.Context, req *computepb.SetMetadataInstanceRequest, opts ...gax.CallOption) (*compute.Operation, error)
	Close() error
}

type subnetworkAPI interface {
	List(ctx context.Context, req *computepb.ListSubnetworksRequest, opts ...gax.CallOption) SubnetworkIterator
	Get(ctx context.Context, req *computepb.GetSubnetworkRequest, opts ...gax.CallOption) (*computepb.Subnetwork, error)
	Close() error
}

type forwardingRulesAPI interface {
	List(ctx context.Context, req *computepb.ListGlobalForwardingRulesRequest, opts ...gax.CallOption) ForwardingRuleIterator
	Close() error
}

type metadataAPI interface {
	InstanceAttributeValue(attr string) (string, error)
	InstanceID() (string, error)
	ProjectID() (string, error)
	Zone() (string, error)
	InstanceName() (string, error)
}

// Operation represents a GCP Operation resource, which is one of: global, regional
// or zonal.
type Operation interface {
	Proto() *computepb.Operation
}

// InstanceIterator iterates over GCP VM instances.
type InstanceIterator interface {
	Next() (*computepb.Instance, error)
}

// SubnetworkIterator iterates over GCP subnetworks.
type SubnetworkIterator interface {
	Next() (*computepb.Subnetwork, error)
}

// ForwardingRuleIterator iterates over GCP forwards rules. Those rules forward
// traffic to a different target when receiving traffic.
type ForwardingRuleIterator interface {
	Next() (*computepb.ForwardingRule, error)
}
