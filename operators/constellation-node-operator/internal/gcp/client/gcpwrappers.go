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

type instanceTemplateClient struct {
	*compute.InstanceTemplatesClient
}

func (c *instanceTemplateClient) Close() error {
	return c.InstanceTemplatesClient.Close()
}

func (c *instanceTemplateClient) Delete(ctx context.Context, req *computepb.DeleteInstanceTemplateRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return c.InstanceTemplatesClient.Delete(ctx, req, opts...)
}

func (c *instanceTemplateClient) Insert(ctx context.Context, req *computepb.InsertInstanceTemplateRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return c.InstanceTemplatesClient.Insert(ctx, req, opts...)
}

type instanceGroupManagersClient struct {
	*compute.InstanceGroupManagersClient
}

func (c *instanceGroupManagersClient) Close() error {
	return c.InstanceGroupManagersClient.Close()
}

func (c *instanceGroupManagersClient) AggregatedList(ctx context.Context,
	req *computepb.AggregatedListInstanceGroupManagersRequest,
	opts ...gax.CallOption,
) InstanceGroupManagerScopedListIterator {
	return c.InstanceGroupManagersClient.AggregatedList(ctx, req, opts...)
}

type regionInstanceGroupManagersClient struct {
	*compute.RegionInstanceGroupManagersClient
}

func (c *regionInstanceGroupManagersClient) Close() error {
	return c.RegionInstanceGroupManagersClient.Close()
}

func (c *regionInstanceGroupManagersClient) Get(ctx context.Context,
	req *computepb.GetRegionInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (*computepb.InstanceGroupManager, error) {
	return c.RegionInstanceGroupManagersClient.Get(ctx, req, opts...)
}

func (c *regionInstanceGroupManagersClient) ListManagedInstances(ctx context.Context,
	req *computepb.ListManagedInstancesRegionInstanceGroupManagersRequest,
	opts ...gax.CallOption,
) ManagedInstanceIterator {
	return c.RegionInstanceGroupManagersClient.ListManagedInstances(ctx, req, opts...)
}

func (c *regionInstanceGroupManagersClient) SetInstanceTemplate(ctx context.Context,
	req *computepb.SetInstanceTemplateRegionInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return c.RegionInstanceGroupManagersClient.SetInstanceTemplate(ctx, req, opts...)
}

func (c *regionInstanceGroupManagersClient) CreateInstances(ctx context.Context,
	req *computepb.CreateInstancesRegionInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return c.RegionInstanceGroupManagersClient.CreateInstances(ctx, req, opts...)
}

func (c *regionInstanceGroupManagersClient) DeleteInstances(ctx context.Context,
	req *computepb.DeleteInstancesRegionInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return c.RegionInstanceGroupManagersClient.DeleteInstances(ctx, req, opts...)
}
