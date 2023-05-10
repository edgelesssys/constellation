/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodeState(t *testing.T) {
	testCases := map[string]struct {
		providerID                string
		describeInstanceStatusOut *ec2.DescribeInstanceStatusOutput
		describeInstanceStatusErr error
		wantState                 updatev1alpha1.CSPNodeState
		wantErr                   bool
	}{
		"getting node state works for running VM": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusOut: &ec2.DescribeInstanceStatusOutput{
				InstanceStatuses: []ec2types.InstanceStatus{
					{
						InstanceState: &ec2types.InstanceState{
							Name: ec2types.InstanceStateNameRunning,
						},
					},
				},
			},
			wantState: updatev1alpha1.NodeStateReady,
		},
		"getting node state works for terminated VM": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusOut: &ec2.DescribeInstanceStatusOutput{
				InstanceStatuses: []ec2types.InstanceStatus{
					{
						InstanceState: &ec2types.InstanceState{
							Name: ec2types.InstanceStateNameTerminated,
						},
					},
				},
			},
			wantState: updatev1alpha1.NodeStateTerminated,
		},
		"getting node state works for stopping VM": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusOut: &ec2.DescribeInstanceStatusOutput{
				InstanceStatuses: []ec2types.InstanceStatus{
					{
						InstanceState: &ec2types.InstanceState{
							Name: ec2types.InstanceStateNameStopping,
						},
					},
				},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"getting node state works for stopped VM": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusOut: &ec2.DescribeInstanceStatusOutput{
				InstanceStatuses: []ec2types.InstanceStatus{
					{
						InstanceState: &ec2types.InstanceState{
							Name: ec2types.InstanceStateNameStopped,
						},
					},
				},
			},
			wantState: updatev1alpha1.NodeStateStopped,
		},
		"getting node state works for pending VM": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusOut: &ec2.DescribeInstanceStatusOutput{
				InstanceStatuses: []ec2types.InstanceStatus{
					{
						InstanceState: &ec2types.InstanceState{
							Name: ec2types.InstanceStateNamePending,
						},
					},
				},
			},
			wantState: updatev1alpha1.NodeStateCreating,
		},
		"getting node state works for shutting-down VM": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusOut: &ec2.DescribeInstanceStatusOutput{
				InstanceStatuses: []ec2types.InstanceStatus{
					{
						InstanceState: &ec2types.InstanceState{
							Name: ec2types.InstanceStateNameShuttingDown,
						},
					},
				},
			},
			wantState: updatev1alpha1.NodeStateTerminating,
		},
		"getting node state fails when the state is unknown": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusOut: &ec2.DescribeInstanceStatusOutput{
				InstanceStatuses: []ec2types.InstanceStatus{
					{
						InstanceState: &ec2types.InstanceState{
							Name: "unknown",
						},
					},
				},
			},
			wantState: updatev1alpha1.NodeStateUnknown,
			wantErr:   true,
		},
		"cannot find instance": {
			providerID:                "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusErr: errors.New("InvalidInstanceID.NotFound"),
			wantState:                 updatev1alpha1.NodeStateTerminated,
		},
		"unknown error when describing the instance error": {
			providerID:                "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusErr: assert.AnError,
			wantState:                 updatev1alpha1.NodeStateUnknown,
			wantErr:                   true,
		},
		"fails when getting no instances": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusOut: &ec2.DescribeInstanceStatusOutput{
				InstanceStatuses: []ec2types.InstanceStatus{},
			},
			wantState: updatev1alpha1.NodeStateUnknown,
			wantErr:   true,
		},
		"fails when the instance state is nil": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstanceStatusOut: &ec2.DescribeInstanceStatusOutput{
				InstanceStatuses: []ec2types.InstanceStatus{
					{
						InstanceState: nil,
					},
				},
			},
			wantState: updatev1alpha1.NodeStateUnknown,
			wantErr:   true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				ec2Client: &stubEC2API{
					describeInstanceStatusOut: tc.describeInstanceStatusOut,
					describeInstanceStatusErr: tc.describeInstanceStatusErr,
				},
			}
			nodeState, err := client.GetNodeState(context.Background(), tc.providerID)
			assert.Equal(tc.wantState, nodeState)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}
