/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
)

// GetNodeState returns the state of the node.
func (c *Client) GetNodeState(ctx context.Context, providerID string) (updatev1alpha1.CSPNodeState, error) {
	_, resourceGroup, scaleSet, instanceID, err := scaleSetInformationFromProviderID(providerID)
	if err != nil {
		return updatev1alpha1.NodeStateUnknown, err
	}
	instanceView, err := c.virtualMachineScaleSetVMsAPI.GetInstanceView(ctx, resourceGroup, scaleSet, instanceID, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == http.StatusNotFound {
			return updatev1alpha1.NodeStateTerminated, nil
		}
		return updatev1alpha1.NodeStateUnknown, err
	}
	return nodeStateFromStatuses(instanceView.Statuses), nil
}
