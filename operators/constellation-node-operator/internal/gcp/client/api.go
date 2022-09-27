/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/googleapis/gax-go/v2"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type projectAPI interface {
	Close() error
	Get(ctx context.Context, req *computepb.GetProjectRequest,
		opts ...gax.CallOption) (*computepb.Project, error)
}

type instanceAPI interface {
	Close() error
	Get(ctx context.Context, req *computepb.GetInstanceRequest,
		opts ...gax.CallOption) (*computepb.Instance, error)
}

type instanceTemplateAPI interface {
	Close() error
	Get(ctx context.Context, req *computepb.GetInstanceTemplateRequest,
		opts ...gax.CallOption) (*computepb.InstanceTemplate, error)
	Delete(ctx context.Context, req *computepb.DeleteInstanceTemplateRequest,
		opts ...gax.CallOption) (Operation, error)
	Insert(ctx context.Context, req *computepb.InsertInstanceTemplateRequest,
		opts ...gax.CallOption) (Operation, error)
}

type instanceGroupManagersAPI interface {
	Close() error
	AggregatedList(ctx context.Context, req *computepb.AggregatedListInstanceGroupManagersRequest,
		opts ...gax.CallOption) InstanceGroupManagerScopedListIterator
}

type regionInstanceGroupManagersAPI interface {
	Close() error
	Get(ctx context.Context, req *computepb.GetRegionInstanceGroupManagerRequest,
		opts ...gax.CallOption) (*computepb.InstanceGroupManager, error)
	ListManagedInstances(ctx context.Context,
		req *computepb.ListManagedInstancesRegionInstanceGroupManagersRequest,
		opts ...gax.CallOption) ManagedInstanceIterator
	SetInstanceTemplate(ctx context.Context, req *computepb.SetInstanceTemplateRegionInstanceGroupManagerRequest,
		opts ...gax.CallOption) (Operation, error)
	CreateInstances(ctx context.Context, req *computepb.CreateInstancesRegionInstanceGroupManagerRequest,
		opts ...gax.CallOption) (Operation, error)
	DeleteInstances(ctx context.Context, req *computepb.DeleteInstancesRegionInstanceGroupManagerRequest,
		opts ...gax.CallOption) (Operation, error)
}

type diskAPI interface {
	Close() error
	Get(ctx context.Context, req *computepb.GetDiskRequest,
		opts ...gax.CallOption) (*computepb.Disk, error)
}

type Operation interface {
	Proto() *computepb.Operation
	Done() bool
	Wait(ctx context.Context, opts ...gax.CallOption) error
}

type InstanceGroupManagerScopedListIterator interface {
	Next() (compute.InstanceGroupManagersScopedListPair, error)
}

type ManagedInstanceIterator interface {
	Next() (*computepb.ManagedInstance, error)
}

type InstanceGroupIterator interface {
	Next() (*computepb.InstanceGroup, error)
}

type prng interface {
	// Intn returns, as an int, a non-negative pseudo-random number in the half-open interval [0,n). It panics if n <= 0.
	Intn(n int) int
}
