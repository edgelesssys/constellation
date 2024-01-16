/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package gcp

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/role"
	gax "github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		goleak.IgnoreAnyFunction("github.com/bazelbuild/rules_go/go/tools/bzltestutil.RegisterTimeoutHandler.func1"),
	)
}

func TestGetInstance(t *testing.T) {
	someErr := errors.New("failed")
	goodInstance := &computepb.Instance{
		Name: proto.String("someInstance"),
		Zone: proto.String("someZone-west3-b"),
		Labels: map[string]string{
			cloud.TagUID:            "1234",
			cloud.TagRole:           role.ControlPlane.String(),
			cloud.TagInitSecretHash: "initSecretHash",
		},
		Metadata: &computepb.Metadata{
			Items: []*computepb.Items{
				{
					Key:   proto.String(cloud.TagInitSecretHash),
					Value: proto.String("initSecretHash"),
				},
			},
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
				Name:      proto.String("nic0"),
				NetworkIP: proto.String("192.0.2.0"),
				AliasIpRanges: []*computepb.AliasIpRange{
					{
						IpCidrRange: proto.String("192.0.3.0/8"),
					},
				},
				Subnetwork: proto.String("projects/someProject/regions/someRegion/subnetworks/someSubnetwork"),
			},
		},
	}

	testCases := map[string]struct {
		projectID, instanceName, zone string
		instanceAPI                   stubInstanceAPI
		subnetAPI                     stubSubnetAPI
		wantErr                       bool
		wantInstance                  metadata.InstanceMetadata
	}{
		"success": {
			instanceName: "someInstance",
			projectID:    "someProject",
			zone:         "someZone-west3-b",
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			subnetAPI: stubSubnetAPI{
				subnet: &computepb.Subnetwork{
					SecondaryIpRanges: []*computepb.SubnetworkSecondaryRange{
						{
							IpCidrRange: proto.String("198.51.100.0/24"),
						},
					},
				},
			},
			wantInstance: metadata.InstanceMetadata{
				Name:             "someInstance",
				Role:             role.ControlPlane,
				ProviderID:       "gce://someProject/someZone-west3-b/someInstance",
				VPCIP:            "192.0.2.0",
				AliasIPRanges:    []string{"192.0.3.0/8"},
				SecondaryIPRange: "198.51.100.0/24",
			},
		},
		"get instance error": {
			instanceName: "someInstance",
			projectID:    "someProject",
			zone:         "someZone-west3-b",
			instanceAPI: stubInstanceAPI{
				instanceErr: someErr,
			},
			subnetAPI: stubSubnetAPI{
				subnet: &computepb.Subnetwork{
					SecondaryIpRanges: []*computepb.SubnetworkSecondaryRange{
						{
							IpCidrRange: proto.String("198.51.100.0/24"),
						},
					},
				},
			},
			wantErr: true,
		},
		"get subnet error": {
			instanceName: "someInstance",
			projectID:    "someProject",
			zone:         "someZone-west3-b",
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			subnetAPI: stubSubnetAPI{
				subnetErr: someErr,
			},
			wantErr: true,
		},
		"invalid instance": {
			instanceName: "someInstance",
			projectID:    "someProject",
			zone:         "someZone-west3-b",
			instanceAPI: stubInstanceAPI{
				instance: nil,
			},
			subnetAPI: stubSubnetAPI{
				subnet: &computepb.Subnetwork{
					SecondaryIpRanges: []*computepb.SubnetworkSecondaryRange{
						{
							IpCidrRange: proto.String("198.51.100.0/24"),
						},
					},
				},
			},
			wantErr: true,
		},
		"invalid zone": {
			instanceName: "someInstance",
			projectID:    "someProject",
			zone:         "invalidZone",
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			subnetAPI: stubSubnetAPI{
				subnet: &computepb.Subnetwork{
					SecondaryIpRanges: []*computepb.SubnetworkSecondaryRange{
						{
							IpCidrRange: proto.String("198.51.100.0/24"),
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
			require := require.New(t)

			cloud := &Cloud{
				instanceAPI: &tc.instanceAPI,
				subnetAPI:   &tc.subnetAPI,
			}
			instance, err := cloud.getInstance(context.Background(), tc.projectID, tc.zone, tc.instanceName)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantInstance, instance)
		})
	}
}

func TestGetLoadbalancerEndpoint(t *testing.T) {
	someErr := errors.New("failed")
	goodInstance := &computepb.Instance{
		Name: proto.String("someInstance"),
		Zone: proto.String("someZone-west3-b"),
		Labels: map[string]string{
			cloud.TagUID:  "1234",
			cloud.TagRole: role.ControlPlane.String(),
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
				Name:      proto.String("nic0"),
				NetworkIP: proto.String("192.0.2.0"),
				AliasIpRanges: []*computepb.AliasIpRange{
					{
						IpCidrRange: proto.String("192.0.3.0/8"),
					},
				},
				Subnetwork: proto.String("projects/someProject/regions/someRegion/subnetworks/someSubnetwork"),
			},
		},
	}

	testCases := map[string]struct {
		imds                       stubIMDS
		instanceAPI                stubInstanceAPI
		globalForwardingRulesAPI   stubGlobalForwardingRulesAPI
		regionalForwardingRulesAPI stubRegionalForwardingRulesAPI
		wantHost                   string
		wantErr                    bool
	}{
		"success global forwarding rule": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					forwardingRules: []*computepb.ForwardingRule{
						{
							PortRange: proto.String("6443"),
							IPAddress: proto.String("192.0.2.255"),
						},
					},
				},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			wantHost: "192.0.2.255",
		},
		"success regional forwarding rule": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					forwardingRules: []*computepb.ForwardingRule{
						{
							PortRange: proto.String("6443"),
							IPAddress: proto.String("192.0.2.255"),
							Region:    proto.String("someRegion"),
						},
					},
				},
			},
			wantHost: "192.0.2.255",
		},
		"regional forwarding rule has no region": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					forwardingRules: []*computepb.ForwardingRule{
						{
							PortRange: proto.String("6443"),
							IPAddress: proto.String("192.0.2.255"),
						},
					},
				},
			},
			wantErr: true,
		},
		"imds error": {
			imds: stubIMDS{
				projectIDErr: someErr,
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					forwardingRules: []*computepb.ForwardingRule{
						{
							PortRange: proto.String("6443"),
							IPAddress: proto.String("192.0.2.255"),
						},
					},
				},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			wantErr: true,
		},
		"global forwarding rule iterator error": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					err: someErr,
				},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			wantErr: true,
		},
		"regional forwarding rule iterator error": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					err: someErr,
				},
			},
			wantErr: true,
		},
		"no forwarding rules": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			wantErr: true,
		},
		"missing port range": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					forwardingRules: []*computepb.ForwardingRule{
						{
							IPAddress: proto.String("192.0.2.255"),
						},
					},
				},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			wantErr: true,
		},
		"missing IP address": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: goodInstance,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					forwardingRules: []*computepb.ForwardingRule{
						{
							PortRange: proto.String("6443"),
						},
					},
				},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			wantErr: true,
		},
		"get instance error": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instanceErr: someErr,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					forwardingRules: []*computepb.ForwardingRule{
						{
							PortRange: proto.String("6443"),
							IPAddress: proto.String("192.0.2.255"),
						},
					},
				},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			wantErr: true,
		},
		"invalid instance": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: nil,
			},
			globalForwardingRulesAPI: stubGlobalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{
					forwardingRules: []*computepb.ForwardingRule{
						{
							PortRange: proto.String("6443"),
							IPAddress: proto.String("192.0.2.255"),
						},
					},
				},
			},
			regionalForwardingRulesAPI: stubRegionalForwardingRulesAPI{
				iterator: &stubForwardingRulesIterator{},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cloud := &Cloud{
				imds:                       &tc.imds,
				instanceAPI:                &tc.instanceAPI,
				globalForwardingRulesAPI:   &tc.globalForwardingRulesAPI,
				regionalForwardingRulesAPI: &tc.regionalForwardingRulesAPI,
			}

			gotHost, gotPort, err := cloud.GetLoadBalancerEndpoint(context.Background())
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

func TestList(t *testing.T) {
	someErr := errors.New("failed")
	goodInstance := &computepb.Instance{
		Name: proto.String("someInstance"),
		Zone: proto.String("someZone-west3-b"),
		Labels: map[string]string{
			cloud.TagUID:  "1234",
			cloud.TagRole: role.ControlPlane.String(),
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
				Name:      proto.String("nic0"),
				NetworkIP: proto.String("192.0.2.0"),
				AliasIpRanges: []*computepb.AliasIpRange{
					{
						IpCidrRange: proto.String("198.51.100.0/24"),
					},
				},
				Subnetwork: proto.String("projects/someProject/regions/someRegion/subnetworks/someSubnetwork"),
			},
		},
	}
	otherZoneInstance := &computepb.Instance{
		Name: proto.String("otherZoneInstance"),
		Zone: proto.String("someZone-east1-a"),
		Labels: map[string]string{
			cloud.TagUID:  "1234",
			cloud.TagRole: role.ControlPlane.String(),
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
				Name:      proto.String("nic0"),
				NetworkIP: proto.String("192.0.2.1"),
				AliasIpRanges: []*computepb.AliasIpRange{
					{
						IpCidrRange: proto.String("198.51.100.0/24"),
					},
				},
				Subnetwork: proto.String("projects/someProject/regions/someRegion/subnetworks/someSubnetwork"),
			},
		},
	}
	goodSubnet := &computepb.Subnetwork{
		SecondaryIpRanges: []*computepb.SubnetworkSecondaryRange{
			{
				IpCidrRange: proto.String("198.51.100.0/24"),
			},
		},
	}

	testCases := map[string]struct {
		imds          stubIMDS
		instanceAPI   instanceAPI
		subnetAPI     stubSubnetAPI
		zoneAPI       stubZoneAPI
		wantErr       bool
		wantInstances []metadata.InstanceMetadata
	}{
		"success": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: &stubInstanceAPI{
				instance: goodInstance,
				iterator: &stubInstanceIterator{
					instances: []*computepb.Instance{
						goodInstance,
					},
				},
			},
			subnetAPI: stubSubnetAPI{
				subnet: goodSubnet,
			},
			zoneAPI: stubZoneAPI{
				iterator: &stubZoneIterator{
					zones: []*computepb.Zone{
						{
							Name: proto.String("someZone-west3-b"),
						},
					},
				},
			},
			wantInstances: []metadata.InstanceMetadata{
				{
					Name:             "someInstance",
					Role:             role.ControlPlane,
					ProviderID:       "gce://someProject/someZone-west3-b/someInstance",
					VPCIP:            "192.0.2.0",
					AliasIPRanges:    []string{"198.51.100.0/24"},
					SecondaryIPRange: "198.51.100.0/24",
				},
			},
		},
		"multi-zone": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: &fakeInstanceAPI{
				instance: goodInstance,
				iterators: map[string]*stubInstanceIterator{
					"someZone-east1-a": {
						instances: []*computepb.Instance{
							otherZoneInstance,
						},
					},
					"someZone-west3-b": {
						instances: []*computepb.Instance{
							goodInstance,
						},
					},
				},
			},
			subnetAPI: stubSubnetAPI{
				subnet: goodSubnet,
			},
			zoneAPI: stubZoneAPI{
				iterator: &stubZoneIterator{
					zones: []*computepb.Zone{
						{Name: proto.String("someZone-east1-a")},
						{Name: proto.String("someZone-west3-b")},
					},
				},
			},
			wantInstances: []metadata.InstanceMetadata{
				{
					Name:             "otherZoneInstance",
					Role:             role.ControlPlane,
					ProviderID:       "gce://someProject/someZone-east1-a/otherZoneInstance",
					VPCIP:            "192.0.2.1",
					AliasIPRanges:    []string{"198.51.100.0/24"},
					SecondaryIPRange: "198.51.100.0/24",
				},
				{
					Name:             "someInstance",
					Role:             role.ControlPlane,
					ProviderID:       "gce://someProject/someZone-west3-b/someInstance",
					VPCIP:            "192.0.2.0",
					AliasIPRanges:    []string{"198.51.100.0/24"},
					SecondaryIPRange: "198.51.100.0/24",
				},
			},
		},
		"list multiple instances": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: &stubInstanceAPI{
				instance: goodInstance,
				iterator: &stubInstanceIterator{
					instances: []*computepb.Instance{
						goodInstance,
						{
							Name: proto.String("anotherInstance"),
							Zone: proto.String("someZone-west3-b"),
							Labels: map[string]string{
								cloud.TagUID:  "1234",
								cloud.TagRole: role.Worker.String(),
							},
							NetworkInterfaces: []*computepb.NetworkInterface{
								{
									Name:      proto.String("nic0"),
									NetworkIP: proto.String("192.0.2.1"),
									AliasIpRanges: []*computepb.AliasIpRange{
										{
											IpCidrRange: proto.String("198.51.100.0/24"),
										},
									},
									Subnetwork: proto.String("projects/someProject/regions/someRegion/subnetworks/someSubnetwork"),
								},
							},
						},
					},
				},
			},
			subnetAPI: stubSubnetAPI{
				subnet: goodSubnet,
			},
			zoneAPI: stubZoneAPI{
				iterator: &stubZoneIterator{
					zones: []*computepb.Zone{
						{
							Name: proto.String("someZone-west3-b"),
						},
					},
				},
			},
			wantInstances: []metadata.InstanceMetadata{
				{
					Name:             "someInstance",
					Role:             role.ControlPlane,
					ProviderID:       "gce://someProject/someZone-west3-b/someInstance",
					VPCIP:            "192.0.2.0",
					AliasIPRanges:    []string{"198.51.100.0/24"},
					SecondaryIPRange: "198.51.100.0/24",
				},
				{
					Name:             "anotherInstance",
					Role:             role.Worker,
					ProviderID:       "gce://someProject/someZone-west3-b/anotherInstance",
					VPCIP:            "192.0.2.1",
					AliasIPRanges:    []string{"198.51.100.0/24"},
					SecondaryIPRange: "198.51.100.0/24",
				},
			},
		},
		"imds error": {
			imds: stubIMDS{
				projectIDErr: someErr,
			},
			instanceAPI: &stubInstanceAPI{
				instance: goodInstance,
				iterator: &stubInstanceIterator{
					instances: []*computepb.Instance{
						goodInstance,
					},
				},
			},
			subnetAPI: stubSubnetAPI{
				subnet: goodSubnet,
			},
			wantErr: true,
		},
		"iterator error": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: &stubInstanceAPI{
				instance: goodInstance,
				iterator: &stubInstanceIterator{
					err: someErr,
				},
			},
			subnetAPI: stubSubnetAPI{
				subnet: goodSubnet,
			},
			zoneAPI: stubZoneAPI{
				iterator: &stubZoneIterator{
					zones: []*computepb.Zone{
						{
							Name: proto.String("someZone-west3-b"),
						},
					},
				},
			},
			wantErr: true,
		},
		"get instance error": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: &stubInstanceAPI{
				instanceErr: someErr,
				iterator: &stubInstanceIterator{
					instances: []*computepb.Instance{
						goodInstance,
					},
				},
			},
			subnetAPI: stubSubnetAPI{
				subnet: goodSubnet,
			},
			zoneAPI: stubZoneAPI{
				iterator: &stubZoneIterator{
					zones: []*computepb.Zone{
						{
							Name: proto.String("someZone-west3-b"),
						},
					},
				},
			},
			wantErr: true,
		},
		"get subnet error": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: &stubInstanceAPI{
				instance: goodInstance,
				iterator: &stubInstanceIterator{
					instances: []*computepb.Instance{
						goodInstance,
					},
				},
			},
			subnetAPI: stubSubnetAPI{
				subnetErr: someErr,
			},
			zoneAPI: stubZoneAPI{
				iterator: &stubZoneIterator{
					zones: []*computepb.Zone{
						{
							Name: proto.String("someZone-west3-b"),
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
			require := require.New(t)

			cloud := &Cloud{
				imds:        &tc.imds,
				instanceAPI: tc.instanceAPI,
				subnetAPI:   &tc.subnetAPI,
				zoneAPI:     &tc.zoneAPI,
			}

			instances, err := cloud.List(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantInstances, instances)
		})
	}
}

