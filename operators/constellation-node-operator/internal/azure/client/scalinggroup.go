/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
)

// GetScalingGroupImage returns the image URI of the scaling group.
func (c *Client) GetScalingGroupImage(ctx context.Context, scalingGroupID string) (string, error) {
	_, resourceGroup, scaleSet, err := splitVMSSID(scalingGroupID)
	if err != nil {
		return "", err
	}
	res, err := c.scaleSetsAPI.Get(ctx, resourceGroup, scaleSet, nil)
	if err != nil {
		return "", err
	}
	if res.Properties == nil ||
		res.Properties.VirtualMachineProfile == nil ||
		res.Properties.VirtualMachineProfile.StorageProfile == nil ||
		res.Properties.VirtualMachineProfile.StorageProfile.ImageReference == nil ||
		res.Properties.VirtualMachineProfile.StorageProfile.ImageReference.ID == nil && res.Properties.VirtualMachineProfile.StorageProfile.ImageReference.CommunityGalleryImageID == nil {
		return "", fmt.Errorf("scalet set %q does not have valid image reference", scalingGroupID)
	}
	if res.Properties.VirtualMachineProfile.StorageProfile.ImageReference.ID != nil {
		return *res.Properties.VirtualMachineProfile.StorageProfile.ImageReference.ID, nil
	} else {
		return *res.Properties.VirtualMachineProfile.StorageProfile.ImageReference.CommunityGalleryImageID, nil
	}
}

// SetScalingGroupImage sets the image URI of the scaling group.
func (c *Client) SetScalingGroupImage(ctx context.Context, scalingGroupID, imageURI string) error {
	_, resourceGroup, scaleSet, err := splitVMSSID(scalingGroupID)
	if err != nil {
		return err
	}
	poller, err := c.scaleSetsAPI.BeginUpdate(ctx, resourceGroup, scaleSet, armcompute.VirtualMachineScaleSetUpdate{
		Properties: &armcompute.VirtualMachineScaleSetUpdateProperties{
			VirtualMachineProfile: &armcompute.VirtualMachineScaleSetUpdateVMProfile{
				StorageProfile: &armcompute.VirtualMachineScaleSetUpdateStorageProfile{
					ImageReference: imageReferenceFromImage(imageURI),
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}
	if _, err := poller.PollUntilDone(ctx, nil); err != nil {
		return err
	}
	return nil
}

// GetScalingGroupName retrieves the name of a scaling group.
func (c *Client) GetScalingGroupName(ctx context.Context, scalingGroupID string) (string, error) {
	_, _, scaleSet, err := splitVMSSID(scalingGroupID)
	if err != nil {
		return "", fmt.Errorf("getting scaling group name: %w", err)
	}
	return strings.ToLower(scaleSet), nil
}

// ListScalingGroups retrieves a list of scaling groups for the cluster.
func (c *Client) ListScalingGroups(ctx context.Context, uid string) (controlPlaneGroupIDs []string, workerGroupIDs []string, err error) {
	scaleSetIDs, err := c.getScaleSets(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("listing scaling groups: %w", err)
	}
	for _, scaleSetID := range scaleSetIDs {
		_, _, scaleSet, err := splitVMSSID(scaleSetID)
		if err != nil {
			return nil, nil, fmt.Errorf("getting scaling group name: %w", err)
		}
		if scaleSet == "constellation-scale-set-controlplanes-"+uid {
			controlPlaneGroupIDs = append(controlPlaneGroupIDs, scaleSetID)
		} else if strings.HasPrefix(scaleSet, "constellation-scale-set-workers-"+uid) {
			workerGroupIDs = append(workerGroupIDs, scaleSetID)
		}
	}
	return controlPlaneGroupIDs, workerGroupIDs, nil
}

func imageReferenceFromImage(img string) *armcompute.ImageReference {
	ref := &armcompute.ImageReference{}

	if strings.HasPrefix(img, "/CommunityGalleries") {
		ref.CommunityGalleryImageID = to.Ptr(img)
	} else {
		ref.ID = to.Ptr(img)
	}

	return ref
}
