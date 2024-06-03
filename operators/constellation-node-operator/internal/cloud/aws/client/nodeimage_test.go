/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	autoscalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodeImage(t *testing.T) {
	ami := "ami-00000000000000000"
	testCases := map[string]struct {
		providerID           string
		describeInstancesErr error
		describeInstancesOut *ec2.DescribeInstancesOutput
		wantImage            string
		wantErr              bool
	}{
		"getting node image works": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstancesOut: &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{
							{
								ImageId: &ami,
							},
						},
					},
				},
			},
			wantImage: ami,
		},
		"no reservations": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstancesOut: &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{},
			},
			wantErr: true,
		},
		"no instances": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstancesOut: &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{},
					},
				},
			},
			wantErr: true,
		},
		"no image": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstancesOut: &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{
							{},
						},
					},
				},
			},
			wantErr: true,
		},
		"error describing instances": {
			providerID:           "aws:///us-east-2a/i-00000000000000000",
			describeInstancesErr: assert.AnError,
			wantErr:              true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				ec2Client: &stubEC2API{
					describeInstancesOut: tc.describeInstancesOut,
					describeInstancesErr: tc.describeInstancesErr,
				},
			}
			gotImage, err := client.GetNodeImage(context.Background(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantImage, gotImage)
		})
	}
}

func TestGetScalingGroupID(t *testing.T) {
	asgName := "my-asg"
	testCases := map[string]struct {
		providerID           string
		describeInstancesErr error
		describeInstancesOut *ec2.DescribeInstancesOutput
		wantASGID            string
		wantErr              bool
	}{
		"getting node's tag works": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstancesOut: &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{
							{
								Tags: []ec2types.Tag{
									{
										Key:   toPtr("aws:autoscaling:groupName"),
										Value: &asgName,
									},
								},
							},
						},
					},
				},
			},
			wantASGID: asgName,
		},
		"no valid tags": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstancesOut: &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{
							{
								Tags: []ec2types.Tag{
									{
										Key:   toPtr("foo"),
										Value: toPtr("bar"),
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"no reservations": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstancesOut: &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{},
			},
			wantErr: true,
		},
		"no instances": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstancesOut: &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{},
					},
				},
			},
			wantErr: true,
		},
		"no image": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeInstancesOut: &ec2.DescribeInstancesOutput{
				Reservations: []ec2types.Reservation{
					{
						Instances: []ec2types.Instance{
							{},
						},
					},
				},
			},
			wantErr: true,
		},
		"error describing instances": {
			providerID:           "aws:///us-east-2a/i-00000000000000000",
			describeInstancesErr: assert.AnError,
			wantErr:              true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				ec2Client: &stubEC2API{
					describeInstancesOut: tc.describeInstancesOut,
					describeInstancesErr: tc.describeInstancesErr,
				},
			}
			gotScalingID, err := client.GetScalingGroupID(context.Background(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantASGID, gotScalingID)
		})
	}
}

func TestCreateNode(t *testing.T) {
	testCases := map[string]struct {
		providerID                   string
		describeAutoscalingOutFirst  *autoscaling.DescribeAutoScalingGroupsOutput
		describeAutoscalingFirstErr  error
		describeAutoscalingOutSecond *autoscaling.DescribeAutoScalingGroupsOutput
		describeAutoscalingSecondErr error
		setDesiredCapacityErr        error
		wantNodeName                 string
		wantProviderID               string
		wantErr                      bool
	}{
		"creating a new node works": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoscalingOutFirst: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []autoscalingtypes.AutoScalingGroup{
					{
						AutoScalingGroupName: toPtr("my-asg"),
						Instances: []autoscalingtypes.Instance{
							{
								InstanceId: toPtr("i-00000000000000000"),
							},
						},
						DesiredCapacity: toPtr(int32(1)),
					},
				},
			},
			describeAutoscalingOutSecond: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []autoscalingtypes.AutoScalingGroup{
					{
						AutoScalingGroupName: toPtr("my-asg"),
						Instances: []autoscalingtypes.Instance{
							{
								InstanceId: toPtr("i-00000000000000000"),
							},
							{
								InstanceId:       toPtr("i-00000000000000001"),
								AvailabilityZone: toPtr("us-east-2a"),
							},
						},
						DesiredCapacity: toPtr(int32(2)),
					},
				},
			},
			wantNodeName:   "i-00000000000000001",
			wantProviderID: "aws:///us-east-2a/i-00000000000000001",
		},
		"creating a new node fails when describing the auto scaling group the first time": {
			providerID:                  "aws:///us-east-2a/i-00000000000000000",
			describeAutoscalingFirstErr: assert.AnError,
			wantErr:                     true,
		},
		"creating a new node fails when describing the auto scaling group the second time": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoscalingOutFirst: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []autoscalingtypes.AutoScalingGroup{
					{
						AutoScalingGroupName: toPtr("my-asg"),
						Instances: []autoscalingtypes.Instance{
							{
								InstanceId: toPtr("i-00000000000000000"),
							},
						},
						DesiredCapacity: toPtr(int32(1)),
					},
				},
			},
			describeAutoscalingSecondErr: assert.AnError,
			wantErr:                      true,
		},
		"creating a new node fails when the auto scaling group is not found": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoscalingOutFirst: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []autoscalingtypes.AutoScalingGroup{},
			},
			wantErr: true,
		},
		"creating a new node fails when set desired capacity fails": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoscalingOutFirst: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []autoscalingtypes.AutoScalingGroup{
					{
						AutoScalingGroupName: toPtr("my-asg"),
						Instances: []autoscalingtypes.Instance{
							{
								InstanceId: toPtr("i-00000000000000000"),
							},
						},
						DesiredCapacity: toPtr(int32(1)),
					},
				},
			},
			setDesiredCapacityErr: assert.AnError,
			wantErr:               true,
		},
		"creating a new node fails when the found vm does not contain an availability zone": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoscalingOutFirst: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []autoscalingtypes.AutoScalingGroup{
					{
						AutoScalingGroupName: toPtr("my-asg"),
						Instances: []autoscalingtypes.Instance{
							{
								InstanceId: toPtr("i-00000000000000000"),
							},
						},
						DesiredCapacity: toPtr(int32(1)),
					},
				},
			},
			describeAutoscalingOutSecond: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []autoscalingtypes.AutoScalingGroup{
					{
						AutoScalingGroupName: toPtr("my-asg"),
						Instances: []autoscalingtypes.Instance{
							{
								InstanceId: toPtr("i-00000000000000000"),
							},
							{
								InstanceId: toPtr("i-00000000000000001"),
							},
						},
						DesiredCapacity: toPtr(int32(2)),
					},
				},
			},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				scalingClient: &stubAutoscalingAPI{
					describeAutoScalingGroupsOut: []*autoscaling.DescribeAutoScalingGroupsOutput{
						tc.describeAutoscalingOutFirst,
						tc.describeAutoscalingOutSecond,
					},
					describeAutoScalingGroupsErr: []error{
						tc.describeAutoscalingFirstErr,
						tc.describeAutoscalingSecondErr,
					},
					setDesiredCapacityErr: tc.setDesiredCapacityErr,
				},
			}
			nodeName, providerID, err := client.CreateNode(context.Background(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantNodeName, nodeName)
			assert.Equal(tc.wantProviderID, providerID)
		})
	}
}

