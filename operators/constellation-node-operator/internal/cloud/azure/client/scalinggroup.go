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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	updatev1alpha1 "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/api/v1alpha1"
	cspapi "github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/v2/internal/cloud/api"
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
	}
	return *res.Properties.VirtualMachineProfile.StorageProfile.ImageReference.CommunityGalleryImageID, nil
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

// GetScalingGroupName retrieves the name of a scaling group, as expected by Kubernetes.
// This keeps the casing of the original name, but Kubernetes requires the name to be lowercase,
// so use strings.ToLower() on the result if using the name in a Kubernetes context.
func (c *Client) GetScalingGroupName(scalingGroupID string) (string, error) {
	_, _, scaleSet, err := splitVMSSID(scalingGroupID)
	if err != nil {
		return "", fmt.Errorf("getting scaling group name: %w", err)
	}
	return scaleSet, nil
}

// GetAutoscalingGroupName retrieves the name of a scaling group as needed by the cluster-autoscaler.
func (c *Client) GetAutoscalingGroupName(scalingGroupID string) (string, error) {
	return c.GetScalingGroupName(scalingGroupID)
}

// ListScalingGroups retrieves a list of scaling groups for the cluster.
func (c *Client) ListScalingGroups(ctx context.Context, uid string) ([]cspapi.ScalingGroup, error) {
	results := []cspapi.ScalingGroup{}
	pager := c.scaleSetsAPI.NewListPager(c.config.ResourceGroup, nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("paging scale sets: %w", err)
		}
		for _, scaleSet := range page.Value {
			if scaleSet == nil || scaleSet.ID == nil {
				continue
			}
			if scaleSet.Tags == nil ||
				scaleSet.Tags["constellation-uid"] == nil ||
				*scaleSet.Tags["constellation-uid"] != uid ||
				scaleSet.Tags["constellation-role"] == nil {
				continue
			}

			role := updatev1alpha1.NodeRoleFromString(*scaleSet.Tags["constellation-role"])

			name, err := c.GetScalingGroupName(*scaleSet.ID)
			if err != nil {
				return nil, fmt.Errorf("getting scaling group name: %w", err)
			}

			var nodeGroupName string
			if scaleSet.Tags["constellation-node-group"] != nil {
				nodeGroupName = *scaleSet.Tags["constellation-node-group"]
			}
			// fallback for legacy clusters
			// TODO(malt3): remove this fallback once we can assume all clusters have the correct labels
			if nodeGroupName == "" {
				switch role {
				case updatev1alpha1.ControlPlaneRole:
					nodeGroupName = constants.ControlPlaneDefault
				case updatev1alpha1.WorkerRole:
					nodeGroupName = constants.WorkerDefault
				}
			}

			autoscalerGroupName, err := c.GetAutoscalingGroupName(*scaleSet.ID)
			if err != nil {
				return nil, fmt.Errorf("getting autoscaling group name: %w", err)
			}

			results = append(results, cspapi.ScalingGroup{
				Name:                 name,
				NodeGroupName:        nodeGroupName,
				GroupID:              *scaleSet.ID,
				AutoscalingGroupName: autoscalerGroupName,
				Role:                 role,
			})
		}
	}
	return results, nil
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
