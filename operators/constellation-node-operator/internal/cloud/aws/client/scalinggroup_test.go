/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	scalingtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	cspapi "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetScalingGroupImage(t *testing.T) {
	testCases := map[string]struct {
		providerID                        string
		describeAutoScalingGroupsOut      *autoscaling.DescribeAutoScalingGroupsOutput
		describeAutoScalingGroupsErr      error
		describeLaunchTemplateVersionsOut *ec2.DescribeLaunchTemplateVersionsOutput
		describeLaunchTemplateVersionsErr error
		wantImage                         string
		wantErr                           bool
	}{
		"getting scaling group image works": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoScalingGroupsOut: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []scalingtypes.AutoScalingGroup{
					{
						LaunchTemplate: &scalingtypes.LaunchTemplateSpecification{
							LaunchTemplateId: toPtr("lt-00000000000000000"),
						},
					},
				},
			},
			describeLaunchTemplateVersionsOut: &ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
					{
						LaunchTemplateData: &ec2types.ResponseLaunchTemplateData{
							ImageId: toPtr("ami-00000000000000000"),
						},
					},
				},
			},
			wantImage: "ami-00000000000000000",
		},
		"fails when describing autoscaling group fails": {
			providerID:                   "aws:///us-east-2a/i-00000000000000000",
			describeAutoScalingGroupsErr: assert.AnError,
			wantErr:                      true,
		},
		"fails when describing launch template versions fails": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoScalingGroupsOut: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []scalingtypes.AutoScalingGroup{
					{
						LaunchTemplate: &scalingtypes.LaunchTemplateSpecification{
							LaunchTemplateId: toPtr("lt-00000000000000000"),
						},
					},
				},
			},
			describeLaunchTemplateVersionsErr: assert.AnError,
			wantErr:                           true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				ec2Client: &stubEC2API{
					describeLaunchTemplateVersionsOut: tc.describeLaunchTemplateVersionsOut,
					describeLaunchTemplateVersionsErr: tc.describeLaunchTemplateVersionsErr,
				},
				scalingClient: &stubAutoscalingAPI{
					describeAutoScalingGroupsOut: []*autoscaling.DescribeAutoScalingGroupsOutput{
						tc.describeAutoScalingGroupsOut,
					},
					describeAutoScalingGroupsErr: []error{
						tc.describeAutoScalingGroupsErr,
					},
				},
			}
			scalingGroupImage, err := client.GetScalingGroupImage(context.Background(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantImage, scalingGroupImage)
		})
	}
}

