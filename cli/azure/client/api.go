package client

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
)

type virtualNetworksCreateOrUpdatePollerResponse interface {
	PollUntilDone(ctx context.Context, freq time.Duration) (armnetwork.VirtualNetworksClientCreateOrUpdateResponse, error)
}

type networksAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		virtualNetworkName string, parameters armnetwork.VirtualNetwork,
		options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (
		virtualNetworksCreateOrUpdatePollerResponse, error)
}

type networkSecurityGroupsCreateOrUpdatePollerResponse interface {
	PollUntilDone(ctx context.Context, freq time.Duration) (armnetwork.SecurityGroupsClientCreateOrUpdateResponse, error)
}

type networkSecurityGroupsAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		networkSecurityGroupName string, parameters armnetwork.SecurityGroup,
		options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (
		networkSecurityGroupsCreateOrUpdatePollerResponse, error)
}

type loadBalancersClientCreateOrUpdatePollerResponse interface {
	PollUntilDone(ctx context.Context, freq time.Duration) (armnetwork.LoadBalancersClientCreateOrUpdateResponse, error)
}

type loadBalancersAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		loadBalancerName string, parameters armnetwork.LoadBalancer,
		options *armnetwork.LoadBalancersClientBeginCreateOrUpdateOptions) (
		loadBalancersClientCreateOrUpdatePollerResponse, error,
	)
}

type virtualMachineScaleSetsCreateOrUpdatePollerResponse interface {
	PollUntilDone(ctx context.Context, freq time.Duration) (armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse, error)
}

type scaleSetsAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
		vmScaleSetName string, parameters armcompute.VirtualMachineScaleSet,
		options *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions) (
		virtualMachineScaleSetsCreateOrUpdatePollerResponse, error)
}

type publicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager interface {
	NextPage(ctx context.Context) bool
	PageResponse() armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse
}

// TODO: deprecate as soon as scale sets are available.
type publicIPAddressesClientCreateOrUpdatePollerResponse interface {
	PollUntilDone(ctx context.Context, freq time.Duration) (armnetwork.PublicIPAddressesClientCreateOrUpdateResponse, error)
}

type publicIPAddressesAPI interface {
	ListVirtualMachineScaleSetVMPublicIPAddresses(resourceGroupName string,
		virtualMachineScaleSetName string, virtualmachineIndex string,
		networkInterfaceName string, ipConfigurationName string,
		options *armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesOptions,
	) publicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager
	// TODO: deprecate as soon as scale sets are available.
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, publicIPAddressName string,
		parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (
		publicIPAddressesClientCreateOrUpdatePollerResponse, error)
	// TODO: deprecate as soon as scale sets are available.
	Get(ctx context.Context, resourceGroupName string, publicIPAddressName string, options *armnetwork.PublicIPAddressesClientGetOptions) (
		armnetwork.PublicIPAddressesClientGetResponse, error)
}

// TODO: deprecate as soon as scale sets are available.
type interfacesClientCreateOrUpdatePollerResponse interface {
	PollUntilDone(ctx context.Context, freq time.Duration) (armnetwork.InterfacesClientCreateOrUpdateResponse, error)
}

type networkInterfacesAPI interface {
	GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string,
		virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
		options *armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceOptions,
	) (armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse, error)
	// TODO: deprecate as soon as scale sets are available
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, networkInterfaceName string,
		parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (
		interfacesClientCreateOrUpdatePollerResponse, error)
}

type resourceGroupsDeletePollerResponse interface {
	PollUntilDone(ctx context.Context, freq time.Duration) (armresources.ResourceGroupsClientDeleteResponse, error)
}

type resourceGroupAPI interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string,
		parameters armresources.ResourceGroup,
		options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (
		armresources.ResourceGroupsClientCreateOrUpdateResponse, error)
	BeginDelete(ctx context.Context, resourceGroupName string,
		options *armresources.ResourceGroupsClientBeginDeleteOptions) (
		resourceGroupsDeletePollerResponse, error)
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
type virtualMachinesClientCreateOrUpdatePollerResponse interface {
	PollUntilDone(ctx context.Context, freq time.Duration) (armcompute.VirtualMachinesClientCreateOrUpdateResponse, error)
}

// TODO: deprecate as soon as scale sets are available.
type virtualMachinesAPI interface {
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, vmName string, parameters armcompute.VirtualMachine,
		options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (virtualMachinesClientCreateOrUpdatePollerResponse, error)
}
