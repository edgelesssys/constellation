package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type imdsAPI interface {
	Retrieve(ctx context.Context) (metadataResponse, error)
}

type virtualNetworksAPI interface {
	NewListPager(resourceGroupName string, options *armnetwork.VirtualNetworksClientListOptions) *runtime.Pager[armnetwork.VirtualNetworksClientListResponse]
}

type securityGroupsAPI interface {
	NewListPager(resourceGroupName string, options *armnetwork.SecurityGroupsClientListOptions) *runtime.Pager[armnetwork.SecurityGroupsClientListResponse]
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
	Get(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string,
		options *armcomputev2.VirtualMachineScaleSetVMsClientGetOptions,
	) (armcomputev2.VirtualMachineScaleSetVMsClientGetResponse, error)
	NewListPager(resourceGroupName string, virtualMachineScaleSetName string,
		options *armcomputev2.VirtualMachineScaleSetVMsClientListOptions,
	) *runtime.Pager[armcomputev2.VirtualMachineScaleSetVMsClientListResponse]
}

type scaleSetsAPI interface {
	NewListPager(resourceGroupName string, options *armcomputev2.VirtualMachineScaleSetsClientListOptions,
	) *runtime.Pager[armcomputev2.VirtualMachineScaleSetsClientListResponse]
}

type loadBalancerAPI interface {
	NewListPager(resourceGroupName string, options *armnetwork.LoadBalancersClientListOptions,
	) *runtime.Pager[armnetwork.LoadBalancersClientListResponse]
}

type tagsAPI interface {
	CreateOrUpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsResource, options *armresources.TagsClientCreateOrUpdateAtScopeOptions) (armresources.TagsClientCreateOrUpdateAtScopeResponse, error)
	UpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsPatchResource, options *armresources.TagsClientUpdateAtScopeOptions) (armresources.TagsClientUpdateAtScopeResponse, error)
}

type applicationInsightsAPI interface {
	Get(ctx context.Context, resourceGroupName string, resourceName string, options *armapplicationinsights.ComponentsClientGetOptions) (armapplicationinsights.ComponentsClientGetResponse, error)
}
