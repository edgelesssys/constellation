package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
)

type networksAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		virtualNetworkName string, parameters armnetwork.VirtualNetwork,
		options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (
		*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error)
}

type networkSecurityGroupsAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		networkSecurityGroupName string, parameters armnetwork.SecurityGroup,
		options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (
		*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error)
}

type loadBalancersAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		loadBalancerName string, parameters armnetwork.LoadBalancer,
		options *armnetwork.LoadBalancersClientBeginCreateOrUpdateOptions) (
		*runtime.Poller[armnetwork.LoadBalancersClientCreateOrUpdateResponse], error,
	)
}

type scaleSetsAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		vmScaleSetName string, parameters armcomputev2.VirtualMachineScaleSet,
		options *armcomputev2.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions) (
		*runtime.Poller[armcomputev2.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error)
}

type publicIPAddressesAPI interface {
	NewListVirtualMachineScaleSetVMPublicIPAddressesPager(
		resourceGroupName string, virtualMachineScaleSetName string,
		virtualmachineIndex string, networkInterfaceName string,
		ipConfigurationName string,
		options *armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesOptions,
	) *runtime.Pager[armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse]
	// TODO: deprecate as soon as scale sets are available.
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, publicIPAddressName string,
		parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (
		*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error)
	// TODO: deprecate as soon as scale sets are available.
	Get(ctx context.Context, resourceGroupName string, publicIPAddressName string, options *armnetwork.PublicIPAddressesClientGetOptions) (
		armnetwork.PublicIPAddressesClientGetResponse, error)
}

type networkInterfacesAPI interface {
	GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string,
		virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
		options *armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceOptions,
	) (armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse, error)
	// TODO: deprecate as soon as scale sets are available
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, networkInterfaceName string,
		parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (
		*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error)
}

type resourceGroupAPI interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string,
		parameters armresources.ResourceGroup,
		options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (
		armresources.ResourceGroupsClientCreateOrUpdateResponse, error)
	BeginDelete(ctx context.Context, resourceGroupName string,
		options *armresources.ResourceGroupsClientBeginDeleteOptions) (
		*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error)
	Get(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientGetOptions) (armresources.ResourceGroupsClientGetResponse, error)
}

type applicationsAPI interface {
	Create(ctx context.Context, parameters graphrbac.ApplicationCreateParameters) (graphrbac.Application, error)
	Delete(ctx context.Context, applicationObjectID string) (autorest.Response, error)
	UpdatePasswordCredentials(ctx context.Context, objectID string, parameters graphrbac.PasswordCredentialsUpdateParameters) (autorest.Response, error)
}

type servicePrincipalsAPI interface {
	Create(ctx context.Context, parameters graphrbac.ServicePrincipalCreateParameters) (graphrbac.ServicePrincipal, error)
}

// the newer version "armauthorization.RoleAssignmentsClient" is currently broken: https://github.com/Azure/azure-sdk-for-go/issues/17071
// TODO: switch to "armauthorization.RoleAssignmentsClient" if issue is resolved.
type roleAssignmentsAPI interface {
	Create(ctx context.Context, scope string, roleAssignmentName string, parameters authorization.RoleAssignmentCreateParameters) (authorization.RoleAssignment, error)
}

// TODO: deprecate as soon as scale sets are available.
type virtualMachinesAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		vmName string, parameters armcomputev2.VirtualMachine,
		options *armcomputev2.VirtualMachinesClientBeginCreateOrUpdateOptions) (
		*runtime.Poller[armcomputev2.VirtualMachinesClientCreateOrUpdateResponse], error)
}

type applicationInsightsAPI interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, insightProperties armapplicationinsights.Component,
		options *armapplicationinsights.ComponentsClientCreateOrUpdateOptions) (armapplicationinsights.ComponentsClientCreateOrUpdateResponse, error)
}
