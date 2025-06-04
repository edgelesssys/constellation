/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v7"
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
		options *armcompute.VirtualMachineScaleSetVMsClientGetOptions,
	) (armcompute.VirtualMachineScaleSetVMsClientGetResponse, error)
	NewListPager(resourceGroupName string, virtualMachineScaleSetName string,
		options *armcompute.VirtualMachineScaleSetVMsClientListOptions,
	) *runtime.Pager[armcompute.VirtualMachineScaleSetVMsClientListResponse]
}

type scaleSetsAPI interface {
	NewListPager(resourceGroupName string, options *armcompute.VirtualMachineScaleSetsClientListOptions,
	) *runtime.Pager[armcompute.VirtualMachineScaleSetsClientListResponse]
}

type loadBalancerAPI interface {
	NewListPager(resourceGroupName string, options *armnetwork.LoadBalancersClientListOptions,
	) *runtime.Pager[armnetwork.LoadBalancersClientListResponse]
}
