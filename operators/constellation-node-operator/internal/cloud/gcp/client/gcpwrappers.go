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

type instanceTemplateClient struct {
	*computeREST.InstanceTemplatesService
}

func (c *instanceTemplateClient) Close() error {
	return nil // no-op
}

func (c *instanceTemplateClient) Get(project, template string) (*computeREST.InstanceTemplate, error) {
	return c.InstanceTemplatesService.Get(project, template).Do()
}

func (c *instanceTemplateClient) Delete(project, template string) (*computeREST.Operation, error) {
	return c.InstanceTemplatesService.Delete(project, template).Do()
}

func (c *instanceTemplateClient) Insert(projectID string, template *computeREST.InstanceTemplate) (*computeREST.Operation, error) {
	return c.InstanceTemplatesService.Insert(projectID, template).Do()
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
