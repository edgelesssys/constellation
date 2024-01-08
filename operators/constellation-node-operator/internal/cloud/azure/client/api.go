/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/poller"
)

type virtualMachineScaleSetVMsAPI interface {
	Get(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string,
		options *armcompute.VirtualMachineScaleSetVMsClientGetOptions,
	) (armcompute.VirtualMachineScaleSetVMsClientGetResponse, error)
	GetInstanceView(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string,
		options *armcompute.VirtualMachineScaleSetVMsClientGetInstanceViewOptions,
	) (armcompute.VirtualMachineScaleSetVMsClientGetInstanceViewResponse, error)
	NewListPager(resourceGroupName string, virtualMachineScaleSetName string,
		options *armcompute.VirtualMachineScaleSetVMsClientListOptions,
	) *runtime.Pager[armcompute.VirtualMachineScaleSetVMsClientListResponse]
}

type scaleSetsAPI interface {
	Get(ctx context.Context, resourceGroupName string, vmScaleSetName string,
		options *armcompute.VirtualMachineScaleSetsClientGetOptions,
	) (armcompute.VirtualMachineScaleSetsClientGetResponse, error)
	BeginUpdate(ctx context.Context, resourceGroupName string, vmScaleSetName string, parameters armcompute.VirtualMachineScaleSetUpdate,
		options *armcompute.VirtualMachineScaleSetsClientBeginUpdateOptions,
	) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientUpdateResponse], error)
	BeginDeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs,
		options *armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions,
	) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse], error)
	NewListPager(resourceGroupName string, options *armcompute.VirtualMachineScaleSetsClientListOptions,
	) *runtime.Pager[armcompute.VirtualMachineScaleSetsClientListResponse]
}

type capacityPoller interface {
	PollUntilDone(context.Context, *poller.PollUntilDoneOptions) (int64, error)
}
