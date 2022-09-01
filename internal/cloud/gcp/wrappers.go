/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/metadata"
	"github.com/googleapis/gax-go/v2"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type instanceClient struct {
	*compute.InstancesClient
}

func (c *instanceClient) Close() error {
	return c.InstancesClient.Close()
}

func (c *instanceClient) List(ctx context.Context, req *computepb.ListInstancesRequest,
	opts ...gax.CallOption,
) InstanceIterator {
	return c.InstancesClient.List(ctx, req)
}

type subnetworkClient struct {
	*compute.SubnetworksClient
}

func (c *subnetworkClient) Close() error {
	return c.SubnetworksClient.Close()
}

func (c *subnetworkClient) List(ctx context.Context, req *computepb.ListSubnetworksRequest,
	opts ...gax.CallOption,
) SubnetworkIterator {
	return c.SubnetworksClient.List(ctx, req)
}

func (c *subnetworkClient) Get(ctx context.Context, req *computepb.GetSubnetworkRequest,
	opts ...gax.CallOption,
) (*computepb.Subnetwork, error) {
	return c.SubnetworksClient.Get(ctx, req)
}

type forwardingRulesClient struct {
	*compute.GlobalForwardingRulesClient
}

func (c *forwardingRulesClient) Close() error {
	return c.GlobalForwardingRulesClient.Close()
}

func (c *forwardingRulesClient) List(ctx context.Context, req *computepb.ListGlobalForwardingRulesRequest,
	opts ...gax.CallOption,
) ForwardingRuleIterator {
	return c.GlobalForwardingRulesClient.List(ctx, req)
}

type metadataClient struct{}

func (c *metadataClient) InstanceAttributeValue(attr string) (string, error) {
	return metadata.InstanceAttributeValue(attr)
}

func (c *metadataClient) ProjectID() (string, error) {
	return metadata.ProjectID()
}

func (c *metadataClient) Zone() (string, error) {
	return metadata.Zone()
}

func (c *metadataClient) InstanceName() (string, error) {
	return metadata.InstanceName()
}

func (c *metadataClient) ProjectAttributeValue(attr string) (string, error) {
	return metadata.ProjectAttributeValue(attr)
}
