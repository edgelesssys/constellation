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

func (c *instanceGroupManagersClient) Get(ctx context.Context, req *computepb.GetInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (*computepb.InstanceGroupManager, error) {
	return c.InstanceGroupManagersClient.Get(ctx, req, opts...)
}

func (c *instanceGroupManagersClient) AggregatedList(ctx context.Context, req *computepb.AggregatedListInstanceGroupManagersRequest,
	opts ...gax.CallOption,
) InstanceGroupManagerScopedListIterator {
	return c.InstanceGroupManagersClient.AggregatedList(ctx, req, opts...)
}

func (c *instanceGroupManagersClient) SetInstanceTemplate(ctx context.Context, req *computepb.SetInstanceTemplateInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return c.InstanceGroupManagersClient.SetInstanceTemplate(ctx, req, opts...)
}

func (c *instanceGroupManagersClient) CreateInstances(ctx context.Context, req *computepb.CreateInstancesInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return c.InstanceGroupManagersClient.CreateInstances(ctx, req, opts...)
}

func (c *instanceGroupManagersClient) DeleteInstances(ctx context.Context, req *computepb.DeleteInstancesInstanceGroupManagerRequest,
	opts ...gax.CallOption,
) (Operation, error) {
	return c.InstanceGroupManagersClient.DeleteInstances(ctx, req, opts...)
}
