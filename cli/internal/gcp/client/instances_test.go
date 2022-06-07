package client

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/cli/cloud/cloudtypes"
	"github.com/stretchr/testify/assert"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

func TestCreateInstances(t *testing.T) {
	testInstances := []*computepb.Instance{
		{
			Name: proto.String("instance-name-1"),
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					AccessConfigs: []*computepb.AccessConfig{
						{NatIP: proto.String("public-ip")},
					},
					NetworkIP: proto.String("private-ip"),
				},
			},
		},
		{
			Name: proto.String("instance-name-2"),
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					AccessConfigs: []*computepb.AccessConfig{
						{NatIP: proto.String("public-ip")},
					},
					NetworkIP: proto.String("private-ip"),
				},
			},
		},
	}
	testManagedInstances := []*computepb.ManagedInstance{
		{CurrentAction: proto.String(computepb.ManagedInstance_NONE.String())},
		{CurrentAction: proto.String(computepb.ManagedInstance_NONE.String())},
	}
	testInput := CreateInstancesInput{
		CountCoordinators: 3,
		CountNodes:        4,
		ImageId:           "img",
		InstanceType:      "n2d-standard-2",
		KubeEnv:           "kube-env",
	}
	someErr := errors.New("failed")

	testCases := map[string]struct {
		instanceAPI              instanceAPI
		operationZoneAPI         operationZoneAPI
		operationGlobalAPI       operationGlobalAPI
		instanceTemplateAPI      instanceTemplateAPI
		instanceGroupManagersAPI instanceGroupManagersAPI
		input                    CreateInstancesInput
		network                  string
		wantErr                  bool
	}{
		"successful create": {
			instanceAPI:              stubInstanceAPI{listIterator: &stubInstanceIterator{instances: testInstances}},
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{listIterator: &stubManagedInstanceIterator{instances: testManagedInstances}},
			network:                  "network",
			input:                    testInput,
		},
		"failed no network": {
			instanceAPI:              stubInstanceAPI{listIterator: &stubInstanceIterator{instances: testInstances}},
			operationZoneAPI:         stubOperationZoneAPI{waitErr: someErr},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{listIterator: &stubManagedInstanceIterator{instances: testManagedInstances}},
			input:                    testInput,
			wantErr:                  true,
		},
		"failed wait zonal op": {
			instanceAPI:              stubInstanceAPI{listIterator: &stubInstanceIterator{instances: testInstances}},
			operationZoneAPI:         stubOperationZoneAPI{waitErr: someErr},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{listIterator: &stubManagedInstanceIterator{instances: testManagedInstances}},
			network:                  "network",
			input:                    testInput,
			wantErr:                  true,
		},
		"failed wait global op": {
			instanceAPI:              stubInstanceAPI{listIterator: &stubInstanceIterator{instances: testInstances}},
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{waitErr: someErr},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{listIterator: &stubManagedInstanceIterator{instances: testManagedInstances}},
			network:                  "network",
			input:                    testInput,
			wantErr:                  true,
		},
		"failed insert template": {
			instanceAPI:              stubInstanceAPI{listIterator: &stubInstanceIterator{instances: testInstances}},
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{insertErr: someErr},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{listIterator: &stubManagedInstanceIterator{instances: testManagedInstances}},
			input:                    testInput,
			network:                  "network",
			wantErr:                  true,
		},
		"failed insert instanceGroupManager": {
			instanceAPI:              stubInstanceAPI{listIterator: &stubInstanceIterator{instances: testInstances}},
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{insertErr: someErr},
			network:                  "network",
			input:                    testInput,
			wantErr:                  true,
		},
		"failed instanceGroupManager iterator": {
			instanceAPI:              stubInstanceAPI{listIterator: &stubInstanceIterator{instances: testInstances}},
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{listIterator: &stubManagedInstanceIterator{nextErr: someErr}},
			network:                  "network",
			input:                    testInput,
			wantErr:                  true,
		},
		"failed instance iterator": {
			instanceAPI:              stubInstanceAPI{listIterator: &stubInstanceIterator{nextErr: someErr}},
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{listIterator: &stubManagedInstanceIterator{instances: testManagedInstances}},
			network:                  "network",
			input:                    testInput,
			wantErr:                  true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:                  "project",
				zone:                     "zone",
				name:                     "name",
				uid:                      "uid",
				network:                  tc.network,
				subnetwork:               "subnetwork",
				secondarySubnetworkRange: "secondary-range",
				instanceAPI:              tc.instanceAPI,
				operationZoneAPI:         tc.operationZoneAPI,
				operationGlobalAPI:       tc.operationGlobalAPI,
				instanceTemplateAPI:      tc.instanceTemplateAPI,
				instanceGroupManagersAPI: tc.instanceGroupManagersAPI,
				nodes:                    make(cloudtypes.Instances),
				coordinators:             make(cloudtypes.Instances),
			}

			if tc.wantErr {
				assert.Error(client.CreateInstances(ctx, tc.input))
			} else {
				assert.NoError(client.CreateInstances(ctx, tc.input))
				assert.Equal([]string{"public-ip", "public-ip"}, client.nodes.PublicIPs())
				assert.Equal([]string{"private-ip", "private-ip"}, client.nodes.PrivateIPs())
				assert.Equal([]string{"public-ip", "public-ip"}, client.coordinators.PublicIPs())
				assert.Equal([]string{"private-ip", "private-ip"}, client.coordinators.PrivateIPs())
				assert.NotNil(client.nodesInstanceGroup)
				assert.NotNil(client.coordinatorInstanceGroup)
				assert.NotNil(client.coordinatorTemplate)
				assert.NotNil(client.nodeTemplate)
			}
		})
	}
}

