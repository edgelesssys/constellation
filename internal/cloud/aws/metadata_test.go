/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
)

func TestSelf(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		imds     *stubIMDS
		ec2      *stubEC2
		wantSelf metadata.InstanceMetadata
		wantErr  bool
	}{
		"success control-plane": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID:       "test-instance-id",
						AvailabilityZone: "test-zone",
						PrivateIP:        "192.0.2.1",
					},
				},
				tags: map[string]string{
					tagName: "test-instance",
					tagRole: "controlplane",
				},
			},
			wantSelf: metadata.InstanceMetadata{
				Name:       "test-instance",
				ProviderID: "aws:///test-zone/test-instance-id",
				Role:       role.ControlPlane,
				VPCIP:      "192.0.2.1",
			},
		},
		"success worker": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID:       "test-instance-id",
						AvailabilityZone: "test-zone",
						PrivateIP:        "192.0.2.1",
					},
				},
				tags: map[string]string{
					tagName: "test-instance",
					tagRole: "worker",
				},
			},
			wantSelf: metadata.InstanceMetadata{
				Name:       "test-instance",
				ProviderID: "aws:///test-zone/test-instance-id",
				Role:       role.Worker,
				VPCIP:      "192.0.2.1",
			},
		},
		"get instance document error": {
			imds: &stubIMDS{
				getInstanceIdentityDocumentErr: someErr,
				tags: map[string]string{
					tagName: "test-instance",
					tagRole: "controlplane",
				},
			},
			wantErr: true,
		},
		"get metadata error": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID:       "test-instance-id",
						AvailabilityZone: "test-zone",
						PrivateIP:        "192.0.2.1",
					},
				},
				getMetadataErr: someErr,
			},
			wantErr: true,
		},
		"name not set": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID:       "test-instance-id",
						AvailabilityZone: "test-zone",
						PrivateIP:        "192.0.2.1",
					},
				},
				tags: map[string]string{
					tagRole: "controlplane",
				},
			},
			wantErr: true,
		},
		"role not set": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID:       "test-instance-id",
						AvailabilityZone: "test-zone",
						PrivateIP:        "192.0.2.1",
					},
				},
				tags: map[string]string{
					tagName: "test-instance",
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			m := &Metadata{imds: tc.imds, ec2: &stubEC2{}}

			self, err := m.Self(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.wantSelf, self)
		})
	}
}

func TestList(t *testing.T) {
	someErr := errors.New("failed")

	successfulResp := &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
						InstanceId:       aws.String("id-1"),
						PrivateIpAddress: aws.String("192.0.2.1"),
						Placement: &types.Placement{
							AvailabilityZone: aws.String("test-zone"),
						},
						Tags: []types.Tag{
							{
								Key:   aws.String(tagName),
								Value: aws.String("name-1"),
							},
							{
								Key:   aws.String(tagRole),
								Value: aws.String("controlplane"),
							},
							{
								Key:   aws.String(tagUID),
								Value: aws.String("uid"),
							},
						},
					},
					{
						State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
						InstanceId:       aws.String("id-2"),
						PrivateIpAddress: aws.String("192.0.2.2"),
						Placement: &types.Placement{
							AvailabilityZone: aws.String("test-zone"),
						},
						Tags: []types.Tag{
							{
								Key:   aws.String(tagName),
								Value: aws.String("name-2"),
							},
							{
								Key:   aws.String(tagRole),
								Value: aws.String("worker"),
							},
							{
								Key:   aws.String(tagUID),
								Value: aws.String("uid"),
							},
						},
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		imds     *stubIMDS
		ec2      *stubEC2
		wantList []metadata.InstanceMetadata
		wantErr  bool
	}{
		"success single page": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagUID: "uid",
				},
			},
			ec2: &stubEC2{
				describeInstancesResp1: successfulResp,
			},
			wantList: []metadata.InstanceMetadata{
				{
					Name:       "name-1",
					Role:       role.ControlPlane,
					ProviderID: "aws:///test-zone/id-1",
					VPCIP:      "192.0.2.1",
				},
				{
					Name:       "name-2",
					Role:       role.Worker,
					ProviderID: "aws:///test-zone/id-2",
					VPCIP:      "192.0.2.2",
				},
			},
		},
		"success multiple pages": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagUID: "uid",
				},
			},
			ec2: &stubEC2{
				describeInstancesResp1: &ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{
									State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
									InstanceId:       aws.String("id-3"),
									PrivateIpAddress: aws.String("192.0.2.3"),
									Placement: &types.Placement{
										AvailabilityZone: aws.String("test-zone-2"),
									},
									Tags: []types.Tag{
										{
											Key:   aws.String(tagName),
											Value: aws.String("name-3"),
										},
										{
											Key:   aws.String(tagRole),
											Value: aws.String("worker"),
										},
										{
											Key:   aws.String(tagUID),
											Value: aws.String("uid"),
										},
									},
								},
							},
						},
					},
					NextToken: aws.String("next-token"),
				},
				describeInstancesResp2: successfulResp,
			},
			wantList: []metadata.InstanceMetadata{
				{
					Name:       "name-3",
					Role:       role.Worker,
					ProviderID: "aws:///test-zone-2/id-3",
					VPCIP:      "192.0.2.3",
				},
				{
					Name:       "name-1",
					Role:       role.ControlPlane,
					ProviderID: "aws:///test-zone/id-1",
					VPCIP:      "192.0.2.1",
				},
				{
					Name:       "name-2",
					Role:       role.Worker,
					ProviderID: "aws:///test-zone/id-2",
					VPCIP:      "192.0.2.2",
				},
			},
		},
		"fail to get UID": {
			imds: &stubIMDS{},
			ec2: &stubEC2{
				describeInstancesResp1: successfulResp,
			},
			wantErr: true,
		},
		"describe instances fails": {
			imds: &stubIMDS{
				tags: map[string]string{
					tagUID: "uid",
				},
			},
			ec2: &stubEC2{
				describeInstancesErr: someErr,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			m := &Metadata{ec2: tc.ec2, imds: tc.imds}

			list, err := m.List(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.wantList, list)
		})
	}
}

