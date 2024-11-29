/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	computeREST "google.golang.org/api/compute/v1"
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
	Get(projectID, template string) (*computeREST.InstanceTemplate, error)
	Delete(projectID, template string) (*computeREST.Operation, error)
	Insert(projectID string, template *computeREST.InstanceTemplate) (*computeREST.Operation, error)
}

type instanceGroupManagersAPI interface {
	Close() error
	Get(ctx context.Context, req *computepb.GetInstanceGroupManagerRequest,
		opts ...gax.CallOption) (*computepb.InstanceGroupManager, error)
	AggregatedList(ctx context.Context, req *computepb.AggregatedListInstanceGroupManagersRequest,
		opts ...gax.CallOption) InstanceGroupManagerScopedListIterator
	SetInstanceTemplate(ctx context.Context, req *computepb.SetInstanceTemplateInstanceGroupManagerRequest,
		opts ...gax.CallOption) (Operation, error)
	CreateInstances(ctx context.Context, req *computepb.CreateInstancesInstanceGroupManagerRequest,
		opts ...gax.CallOption) (Operation, error)
	DeleteInstances(ctx context.Context, req *computepb.DeleteInstancesInstanceGroupManagerRequest,
		opts ...gax.CallOption) (Operation, error)
}

type diskAPI interface {
	Close() error
	Get(ctx context.Context, req *computepb.GetDiskRequest,
		opts ...gax.CallOption) (*computepb.Disk, error)
}

// Operation describes a generic protobuf operation that can be waited for.
type Operation interface {
	Proto() *computepb.Operation
	Done() bool
	Wait(ctx context.Context, opts ...gax.CallOption) error
}

// InstanceGroupManagerScopedListIterator can list the Next InstanceGroupManagersScopedListPair.
type InstanceGroupManagerScopedListIterator interface {
	Next() (compute.InstanceGroupManagersScopedListPair, error)
}

type prng interface {
	// Intn returns, as an int, a non-negative pseudo-random number in the half-open interval [0,n). It panics if n <= 0.
	Intn(n int) int
}