func TestRegion(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		imds       stubIMDS
		wantRegion string
		wantErr    bool
	}{
		"success": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someregion-west3-b",
				instanceName: "someInstance",
			},
			wantRegion: "someregion-west3",
		},
		"get zone error": {
			imds: stubIMDS{
				zoneErr: someErr,
			},
			wantErr: true,
		},
		"invalid zone format": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "invalid",
				instanceName: "someInstance",
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cloud := &Cloud{
				imds: &tc.imds,
			}

			assert.Empty(cloud.regionCache)

			gotRegion, err := cloud.region()
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantRegion, gotRegion)
			assert.Equal(tc.wantRegion, cloud.regionCache)
		})
	}
}

func TestZones(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		zoneAPI   stubZoneAPI
		wantZones []string
		wantErr   bool
	}{
		"success": {
			zoneAPI: stubZoneAPI{
				&stubZoneIterator{
					zones: []*computepb.Zone{
						{Name: proto.String("someregion-west3-b")},
						{},  // missing name (should be ignored)
						nil, // nil (should be ignored)
					},
				},
			},
			wantZones: []string{"someregion-west3-b"},
		},
		"get zones error": {
			zoneAPI: stubZoneAPI{
				&stubZoneIterator{
					err: someErr,
				},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cloud := &Cloud{
				zoneAPI: &tc.zoneAPI,
			}

			assert.Empty(cloud.zoneCache)

			gotZones, err := cloud.zones(context.Background(), "someProject", "someregion-west3")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantZones, gotZones)
			assert.Equal(tc.wantZones, cloud.zoneCache["someregion-west3"].zones)
		})
	}
}

