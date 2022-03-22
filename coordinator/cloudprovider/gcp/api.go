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

type metadataAPI interface {
	InstanceAttributeValue(attr string) (string, error)
	ProjectID() (string, error)
	Zone() (string, error)
	InstanceName() (string, error)
}

type Operation interface {
	Proto() *computepb.Operation
}

type InstanceIterator interface {
	Next() (*computepb.Instance, error)
}
