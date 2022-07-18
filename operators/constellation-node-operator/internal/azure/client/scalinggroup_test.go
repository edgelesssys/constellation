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
