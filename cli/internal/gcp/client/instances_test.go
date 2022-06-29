package client

import (
	"context"
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
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
		CountControlPlanes: 3,
		CountWorkers:       4,
		ImageId:            "img",
		InstanceType:       "n2d-standard-2",
		KubeEnv:            "kube-env",
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
				workers:                  make(cloudtypes.Instances),
				controlPlanes:            make(cloudtypes.Instances),
			}

			if tc.wantErr {
				assert.Error(client.CreateInstances(ctx, tc.input))
			} else {
				assert.NoError(client.CreateInstances(ctx, tc.input))
				assert.Equal([]string{"public-ip", "public-ip"}, client.workers.PublicIPs())
				assert.Equal([]string{"private-ip", "private-ip"}, client.workers.PrivateIPs())
				assert.Equal([]string{"public-ip", "public-ip"}, client.controlPlanes.PublicIPs())
				assert.Equal([]string{"private-ip", "private-ip"}, client.controlPlanes.PrivateIPs())
				assert.NotNil(client.workerInstanceGroup)
				assert.NotNil(client.controlPlaneInstanceGroup)
				assert.NotNil(client.controlPlaneTemplate)
				assert.NotNil(client.workerTemplate)
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

		missingWorkerInstanceGroup bool
		wantErr                    bool
	}{
		"successful terminate": {
			operationZoneAPI:         stubOperationZoneAPI{},
			operationGlobalAPI:       stubOperationGlobalAPI{},
			instanceTemplateAPI:      stubInstanceTemplateAPI{},
			instanceGroupManagersAPI: stubInstanceGroupManagersAPI{},
		},
		"successful terminate with missing worker instance group": {
			operationZoneAPI:           stubOperationZoneAPI{},
			operationGlobalAPI:         stubOperationGlobalAPI{},
			instanceTemplateAPI:        stubInstanceTemplateAPI{},
			instanceGroupManagersAPI:   stubInstanceGroupManagersAPI{},
			missingWorkerInstanceGroup: true,
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
				project:                   "project",
				zone:                      "zone",
				name:                      "name",
				uid:                       "uid",
				operationZoneAPI:          tc.operationZoneAPI,
				operationGlobalAPI:        tc.operationGlobalAPI,
				instanceTemplateAPI:       tc.instanceTemplateAPI,
				instanceGroupManagersAPI:  tc.instanceGroupManagersAPI,
				workers:                   cloudtypes.Instances{"worker-id-1": cloudtypes.Instance{}, "worker-id-2": cloudtypes.Instance{}},
				controlPlanes:             cloudtypes.Instances{"controlplane-id-1": cloudtypes.Instance{}},
				firewalls:                 []string{"firewall-1", "firewall-2"},
				network:                   "network-id-1",
				workerInstanceGroup:       "workerInstanceGroup-id-1",
				controlPlaneInstanceGroup: "controlplaneInstanceGroup-id-1",
				workerTemplate:            "template-id-1",
				controlPlaneTemplate:      "template-id-1",
			}
			if tc.missingWorkerInstanceGroup {
				client.workerInstanceGroup = ""
				client.workers = cloudtypes.Instances{}
			}

			if tc.wantErr {
				assert.Error(client.TerminateInstances(ctx))
			} else {
				assert.NoError(client.TerminateInstances(ctx))
				assert.Nil(client.workers.PublicIPs())
				assert.Nil(client.workers.PrivateIPs())
				assert.Nil(client.controlPlanes.PublicIPs())
				assert.Nil(client.controlPlanes.PrivateIPs())
				assert.Empty(client.workerInstanceGroup)
				assert.Empty(client.controlPlaneInstanceGroup)
				assert.Empty(client.controlPlaneTemplate)
				assert.Empty(client.workerTemplate)
			}
		})
	}
}