func TestTerminateInstances(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		operationZoneAPI         operationZoneAPI
		operationGlobalAPI       operationGlobalAPI
		instanceTemplateAPI      instanceTemplateAPI
		instanceGroupManagersAPI instanceGroupManagersAPI

		missingNodeInstanceGroup bool
		wantErr                  bool
	}{
		"successful terminate": {
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{},
		},
		"successful terminate with missing node instance group": {
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{},
			missingNodeInstanceGroup: true,
		},
		"fail delete instanceGroupManager": {
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{deleteErr: someErr},
			wantErr:                  true,
		},
		"fail delete instanceTemplate": {
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{deleteErr: someErr},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{},
			wantErr:                  true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()
			client := Client{
				project:                  "project",
				zone:                     "zone",
				name:                     "name",
				uid:                      "uid",
				operationZoneAPI:         tc.operationZoneAPI,
				operationGlobalAPI:       tc.operationGlobalAPI,
				instanceTemplateAPI:      tc.instanceTemplateAPI,
				instanceGroupManagersAPI: tc.instanceGroupManagersAPI,
				nodes:                    cloudtypes.Instances{"node-id-1": cloudtypes.Instance{}, "node-id-2": cloudtypes.Instance{}},
				coordinators:             cloudtypes.Instances{"coordinator-id-1": cloudtypes.Instance{}},
				firewalls:                []string{"firewall-1", "firewall-2"},
				network:                  "network-id-1",
				nodesInstanceGroup:       "nodeInstanceGroup-id-1",
				coordinatorInstanceGroup: "coordinatorInstanceGroup-id-1",
				nodeTemplate:             "template-id-1",
				coordinatorTemplate:      "template-id-1",
			}
			if tc.missingNodeInstanceGroup {
				client.nodesInstanceGroup = ""
				client.nodes = cloudtypes.Instances{}
			}

			if tc.wantErr {
				assert.Error(client.TerminateInstances(ctx))
			} else {
				assert.NoError(client.TerminateInstances(ctx))
				assert.Nil(client.nodes.PublicIPs())
				assert.Nil(client.nodes.PrivateIPs())
				assert.Nil(client.coordinators.PublicIPs())
				assert.Nil(client.coordinators.PrivateIPs())
				assert.Empty(client.nodesInstanceGroup)
				assert.Empty(client.coordinatorInstanceGroup)
				assert.Empty(client.coordinatorTemplate)
				assert.Empty(client.nodeTemplate)
			}
		})
	}
}
