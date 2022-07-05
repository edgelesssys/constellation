package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
	"google.golang.org/protobuf/proto"
)

func TestGetScalingGroupImage(t *testing.T) {
	testCases := map[string]struct {
		scalingGroupID                 string
		instanceGroupManagerTemplateID *string
		instanceTemplate               *computepb.InstanceTemplate
		getInstanceGroupManagerErr     error
		getInstanceTemplateErr         error
		wantImage                      string
		wantErr                        bool
	}{
		"getting image works": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computepb.InstanceTemplate{
				Properties: &computepb.InstanceProperties{
					Disks: []*computepb.AttachedDisk{
						{
							InitializeParams: &computepb.AttachedDiskInitializeParams{
								SourceImage: proto.String("https://www.googleapis.com/compute/v1/projects/project/global/images/image"),
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
			instanceTemplate: &computepb.InstanceTemplate{
				Properties: &computepb.InstanceProperties{},
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
		instanceTemplate               *computepb.InstanceTemplate
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
			instanceTemplate: &computepb.InstanceTemplate{
				Name: proto.String("instance-template"),
				Properties: &computepb.InstanceProperties{
					Disks: []*computepb.AttachedDisk{
						{
							InitializeParams: &computepb.AttachedDiskInitializeParams{
								SourceImage: proto.String("https://www.googleapis.com/compute/v1/projects/project/global/images/image-1"),
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
			instanceTemplate: &computepb.InstanceTemplate{
				Name: proto.String("instance-template"),
				Properties: &computepb.InstanceProperties{
					Disks: []*computepb.AttachedDisk{
						{
							InitializeParams: &computepb.AttachedDiskInitializeParams{
								SourceImage: proto.String("https://www.googleapis.com/compute/v1/projects/project/global/images/image"),
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
			instanceTemplate: &computepb.InstanceTemplate{
				Properties: &computepb.InstanceProperties{},
			},
			wantErr: true,
		},
		"instance template has no name": {
			scalingGroupID:                 "projects/project/zones/zone/instanceGroupManagers/instance-group",
			imageURI:                       "projects/project/global/images/image-2",
			instanceGroupManagerTemplateID: proto.String("projects/project/global/instanceTemplates/instance-template"),
			instanceTemplate: &computepb.InstanceTemplate{
				Properties: &computepb.InstanceProperties{
					Disks: []*computepb.AttachedDisk{
						{
							InitializeParams: &computepb.AttachedDiskInitializeParams{
								SourceImage: proto.String("https://www.googleapis.com/compute/v1/projects/project/global/images/image-1"),
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
			instanceTemplate: &computepb.InstanceTemplate{
				Name: proto.String("instance-template-999999999999999999999"),
				Properties: &computepb.InstanceProperties{
					Disks: []*computepb.AttachedDisk{
						{
							InitializeParams: &computepb.AttachedDiskInitializeParams{
								SourceImage: proto.String("https://www.googleapis.com/compute/v1/projects/project/global/images/image-1"),
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
			instanceTemplate: &computepb.InstanceTemplate{
				Name: proto.String("instance-template"),
				Properties: &computepb.InstanceProperties{
					Disks: []*computepb.AttachedDisk{
						{
							InitializeParams: &computepb.AttachedDiskInitializeParams{
								SourceImage: proto.String("https://www.googleapis.com/compute/v1/projects/project/global/images/image-1"),
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
			instanceTemplate: &computepb.InstanceTemplate{
				Name: proto.String("instance-template"),
				Properties: &computepb.InstanceProperties{
					Disks: []*computepb.AttachedDisk{
						{
							InitializeParams: &computepb.AttachedDiskInitializeParams{
								SourceImage: proto.String("https://www.googleapis.com/compute/v1/projects/project/global/images/image-1"),
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
