package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type imdsAPI interface {
	Retrieve(ctx context.Context) (metadataResponse, error)
}

type virtualNetworksClientListPager interface {
	NextPage(ctx context.Context) bool
	PageResponse() armnetwork.VirtualNetworksClientListResponse
}

type virtualNetworksAPI interface {
	List(resourceGroupName string, options *armnetwork.VirtualNetworksClientListOptions) virtualNetworksClientListPager
}

type securityGroupsClientListPager interface {
	NextPage(ctx context.Context) bool
	PageResponse() armnetwork.SecurityGroupsClientListResponse
}

type securityGroupsAPI interface {
	List(resourceGroupName string, options *armnetwork.SecurityGroupsClientListOptions) securityGroupsClientListPager
}

type networkInterfacesAPI interface {
	GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string,
		virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
		options *armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceOptions,
	) (armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse, error)
	Get(ctx context.Context, resourceGroupName string, networkInterfaceName string,
		options *armnetwork.InterfacesClientGetOptions) (armnetwork.InterfacesClientGetResponse, error)
}

type publicIPAddressesAPI interface {
	GetVirtualMachineScaleSetPublicIPAddress(ctx context.Context, resourceGroupName string,
		virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
		ipConfigurationName string, publicIPAddressName string,
		options *armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressOptions,
	) (armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse, error)
	Get(ctx context.Context, resourceGroupName string, publicIPAddressName string,
		options *armnetwork.PublicIPAddressesClientGetOptions) (armnetwork.PublicIPAddressesClientGetResponse, error)
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

type loadBalancersClientListPager interface {
	NextPage(ctx context.Context) bool
	PageResponse() armnetwork.LoadBalancersClientListResponse
}

type loadBalancerAPI interface {
	List(resourceGroupName string, options *armnetwork.LoadBalancersClientListOptions) loadBalancersClientListPager
}

type virtualMachineScaleSetsClientListPager interface {
	NextPage(ctx context.Context) bool
	PageResponse() armcompute.VirtualMachineScaleSetsClientListResponse
}

type tagsAPI interface {
	CreateOrUpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsResource, options *armresources.TagsClientCreateOrUpdateAtScopeOptions) (armresources.TagsClientCreateOrUpdateAtScopeResponse, error)
	UpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsPatchResource, options *armresources.TagsClientUpdateAtScopeOptions) (armresources.TagsClientUpdateAtScopeResponse, error)
}

type applicationInsightsAPI interface {
	Get(ctx context.Context, resourceGroupName string, resourceName string, options *armapplicationinsights.ComponentsClientGetOptions) (armapplicationinsights.ComponentsClientGetResponse, error)
}
