/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"net/http"

	"cloud.google.com/go/compute/apiv1/computepb"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	"google.golang.org/api/googleapi"
)

// GetNodeState returns the state of the node.
func (c *Client) GetNodeState(ctx context.Context, providerID string) (updatev1alpha1.CSPNodeState, error) {
	project, zone, instanceName, err := splitProviderID(providerID)
	if err != nil {
		return "", err
	}
	instance, err := c.instanceAPI.Get(ctx, &computepb.GetInstanceRequest{
		Instance: instanceName,
		Project:  project,
		Zone:     zone,
	})
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) {
			if apiErr.Code == http.StatusNotFound {
				return updatev1alpha1.NodeStateTerminated, nil
			}
		}
		return "", err
	}

	if instance.Status == nil {
		return updatev1alpha1.NodeStateUnknown, nil
	}

	// reference: https://cloud.google.com/compute/docs/instances/instance-life-cycle
	switch *instance.Status {
	case computepb.Instance_PROVISIONING.String():
		fallthrough
	case computepb.Instance_STAGING.String():
		return updatev1alpha1.NodeStateCreating, nil
	case computepb.Instance_RUNNING.String():
		return updatev1alpha1.NodeStateReady, nil
	case computepb.Instance_STOPPING.String():
		fallthrough
	case computepb.Instance_SUSPENDING.String():
		fallthrough
	case computepb.Instance_SUSPENDED.String():
		fallthrough
	case computepb.Instance_REPAIRING.String():
		fallthrough
	case computepb.Instance_TERMINATED.String(): // this is stopped in GCP terms
		return updatev1alpha1.NodeStateStopped, nil
	}
	return updatev1alpha1.NodeStateUnknown, nil
}