func TestSetScalingGroupImage(t *testing.T) {
	testCases := map[string]struct {
		providerID                        string
		describeAutoScalingGroupsOut      *autoscaling.DescribeAutoScalingGroupsOutput
		describeAutoScalingGroupsErr      error
		describeLaunchTemplateVersionsOut *ec2.DescribeLaunchTemplateVersionsOutput
		describeLaunchTemplateVersionsErr error
		createLaunchTemplateVersionOut    *ec2.CreateLaunchTemplateVersionOutput
		createLaunchTemplateVersionErr    error
		modifyLaunchTemplateErr           error
		imageURI                          string
		wantErr                           bool
	}{
		"getting scaling group image works": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoScalingGroupsOut: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []scalingtypes.AutoScalingGroup{
					{
						LaunchTemplate: &scalingtypes.LaunchTemplateSpecification{
							LaunchTemplateId: toPtr("lt-00000000000000000"),
						},
					},
				},
			},
			describeLaunchTemplateVersionsOut: &ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
					{
						LaunchTemplateData: &ec2types.ResponseLaunchTemplateData{
							ImageId: toPtr("ami-00000000000000000"),
						},
						VersionNumber: toPtr(int64(1)),
					},
				},
			},
			createLaunchTemplateVersionOut: &ec2.CreateLaunchTemplateVersionOutput{
				LaunchTemplateVersion: &ec2types.LaunchTemplateVersion{
					VersionNumber: toPtr(int64(2)),
				},
			},
			imageURI: "ami-00000000000000000",
		},
		"fails when creating launch template version fails": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoScalingGroupsOut: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []scalingtypes.AutoScalingGroup{
					{
						LaunchTemplate: &scalingtypes.LaunchTemplateSpecification{
							LaunchTemplateId: toPtr("lt-00000000000000000"),
						},
					},
				},
			},
			describeLaunchTemplateVersionsOut: &ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
					{
						LaunchTemplateData: &ec2types.ResponseLaunchTemplateData{
							ImageId: toPtr("ami-00000000000000000"),
						},
						VersionNumber: toPtr(int64(1)),
					},
				},
			},
			imageURI:                       "ami-00000000000000000",
			createLaunchTemplateVersionErr: assert.AnError,
			wantErr:                        true,
		},
		"fails when modifying launch template fails": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoScalingGroupsOut: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []scalingtypes.AutoScalingGroup{
					{
						LaunchTemplate: &scalingtypes.LaunchTemplateSpecification{
							LaunchTemplateId: toPtr("lt-00000000000000000"),
						},
					},
				},
			},
			describeLaunchTemplateVersionsOut: &ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
					{
						LaunchTemplateData: &ec2types.ResponseLaunchTemplateData{
							ImageId: toPtr("ami-00000000000000000"),
						},
						VersionNumber: toPtr(int64(1)),
					},
				},
			},
			imageURI:                "ami-00000000000000000",
			modifyLaunchTemplateErr: assert.AnError,
			wantErr:                 true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				ec2Client: &stubEC2API{
					describeLaunchTemplateVersionsOut: tc.describeLaunchTemplateVersionsOut,
					describeLaunchTemplateVersionsErr: tc.describeLaunchTemplateVersionsErr,
					createLaunchTemplateVersionOut:    tc.createLaunchTemplateVersionOut,
					createLaunchTemplateVersionErr:    tc.createLaunchTemplateVersionErr,
					modifyLaunchTemplateErr:           tc.modifyLaunchTemplateErr,
				},
				scalingClient: &stubAutoscalingAPI{
					describeAutoScalingGroupsOut: []*autoscaling.DescribeAutoScalingGroupsOutput{
						tc.describeAutoScalingGroupsOut,
					},
					describeAutoScalingGroupsErr: []error{
						tc.describeAutoScalingGroupsErr,
					},
				},
			}
			err := client.SetScalingGroupImage(context.Background(), tc.providerID, tc.imageURI)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestListScalingGroups(t *testing.T) {
	testCases := map[string]struct {
		providerID                   string
		describeAutoScalingGroupsOut []*autoscaling.DescribeAutoScalingGroupsOutput
		describeAutoScalingGroupsErr []error
		wantGroups                   []cspapi.ScalingGroup
		wantErr                      bool
	}{
		"listing scaling groups work": {
			providerID: "aws:///us-east-2a/i-00000000000000000",
			describeAutoScalingGroupsOut: []*autoscaling.DescribeAutoScalingGroupsOutput{
				{
					AutoScalingGroups: []scalingtypes.AutoScalingGroup{
						{
							AutoScalingGroupName: toPtr("control-plane-asg"),
							Tags: []scalingtypes.TagDescription{
								{
									Key:   toPtr("constellation-role"),
									Value: toPtr("control-plane"),
								},
							},
						},
						{
							AutoScalingGroupName: toPtr("worker-asg"),
							Tags: []scalingtypes.TagDescription{
								{
									Key:   toPtr("constellation-role"),
									Value: toPtr("worker"),
								},
							},
						},
						{
							AutoScalingGroupName: toPtr("worker-asg-2"),
							Tags: []scalingtypes.TagDescription{
								{
									Key:   toPtr("constellation-role"),
									Value: toPtr("worker"),
								},
								{
									Key:   toPtr("constellation-node-group"),
									Value: toPtr("foo-group"),
								},
							},
						},
						{
							AutoScalingGroupName: toPtr("other-asg"),
						},
					},
				},
			},
			describeAutoScalingGroupsErr: []error{nil},
			wantGroups: []cspapi.ScalingGroup{
				{
					Name:                 "control-plane-asg",
					NodeGroupName:        constants.ControlPlaneDefault,
					GroupID:              "control-plane-asg",
					AutoscalingGroupName: "control-plane-asg",
					Role:                 "ControlPlane",
				},
				{
					Name:                 "worker-asg",
					NodeGroupName:        constants.WorkerDefault,
					GroupID:              "worker-asg",
					AutoscalingGroupName: "worker-asg",
					Role:                 "Worker",
				},
				{
					Name:                 "worker-asg-2",
					NodeGroupName:        "foo-group",
					GroupID:              "worker-asg-2",
					AutoscalingGroupName: "worker-asg-2",
					Role:                 "Worker",
				},
			},
		},
		"fails when describing scaling groups fails": {
			providerID:                   "aws:///us-east-2a/i-00000000000000000",
			describeAutoScalingGroupsOut: []*autoscaling.DescribeAutoScalingGroupsOutput{nil},
			describeAutoScalingGroupsErr: []error{assert.AnError},
			wantErr:                      true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				scalingClient: &stubAutoscalingAPI{
					describeAutoScalingGroupsOut: tc.describeAutoScalingGroupsOut,
					describeAutoScalingGroupsErr: tc.describeAutoScalingGroupsErr,
				},
			}
			gotGroups, err := client.ListScalingGroups(context.Background(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantGroups, gotGroups)
		})
	}
}
