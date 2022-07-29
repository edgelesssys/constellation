package client

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/edgelesssys/constellation/operators/constellation-node-operator/internal/poller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodeImage(t *testing.T) {
	testCases := map[string]struct {
		providerID       string
		vm               armcomputev2.VirtualMachineScaleSetVM
		getScaleSetVMErr error
		wantImage        string
		wantErr          bool
	}{
		"getting node image works": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			vm: armcomputev2.VirtualMachineScaleSetVM{
				Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
					StorageProfile: &armcomputev2.StorageProfile{
						ImageReference: &armcomputev2.ImageReference{
							ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/images/image-name"),
						},
					},
				},
			},
			wantImage: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/images/image-name",
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
			vm:         armcomputev2.VirtualMachineScaleSetVM{},
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				virtualMachineScaleSetVMsAPI: &stubvirtualMachineScaleSetVMsAPI{
					scaleSetVM: armcomputev2.VirtualMachineScaleSetVMsClientGetResponse{
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
		sku               *armcomputev2.SKU
		preexistingVMs    []armcomputev2.VirtualMachineScaleSetVM
		newVM             *armcomputev2.VirtualMachineScaleSetVM
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
			sku: &armcomputev2.SKU{
				Capacity: to.Ptr(int64(0)),
			},
			newVM: &armcomputev2.VirtualMachineScaleSetVM{
				ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/node-name"),
				Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
					OSProfile: &armcomputev2.OSProfile{
						ComputerName: to.Ptr("node-name"),
					},
				},
			},
			wantNodeName:   "node-name",
			wantProviderID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/node-name",
		},
		"creating node works with existing nodes": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			sku: &armcomputev2.SKU{
				Capacity: to.Ptr(int64(1)),
			},
			preexistingVMs: []armcomputev2.VirtualMachineScaleSetVM{
				{ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/preexisting-node")},
			},
			newVM: &armcomputev2.VirtualMachineScaleSetVM{
				ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/new-node"),
				Properties: &armcomputev2.VirtualMachineScaleSetVMProperties{
					OSProfile: &armcomputev2.OSProfile{
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
			sku:               &armcomputev2.SKU{Capacity: to.Ptr(int64(0))},
			updateScaleSetErr: errors.New("update scale set error"),
			wantEarlyErr:      true,
		},
		"polling for increased capacity fails": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			sku:            &armcomputev2.SKU{Capacity: to.Ptr(int64(0))},
			pollErr:        errors.New("poll error"),
			wantLateErr:    true,
		},
		"new node cannot be found": {
			scalingGroupID: "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			sku:            &armcomputev2.SKU{Capacity: to.Ptr(int64(0))},
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
			poller := NewStubCapacityPoller(tc.pollErr)
			client := Client{
				virtualMachineScaleSetVMsAPI: &stubvirtualMachineScaleSetVMsAPI{
					pager: pager,
				},
				scaleSetsAPI: &stubScaleSetsAPI{
					scaleSet: armcomputev2.VirtualMachineScaleSetsClientGetResponse{
						VirtualMachineScaleSet: armcomputev2.VirtualMachineScaleSet{
							SKU: tc.sku,
						},
					},
					getErr:    tc.getSKUCapacityErr,
					updateErr: tc.updateScaleSetErr,
				},
				capacityPollerGenerator: func(resourceGroup, scaleSet string, wantedCapacity int64) capacityPoller {
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

// TODO: test capacityPollingHandler

func TestCapacityPollingHandler(t *testing.T) {
	assert := assert.New(t)
	wantCapacity := int64(1)
	var gotCapacity int64
	handler := capacityPollingHandler{
		scaleSetsAPI: &stubScaleSetsAPI{
			scaleSet: armcomputev2.VirtualMachineScaleSetsClientGetResponse{
				VirtualMachineScaleSet: armcomputev2.VirtualMachineScaleSet{
					SKU: &armcomputev2.SKU{Capacity: to.Ptr(int64(0))},
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
	handler.scaleSetsAPI.(*stubScaleSetsAPI).scaleSet.SKU = &armcomputev2.SKU{Capacity: to.Ptr(wantCapacity)}
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

func NewStubCapacityPoller(pollErr error) *stubCapacityPoller {
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
