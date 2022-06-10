package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type virtualNetworksClient struct {
	*armnetwork.VirtualNetworksClient
}

func (c *virtualNetworksClient) List(resourceGroupName string, options *armnetwork.VirtualNetworksClientListOptions) virtualNetworksClientListPager {
	return c.VirtualNetworksClient.List(resourceGroupName, options)
}

type securityGroupsClient struct {
	*armnetwork.SecurityGroupsClient
}

func (c *securityGroupsClient) List(resourceGroupName string, options *armnetwork.SecurityGroupsClientListOptions) securityGroupsClientListPager {
	return c.SecurityGroupsClient.List(resourceGroupName, options)
}

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

type publicIPAddressesClient struct {
	*armnetwork.PublicIPAddressesClient
}

func (c *publicIPAddressesClient) GetVirtualMachineScaleSetPublicIPAddress(ctx context.Context, resourceGroupName string,
	virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
	ipConfigurationName string, publicIPAddressName string,
	options *armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressOptions,
) (armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse, error) {
	return c.PublicIPAddressesClient.GetVirtualMachineScaleSetPublicIPAddress(ctx, resourceGroupName, virtualMachineScaleSetName, virtualmachineIndex, networkInterfaceName, ipConfigurationName, publicIPAddressName, options)
}

func (c *publicIPAddressesClient) Get(ctx context.Context, resourceGroupName string, publicIPAddressName string,
	options *armnetwork.PublicIPAddressesClientGetOptions,
) (armnetwork.PublicIPAddressesClientGetResponse, error) {
	return c.PublicIPAddressesClient.Get(ctx, resourceGroupName, publicIPAddressName, options)
}

type loadBalancersClient struct {
	*armnetwork.LoadBalancersClient
}

func (c *loadBalancersClient) List(resourceGroupName string, options *armnetwork.LoadBalancersClientListOptions) loadBalancersClientListPager {
	return c.LoadBalancersClient.List(resourceGroupName, options)
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

type applicationInsightsClient struct {
	*armapplicationinsights.ComponentsClient
}

func (c *applicationInsightsClient) Get(ctx context.Context, resourceGroupName string, resourceName string, options *armapplicationinsights.ComponentsClientGetOptions) (armapplicationinsights.ComponentsClientGetResponse, error) {
	return c.ComponentsClient.Get(ctx, resourceGroupName, resourceName, options)
}
