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

type forwardingRuleIterator interface {
	Next() (*computepb.ForwardingRule, error)
}

type instanceIterator interface {
	Next() (*computepb.Instance, error)
}

type forwardingRulesClient struct {
	*compute.GlobalForwardingRulesClient
}

func (c *forwardingRulesClient) Close() error {
	return c.GlobalForwardingRulesClient.Close()
}

func (c *forwardingRulesClient) List(ctx context.Context, req *computepb.ListGlobalForwardingRulesRequest,
	opts ...gax.CallOption,
) forwardingRuleIterator {
	return c.GlobalForwardingRulesClient.List(ctx, req)
}

type instanceClient struct {
	*compute.InstancesClient
}

func (c *instanceClient) Close() error {
	return c.InstancesClient.Close()
}

func (c *instanceClient) List(ctx context.Context, req *computepb.ListInstancesRequest,
	opts ...gax.CallOption,
) instanceIterator {
	return c.InstancesClient.List(ctx, req)
}
