/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	cspapi "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	computeREST "google.golang.org/api/compute/v1"
	"google.golang.org/protobuf/proto"
)

func TestGetScalingGroupImage(t *testing.T) {
	testCases := map[string]struct {
		scalingGroupID                 string
		instanceGroupManagerTemplateID *string
		instanceTemplate               *computeREST.InstanceTemplate
		getInstanceGroupManagerErr     error
		getInstanceTemplateErr         error
		wantImage                      string
		wantErr                        bool
	}{
		"getting image works": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computeREST.InstanceTemplate{
				Properties: &computeREST.InstanceProperties{
					Disks: []*computeREST.AttachedDisk{
						{
							InitializeParams: &computeREST.AttachedDiskInitializeParams{
								SourceImage: "https://www.googleapis.com/compute/v1/projects/project/global/images/image",
							},
						},
					},
				},
			},
			wantImage: "projects/project/global/images/image",
		},
		"splitting scalingGroupID fails": {
			scalingGroupID: "invalid",
			wantErr:        true,
		},
		"get instance fails": {
			scalingGroupID:             "projects/project/zones/zone/instanceGroupManagers/instance-group",
			getInstanceGroupManagerErr: errors.New("get instance error"),
			wantErr:                    true,
		},
		"instance group manager has no template": {
			scalingGroupID: "projects/project/zones/zone/instanceGroupManagers/instance-group",
			wantErr:        true,
		},
		"instance group manager template id is invalid": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			instanceGroupManagerTemplateID: proto.String("invalid"),
			wantErr:                        true,
		},
		"get instance template fails": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			getInstanceTemplateErr:         errors.New("get instance template error"),
			wantErr:                        true,
		},
		"instance template has no disks": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computeREST.InstanceTemplate{
				Properties: &computeREST.InstanceProperties{},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				instanceGroupManagersAPI: &stubInstanceGroupManagersAPI{
					getErr: tc.getInstanceGroupManagerErr,
					instanceGroupManager: &computepb.InstanceGroupManager{
						InstanceTemplate: tc.instanceGroupManagerTemplateID,
					},
				},
				instanceTemplateAPI: &stubInstanceTemplateAPI{
					getErr:   tc.getInstanceTemplateErr,
					template: tc.instanceTemplate,
				},
			}
			gotImage, err := client.GetScalingGroupImage(context.Background(), tc.scalingGroupID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantImage, gotImage)
		})
	}
}

