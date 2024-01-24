/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/edgelesssys/constellation/v2/internal/mpimage"
)

// GetNodeImage returns the image name of the node.
func (c *Client) GetNodeImage(ctx context.Context, providerID string) (string, error) {
	_, resourceGroup, scaleSet, instanceID, err := scaleSetInformationFromProviderID(providerID)
	if err != nil {
		return "", err
	}
	resp, err := c.virtualMachineScaleSetVMsAPI.Get(ctx, resourceGroup, scaleSet, instanceID, nil)
	if err != nil {
		return "", err
	}
	if resp.Properties == nil ||
		resp.Properties.StorageProfile == nil ||
		resp.Properties.StorageProfile.ImageReference == nil ||
		resp.Properties.StorageProfile.ImageReference.ID == nil &&
			resp.Properties.StorageProfile.ImageReference.CommunityGalleryImageID == nil &&
			(resp.Properties.StorageProfile.ImageReference.Publisher == nil ||
				resp.Properties.StorageProfile.ImageReference.Offer == nil ||
				resp.Properties.StorageProfile.ImageReference.SKU == nil ||
				resp.Properties.StorageProfile.ImageReference.Version == nil) {
		return "", fmt.Errorf("node %q does not have valid image reference", providerID)
	}

	// Image ID is set, return it.
	if resp.Properties.StorageProfile.ImageReference.ID != nil {
		return *resp.Properties.StorageProfile.ImageReference.ID, nil
	}

	// Community Gallery image ID is set, return it.
	if resp.Properties.StorageProfile.ImageReference.CommunityGalleryImageID != nil {
		return *resp.Properties.StorageProfile.ImageReference.CommunityGalleryImageID, nil
	}

	// Last possible option: Marketplace Image is used, format it to an URI and return it.
	return mpimage.AzureMarketplaceImage{
		Publisher: *resp.Properties.StorageProfile.ImageReference.Publisher,
		Offer:     *resp.Properties.StorageProfile.ImageReference.Offer,
		SKU:       *resp.Properties.StorageProfile.ImageReference.SKU,
		Version:   *resp.Properties.StorageProfile.ImageReference.Version,
	}.URI(), nil
}

// GetScalingGroupID returns the scaling group ID of the node.
func (c *Client) GetScalingGroupID(_ context.Context, providerID string) (string, error) {
	subscriptionID, resourceGroup, scaleSet, _, err := scaleSetInformationFromProviderID(providerID)
	if err != nil {
		return "", err
	}
	return joinVMSSID(subscriptionID, resourceGroup, scaleSet), nil
}

// CreateNode creates a node in the specified scaling group.
func (c *Client) CreateNode(ctx context.Context, scalingGroupID string) (nodeName, providerID string, err error) {
	_, resourceGroup, scaleSet, err := splitVMSSID(scalingGroupID)
	if err != nil {
		return "", "", err
	}

	// get list of instance IDs before scaling,
	var oldVMIDs []string
	pager := c.virtualMachineScaleSetVMsAPI.NewListPager(resourceGroup, scaleSet, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return "", "", err
		}
		for _, vm := range page.Value {
			if vm == nil || vm.ID == nil {
				continue
			}
			oldVMIDs = append(oldVMIDs, *vm.ID)
		}
	}

	// increase the number of instances by one
	resp, err := c.scaleSetsAPI.Get(ctx, resourceGroup, scaleSet, nil)
	if err != nil {
		return "", "", err
	}
	if resp.SKU == nil || resp.SKU.Capacity == nil {
		return "", "", fmt.Errorf("scale set %q does not have valid capacity", scaleSet)
	}
	wantedCapacity := *resp.SKU.Capacity + 1
	_, err = c.scaleSetsAPI.BeginUpdate(ctx, resourceGroup, scaleSet, armcompute.VirtualMachineScaleSetUpdate{
		SKU: &armcompute.SKU{
			Capacity: &wantedCapacity,
		},
	}, nil)
	if err != nil {
		return "", "", err
	}

	poller := c.capacityPollerGenerator(resourceGroup, scaleSet, wantedCapacity)
	if _, err := poller.PollUntilDone(ctx, c.pollerOptions); err != nil {
		return "", "", err
	}

	// get the list of instances again
	// and find the new instance id by comparing the old and new lists
	pager = c.virtualMachineScaleSetVMsAPI.NewListPager(resourceGroup, scaleSet, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return "", "", err
		}
		for _, vm := range page.Value {
			// check if the instance already existed in the old list
			if !hasKnownVMID(vm, oldVMIDs) {
				return strings.ToLower(*vm.Properties.OSProfile.ComputerName), "azure://" + *vm.ID, nil
			}
		}
	}
	return "", "", fmt.Errorf("failed to find new node after scaling up")
}

// DeleteNode deletes a node specified by its provider ID.
func (c *Client) DeleteNode(ctx context.Context, providerID string) error {
	_, resourceGroup, scaleSet, instanceID, err := scaleSetInformationFromProviderID(providerID)
	if err != nil {
		return err
	}
	ids := armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{InstanceIDs: []*string{&instanceID}}
	_, err = c.scaleSetsAPI.BeginDeleteInstances(ctx, resourceGroup, scaleSet, ids, nil)
	return err
}

// capacityPollingHandler polls a scale set
// until its capacity reaches the desired value.
type capacityPollingHandler struct {
	done           bool
	wantedCapacity int64
	resourceGroup  string
	scaleSet       string
	scaleSetsAPI
}

func (h *capacityPollingHandler) Done() bool {
	return h.done
}

func (h *capacityPollingHandler) Poll(ctx context.Context) error {
	resp, err := h.scaleSetsAPI.Get(ctx, h.resourceGroup, h.scaleSet, nil)
	if err != nil {
		return err
	}
	if resp.SKU == nil || resp.SKU.Capacity == nil {
		return fmt.Errorf("scale set %q does not have valid capacity", h.scaleSet)
	}
	h.done = *resp.SKU.Capacity == h.wantedCapacity
	return nil
}

func (h *capacityPollingHandler) Result(_ context.Context, out *int64) error {
	if !h.done {
		return fmt.Errorf("failed to scale up")
	}
	*out = h.wantedCapacity
	return nil
}

// hasKnownVMID returns true if the vmID is found in the vm ID list.
func hasKnownVMID(vm *armcompute.VirtualMachineScaleSetVM, vmIDs []string) bool {
	for _, id := range vmIDs {
		if vm != nil && vm.ID != nil && *vm.ID == id {
			return true
		}
	}
	return false
}
