/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"

	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
)

const (
	controlPlanesID = "control-planes-id"
	workersID       = "workers-id"
)

// Client is a stub client providing the minimal implementation to set up the initial resources.
type Client struct{}

// GetNodeImage retrieves the image currently used by a node.
func (c *Client) GetNodeImage(ctx context.Context, providerID string) (string, error) {
	panic("not implemented")
}

// GetScalingGroupID retrieves the scaling group that a node is part of.
func (c *Client) GetScalingGroupID(ctx context.Context, providerID string) (string, error) {
	panic("not implemented")
}

// CreateNode creates a new node inside a specified scaling group at the CSP and returns its future name and provider id.
func (c *Client) CreateNode(ctx context.Context, scalingGroupID string) (nodeName, providerID string, err error) {
	panic("not implemented")
}

// DeleteNode starts the termination of the node at the CSP.
func (c *Client) DeleteNode(ctx context.Context, providerID string) error {
	panic("not implemented")
}

// GetNodeState retrieves the state of a pending node from a CSP.
func (c *Client) GetNodeState(ctx context.Context, providerID string) (updatev1alpha1.CSPNodeState, error) {
	panic("not implemented")
}

// SetScalingGroupImage sets the image to be used by newly created nodes in a scaling group.
func (c *Client) SetScalingGroupImage(ctx context.Context, scalingGroupID, image string) error {
	panic("not implemented")
}

// GetScalingGroupImage retrieves the image currently used by a scaling group.
func (c *Client) GetScalingGroupImage(ctx context.Context, scalingGroupID string) (string, error) {
	return "unsupportedCSP", nil
}

// GetScalingGroupName retrieves the name of a scaling group.
func (c *Client) GetScalingGroupName(scalingGroupID string) (string, error) {
	switch scalingGroupID {
	case controlPlanesID:
		return controlPlanesID, nil
	case workersID:
		return workersID, nil
	default:
		return "", fmt.Errorf("unknown scaling group id %s", scalingGroupID)
	}
}

// GetAutoscalingGroupName retrieves the name of a scaling group as needed by the cluster-autoscaler.
func (c *Client) GetAutoscalingGroupName(scalingGroupID string) (string, error) {
	switch scalingGroupID {
	case controlPlanesID:
		return controlPlanesID, nil
	case workersID:
		return workersID, nil
	default:
		return "", fmt.Errorf("unknown scaling group id %s", scalingGroupID)
	}
}

// ListScalingGroups retrieves a list of scaling groups for the cluster.
func (c *Client) ListScalingGroups(ctx context.Context, uid string) (controlPlaneGroupIDs []string, workerGroupIDs []string, err error) {
	return []string{controlPlanesID}, []string{workersID}, nil
}

// AutoscalingCloudProvider returns the cloud-provider name as used by k8s cluster-autoscaler.
func (c *Client) AutoscalingCloudProvider() string {
	return "unsupportedCSP"
}
