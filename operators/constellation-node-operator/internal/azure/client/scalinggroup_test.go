/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetScalingGroupImage(t *testing.T) {
	testCases := map[string]struct {
		scalingGroupID string
		scaleSet       armcomputev2.VirtualMachineScaleSet
		getScaleSetErr error
		wantImage      string
		wantErr        bool
	}{
		"getting image works": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			scaleSet: armcomputev2.VirtualMachineScaleSet{
				Properties: &armcomputev2.VirtualMachineScaleSetProperties{
					VirtualMachineProfile: &armcomputev2.VirtualMachineScaleSetVMProfile{
						StorageProfile: &armcomputev2.VirtualMachineScaleSetStorageProfile{
							ImageReference: &armcomputev2.ImageReference{
								ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/images/image-name"),
							},
						},
					},
				},
			},
			wantImage: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/images/image-name",
		},
		"getting community image works": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			scaleSet: armcomputev2.VirtualMachineScaleSet{
				Properties: &armcomputev2.VirtualMachineScaleSetProperties{
					VirtualMachineProfile: &armcomputev2.VirtualMachineScaleSetVMProfile{
						StorageProfile: &armcomputev2.VirtualMachineScaleSetStorageProfile{
							ImageReference: &armcomputev2.ImageReference{
								CommunityGalleryImageID: to.Ptr("/CommunityGalleries/gallery-name/Images/image-name/Versions/1.2.3"),
							},
						},
					},
				},
			},
			wantImage: "/CommunityGalleries/gallery-name/Images/image-name/Versions/1.2.3",
		},
		"splitting scalingGroupID fails": {
			scalingGroupID: "invalid",
			wantErr:        true,
		},
		"get scale set fails": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			getScaleSetErr: errors.New("get scale set error"),
			wantErr:        true,
		},
		"scale set is invalid": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				scaleSetsAPI: &stubScaleSetsAPI{
					scaleSet: armcomputev2.VirtualMachineScaleSetsClientGetResponse{
						VirtualMachineScaleSet: tc.scaleSet,
					},
					getErr: tc.getScaleSetErr,
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
		scalingGroupID string
		imageURI       string
		updateErr      error
		resultErr      error
		wantErr        bool
	}{
		"setting image works": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			imageURI:       "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/images/image-name-2",
		},
		"splitting scalingGroupID fails": {
			scalingGroupID: "invalid",
			wantErr:        true,
		},
		"beginning update fails": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			imageURI:       "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/images/image-name-2",
			updateErr:      errors.New("update error"),
			wantErr:        true,
		},
		"retrieving polling result fails": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			imageURI:       "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/images/image-name-2",
			resultErr:      errors.New("result error"),
			wantErr:        true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				scaleSetsAPI: &stubScaleSetsAPI{
					updateErr: tc.updateErr,
					resultErr: tc.resultErr,
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
		"getting name works": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			wantName:       "scale-set-name",
		},
		"uppercase name isn't lowercased": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/SCALE-SET-NAME",
			wantName:       "SCALE-SET-NAME",
		},
		"splitting scalingGroupID fails": {
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
		scaleSet          armcomputev2.VirtualMachineScaleSet
		fetchPageErr      error
		wantControlPlanes []string
		wantWorkers       []string
		wantErr           bool
	}{
		"listing control-plane works": {
			scaleSet: armcomputev2.VirtualMachineScaleSet{
				ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/constellation-scale-set-control-planes-uid"),
				Tags: map[string]*string{
					"constellation-uid":  to.Ptr("uid"),
					"constellation-role": to.Ptr("control-plane"),
				},
			},
			wantControlPlanes: []string{"/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/constellation-scale-set-control-planes-uid"},
		},
		"listing worker works": {
			scaleSet: armcomputev2.VirtualMachineScaleSet{
				ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/constellation-scale-set-workers-uid"),
				Tags: map[string]*string{
					"constellation-uid":  to.Ptr("uid"),
					"constellation-role": to.Ptr("worker"),
				},
			},
			wantWorkers: []string{"/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/constellation-scale-set-workers-uid"},
		},
		"listing is not dependent on resource name": {
			scaleSet: armcomputev2.VirtualMachineScaleSet{
				ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/some-scale-set"),
				Tags: map[string]*string{
					"constellation-uid":  to.Ptr("uid"),
					"constellation-role": to.Ptr("control-plane"),
				},
			},
			wantControlPlanes: []string{"/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/some-scale-set"},
		},
		"listing other works": {
			scaleSet: armcomputev2.VirtualMachineScaleSet{
				ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/other"),
			},
		},
		"fetching scale sets fails": {
			fetchPageErr: errors.New("fetch page error"),
			wantErr:      true,
		},
		"scale set is invalid": {},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				scaleSetsAPI: &stubScaleSetsAPI{
					pager: &stubVMSSPager{
						list:     []armcomputev2.VirtualMachineScaleSet{tc.scaleSet},
						fetchErr: tc.fetchPageErr,
					},
				},
			}
			gotControlPlanes, gotWorkers, err := client.ListScalingGroups(context.Background(), "uid")
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantControlPlanes, gotControlPlanes)
			assert.ElementsMatch(tc.wantWorkers, gotWorkers)
		})
	}
}

func TestImageReferenceFromImage(t *testing.T) {
	testCases := map[string]struct {
		img             string
		wantID          *string
		wantCommunityID *string
	}{
		"ID": {
			img:             "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/1.5.0",
			wantID:          to.Ptr("/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/1.5.0"),
			wantCommunityID: nil,
		},
		"Community": {
			img:             "/CommunityGalleries/ConstellationCVM-728bd310-e898-4450-a1ed-21cf2fb0d735/Images/feat-azure-cvm-sharing/Versions/2022.0826.084922",
			wantID:          nil,
			wantCommunityID: to.Ptr("/CommunityGalleries/ConstellationCVM-728bd310-e898-4450-a1ed-21cf2fb0d735/Images/feat-azure-cvm-sharing/Versions/2022.0826.084922"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ref := imageReferenceFromImage(tc.img)

			assert.Equal(tc.wantID, ref.ID)
			assert.Equal(tc.wantCommunityID, ref.CommunityGalleryImageID)
		})
	}
}
