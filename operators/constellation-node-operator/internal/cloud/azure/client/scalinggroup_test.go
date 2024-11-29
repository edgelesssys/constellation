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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	cspapi "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/cloud/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetScalingGroupImage(t *testing.T) {
	testCases := map[string]struct {
		scalingGroupID string
		scaleSet       armcompute.VirtualMachineScaleSet
		getScaleSetErr error
		wantImage      string
		wantErr        bool
	}{
		"getting image works": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			scaleSet: armcompute.VirtualMachineScaleSet{
				Properties: &armcompute.VirtualMachineScaleSetProperties{
					VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
						StorageProfile: &armcompute.VirtualMachineScaleSetStorageProfile{
							ImageReference: &armcompute.ImageReference{
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
			scaleSet: armcompute.VirtualMachineScaleSet{
				Properties: &armcompute.VirtualMachineScaleSetProperties{
					VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
						StorageProfile: &armcompute.VirtualMachineScaleSetStorageProfile{
							ImageReference: &armcompute.ImageReference{
								CommunityGalleryImageID: to.Ptr("/communityGalleries/gallery-name/Images/image-name/Versions/1.2.3"),
							},
						},
					},
				},
			},
			wantImage: "/communityGalleries/gallery-name/Images/image-name/Versions/1.2.3",
		},
		"getting marketplace image works": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			scaleSet: armcompute.VirtualMachineScaleSet{
				Properties: &armcompute.VirtualMachineScaleSetProperties{
					VirtualMachineProfile: &armcompute.VirtualMachineScaleSetVMProfile{
						StorageProfile: &armcompute.VirtualMachineScaleSetStorageProfile{
							ImageReference: &armcompute.ImageReference{
								Publisher: to.Ptr("edgelesssystems"),
								Offer:     to.Ptr("constellation"),
								SKU:       to.Ptr("constellation"),
								Version:   to.Ptr("2.14.2"),
							},
						},
					},
				},
			},
			wantImage: "constellation-marketplace-image://Azure?offer=constellation&publisher=edgelesssystems&sku=constellation&version=2.14.2",
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
					scaleSet: armcompute.VirtualMachineScaleSetsClientGetResponse{
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
		scaleSet     armcompute.VirtualMachineScaleSet
		fetchPageErr error
		wantGroups   []cspapi.ScalingGroup
		wantErr      bool
	}{
		"listing control-plane works": {
			scaleSet: armcompute.VirtualMachineScaleSet{
				Name: to.Ptr("constellation-scale-set-control-planes-uid"),
				ID:   to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/constellation-scale-set-control-planes-uid"),
				Tags: map[string]*string{
					"constellation-uid":  to.Ptr("uid"),
					"constellation-role": to.Ptr("control-plane"),
				},
			},
			wantGroups: []cspapi.ScalingGroup{
				{
					Name:                 "constellation-scale-set-control-planes-uid",
					NodeGroupName:        constants.ControlPlaneDefault,
					GroupID:              "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/constellation-scale-set-control-planes-uid",
					AutoscalingGroupName: "constellation-scale-set-control-planes-uid",
					Role:                 "ControlPlane",
				},
			},
		},
		"listing worker works": {
			scaleSet: armcompute.VirtualMachineScaleSet{
				Name: to.Ptr("constellation-scale-set-workers-uid"),
				ID:   to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/constellation-scale-set-workers-uid"),
				Tags: map[string]*string{
					"constellation-uid":  to.Ptr("uid"),
					"constellation-role": to.Ptr("worker"),
				},
			},
			wantGroups: []cspapi.ScalingGroup{
				{
					Name:                 "constellation-scale-set-workers-uid",
					NodeGroupName:        constants.WorkerDefault,
					GroupID:              "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/constellation-scale-set-workers-uid",
					AutoscalingGroupName: "constellation-scale-set-workers-uid",
					Role:                 "Worker",
				},
			},
		},
		"listing is not dependent on resource name": {
			scaleSet: armcompute.VirtualMachineScaleSet{
				Name: to.Ptr("foo-bar"),
				ID:   to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/some-scale-set"),
				Tags: map[string]*string{
					"constellation-uid":  to.Ptr("uid"),
					"constellation-role": to.Ptr("control-plane"),
				},
			},
			wantGroups: []cspapi.ScalingGroup{
				{
					Name:                 "some-scale-set",
					NodeGroupName:        constants.ControlPlaneDefault,
					GroupID:              "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/some-scale-set",
					AutoscalingGroupName: "some-scale-set",
					Role:                 "ControlPlane",
				},
			},
		},
		"listing other works": {
			scaleSet: armcompute.VirtualMachineScaleSet{
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
						list:     []armcompute.VirtualMachineScaleSet{tc.scaleSet},
						fetchErr: tc.fetchPageErr,
					},
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

func TestImageReferenceFromImage(t *testing.T) {
	testCases := map[string]struct {
		img             string
		wantID          *string
		wantCommunityID *string
		wantPublisher   *string
		wantOffer       *string
		wantSKU         *string
		wantVersion     *string
	}{
		"ID": {
			img:    "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/1.5.0",
			wantID: to.Ptr("/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/constellation-images/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/1.5.0"),
		},
		"Community": {
			img:             "/communityGalleries/ConstellationCVM-728bd310-e898-4450-a1ed-21cf2fb0d735/Images/feat-azure-cvm-sharing/Versions/2022.0826.084922",
			wantCommunityID: to.Ptr("/communityGalleries/ConstellationCVM-728bd310-e898-4450-a1ed-21cf2fb0d735/Images/feat-azure-cvm-sharing/Versions/2022.0826.084922"),
		},
		"Marketplace": {
			img:           "constellation-marketplace-image://Azure?offer=constellation&publisher=edgelesssystems&sku=constellation&version=1.2.3",
			wantPublisher: to.Ptr("edgelesssystems"),
			wantOffer:     to.Ptr("constellation"),
			wantSKU:       to.Ptr("constellation"),
			wantVersion:   to.Ptr("1.2.3"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ref, err := imageReferenceFromImage(tc.img)
			assert.NoError(err)

			assert.Equal(tc.wantID, ref.ID)
			assert.Equal(tc.wantCommunityID, ref.CommunityGalleryImageID)
			assert.Equal(tc.wantPublisher, ref.Publisher)
			assert.Equal(tc.wantOffer, ref.Offer)
			assert.Equal(tc.wantSKU, ref.SKU)
			assert.Equal(tc.wantVersion, ref.Version)
		})
	}
}
