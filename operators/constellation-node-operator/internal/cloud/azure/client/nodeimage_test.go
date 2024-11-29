/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/poller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodeImage(t *testing.T) {
	testCases := map[string]struct {
		providerID       string
		vm               armcompute.VirtualMachineScaleSetVM
		getScaleSetVMErr error
		wantImage        string
		wantErr          bool
	}{
		"getting node image works": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			vm: armcompute.VirtualMachineScaleSetVM{
				Properties: &armcompute.VirtualMachineScaleSetVMProperties{
					StorageProfile: &armcompute.StorageProfile{
						ImageReference: &armcompute.ImageReference{
							ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/images/image-name"),
						},
					},
				},
			},
			wantImage: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/images/image-name",
		},
		"getting community node image works": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			vm: armcompute.VirtualMachineScaleSetVM{
				Properties: &armcompute.VirtualMachineScaleSetVMProperties{
					StorageProfile: &armcompute.StorageProfile{
						ImageReference: &armcompute.ImageReference{
							CommunityGalleryImageID: to.Ptr("/CommunityGalleries/gallery-name/Images/image-name/Versions/1.2.3"),
						},
					},
				},
			},
			wantImage: "/CommunityGalleries/gallery-name/Images/image-name/Versions/1.2.3",
		},
		"getting marketplace image works": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			vm: armcompute.VirtualMachineScaleSetVM{
				Properties: &armcompute.VirtualMachineScaleSetVMProperties{
					StorageProfile: &armcompute.StorageProfile{
						ImageReference: &armcompute.ImageReference{
							Publisher: to.Ptr("edgelesssystems"),
							Offer:     to.Ptr("constellation"),
							SKU:       to.Ptr("constellation"),
							Version:   to.Ptr("2.14.2"),
						},
					},
				},
			},
			wantImage: "constellation-marketplace-image://Azure?offer=constellation&publisher=edgelesssystems&sku=constellation&version=2.14.2",
		},
		"splitting providerID fails": {
			providerID: "invalid",
			wantErr:    true,
		},
		"get scale set vm fails": {
			providerID:       "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			getScaleSetVMErr: errors.New("get scale set vm error"),
			wantErr:          true,
		},
		"scale set vm does not have valid image reference": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			vm:         armcompute.VirtualMachineScaleSetVM{},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				virtualMachineScaleSetVMsAPI: &stubvirtualMachineScaleSetVMsAPI{
					scaleSetVM: armcompute.VirtualMachineScaleSetVMsClientGetResponse{
						VirtualMachineScaleSetVM: tc.vm,
					},
					getErr: tc.getScaleSetVMErr,
				},
			}
			gotImage, err := client.GetNodeImage(context.Background(), tc.providerID)
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
		wantScalingGroupID string
		wantErr            bool
	}{
		"getting scaling group id works": {
			providerID:         "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			wantScalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
		},
		"splitting providerID fails": {
			providerID: "invalid",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{}
			gotScalingGroupID, err := client.GetScalingGroupID(context.Background(), tc.providerID)
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
		scalingGroupID    string
		sku               *armcompute.SKU
		preexistingVMs    []armcompute.VirtualMachineScaleSetVM
		newVM             *armcompute.VirtualMachineScaleSetVM
		fetchErr          error
		pollErr           error
		getSKUCapacityErr error
		updateScaleSetErr error
		wantNodeName      string
		wantProviderID    string
		wantEarlyErr      bool
		wantLateErr       bool
	}{
		"creating node works": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			sku: &armcompute.SKU{
				Capacity: to.Ptr(int64(0)),
			},
			newVM: &armcompute.VirtualMachineScaleSetVM{
				ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/node-name"),
				Properties: &armcompute.VirtualMachineScaleSetVMProperties{
					OSProfile: &armcompute.OSProfile{
						ComputerName: to.Ptr("node-name"),
					},
				},
			},
			wantNodeName:   "node-name",
			wantProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/node-name",
		},
		"creating node works with existing nodes": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			sku: &armcompute.SKU{
				Capacity: to.Ptr(int64(1)),
			},
			preexistingVMs: []armcompute.VirtualMachineScaleSetVM{
				{ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/preexisting-node")},
			},
			newVM: &armcompute.VirtualMachineScaleSetVM{
				ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/new-node"),
				Properties: &armcompute.VirtualMachineScaleSetVMProperties{
					OSProfile: &armcompute.OSProfile{
						ComputerName: to.Ptr("new-node"),
					},
				},
			},
			wantNodeName:   "new-node",
			wantProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/new-node",
		},
		"splitting scalingGroupID fails": {
			scalingGroupID: "invalid",
			wantEarlyErr:   true,
		},
		"getting node list fails": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			fetchErr:       errors.New("get node list error"),
			wantEarlyErr:   true,
		},
		"getting SKU capacity fails": {
			scalingGroupID:    "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			getSKUCapacityErr: errors.New("get sku capacity error"),
			wantEarlyErr:      true,
		},
		"sku is invalid": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			wantEarlyErr:   true,
		},
		"updating scale set capacity fails": {
			scalingGroupID:    "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			sku:               &armcompute.SKU{Capacity: to.Ptr(int64(0))},
			updateScaleSetErr: errors.New("update scale set error"),
			wantEarlyErr:      true,
		},
		"polling for increased capacity fails": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			sku:            &armcompute.SKU{Capacity: to.Ptr(int64(0))},
			pollErr:        errors.New("poll error"),
			wantLateErr:    true,
		},
		"new node cannot be found": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			sku:            &armcompute.SKU{Capacity: to.Ptr(int64(0))},
			wantLateErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			pager := &stubVMSSVMPager{
				list:     tc.preexistingVMs,
				fetchErr: tc.fetchErr,
			}
			poller := newStubCapacityPoller(tc.pollErr)
			client := Client{
				virtualMachineScaleSetVMsAPI: &stubvirtualMachineScaleSetVMsAPI{
					pager: pager,
				},
				scaleSetsAPI: &stubScaleSetsAPI{
					scaleSet: armcompute.VirtualMachineScaleSetsClientGetResponse{
						VirtualMachineScaleSet: armcompute.VirtualMachineScaleSet{
							SKU: tc.sku,
						},
					},
					getErr:    tc.getSKUCapacityErr,
					updateErr: tc.updateScaleSetErr,
				},
				capacityPollerGenerator: func(_, _ string, _ int64) capacityPoller {
					return poller
				},
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			var gotNodeName, gotProviderID string
			var createErr error
			go func() {
				defer wg.Done()
				gotNodeName, gotProviderID, createErr = client.CreateNode(context.Background(), tc.scalingGroupID)
			}()

			// want error before PollUntilDone is called
			if tc.wantEarlyErr {
				wg.Wait()
				assert.Error(createErr)
				return
			}

			// wait for PollUntilDone to be called
			<-poller.pollC
			// update list of nodes
			if tc.newVM != nil {
				pager.list = append(pager.list, *tc.newVM)
			}
			// let PollUntilDone finish
			poller.doneC <- struct{}{}

			wg.Wait()
			if tc.wantLateErr {
				assert.Error(createErr)
				return
			}
			require.NoError(createErr)
			assert.Equal(tc.wantNodeName, gotNodeName)
			assert.Equal(tc.wantProviderID, gotProviderID)
		})
	}
}