func TestSetScalingGroupImage(t *testing.T) {
	testCases := map[string]struct {
		scalingGroupID                 string
		imageURI                       string
		instanceGroupManagerTemplateID *string
		instanceTemplate               *computeREST.InstanceTemplate
		getInstanceGroupManagerErr     error
		getInstanceTemplateErr         error
		setInstanceTemplateErr         error
		insertInstanceTemplateErr      error
		wantErr                        bool
	}{
		"setting image works": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			imageURI:                       "projects/project/global/images/image-2",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computeREST.InstanceTemplate{
				Name: "instance-template",
				Properties: &computeREST.InstanceProperties{
					Disks: []*computeREST.AttachedDisk{
						{
							InitializeParams: &computeREST.AttachedDiskInitializeParams{
								SourceImage: "https://www.googleapis.com/compute/v1/projects/project/global/images/image-1",
							},
						},
					},
				},
			},
		},
		"same image already in use": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			imageURI:                       "projects/project/global/images/image",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computeREST.InstanceTemplate{
				Name: "instance-template",
				Properties: &computeREST.InstanceProperties{
					Disks: []*computeREST.AttachedDisk{
						{
							InitializeParams: &computeREST.AttachedDiskInitializeParams{
								SourceImage: "https://www.googleapis.com/compute/v1/projects/project/global/images/image",
							},
						},
					},
				},
			},
			// will not be triggered
			insertInstanceTemplateErr: errors.New("insert instance template error"),
		},
		"splitting scalingGroupID fails": {
			scalingGroupID: "invalid",
			wantErr:        true,
		},
		"get instance fails": {
			scalingGroupID:             "projects/project/zones/zone/instanceGroupManagers/instance-group",
			getInstanceGroupManagerErr: errors.New("get instance error"),
			wantErr:                    true,
		},
		"instance group manager has no template": {
			scalingGroupID: "projects/project/zones/zone/instanceGroupManagers/instance-group",
			wantErr:        true,
		},
		"instance group manager template id is invalid": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			instanceGroupManagerTemplateID: proto.String("invalid"),
			wantErr:                        true,
		},
		"get instance template fails": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			getInstanceTemplateErr:         errors.New("get instance template error"),
			wantErr:                        true,
		},
		"instance template has no disks": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computeREST.InstanceTemplate{
				Properties: &computeREST.InstanceProperties{},
			},
			wantErr: true,
		},
		"instance template has no name": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			imageURI:                       "projects/project/global/images/image-2",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computeREST.InstanceTemplate{
				Properties: &computeREST.InstanceProperties{
					Disks: []*computeREST.AttachedDisk{
						{
							InitializeParams: &computeREST.AttachedDiskInitializeParams{
								SourceImage: "https://www.googleapis.com/compute/v1/projects/project/global/images/image-1",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"instance template name generation fails": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			imageURI:                       "projects/project/global/images/image-2",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computeREST.InstanceTemplate{
				Name: "instance-template-999999999999999999999",
				Properties: &computeREST.InstanceProperties{
					Disks: []*computeREST.AttachedDisk{
						{
							InitializeParams: &computeREST.AttachedDiskInitializeParams{
								SourceImage: "https://www.googleapis.com/compute/v1/projects/project/global/images/image-1",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		"instance template insert fails": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			imageURI:                       "projects/project/global/images/image-2",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computeREST.InstanceTemplate{
				Name: "instance-template",
				Properties: &computeREST.InstanceProperties{
					Disks: []*computeREST.AttachedDisk{
						{
							InitializeParams: &computeREST.AttachedDiskInitializeParams{
								SourceImage: "https://www.googleapis.com/compute/v1/projects/project/global/images/image-1",
							},
						},
					},
				},
			},
			insertInstanceTemplateErr: errors.New("insert instance template error"),
			wantErr:                   true,
		},
		"setting instance template fails": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			imageURI:                       "projects/project/global/images/image-2",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computeREST.InstanceTemplate{
				Name: "instance-template",
				Properties: &computeREST.InstanceProperties{
					Disks: []*computeREST.AttachedDisk{
						{
							InitializeParams: &computeREST.AttachedDiskInitializeParams{
								SourceImage: "https://www.googleapis.com/compute/v1/projects/project/global/images/image-1",
							},
						},
					},
				},
			},
			setInstanceTemplateErr: errors.New("setting instance template error"),
			wantErr:                true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				instanceGroupManagersAPI: &stubInstanceGroupManagersAPI{
					getErr:                 tc.getInstanceGroupManagerErr,
					setInstanceTemplateErr: tc.setInstanceTemplateErr,
					instanceGroupManager: &computepb.InstanceGroupManager{
						InstanceTemplate: tc.instanceGroupManagerTemplateID,
					},
				},
				instanceTemplateAPI: &stubInstanceTemplateAPI{
					getErr:    tc.getInstanceTemplateErr,
					insertErr: tc.insertInstanceTemplateErr,
					template:  tc.instanceTemplate,
				},
			}
			err := client.SetScalingGroupImage(context.Background(), tc.scalingGroupID, tc.imageURI)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestGetScalingGroupName(t *testing.T) {
	testCases := map[string]struct {
		scalingGroupID string
		wantName       string
		wantErr        bool
	}{
		"valid scaling group ID": {
			scalingGroupID: "projects/project/zones/zone/instanceGroupManagers/instance-group",
			wantName:       "instance-group",
		},
		"invalid scaling group ID": {
			scalingGroupID: "invalid",
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{}
			gotName, err := client.GetScalingGroupName(tc.scalingGroupID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantName, gotName)
		})
	}
}

func TestListScalingGroups(t *testing.T) {
	testCases := map[string]struct {
		name                         *string
		groupID                      *string
		templateRef                  *string
		templateLabels               map[string]string
		listInstanceGroupManagersErr error
		templateGetErr               error
		wantGroups                   []cspapi.ScalingGroup
		wantErr                      bool
	}{
		"list instance group managers fails": {
			listInstanceGroupManagersErr: errors.New("list instance group managers error"),
			wantErr:                      true,
		},
		"get instance template fails": {
			name:           proto.String("test-control-plane-uid"),
			groupID:        proto.String("projects/project/zones/zone/instanceGroupManagers/test-control-plane-uid"),
			templateRef:    proto.String("projects/project/global/instanceTemplates/test-control-plane-uid"),
			templateGetErr: errors.New("get instance template error"),
			wantErr:        true,
		},
		"list instance group managers for control plane": {
			name:        proto.String("test-control-plane-uid"),
			groupID:     proto.String("projects/project/zones/zone/instanceGroupManagers/test-control-plane-uid"),
			templateRef: proto.String("projects/project/global/instanceTemplates/test-control-plane-uid"),
			templateLabels: map[string]string{
				"constellation-uid":  "uid",
				"constellation-role": "control-plane",
			},
			wantGroups: []cspapi.ScalingGroup{
				{
					Name:                 "test-control-plane-uid",
					NodeGroupName:        constants.ControlPlaneDefault,
					GroupID:              "projects/project/zones/zone/instanceGroupManagers/test-control-plane-uid",
					AutoscalingGroupName: "https://www.googleapis.com/compute/v1/projects/project/zones/zone/instanceGroups/test-control-plane-uid",
					Role:                 "ControlPlane",
				},
			},
		},
		"list instance group managers for worker": {
			name:        proto.String("test-worker-uid"),
			groupID:     proto.String("projects/project/zones/zone/instanceGroupManagers/test-worker-uid"),
			templateRef: proto.String("projects/project/global/instanceTemplates/test-control-plane-uid"),
			templateLabels: map[string]string{
				"constellation-uid":  "uid",
				"constellation-role": "worker",
			},
			wantGroups: []cspapi.ScalingGroup{
				{
					Name:                 "test-worker-uid",
					NodeGroupName:        constants.WorkerDefault,
					GroupID:              "projects/project/zones/zone/instanceGroupManagers/test-worker-uid",
					AutoscalingGroupName: "https://www.googleapis.com/compute/v1/projects/project/zones/zone/instanceGroups/test-worker-uid",
					Role:                 "Worker",
				},
			},
		},
		"list instance group managers with custom group name": {
			name:        proto.String("test-worker-uid"),
			groupID:     proto.String("projects/project/zones/zone/instanceGroupManagers/test-worker-uid"),
			templateRef: proto.String("projects/project/global/instanceTemplates/test-control-plane-uid"),
			templateLabels: map[string]string{
				"constellation-uid":        "uid",
				"constellation-role":       "worker",
				"constellation-node-group": "custom-group-name",
			},
			wantGroups: []cspapi.ScalingGroup{
				{
					Name:                 "test-worker-uid",
					NodeGroupName:        "custom-group-name",
					GroupID:              "projects/project/zones/zone/instanceGroupManagers/test-worker-uid",
					AutoscalingGroupName: "https://www.googleapis.com/compute/v1/projects/project/zones/zone/instanceGroups/test-worker-uid",
					Role:                 "Worker",
				},
			},
		},
		"listing instance group managers is not dependant on resource name": {
			name:        proto.String("some-instance-group-manager"),
			groupID:     proto.String("projects/project/zones/zone/instanceGroupManagers/some-instance-group-manager"),
			templateRef: proto.String("projects/project/global/instanceTemplates/some-instance-group-template"),
			templateLabels: map[string]string{
				"constellation-uid":  "uid",
				"constellation-role": "control-plane",
			},
			wantGroups: []cspapi.ScalingGroup{
				{
					Name:                 "some-instance-group-manager",
					NodeGroupName:        constants.ControlPlaneDefault,
					GroupID:              "projects/project/zones/zone/instanceGroupManagers/some-instance-group-manager",
					AutoscalingGroupName: "https://www.googleapis.com/compute/v1/projects/project/zones/zone/instanceGroups/some-instance-group-manager",
					Role:                 "ControlPlane",
				},
			},
		},
		"unrelated instance group manager": {
			name:        proto.String("test-control-plane-uid"),
			groupID:     proto.String("projects/project/zones/zone/instanceGroupManagers/test-unrelated-uid"),
			templateRef: proto.String("projects/project/global/instanceTemplates/test-control-plane-uid"),
			templateLabels: map[string]string{
				"label": "value",
			},
		},
		"invalid instance group manager": {},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				instanceGroupManagersAPI: &stubInstanceGroupManagersAPI{
					aggregatedListErr: tc.listInstanceGroupManagersErr,
					instanceGroupManager: &computepb.InstanceGroupManager{
						Name:             tc.name,
						SelfLink:         tc.groupID,
						InstanceTemplate: tc.templateRef,
					},
				},
				instanceTemplateAPI: &stubInstanceTemplateAPI{
					template: &computeREST.InstanceTemplate{
						Properties: &computeREST.InstanceProperties{
							Labels: tc.templateLabels,
						},
					},
					getErr: tc.templateGetErr,
				},
			}
			gotGroups, err := client.ListScalingGroups(context.Background(), "uid")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantGroups, gotGroups)
		})
	}
}
