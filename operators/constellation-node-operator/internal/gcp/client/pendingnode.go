package client

import (
	"context"
	"errors"
	"net/http"

	"github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
	updatev1alpha1 "github.com/edgelesssys/constellation/operators/constellation-node-operator/api/v1alpha1"
	"google.golang.org/api/googleapi"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
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
				return v1alpha1.NodeStateTerminated, nil
			}
		}
		return "", err
	}

	if instance.Status == nil {
		return v1alpha1.NodeStateUnknown, nil
	}

	// reference: https://cloud.google.com/compute/docs/instances/instance-life-cycle
	switch *instance.Status {
	case computepb.Instance_PROVISIONING.String():
		fallthrough
	case computepb.Instance_STAGING.String():
		return v1alpha1.NodeStateCreating, nil
	case computepb.Instance_RUNNING.String():
		return v1alpha1.NodeStateReady, nil
	case computepb.Instance_STOPPING.String():
		fallthrough
	case computepb.Instance_SUSPENDING.String():
		fallthrough
	case computepb.Instance_SUSPENDED.String():
		fallthrough
	case computepb.Instance_REPAIRING.String():
		fallthrough
	case computepb.Instance_TERMINATED.String(): // this is stopped in GCP terms
		return v1alpha1.NodeStateStopped, nil
	}
	return v1alpha1.NodeStateUnknown, nil
}
