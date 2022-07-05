package client

import (
	"context"

	"github.com/googleapis/gax-go/v2"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

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
	Get(ctx context.Context, req *computepb.GetInstanceGroupManagerRequest,
		opts ...gax.CallOption) (*computepb.InstanceGroupManager, error)
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

type Operation interface {
	Proto() *computepb.Operation
	Done() bool
	Wait(ctx context.Context, opts ...gax.CallOption) error
}

type prng interface {
	// Intn returns, as an int, a non-negative pseudo-random number in the half-open interval [0,n). It panics if n <= 0.
	Intn(n int) int
}
