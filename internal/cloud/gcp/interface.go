/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
)

type globalForwardingRulesAPI interface {
	List(ctx context.Context, req *computepb.ListGlobalForwardingRulesRequest, opts ...gax.CallOption) forwardingRuleIterator
	Close() error
}

type regionalForwardingRulesAPI interface {
	List(ctx context.Context, req *computepb.ListForwardingRulesRequest, opts ...gax.CallOption) forwardingRuleIterator
	Close() error
}

type imdsAPI interface {
	InstanceID() (string, error)
	ProjectID() (string, error)
	Zone() (string, error)
	InstanceName() (string, error)
}

type instanceAPI interface {
	Get(ctx context.Context, req *computepb.GetInstanceRequest, opts ...gax.CallOption) (*computepb.Instance, error)
	List(ctx context.Context, req *computepb.ListInstancesRequest, opts ...gax.CallOption) instanceIterator
	Close() error
}

type subnetAPI interface {
	Get(ctx context.Context, req *computepb.GetSubnetworkRequest, opts ...gax.CallOption) (*computepb.Subnetwork, error)
	Close() error
}
