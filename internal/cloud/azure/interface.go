/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

type imdsAPI interface {
	providerID(ctx context.Context) (string, error)
	name(ctx context.Context) (string, error)
	resourceGroup(ctx context.Context) (string, error)
	subscriptionID(ctx context.Context) (string, error)
	uid(ctx context.Context) (string, error)
	initSecretHash(ctx context.Context) (string, error)
}

type virtualNetworksAPI interface {
	NewListPager(resourceGroupName string,
		options *armnetwork.VirtualNetworksClientListOptions,
	) *runtime.Pager[armnetwork.VirtualNetworksClientListResponse]
}

type securityGroupsAPI interface {
	NewListPager(resourceGroupName string,
		options *armnetwork.SecurityGroupsClientListOptions,
	) *runtime.Pager[armnetwork.SecurityGroupsClientListResponse]
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
		options *armnetwork.PublicIPAddressesClientGetOptions,
	) (armnetwork.PublicIPAddressesClientGetResponse, error)
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

type applicationInsightsAPI interface {
	NewListByResourceGroupPager(resourceGroupName string,
		options *armapplicationinsights.ComponentsClientListByResourceGroupOptions,
	) *runtime.Pager[armapplicationinsights.ComponentsClientListByResourceGroupResponse]
}