func TestDeleteNode(t *testing.T) {
	testCases := map[string]struct {
		providerID           string
		terminateInstanceErr error
		wantErr              bool
	}{
		"deleting node works": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
		},
		"deleting node fails when terminating the instance fails": {
			providerID:           "aws:///us-east-2a/i-00000000000000000",
			terminateInstanceErr: assert.AnError,
			wantErr:              true,
		},
		"deleting node succeeds when the instance does not exist": {
			providerID:           "aws:///us-east-2a/i-00000000000000000",
			terminateInstanceErr: fmt.Errorf("Instance Id not found - No managed instance found for instance ID: i-00000000000000000"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				scalingClient: &stubAutoscalingAPI{
					terminateInstanceErr: tc.terminateInstanceErr,
				},
			}
			err := client.DeleteNode(context.Background(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

type stubEC2API struct {
	describeInstancesOut              *ec2.DescribeInstancesOutput
	describeInstancesErr              error
	describeInstanceStatusOut         *ec2.DescribeInstanceStatusOutput
	describeInstanceStatusErr         error
	describeLaunchTemplateVersionsOut *ec2.DescribeLaunchTemplateVersionsOutput
	describeLaunchTemplateVersionsErr error
	createLaunchTemplateVersionOut    *ec2.CreateLaunchTemplateVersionOutput
	createLaunchTemplateVersionErr    error
	modifyLaunchTemplateErr           error
}

func (a *stubEC2API) DescribeInstances(_ context.Context, _ *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return a.describeInstancesOut, a.describeInstancesErr
}

func (a *stubEC2API) DescribeInstanceStatus(_ context.Context, _ *ec2.DescribeInstanceStatusInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstanceStatusOutput, error) {
	return a.describeInstanceStatusOut, a.describeInstanceStatusErr
}

func (a *stubEC2API) CreateLaunchTemplateVersion(_ context.Context, _ *ec2.CreateLaunchTemplateVersionInput, _ ...func(*ec2.Options)) (*ec2.CreateLaunchTemplateVersionOutput, error) {
	return a.createLaunchTemplateVersionOut, a.createLaunchTemplateVersionErr
}

func (a *stubEC2API) ModifyLaunchTemplate(_ context.Context, _ *ec2.ModifyLaunchTemplateInput, _ ...func(*ec2.Options)) (*ec2.ModifyLaunchTemplateOutput, error) {
	return nil, a.modifyLaunchTemplateErr
}

func (a *stubEC2API) DescribeLaunchTemplateVersions(_ context.Context, _ *ec2.DescribeLaunchTemplateVersionsInput, _ ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	return a.describeLaunchTemplateVersionsOut, a.describeLaunchTemplateVersionsErr
}

type stubAutoscalingAPI struct {
	describeAutoScalingGroupsOut []*autoscaling.DescribeAutoScalingGroupsOutput
	describeAutoScalingGroupsErr []error
	describeCounter              int
	setDesiredCapacityErr        error
	terminateInstanceErr         error
}

func (a *stubAutoscalingAPI) DescribeAutoScalingGroups(_ context.Context, _ *autoscaling.DescribeAutoScalingGroupsInput, _ ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	out := a.describeAutoScalingGroupsOut[a.describeCounter]
	err := a.describeAutoScalingGroupsErr[a.describeCounter]
	a.describeCounter++
	return out, err
}

func (a *stubAutoscalingAPI) SetDesiredCapacity(_ context.Context, _ *autoscaling.SetDesiredCapacityInput, _ ...func(*autoscaling.Options)) (*autoscaling.SetDesiredCapacityOutput, error) {
	return nil, a.setDesiredCapacityErr
}

func (a *stubAutoscalingAPI) TerminateInstanceInAutoScalingGroup(_ context.Context, _ *autoscaling.TerminateInstanceInAutoScalingGroupInput, _ ...func(*autoscaling.Options)) (*autoscaling.TerminateInstanceInAutoScalingGroupOutput, error) {
	return nil, a.terminateInstanceErr
}
