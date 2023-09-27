/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	rgtTypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"

	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/stretchr/testify/assert"
)

func TestSelf(t *testing.T) {
	testCases := map[string]struct {
		imds     *stubIMDS
		ec2API   *stubEC2
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
			},
			ec2API: &stubEC2{
				selfInstance: &ec2.DescribeInstancesOutput{
					Reservations: []ec2Types.Reservation{
						{
							Instances: []ec2Types.Instance{
								{
									InstanceId: aws.String("test-instance-id"),
									Tags: []ec2Types.Tag{
										{
											Key:   aws.String(cloud.TagRole),
											Value: aws.String("controlplane"),
										},
									},
								},
							},
						},
					},
				},
			},
			wantSelf: metadata.InstanceMetadata{
				Name:       "test-instance-id",
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
			},
			ec2API: &stubEC2{
				selfInstance: &ec2.DescribeInstancesOutput{
					Reservations: []ec2Types.Reservation{
						{
							Instances: []ec2Types.Instance{
								{
									InstanceId: aws.String("test-instance-id"),
									Tags: []ec2Types.Tag{
										{
											Key:   aws.String(cloud.TagRole),
											Value: aws.String("worker"),
										},
										{
											Key:   aws.String(cloud.TagInitSecretHash),
											Value: aws.String("initSecretHash"),
										},
									},
								},
							},
						},
					},
				},
			},
			wantSelf: metadata.InstanceMetadata{
				Name:       "test-instance-id",
				ProviderID: "aws:///test-zone/test-instance-id",
				Role:       role.Worker,
				VPCIP:      "192.0.2.1",
			},
		},
		"get instance document error": {
			imds: &stubIMDS{
				getInstanceIdentityDocumentErr: assert.AnError,
			},
			ec2API: &stubEC2{
				selfInstance: &ec2.DescribeInstancesOutput{
					Reservations: []ec2Types.Reservation{
						{
							Instances: []ec2Types.Instance{
								{
									InstanceId: aws.String("test-instance-id"),
									Tags: []ec2Types.Tag{
										{
											Key:   aws.String(cloud.TagRole),
											Value: aws.String("controlplane"),
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"get instance error": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID:       "test-instance-id",
						AvailabilityZone: "test-zone",
						PrivateIP:        "192.0.2.1",
					},
				},
			},
			ec2API: &stubEC2{
				describeInstancesErr: assert.AnError,
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
			},
			ec2API: &stubEC2{
				selfInstance: &ec2.DescribeInstancesOutput{
					Reservations: []ec2Types.Reservation{
						{
							Instances: []ec2Types.Instance{
								{
									InstanceId: aws.String("test-instance-id"),
									Tags:       []ec2Types.Tag{},
								},
							},
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
			m := &Cloud{
				imds: tc.imds,
				ec2:  tc.ec2API,
			}

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
		Reservations: []ec2Types.Reservation{
			{
				Instances: []ec2Types.Instance{
					{
						State:            &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameRunning},
						InstanceId:       aws.String("id-1"),
						PrivateIpAddress: aws.String("192.0.2.1"),
						Placement: &ec2Types.Placement{
							AvailabilityZone: aws.String("test-zone"),
						},
						Tags: []ec2Types.Tag{
							{
								Key:   aws.String(cloud.TagRole),
								Value: aws.String("controlplane"),
							},
							{
								Key:   aws.String(cloud.TagUID),
								Value: aws.String("uid"),
							},
						},
					},
					{
						State:            &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameRunning},
						InstanceId:       aws.String("id-2"),
						PrivateIpAddress: aws.String("192.0.2.2"),
						Placement: &ec2Types.Placement{
							AvailabilityZone: aws.String("test-zone"),
						},
						Tags: []ec2Types.Tag{
							{
								Key:   aws.String(cloud.TagRole),
								Value: aws.String("worker"),
							},
							{
								Key:   aws.String(cloud.TagUID),
								Value: aws.String("uid"),
							},
						},
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		imdsAPI  *stubIMDS
		ec2      *stubEC2
		wantList []metadata.InstanceMetadata
		wantErr  bool
	}{
		"success single page": {
			imdsAPI: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID: "id-1",
					},
				},
			},
			ec2: &stubEC2{
				selfInstance: &ec2.DescribeInstancesOutput{
					Reservations: []ec2Types.Reservation{
						{
							Instances: []ec2Types.Instance{
								{
									InstanceId: aws.String("id-1"),
									Tags: []ec2Types.Tag{
										{
											Key:   aws.String(cloud.TagRole),
											Value: aws.String("controlplane"),
										},
										{
											Key:   aws.String(cloud.TagUID),
											Value: aws.String("uid"),
										},
									},
								},
							},
						},
					},
				},
				describeInstancesResp1: successfulResp,
			},
			wantList: []metadata.InstanceMetadata{
				{
					Name:       "id-1",
					Role:       role.ControlPlane,
					ProviderID: "aws:///test-zone/id-1",
					VPCIP:      "192.0.2.1",
				},
				{
					Name:       "id-2",
					Role:       role.Worker,
					ProviderID: "aws:///test-zone/id-2",
					VPCIP:      "192.0.2.2",
				},
			},
		},
		"success multiple pages": {
			imdsAPI: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID: "id-1",
					},
				},
			},
			ec2: &stubEC2{
				selfInstance: &ec2.DescribeInstancesOutput{
					Reservations: []ec2Types.Reservation{
						{
							Instances: []ec2Types.Instance{
								{
									InstanceId: aws.String("id-1"),
									Tags: []ec2Types.Tag{
										{
											Key:   aws.String(cloud.TagRole),
											Value: aws.String("controlplane"),
										},
										{
											Key:   aws.String(cloud.TagUID),
											Value: aws.String("uid"),
										},
									},
								},
							},
						},
					},
				},
				describeInstancesResp1: &ec2.DescribeInstancesOutput{
					Reservations: []ec2Types.Reservation{
						{
							Instances: []ec2Types.Instance{
								{
									State:            &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameRunning},
									InstanceId:       aws.String("id-3"),
									PrivateIpAddress: aws.String("192.0.2.3"),
									Placement: &ec2Types.Placement{
										AvailabilityZone: aws.String("test-zone-2"),
									},
									Tags: []ec2Types.Tag{
										{
											Key:   aws.String(cloud.TagRole),
											Value: aws.String("worker"),
										},
										{
											Key:   aws.String(cloud.TagUID),
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
					Name:       "id-3",
					Role:       role.Worker,
					ProviderID: "aws:///test-zone-2/id-3",
					VPCIP:      "192.0.2.3",
				},
				{
					Name:       "id-1",
					Role:       role.ControlPlane,
					ProviderID: "aws:///test-zone/id-1",
					VPCIP:      "192.0.2.1",
				},
				{
					Name:       "id-2",
					Role:       role.Worker,
					ProviderID: "aws:///test-zone/id-2",
					VPCIP:      "192.0.2.2",
				},
			},
		},
		"fail to get UID": {
			imdsAPI: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID: "id-1",
					},
				},
			},
			ec2: &stubEC2{
				selfInstance: &ec2.DescribeInstancesOutput{
					Reservations: []ec2Types.Reservation{
						{
							Instances: []ec2Types.Instance{
								{
									InstanceId: aws.String("id-1"),
									Tags: []ec2Types.Tag{
										{
											Key:   aws.String(cloud.TagRole),
											Value: aws.String("controlplane"),
										},
									},
								},
							},
						},
					},
				},
				describeInstancesResp1: successfulResp,
			},
			wantErr: true,
		},
		"describe instances fails": {
			imdsAPI: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						InstanceID: "id-1",
					},
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
			m := &Cloud{
				imds: tc.imdsAPI,
				ec2:  tc.ec2,
			}

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

func TestGetLoadBalancerEndpoint(t *testing.T) {
	lbAddr := "192.0.2.1"
	successfulEC2 := &stubEC2{
		selfInstance: &ec2.DescribeInstancesOutput{
			Reservations: []ec2Types.Reservation{
				{
					Instances: []ec2Types.Instance{
						{
							InstanceId: aws.String("id-1"),
							Tags: []ec2Types.Tag{
								{
									Key:   aws.String(cloud.TagRole),
									Value: aws.String("controlplane"),
								},
								{
									Key:   aws.String(cloud.TagUID),
									Value: aws.String("uid"),
								},
							},
						},
					},
				},
			},
		},
	}

	testCases := map[string]struct {
		imds         *stubIMDS
		loadbalancer *stubLoadbalancer
		resourceapi  *stubResourceGroupTagging
		wantHost     string
		wantErr      bool
	}{
		"success": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						AvailabilityZone: "test-zone",
					},
				},
			},
			loadbalancer: &stubLoadbalancer{
				describeLoadBalancersOut: &elasticloadbalancingv2.DescribeLoadBalancersOutput{
					LoadBalancers: []elbTypes.LoadBalancer{
						{
							LoadBalancerName: aws.String("test-lb"),
							AvailabilityZones: []elbTypes.AvailabilityZone{
								{
									ZoneName: aws.String("test-zone"),
								},
							},
							DNSName: aws.String(lbAddr),
						},
					},
				},
			},
			resourceapi: &stubResourceGroupTagging{
				getResourcesOut1: &resourcegroupstaggingapi.GetResourcesOutput{
					PaginationToken: aws.String("next-token"),
				},
				getResourcesOut2: &resourcegroupstaggingapi.GetResourcesOutput{
					ResourceTagMappingList: []rgtTypes.ResourceTagMapping{
						{
							ResourceARN: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test-lb/1234567890abcdef"),
							Tags: []rgtTypes.Tag{
								{
									Key:   aws.String(cloud.TagUID),
									Value: aws.String("uid"),
								},
							},
						},
					},
				},
			},
			wantHost: lbAddr,
		},
		"no load balancer found": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						AvailabilityZone: "test-zone",
					},
				},
			},
			loadbalancer: &stubLoadbalancer{
				describeLoadBalancersOut: &elasticloadbalancingv2.DescribeLoadBalancersOutput{
					LoadBalancers: []elbTypes.LoadBalancer{},
				},
			},
			resourceapi: &stubResourceGroupTagging{
				getResourcesOut1: &resourcegroupstaggingapi.GetResourcesOutput{
					PaginationToken: aws.String("next-token"),
				},
				getResourcesOut2: &resourcegroupstaggingapi.GetResourcesOutput{
					ResourceTagMappingList: []rgtTypes.ResourceTagMapping{
						{
							ResourceARN: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test-lb/1234567890abcdef"),
							Tags: []rgtTypes.Tag{
								{
									Key:   aws.String(cloud.TagUID),
									Value: aws.String("uid"),
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"no load balancer DNS name": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						AvailabilityZone: "test-zone",
					},
				},
			},
			loadbalancer: &stubLoadbalancer{
				describeLoadBalancersOut: &elasticloadbalancingv2.DescribeLoadBalancersOutput{
					LoadBalancers: []elbTypes.LoadBalancer{
						{
							LoadBalancerName: aws.String("test-lb"),
							AvailabilityZones: []elbTypes.AvailabilityZone{
								{
									ZoneName: aws.String("test-zone"),
								},
							},
						},
					},
				},
			},
			resourceapi: &stubResourceGroupTagging{
				getResourcesOut1: &resourcegroupstaggingapi.GetResourcesOutput{
					PaginationToken: aws.String("next-token"),
				},
				getResourcesOut2: &resourcegroupstaggingapi.GetResourcesOutput{
					ResourceTagMappingList: []rgtTypes.ResourceTagMapping{
						{
							ResourceARN: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test-lb/1234567890abcdef"),
							Tags: []rgtTypes.Tag{
								{
									Key:   aws.String(cloud.TagUID),
									Value: aws.String("uid"),
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"describe load balancers fails": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{
						AvailabilityZone: "test-zone",
					},
				},
			},
			resourceapi: &stubResourceGroupTagging{
				getResourcesOut1: &resourcegroupstaggingapi.GetResourcesOutput{
					PaginationToken: aws.String("next-token"),
				},
				getResourcesOut2: &resourcegroupstaggingapi.GetResourcesOutput{
					ResourceTagMappingList: []rgtTypes.ResourceTagMapping{
						{
							ResourceARN: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test-lb/1234567890abcdef"),
							Tags: []rgtTypes.Tag{
								{
									Key:   aws.String(cloud.TagUID),
									Value: aws.String("uid"),
								},
							},
						},
					},
				},
			},
			loadbalancer: &stubLoadbalancer{
				describeLoadBalancersErr: assert.AnError,
			},
			wantErr: true,
		},
		"get resources fails": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{},
				},
			},
			loadbalancer: &stubLoadbalancer{
				describeLoadBalancersOut: &elasticloadbalancingv2.DescribeLoadBalancersOutput{
					LoadBalancers: []elbTypes.LoadBalancer{
						{
							LoadBalancerName: aws.String("test-lb"),
							AvailabilityZones: []elbTypes.AvailabilityZone{
								{
									ZoneName: aws.String("test-zone"),
								},
							},
							DNSName: aws.String(lbAddr),
						},
					},
				},
			},
			resourceapi: &stubResourceGroupTagging{
				getResourcesErr: assert.AnError,
			},
			wantErr: true,
		},
		"no resources found": {
			imds: &stubIMDS{
				instanceDocumentResp: &imds.GetInstanceIdentityDocumentOutput{
					InstanceIdentityDocument: imds.InstanceIdentityDocument{},
				},
			},
			loadbalancer: &stubLoadbalancer{
				describeLoadBalancersOut: &elasticloadbalancingv2.DescribeLoadBalancersOutput{
					LoadBalancers: []elbTypes.LoadBalancer{
						{
							LoadBalancerName: aws.String("test-lb"),
							AvailabilityZones: []elbTypes.AvailabilityZone{
								{
									ZoneName: aws.String("test-zone"),
								},
							},
							DNSName: aws.String(lbAddr),
						},
					},
				},
			},
			resourceapi: &stubResourceGroupTagging{
				getResourcesOut1: &resourcegroupstaggingapi.GetResourcesOutput{
					PaginationToken: aws.String("next-token"),
				},
				getResourcesOut2: &resourcegroupstaggingapi.GetResourcesOutput{
					ResourceTagMappingList: []rgtTypes.ResourceTagMapping{},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			m := &Cloud{
				imds:              tc.imds,
				loadbalancer:      tc.loadbalancer,
				resourceapiClient: tc.resourceapi,
				ec2:               successfulEC2,
			}

			gotHost, gotPort, err := m.GetLoadBalancerEndpoint(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.wantHost, gotHost)
			assert.Equal("6443", gotPort)
		})
	}
}

func TestConvertToMetadataInstance(t *testing.T) {
	testCases := map[string]struct {
		in            []ec2Types.Instance
		wantInstances []metadata.InstanceMetadata
		wantErr       bool
	}{
		"success": {
			in: []ec2Types.Instance{
				{
					State:            &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameRunning},
					InstanceId:       aws.String("id-1"),
					PrivateIpAddress: aws.String("192.0.2.1"),
					Placement: &ec2Types.Placement{
						AvailabilityZone: aws.String("test-zone"),
					},
					Tags: []ec2Types.Tag{
						{
							Key:   aws.String(cloud.TagRole),
							Value: aws.String("controlplane"),
						},
					},
				},
			},
			wantInstances: []metadata.InstanceMetadata{
				{
					Name:       "id-1",
					Role:       role.ControlPlane,
					ProviderID: "aws:///test-zone/id-1",
					VPCIP:      "192.0.2.1",
				},
			},
		},
		"fallback to instance ID": {
			in: []ec2Types.Instance{
				{
					State:            &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameRunning},
					InstanceId:       aws.String("id-1"),
					PrivateIpAddress: aws.String("192.0.2.1"),
					Tags: []ec2Types.Tag{
						{
							Key:   aws.String(cloud.TagRole),
							Value: aws.String("controlplane"),
						},
					},
				},
			},

			wantInstances: []metadata.InstanceMetadata{
				{
					Name:       "id-1",
					Role:       role.ControlPlane,
					ProviderID: "aws:///id-1",
					VPCIP:      "192.0.2.1",
				},
			},
		},
		"non running instances are ignored": {
			in: []ec2Types.Instance{
				{
					State: &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameStopped},
				},
				{
					State: &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameTerminated},
				},
			},
		},
		"no instance ID": {
			in: []ec2Types.Instance{
				{
					State:            &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameRunning},
					PrivateIpAddress: aws.String("192.0.2.1"),
					Placement: &ec2Types.Placement{
						AvailabilityZone: aws.String("test-zone"),
					},
					Tags: []ec2Types.Tag{
						{
							Key:   aws.String(cloud.TagRole),
							Value: aws.String("controlplane"),
						},
					},
				},
			},
			wantErr: true,
		},
		"no private IP": {
			in: []ec2Types.Instance{
				{
					State:      &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameRunning},
					InstanceId: aws.String("id-1"),
					Placement: &ec2Types.Placement{
						AvailabilityZone: aws.String("test-zone"),
					},
					Tags: []ec2Types.Tag{
						{
							Key:   aws.String(cloud.TagRole),
							Value: aws.String("controlplane"),
						},
					},
				},
			},
			wantErr: true,
		},
		"missing role tag": {
			in: []ec2Types.Instance{
				{
					State:            &ec2Types.InstanceState{Name: ec2Types.InstanceStateNameRunning},
					InstanceId:       aws.String("id-1"),
					PrivateIpAddress: aws.String("192.0.2.1"),
					Placement: &ec2Types.Placement{
						AvailabilityZone: aws.String("test-zone"),
					},
					Tags: []ec2Types.Tag{},
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			m := &Cloud{}

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
	instanceDocumentResp           *imds.GetInstanceIdentityDocumentOutput
	getInstanceIdentityDocumentErr error
}

func (s *stubIMDS) GetInstanceIdentityDocument(context.Context, *imds.GetInstanceIdentityDocumentInput, ...func(*imds.Options)) (*imds.GetInstanceIdentityDocumentOutput, error) {
	return s.instanceDocumentResp, s.getInstanceIdentityDocumentErr
}

type stubEC2 struct {
	describeInstancesErr   error
	selfInstance           *ec2.DescribeInstancesOutput
	describeInstancesResp1 *ec2.DescribeInstancesOutput
	describeInstancesResp2 *ec2.DescribeInstancesOutput
	describeAddressesErr   error
	describeAddressesResp  *ec2.DescribeAddressesOutput
}

func (s *stubEC2) DescribeInstances(_ context.Context, in *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if len(in.InstanceIds) == 1 {
		return s.selfInstance, s.describeInstancesErr
	}
	if in.NextToken == nil {
		return s.describeInstancesResp1, s.describeInstancesErr
	}
	return s.describeInstancesResp2, s.describeInstancesErr
}

func (s *stubEC2) DescribeAddresses(context.Context, *ec2.DescribeAddressesInput, ...func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error) {
	return s.describeAddressesResp, s.describeAddressesErr
}

type stubLoadbalancer struct {
	describeLoadBalancersErr error
	describeLoadBalancersOut *elasticloadbalancingv2.DescribeLoadBalancersOutput
}

func (s *stubLoadbalancer) DescribeLoadBalancers(_ context.Context,
	_ *elasticloadbalancingv2.DescribeLoadBalancersInput,
	_ ...func(*elasticloadbalancingv2.Options)) (
	*elasticloadbalancingv2.DescribeLoadBalancersOutput, error,
) {
	return s.describeLoadBalancersOut, s.describeLoadBalancersErr
}

type stubResourceGroupTagging struct {
	getResourcesErr  error
	getResourcesOut1 *resourcegroupstaggingapi.GetResourcesOutput
	getResourcesOut2 *resourcegroupstaggingapi.GetResourcesOutput
}

func (s *stubResourceGroupTagging) GetResources(_ context.Context,
	in *resourcegroupstaggingapi.GetResourcesInput,
	_ ...func(*resourcegroupstaggingapi.Options)) (
	*resourcegroupstaggingapi.GetResourcesOutput, error,
) {
	if in.PaginationToken == nil {
		return s.getResourcesOut1, s.getResourcesErr
	}
	return s.getResourcesOut2, s.getResourcesErr
}