func TestRetrieveInstanceInfo(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		imds    stubIMDS
		wantErr bool
	}{
		"success": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
		},
		"get project id error": {
			imds: stubIMDS{
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
				projectIDErr: someErr,
			},
			wantErr: true,
		},
		"get zone error": {
			imds: stubIMDS{
				projectID:    "someProject",
				instanceName: "someInstance",
				zoneErr:      someErr,
			},
			wantErr: true,
		},
		"get instance name error": {
			imds: stubIMDS{
				projectID:       "someProject",
				zone:            "someZone-west3-b",
				instanceNameErr: someErr,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cloud := &Cloud{
				imds: &tc.imds,
			}

			project, zone, instance, err := cloud.retrieveInstanceInfo()
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.imds.projectID, project)
			assert.Equal(tc.imds.zone, zone)
			assert.Equal(tc.imds.instanceName, instance)
		})
	}
}

func TestUID(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		imds        stubIMDS
		instanceAPI stubInstanceAPI
		wantUID     string
		wantErr     bool
	}{
		"success": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: &computepb.Instance{
					Name: proto.String("someInstance"),
					Zone: proto.String("someZone-west3-b"),
					Labels: map[string]string{
						cloud.TagUID:  "1234",
						cloud.TagRole: role.ControlPlane.String(),
					},
				},
			},
			wantUID: "1234",
		},
		"imds error": {
			imds: stubIMDS{
				projectIDErr: someErr,
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: &computepb.Instance{
					Name: proto.String("someInstance"),
					Zone: proto.String("someZone-west3-b"),
					Labels: map[string]string{
						cloud.TagUID:  "1234",
						cloud.TagRole: role.ControlPlane.String(),
					},
				},
			},
			wantErr: true,
		},
		"instance error": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instanceErr: someErr,
			},
			wantErr: true,
		},
		"invalid instance": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: nil,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cloud := &Cloud{
				imds:        &tc.imds,
				instanceAPI: &tc.instanceAPI,
			}

			uid, err := cloud.UID(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantUID, uid)
		})
	}
}