func TestDeleteNode(t *testing.T) {
	testCases := map[string]struct {
		providerID string
		deleteErr  error
		wantErr    bool
	}{
		"deleting node works": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
		},
		"invalid providerID": {
			providerID: "invalid",
			wantErr:    true,
		},
		"deleting node fails": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			deleteErr:  errors.New("delete error"),
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			client := Client{
				scaleSetsAPI: &stubScaleSetsAPI{deleteErr: tc.deleteErr},
			}
			err := client.DeleteNode(context.Background(), tc.providerID)
			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

func TestCapacityPollingHandler(t *testing.T) {
	assert := assert.New(t)
	wantCapacity := int64(1)
	var gotCapacity int64
	handler := capacityPollingHandler{
		scaleSetsAPI: &stubScaleSetsAPI{
			scaleSet: armcompute.VirtualMachineScaleSetsClientGetResponse{
				VirtualMachineScaleSet: armcompute.VirtualMachineScaleSet{
					SKU: &armcompute.SKU{Capacity: to.Ptr(int64(0))},
				},
			},
		},
		wantedCapacity: wantCapacity,
	}
	assert.NoError(handler.Poll(context.Background()))
	assert.False(handler.Done())
	// Calling Result early should error
	assert.Error(handler.Result(context.Background(), &gotCapacity))

	// let scaleSet API error
	handler.scaleSetsAPI.(*stubScaleSetsAPI).getErr = errors.New("get error")
	assert.Error(handler.Poll(context.Background()))
	handler.scaleSetsAPI.(*stubScaleSetsAPI).getErr = nil

	// let scaleSet API return invalid SKU
	handler.scaleSetsAPI.(*stubScaleSetsAPI).scaleSet.SKU = nil
	assert.Error(handler.Poll(context.Background()))

	// let Poll finish
	handler.scaleSetsAPI.(*stubScaleSetsAPI).scaleSet.SKU = &armcompute.SKU{Capacity: to.Ptr(wantCapacity)}
	assert.NoError(handler.Poll(context.Background()))
	assert.True(handler.Done())
	assert.NoError(handler.Result(context.Background(), &gotCapacity))
	assert.Equal(wantCapacity, gotCapacity)
}

type stubCapacityPoller struct {
	pollErr error
	pollC   chan struct{}
	doneC   chan struct{}
}

func newStubCapacityPoller(pollErr error) *stubCapacityPoller {
	return &stubCapacityPoller{
		pollErr: pollErr,
		pollC:   make(chan struct{}),
		doneC:   make(chan struct{}),
	}
}

func (p *stubCapacityPoller) PollUntilDone(context.Context, *poller.PollUntilDoneOptions) (int64, error) {
	p.pollC <- struct{}{}
	<-p.doneC
	return 0, p.pollErr
}
