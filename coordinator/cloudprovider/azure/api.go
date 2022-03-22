package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type imdsAPI interface {
	Retrieve(ctx context.Context) (metadataResponse, error)
}

type networkInterfacesAPI interface {
	GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string,
		virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
		options *armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceOptions,
	) (armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse, error)
	Get(ctx context.Context, resourceGroupName string, networkInterfaceName string,
		options *armnetwork.InterfacesClientGetOptions) (armnetwork.InterfacesClientGetResponse, error)
}

type virtualMachinesAPI interface {
	Get(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientGetOptions) (armcompute.VirtualMachinesClientGetResponse, error)
	List(resourceGroupName string, options *armcompute.VirtualMachinesClientListOptions) virtualMachinesClientListPager
}

type virtualMachinesClientListPager interface {
	NextPage(ctx context.Context) bool
	PageResponse() armcompute.VirtualMachinesClientListResponse
}

type virtualMachineScaleSetVMsAPI interface {
	Get(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string, options *armcompute.VirtualMachineScaleSetVMsClientGetOptions) (armcompute.VirtualMachineScaleSetVMsClientGetResponse, error)
	List(resourceGroupName string, virtualMachineScaleSetName string, options *armcompute.VirtualMachineScaleSetVMsClientListOptions) virtualMachineScaleSetVMsClientListPager
}

type virtualMachineScaleSetVMsClientListPager interface {
	NextPage(ctx context.Context) bool
	PageResponse() armcompute.VirtualMachineScaleSetVMsClientListResponse
}

type scaleSetsAPI interface {
	List(resourceGroupName string, options *armcompute.VirtualMachineScaleSetsClientListOptions) virtualMachineScaleSetsClientListPager
}

type virtualMachineScaleSetsClientListPager interface {
	NextPage(ctx context.Context) bool
	PageResponse() armcompute.VirtualMachineScaleSetsClientListResponse
}

type tagsAPI interface {
	CreateOrUpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsResource, options *armresources.TagsClientCreateOrUpdateAtScopeOptions) (armresources.TagsClientCreateOrUpdateAtScopeResponse, error)
	UpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsPatchResource, options *armresources.TagsClientUpdateAtScopeOptions) (armresources.TagsClientUpdateAtScopeResponse, error)
}
