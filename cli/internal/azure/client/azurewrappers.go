package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
)

type networksClient struct {
	*armnetwork.VirtualNetworksClient
}

func (c *networksClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	virtualNetworkName string, parameters armnetwork.VirtualNetwork,
	options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (
	virtualNetworksCreateOrUpdatePollerResponse, error,
) {
	return c.VirtualNetworksClient.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, parameters, options)
}

// TODO: deprecate as soon as scale sets are available.
type networkInterfacesClient struct {
	*armnetwork.InterfacesClient
}

// TODO: deprecate as soon as scale sets are available.
func (c *networkInterfacesClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, networkInterfaceName string,
	parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions,
) (interfacesClientCreateOrUpdatePollerResponse, error) {
	return c.InterfacesClient.BeginCreateOrUpdate(ctx, resourceGroupName, networkInterfaceName, parameters, options)
}

type loadBalancersClient struct {
	*armnetwork.LoadBalancersClient
}

func (c *loadBalancersClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, loadBalancerName string,
	parameters armnetwork.LoadBalancer, options *armnetwork.LoadBalancersClientBeginCreateOrUpdateOptions) (
	loadBalancersClientCreateOrUpdatePollerResponse, error,
) {
	return c.LoadBalancersClient.BeginCreateOrUpdate(ctx, resourceGroupName, loadBalancerName, parameters, options)
}

type networkSecurityGroupsClient struct {
	*armnetwork.SecurityGroupsClient
}

func (c *networkSecurityGroupsClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	networkSecurityGroupName string, parameters armnetwork.SecurityGroup,
	options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (
	networkSecurityGroupsCreateOrUpdatePollerResponse, error,
) {
	return c.SecurityGroupsClient.BeginCreateOrUpdate(ctx, resourceGroupName, networkSecurityGroupName, parameters, options)
}

type publicIPAddressesClient struct {
	*armnetwork.PublicIPAddressesClient
}

func (c *publicIPAddressesClient) ListVirtualMachineScaleSetVMPublicIPAddresses(resourceGroupName string,
	virtualMachineScaleSetName string, virtualmachineIndex string,
	networkInterfaceName string, ipConfigurationName string,
	options *armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesOptions,
) publicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager {
	return c.PublicIPAddressesClient.ListVirtualMachineScaleSetVMPublicIPAddresses(resourceGroupName, virtualMachineScaleSetName,
		virtualmachineIndex, networkInterfaceName, ipConfigurationName, options)
}

// TODO: deprecate as soon as scale sets are available.
func (c *publicIPAddressesClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, publicIPAddressName string,
	parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (
	publicIPAddressesClientCreateOrUpdatePollerResponse, error,
) {
	return c.PublicIPAddressesClient.BeginCreateOrUpdate(ctx, resourceGroupName, publicIPAddressName, parameters, options)
}

type virtualMachineScaleSetsClient struct {
	*armcompute.VirtualMachineScaleSetsClient
}

func (c *virtualMachineScaleSetsClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	vmScaleSetName string, parameters armcompute.VirtualMachineScaleSet,
	options *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions) (
	virtualMachineScaleSetsCreateOrUpdatePollerResponse, error,
) {
	return c.VirtualMachineScaleSetsClient.BeginCreateOrUpdate(ctx, resourceGroupName, vmScaleSetName, parameters, options)
}

type resourceGroupsClient struct {
	*armresources.ResourceGroupsClient
}

func (c *resourceGroupsClient) BeginDelete(ctx context.Context, resourceGroupName string,
	options *armresources.ResourceGroupsClientBeginDeleteOptions) (
	resourceGroupsDeletePollerResponse, error,
) {
	return c.ResourceGroupsClient.BeginDelete(ctx, resourceGroupName, options)
}

// TODO: deprecate as soon as scale sets are available.
type virtualMachinesClient struct {
	*armcompute.VirtualMachinesClient
}

// TODO: deprecate as soon as scale sets are available.
func (c *virtualMachinesClient) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, vmName string, parameters armcompute.VirtualMachine,
	options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions,
) (virtualMachinesClientCreateOrUpdatePollerResponse, error) {
	return c.VirtualMachinesClient.BeginCreateOrUpdate(ctx, resourceGroupName, vmName, parameters, options)
}

type applicationsClient struct {
	*graphrbac.ApplicationsClient
}

func (c *applicationsClient) Create(ctx context.Context, parameters graphrbac.ApplicationCreateParameters) (graphrbac.Application, error) {
	return c.ApplicationsClient.Create(ctx, parameters)
}

func (c *applicationsClient) Delete(ctx context.Context, applicationObjectID string) (autorest.Response, error) {
	return c.ApplicationsClient.Delete(ctx, applicationObjectID)
}

func (c *applicationsClient) UpdatePasswordCredentials(ctx context.Context, objectID string, parameters graphrbac.PasswordCredentialsUpdateParameters) (autorest.Response, error) {
	return c.ApplicationsClient.UpdatePasswordCredentials(ctx, objectID, parameters)
}

type servicePrincipalsClient struct {
	*graphrbac.ServicePrincipalsClient
}

func (c *servicePrincipalsClient) Create(ctx context.Context, parameters graphrbac.ServicePrincipalCreateParameters) (graphrbac.ServicePrincipal, error) {
	return c.ServicePrincipalsClient.Create(ctx, parameters)
}

type roleAssignmentsClient struct {
	*authorization.RoleAssignmentsClient
}

func (c *roleAssignmentsClient) Create(ctx context.Context, scope string, roleAssignmentName string, parameters authorization.RoleAssignmentCreateParameters) (authorization.RoleAssignment, error) {
	return c.RoleAssignmentsClient.Create(ctx, scope, roleAssignmentName, parameters)
}