func TestInitSecretHash(t *testing.T) {
	someErr := errors.New("failed")

	testCases := map[string]struct {
		imds               stubIMDS
		instanceAPI        stubInstanceAPI
		wantInitSecretHash string
		wantErr            bool
	}{
		"success": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: &computepb.Instance{
					Name: proto.String("someInstance"),
					Zone: proto.String("someZone-west3-b"),
					Labels: map[string]string{
						cloud.TagRole: role.ControlPlane.String(),
					},
					Metadata: &computepb.Metadata{
						Items: []*computepb.Items{
							{
								Key:   proto.String(cloud.TagInitSecretHash),
								Value: proto.String("initSecretHash"),
							},
						},
					},
				},
			},
			wantInitSecretHash: "initSecretHash",
		},
		"imds error": {
			imds: stubIMDS{
				projectIDErr: someErr,
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: &computepb.Instance{
					Name: proto.String("someInstance"),
					Zone: proto.String("someZone-west3-b"),
					Labels: map[string]string{
						cloud.TagInitSecretHash: "initSecretHash",
						cloud.TagRole:           role.ControlPlane.String(),
					},
					Metadata: &computepb.Metadata{
						Items: []*computepb.Items{
							{
								Key:   proto.String(cloud.TagInitSecretHash),
								Value: proto.String("initSecretHash"),
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"instance error": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instanceErr: someErr,
			},
			wantErr: true,
		},
		"invalid instance": {
			imds: stubIMDS{
				projectID:    "someProject",
				zone:         "someZone-west3-b",
				instanceName: "someInstance",
			},
			instanceAPI: stubInstanceAPI{
				instance: nil,
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			cloud := &Cloud{
				imds:        &tc.imds,
				instanceAPI: &tc.instanceAPI,
			}

			initSecretHash, err := cloud.InitSecretHash(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal([]byte(tc.wantInitSecretHash), initSecretHash)
		})
	}
}

type stubGlobalForwardingRulesAPI struct {
	iterator forwardingRuleIterator
}

func (s *stubGlobalForwardingRulesAPI) List(
	_ context.Context, _ *computepb.ListGlobalForwardingRulesRequest, _ ...gax.CallOption,
) forwardingRuleIterator {
	return s.iterator
}

func (s *stubGlobalForwardingRulesAPI) Close() error { return nil }

type stubRegionalForwardingRulesAPI struct {
	iterator forwardingRuleIterator
}

func (s *stubRegionalForwardingRulesAPI) List(
	_ context.Context, _ *computepb.ListForwardingRulesRequest, _ ...gax.CallOption,
) forwardingRuleIterator {
	return s.iterator
}

func (s *stubRegionalForwardingRulesAPI) Close() error { return nil }

type stubForwardingRulesIterator struct {
	ctr             int
	forwardingRules []*computepb.ForwardingRule
	err             error
}

func (s *stubForwardingRulesIterator) Next() (*computepb.ForwardingRule, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.ctr >= len(s.forwardingRules) {
		return nil, iterator.Done
	}
	s.ctr++
	return s.forwardingRules[s.ctr-1], nil
}

type stubIMDS struct {
	instanceID      string
	projectID       string
	zone            string
	instanceName    string
	instanceIDErr   error
	projectIDErr    error
	zoneErr         error
	instanceNameErr error
}

func (s *stubIMDS) InstanceID() (string, error) { return s.instanceID, s.instanceIDErr }

func (s *stubIMDS) ProjectID() (string, error) { return s.projectID, s.projectIDErr }

func (s *stubIMDS) Zone() (string, error) { return s.zone, s.zoneErr }

func (s *stubIMDS) InstanceName() (string, error) { return s.instanceName, s.instanceNameErr }

type stubInstanceAPI struct {
	instance    *computepb.Instance
	instanceErr error
	iterator    *stubInstanceIterator
}

func (s *stubInstanceAPI) Get(
	_ context.Context, _ *computepb.GetInstanceRequest, _ ...gax.CallOption,
) (*computepb.Instance, error) {
	return s.instance, s.instanceErr
}

func (s *stubInstanceAPI) List(
	_ context.Context, _ *computepb.ListInstancesRequest, _ ...gax.CallOption,
) instanceIterator {
	return s.iterator
}

func (s *stubInstanceAPI) Close() error { return nil }

type fakeInstanceAPI struct {
	instance    *computepb.Instance
	instanceErr error
	// iterators is a map of zone to instance iterator.
	iterators map[string]*stubInstanceIterator
}

func (s *fakeInstanceAPI) Get(
	_ context.Context, _ *computepb.GetInstanceRequest, _ ...gax.CallOption,
) (*computepb.Instance, error) {
	return s.instance, s.instanceErr
}

func (s *fakeInstanceAPI) List(
	_ context.Context, req *computepb.ListInstancesRequest, _ ...gax.CallOption,
) instanceIterator {
	return s.iterators[req.GetZone()]
}

func (s *fakeInstanceAPI) Close() error { return nil }

type stubInstanceIterator struct {
	ctr       int
	instances []*computepb.Instance
	err       error
}

func (s *stubInstanceIterator) Next() (*computepb.Instance, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.ctr >= len(s.instances) {
		return nil, iterator.Done
	}
	s.ctr++
	return s.instances[s.ctr-1], nil
}

type stubSubnetAPI struct {
	subnet    *computepb.Subnetwork
	subnetErr error
}

func (s *stubSubnetAPI) Get(
	_ context.Context, _ *computepb.GetSubnetworkRequest, _ ...gax.CallOption,
) (*computepb.Subnetwork, error) {
	return s.subnet, s.subnetErr
}
func (s *stubSubnetAPI) Close() error { return nil }

type stubZoneAPI struct {
	iterator *stubZoneIterator
}

func (s *stubZoneAPI) List(_ context.Context,
	_ *computepb.ListZonesRequest, _ ...gax.CallOption,
) zoneIterator {
	return s.iterator
}

type stubZoneIterator struct {
	ctr   int
	zones []*computepb.Zone
	err   error
}

func (s *stubZoneIterator) Next() (*computepb.Zone, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.ctr >= len(s.zones) {
		return nil, iterator.Done
	}
	s.ctr++
	return s.zones[s.ctr-1], nil
}