func TestConvertToMetadataInstance(t *testing.T) {
	testCases := map[string]struct {
		in            []types.Instance
		wantInstances []metadata.InstanceMetadata
		wantErr       bool
	}{
		"success": {
			in: []types.Instance{
				{
					State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
					InstanceId:       aws.String("id-1"),
					PrivateIpAddress: aws.String("192.0.2.1"),
					Placement: &types.Placement{
						AvailabilityZone: aws.String("test-zone"),
					},
					Tags: []types.Tag{
						{
							Key:   aws.String(tagName),
							Value: aws.String("name-1"),
						},
						{
							Key:   aws.String(tagRole),
							Value: aws.String("controlplane"),
						},
					},
				},
			},
			wantInstances: []metadata.InstanceMetadata{
				{
					Name:       "name-1",
					Role:       role.ControlPlane,
					ProviderID: "aws:///test-zone/id-1",
					VPCIP:      "192.0.2.1",
				},
			},
		},
		"fallback to instance ID": {
			in: []types.Instance{
				{
					State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
					InstanceId:       aws.String("id-1"),
					PrivateIpAddress: aws.String("192.0.2.1"),
					Tags: []types.Tag{
						{
							Key:   aws.String(tagName),
							Value: aws.String("name-1"),
						},
						{
							Key:   aws.String(tagRole),
							Value: aws.String("controlplane"),
						},
					},
				},
			},

			wantInstances: []metadata.InstanceMetadata{
				{
					Name:       "name-1",
					Role:       role.ControlPlane,
					ProviderID: "aws:///id-1",
					VPCIP:      "192.0.2.1",
				},
			},
		},
		"non running instances are ignored": {
			in: []types.Instance{
				{
					State: &types.InstanceState{Name: types.InstanceStateNameStopped},
				},
				{
					State: &types.InstanceState{Name: types.InstanceStateNameTerminated},
				},
			},
		},
		"no instance ID": {
			in: []types.Instance{
				{
					State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
					PrivateIpAddress: aws.String("192.0.2.1"),
					Placement: &types.Placement{
						AvailabilityZone: aws.String("test-zone"),
					},
					Tags: []types.Tag{
						{
							Key:   aws.String(tagName),
							Value: aws.String("name-1"),
						},
						{
							Key:   aws.String(tagRole),
							Value: aws.String("controlplane"),
						},
					},
				},
			},
			wantErr: true,
		},
		"no private IP": {
			in: []types.Instance{
				{
					State:      &types.InstanceState{Name: types.InstanceStateNameRunning},
					InstanceId: aws.String("id-1"),
					Placement: &types.Placement{
						AvailabilityZone: aws.String("test-zone"),
					},
					Tags: []types.Tag{
						{
							Key:   aws.String(tagName),
							Value: aws.String("name-1"),
						},
						{
							Key:   aws.String(tagRole),
							Value: aws.String("controlplane"),
						},
					},
				},
			},
			wantErr: true,
		},
		"missing name tag": {
			in: []types.Instance{
				{
					State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
					InstanceId:       aws.String("id-1"),
					PrivateIpAddress: aws.String("192.0.2.1"),
					Placement: &types.Placement{
						AvailabilityZone: aws.String("test-zone"),
					},
					Tags: []types.Tag{
						{
							Key:   aws.String(tagRole),
							Value: aws.String("controlplane"),
						},
					},
				},
			},
			wantErr: true,
		},
		"missing role tag": {
			in: []types.Instance{
				{
					State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
					InstanceId:       aws.String("id-1"),
					PrivateIpAddress: aws.String("192.0.2.1"),
					Placement: &types.Placement{
						AvailabilityZone: aws.String("test-zone"),
					},
					Tags: []types.Tag{
						{
							Key:   aws.String(tagName),
							Value: aws.String("name-1"),
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			m := &Metadata{}

			instances, err := m.convertToMetadataInstance(tc.in)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.wantInstances, instances)
		})
	}
}

type stubIMDS struct {
	getInstanceIdentityDocumentErr error
	getMetadataErr                 error
	instanceDocumentResp           *imds.GetInstanceIdentityDocumentOutput
	tags                           map[string]string
}

func (s *stubIMDS) GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error) {
	return s.instanceDocumentResp, s.getInstanceIdentityDocumentErr
}

func (s *stubIMDS) GetMetadata(_ context.Context, in *imds.GetMetadataInput, _ ...func(*imds.Options)) (*imds.GetMetadataOutput, error) {
	tag, ok := s.tags[strings.TrimPrefix(in.Path, "/tags/instance/")]
	if !ok {
		return nil, errors.New("not found")
	}
	return &imds.GetMetadataOutput{
		Content: io.NopCloser(
			strings.NewReader(
				tag,
			),
		),
	}, s.getMetadataErr
}

type stubEC2 struct {
	describeInstancesErr   error
	describeInstancesResp1 *ec2.DescribeInstancesOutput
	describeInstancesResp2 *ec2.DescribeInstancesOutput
}

func (s *stubEC2) DescribeInstances(_ context.Context, in *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if in.NextToken == nil {
		return s.describeInstancesResp1, s.describeInstancesErr
	}
	return s.describeInstancesResp2, s.describeInstancesErr
}
