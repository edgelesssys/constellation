/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"errors"
	"math/rand"
	"testing"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGetNodeImage(t *testing.T) {
	testCases := map[string]struct {
		providerID     string
		attachedDisks  []*computepb.AttachedDisk
		disk           *computepb.Disk
		getInstanceErr error
		getDiskErr     error
		wantImage      string
		wantErr        bool
	}{
		"boot disk is found": {
			providerID: "gce://project/zone/instance-name",
			attachedDisks: []*computepb.AttachedDisk{
				{
					Source: proto.String("https://www.googleapis.com/compute/v1/projects/project/zones/zone/disks/disk"),
				},
			},
			disk: &computepb.Disk{
				SourceImage: proto.String("https://www.googleapis.com/compute/v1/projects/project/global/images/image"),
			},
			wantImage: "projects/project/global/images/image",
		},
		"splitting providerID fails": {
			providerID: "invalid",
			wantErr:    true,
		},
		"get instance fails": {
			providerID:     "gce://project/zone/instance-name",
			getInstanceErr: errors.New("get instance error"),
			wantErr:        true,
		},
		"instance has no disks": {
			providerID: "gce://project/zone/instance-name",
			wantErr:    true,
		},
		"attached disk is invalid": {
			providerID:    "gce://project/zone/instance-name",
			attachedDisks: []*computepb.AttachedDisk{{}},
			wantErr:       true,
		},
		"boot disk reference is invalid": {
			providerID: "gce://project/zone/instance-name",
			attachedDisks: []*computepb.AttachedDisk{{
				Source: proto.String("invalid"),
			}},
			wantErr: true,
		},
		"get disk fails": {
			providerID: "gce://project/zone/instance-name",
			attachedDisks: []*computepb.AttachedDisk{{
				Source: proto.String("https://www.googleapis.com/compute/v1/projects/project/zones/zone/disks/disk"),
			}},
			getDiskErr: errors.New("get disk error"),
			wantErr:    true,
		},
		"disk has no source image": {
			providerID: "gce://project/zone/instance-name",
			attachedDisks: []*computepb.AttachedDisk{{
				Source: proto.String("https://www.googleapis.com/compute/v1/projects/project/zones/zone/disks/disk"),
			}},
			disk:    &computepb.Disk{},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				instanceAPI: &stubInstanceAPI{
					getErr: tc.getInstanceErr,
					instance: &computepb.Instance{
						Disks: tc.attachedDisks,
					},
				},
				diskAPI: &stubDiskAPI{
					getErr: tc.getDiskErr,
					disk:   tc.disk,
				},
			}
			gotImage, err := client.GetNodeImage(t.Context(), tc.providerID)
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
	testCases := map[string]struct {
		providerID         string
		createdBy          string
		getInstanceErr     error
		wantScalingGroupID string
		wantErr            bool
	}{
		"scaling group is found": {
			providerID:         "gce://project/zone/instance-name",
			createdBy:          "projects/project/zones/zone/instanceGroupManagers/instance-group",
			wantScalingGroupID: "projects/project/zones/zone/instanceGroupManagers/instance-group",
		},
		"splitting providerID fails": {
			providerID: "invalid",
			wantErr:    true,
		},
		"get instance fails": {
			providerID:     "gce://project/zone/instance-name",
			getInstanceErr: errors.New("get instance error"),
			wantErr:        true,
		},
		"instance has no created-by": {
			providerID: "gce://project/zone/instance-name",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			instance := computepb.Instance{}
			if tc.createdBy != "" {
				instance.Metadata = &computepb.Metadata{
					Items: []*computepb.Items{
						{
							Key:   proto.String("created-by"),
							Value: proto.String(tc.createdBy),
						},
					},
				}
			}
			client := Client{
				instanceAPI: &stubInstanceAPI{
					getErr:   tc.getInstanceErr,
					instance: &instance,
				},
			}
			gotScalingGroupID, err := client.GetScalingGroupID(t.Context(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantScalingGroupID, gotScalingGroupID)
		})
	}
}

func TestCreateNode(t *testing.T) {
	testCases := map[string]struct {
		scalingGroupID             string
		baseInstanceName           *string
		getInstanceGroupManagerErr error
		createInstanceErr          error
		wantErr                    bool
	}{
		"scaling group is found": {
			scalingGroupID:   "projects/project/zones/zone/instanceGroupManagers/instance-group",
			baseInstanceName: proto.String("base-name"),
		},
		"splitting scalingGroupID fails": {
			scalingGroupID: "invalid",
			wantErr:        true,
		},
		"get instance group manager fails": {
			scalingGroupID:             "projects/project/zones/zone/instanceGroupManagers/instance-group",
			getInstanceGroupManagerErr: errors.New("get instance group manager error"),
			wantErr:                    true,
		},
		"instance group manager has no base instance name": {
			scalingGroupID: "projects/project/zones/zone/instanceGroupManagers/instance-group",
			wantErr:        true,
		},
		"create instance fails": {
			scalingGroupID:    "projects/project/zones/zone/instanceGroupManagers/instance-group",
			baseInstanceName:  proto.String("base-name"),
			createInstanceErr: errors.New("create instance error"),
			wantErr:           true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				instanceGroupManagersAPI: &stubInstanceGroupManagersAPI{
					getErr:             tc.getInstanceGroupManagerErr,
					createInstancesErr: tc.createInstanceErr,
					instanceGroupManager: &computepb.InstanceGroupManager{
						BaseInstanceName: tc.baseInstanceName,
					},
				},
				prng: rand.New(rand.NewSource(int64(time.Now().Nanosecond()))),
			}
			instanceName, providerID, err := client.CreateNode(t.Context(), tc.scalingGroupID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Contains(instanceName, "base-name")
			assert.Contains(providerID, "base-name")
		})
	}
}

func TestDeleteNode(t *testing.T) {
	testCases := map[string]struct {
		providerID        string
		scalingGroupID    string
		getInstanceErr    error
		deleteInstanceErr error
		wantErr           bool
	}{
		"node is deleted": {
			providerID:     "gce://project/zone/instance-name",
			scalingGroupID: "projects/project/zones/zone/instanceGroupManagers/instance-group",
		},
		"splitting providerID fails": {
			providerID: "invalid",
			wantErr:    true,
		},
		"get instance fails": {
			providerID:     "gce://project/zone/instance-name",
			getInstanceErr: errors.New("get instance error"),
			wantErr:        true,
		},
		"splitting scalingGroupID fails": {
			providerID:     "gce://project/zone/instance-name",
			scalingGroupID: "invalid",
			wantErr:        true,
		},
		"delete instance fails": {
			providerID:        "gce://project/zone/instance-name",
			scalingGroupID:    "projects/project/zones/zone/instanceGroupManagers/instance-group",
			deleteInstanceErr: errors.New("delete instance error"),
			wantErr:           true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				instanceGroupManagersAPI: &stubInstanceGroupManagersAPI{
					deleteInstancesErr: tc.deleteInstanceErr,
				},
				instanceAPI: &stubInstanceAPI{
					getErr: tc.getInstanceErr,
					instance: &computepb.Instance{
						Metadata: &computepb.Metadata{
							Items: []*computepb.Items{
								{Key: proto.String("created-by"), Value: &tc.scalingGroupID},
							},
						},
					},
				},
			}
			err := client.DeleteNode(t.Context(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}
