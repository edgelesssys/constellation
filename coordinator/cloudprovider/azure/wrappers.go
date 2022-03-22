package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type networkInterfacesClient struct {
	*armnetwork.InterfacesClient
}

func (c *networkInterfacesClient) GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string,
	virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
	options *armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceOptions,
) (armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse, error) {
	return c.InterfacesClient.GetVirtualMachineScaleSetNetworkInterface(ctx, resourceGroupName, virtualMachineScaleSetName, virtualmachineIndex, networkInterfaceName, options)
}

func (c *networkInterfacesClient) Get(ctx context.Context, resourceGroupName string, networkInterfaceName string,
	options *armnetwork.InterfacesClientGetOptions,
) (armnetwork.InterfacesClientGetResponse, error) {
	return c.InterfacesClient.Get(ctx, resourceGroupName, networkInterfaceName, options)
}

type virtualMachinesClient struct {
	*armcompute.VirtualMachinesClient
}

func (c *virtualMachinesClient) Get(ctx context.Context, resourceGroupName, vmName string, options *armcompute.VirtualMachinesClientGetOptions) (armcompute.VirtualMachinesClientGetResponse, error) {
	return c.VirtualMachinesClient.Get(ctx, resourceGroupName, vmName, options)
}

func (c *virtualMachinesClient) List(resourceGroupName string, options *armcompute.VirtualMachinesClientListOptions) virtualMachinesClientListPager {
	return c.VirtualMachinesClient.List(resourceGroupName, options)
}

type virtualMachineScaleSetVMsClient struct {
	*armcompute.VirtualMachineScaleSetVMsClient
}

func (c *virtualMachineScaleSetVMsClient) Get(ctx context.Context, resourceGroupName, vmScaleSetName, instanceID string, options *armcompute.VirtualMachineScaleSetVMsClientGetOptions) (armcompute.VirtualMachineScaleSetVMsClientGetResponse, error) {
	return c.VirtualMachineScaleSetVMsClient.Get(ctx, resourceGroupName, vmScaleSetName, instanceID, options)
}

func (c *virtualMachineScaleSetVMsClient) List(resourceGroupName, virtualMachineScaleSetName string, options *armcompute.VirtualMachineScaleSetVMsClientListOptions) virtualMachineScaleSetVMsClientListPager {
	return c.VirtualMachineScaleSetVMsClient.List(resourceGroupName, virtualMachineScaleSetName, options)
}

type tagsClient struct {
	*armresources.TagsClient
}

func (c *tagsClient) CreateOrUpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsResource, options *armresources.TagsClientCreateOrUpdateAtScopeOptions) (armresources.TagsClientCreateOrUpdateAtScopeResponse, error) {
	return c.TagsClient.CreateOrUpdateAtScope(ctx, scope, parameters, options)
}

func (c *tagsClient) UpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsPatchResource, options *armresources.TagsClientUpdateAtScopeOptions) (armresources.TagsClientUpdateAtScopeResponse, error) {
	return c.TagsClient.UpdateAtScope(ctx, scope, parameters, options)
}

type scaleSetsClient struct {
	*armcompute.VirtualMachineScaleSetsClient
}

func (c *scaleSetsClient) List(resourceGroupName string, options *armcompute.VirtualMachineScaleSetsClientListOptions) virtualMachineScaleSetsClientListPager {
	return c.VirtualMachineScaleSetsClient.List(resourceGroupName, options)
}
