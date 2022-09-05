/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/edgelesssys/constellation/operators/constellation-node-operator/internal/poller"
)

type virtualMachineScaleSetVMsAPI interface {
	Get(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string,
		options *armcomputev2.VirtualMachineScaleSetVMsClientGetOptions,
	) (armcomputev2.VirtualMachineScaleSetVMsClientGetResponse, error)
	GetInstanceView(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string,
		options *armcomputev2.VirtualMachineScaleSetVMsClientGetInstanceViewOptions,
	) (armcomputev2.VirtualMachineScaleSetVMsClientGetInstanceViewResponse, error)
	NewListPager(resourceGroupName string, virtualMachineScaleSetName string,
		options *armcomputev2.VirtualMachineScaleSetVMsClientListOptions,
	) *runtime.Pager[armcomputev2.VirtualMachineScaleSetVMsClientListResponse]
}

type scaleSetsAPI interface {
	Get(ctx context.Context, resourceGroupName string, vmScaleSetName string,
		options *armcomputev2.VirtualMachineScaleSetsClientGetOptions,
	) (armcomputev2.VirtualMachineScaleSetsClientGetResponse, error)
	BeginUpdate(ctx context.Context, resourceGroupName string, vmScaleSetName string, parameters armcomputev2.VirtualMachineScaleSetUpdate,
		options *armcomputev2.VirtualMachineScaleSetsClientBeginUpdateOptions,
	) (*runtime.Poller[armcomputev2.VirtualMachineScaleSetsClientUpdateResponse], error)
	BeginDeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs armcomputev2.VirtualMachineScaleSetVMInstanceRequiredIDs,
		options *armcomputev2.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions,
	) (*runtime.Poller[armcomputev2.VirtualMachineScaleSetsClientDeleteInstancesResponse], error)
	NewListPager(resourceGroupName string, options *armcomputev2.VirtualMachineScaleSetsClientListOptions,
	) *runtime.Pager[armcomputev2.VirtualMachineScaleSetsClientListResponse]
}

type capacityPoller interface {
	PollUntilDone(context.Context, *poller.PollUntilDoneOptions) (int64, error)
}
