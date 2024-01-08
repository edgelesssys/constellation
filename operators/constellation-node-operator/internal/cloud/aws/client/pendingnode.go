/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// GetNodeState returns the state of the node.
func (c *Client) GetNodeState(ctx context.Context, providerID string) (updatev1alpha1.CSPNodeState, error) {
	logr := log.FromContext(ctx)
	logr.Info("GetNodeState", "providerID", providerID)
	instanceName, err := getInstanceNameFromProviderID(providerID)
	if err != nil {
		return updatev1alpha1.NodeStateUnknown, fmt.Errorf("failed to get instance name from providerID: %w", err)
	}

	statusOut, err := c.ec2Client.DescribeInstanceStatus(ctx, &ec2.DescribeInstanceStatusInput{
		InstanceIds:         []string{instanceName},
		IncludeAllInstances: toPtr(true),
	})
	if err != nil {
		if strings.Contains(err.Error(), "InvalidInstanceID.NotFound") {
			return updatev1alpha1.NodeStateTerminated, nil
		}
		return updatev1alpha1.NodeStateUnknown, err
	}

	if len(statusOut.InstanceStatuses) != 1 {
		return updatev1alpha1.NodeStateUnknown, fmt.Errorf("expected 1 instance status, got %d", len(statusOut.InstanceStatuses))
	}

	if statusOut.InstanceStatuses[0].InstanceState == nil {
		return updatev1alpha1.NodeStateUnknown, errors.New("instance state is nil")
	}

	// Translate AWS instance state to node state.
	switch statusOut.InstanceStatuses[0].InstanceState.Name {
	case ec2types.InstanceStateNameRunning:
		return updatev1alpha1.NodeStateReady, nil
	case ec2types.InstanceStateNameTerminated:
		return updatev1alpha1.NodeStateTerminated, nil
	case ec2types.InstanceStateNameShuttingDown:
		return updatev1alpha1.NodeStateTerminating, nil
	case ec2types.InstanceStateNameStopped:
		return updatev1alpha1.NodeStateStopped, nil
	// For "Stopping" we can only know the next state in the state machine
	// so we preemptively set it to "Stopped".
	case ec2types.InstanceStateNameStopping:
		return updatev1alpha1.NodeStateStopped, nil
	case ec2types.InstanceStateNamePending:
		return updatev1alpha1.NodeStateCreating, nil
	default:
		return updatev1alpha1.NodeStateUnknown, fmt.Errorf("unknown instance state %q", statusOut.InstanceStatuses[0].InstanceState.Name)
	}
}
