package client

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
)

type stubNetworksAPI struct {
	createErr    error
	stubResponse stubVirtualNetworksCreateOrUpdatePollerResponse
}

type stubVirtualNetworksCreateOrUpdatePollerResponse struct {
	armnetwork.VirtualNetworksClientCreateOrUpdatePollerResponse
	pollerErr error
}

func (r stubVirtualNetworksCreateOrUpdatePollerResponse) PollUntilDone(ctx context.Context, freq time.Duration,
) (armnetwork.VirtualNetworksClientCreateOrUpdateResponse, error) {
	return armnetwork.VirtualNetworksClientCreateOrUpdateResponse{
		VirtualNetworksClientCreateOrUpdateResult: armnetwork.VirtualNetworksClientCreateOrUpdateResult{
			VirtualNetwork: armnetwork.VirtualNetwork{
				Properties: &armnetwork.VirtualNetworkPropertiesFormat{
					Subnets: []*armnetwork.Subnet{
						{
							ID: to.StringPtr("virtual-network-subnet-id"),
						},
					},
				},
			},
		},
	}, r.pollerErr
}

func (a stubNetworksAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	virtualNetworkName string, parameters armnetwork.VirtualNetwork,
	options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (
	virtualNetworksCreateOrUpdatePollerResponse, error,
) {
	return a.stubResponse, a.createErr
}

type stubNetworkSecurityGroupsCreateOrUpdatePollerResponse struct {
	armnetwork.SecurityGroupsClientCreateOrUpdatePollerResponse
	pollerErr error
}

func (r stubNetworkSecurityGroupsCreateOrUpdatePollerResponse) PollUntilDone(ctx context.Context, freq time.Duration,
) (armnetwork.SecurityGroupsClientCreateOrUpdateResponse, error) {
	return armnetwork.SecurityGroupsClientCreateOrUpdateResponse{
		SecurityGroupsClientCreateOrUpdateResult: armnetwork.SecurityGroupsClientCreateOrUpdateResult{
			SecurityGroup: armnetwork.SecurityGroup{
				ID: to.StringPtr("network-security-group-id"),
			},
		},
	}, r.pollerErr
}

type stubNetworkSecurityGroupsAPI struct {
	createErr  error
	stubPoller stubNetworkSecurityGroupsCreateOrUpdatePollerResponse
}

func (a stubNetworkSecurityGroupsAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	networkSecurityGroupName string, parameters armnetwork.SecurityGroup,
	options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (
	networkSecurityGroupsCreateOrUpdatePollerResponse, error,
) {
	return a.stubPoller, a.createErr
}

type stubResourceGroupAPI struct {
	terminateErr     error
	createErr        error
	getErr           error
	getResourceGroup armresources.ResourceGroup
	stubResponse     stubResourceGroupsDeletePollerResponse
}

func (a stubResourceGroupAPI) CreateOrUpdate(ctx context.Context, resourceGroupName string,
	parameters armresources.ResourceGroup,
	options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (
	armresources.ResourceGroupsClientCreateOrUpdateResponse, error,
) {
	return armresources.ResourceGroupsClientCreateOrUpdateResponse{}, a.createErr
}

func (a stubResourceGroupAPI) Get(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientGetOptions) (armresources.ResourceGroupsClientGetResponse, error) {
	return armresources.ResourceGroupsClientGetResponse{
		ResourceGroupsClientGetResult: armresources.ResourceGroupsClientGetResult{
			ResourceGroup: a.getResourceGroup,
		},
	}, a.getErr
}

type stubResourceGroupsDeletePollerResponse struct {
	armresources.ResourceGroupsClientDeletePollerResponse
	pollerErr error
}

func (r stubResourceGroupsDeletePollerResponse) PollUntilDone(ctx context.Context, freq time.Duration) (
	armresources.ResourceGroupsClientDeleteResponse, error,
) {
	return armresources.ResourceGroupsClientDeleteResponse{}, r.pollerErr
}

func (a stubResourceGroupAPI) BeginDelete(ctx context.Context, resourceGroupName string,
	options *armresources.ResourceGroupsClientBeginDeleteOptions) (
	resourceGroupsDeletePollerResponse, error,
) {
	return a.stubResponse, a.terminateErr
}

type stubScaleSetsAPI struct {
	createErr    error
	stubResponse stubVirtualMachineScaleSetsCreateOrUpdatePollerResponse
}

type stubVirtualMachineScaleSetsCreateOrUpdatePollerResponse struct {
	pollResponse armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse
	pollErr      error
}

func (r stubVirtualMachineScaleSetsCreateOrUpdatePollerResponse) PollUntilDone(ctx context.Context, freq time.Duration) (
	armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse, error,
) {
	return r.pollResponse, r.pollErr
}

func (a stubScaleSetsAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	vmScaleSetName string, parameters armcompute.VirtualMachineScaleSet,
	options *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions) (
	virtualMachineScaleSetsCreateOrUpdatePollerResponse, error,
) {
	return a.stubResponse, a.createErr
}

// TODO: deprecate as soon as scale sets are available.
type stubPublicIPAddressesAPI struct {
	// TODO: deprecate as soon as scale sets are available.
	createErr error
	// TODO: deprecate as soon as scale sets are available.
	getErr error
	// TODO: deprecate as soon as scale sets are available.
	stubCreateResponse stubPublicIPAddressesClientCreateOrUpdatePollerResponse
}

// TODO: deprecate as soon as scale sets are available.
type stubPublicIPAddressesClientCreateOrUpdatePollerResponse struct {
	armnetwork.PublicIPAddressesClientCreateOrUpdatePollerResponse
	pollErr error
}

// TODO: deprecate as soon as scale sets are available.
func (r stubPublicIPAddressesClientCreateOrUpdatePollerResponse) PollUntilDone(ctx context.Context, freq time.Duration) (
	armnetwork.PublicIPAddressesClientCreateOrUpdateResponse, error,
) {
	return armnetwork.PublicIPAddressesClientCreateOrUpdateResponse{
		PublicIPAddressesClientCreateOrUpdateResult: armnetwork.PublicIPAddressesClientCreateOrUpdateResult{
			PublicIPAddress: armnetwork.PublicIPAddress{
				ID: to.StringPtr("pubIP-id"),
			},
		},
	}, r.pollErr
}

type stubPublicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager struct {
	pagesCounter int
	PagesMax     int
}

func (p *stubPublicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager) NextPage(ctx context.Context) bool {
	p.pagesCounter++
	return p.pagesCounter <= p.PagesMax
}

func (p *stubPublicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager) PageResponse() armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse {
	return armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse{
		PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResult: armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResult{
			PublicIPAddressListResult: armnetwork.PublicIPAddressListResult{
				Value: []*armnetwork.PublicIPAddress{
					{
						Properties: &armnetwork.PublicIPAddressPropertiesFormat{
							IPAddress: to.StringPtr("192.0.2.1"),
						},
					},
				},
			},
		},
	}
}

func (a stubPublicIPAddressesAPI) ListVirtualMachineScaleSetVMPublicIPAddresses(resourceGroupName string,
	virtualMachineScaleSetName string, virtualmachineIndex string,
	networkInterfaceName string, ipConfigurationName string,
	options *armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesOptions,
) publicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager {
	return &stubPublicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager{pagesCounter: 0, PagesMax: 1}
}

// TODO: deprecate as soon as scale sets are available.
func (a stubPublicIPAddressesAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, publicIPAddressName string,
	parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (
	publicIPAddressesClientCreateOrUpdatePollerResponse, error,
) {
	return a.stubCreateResponse, a.createErr
}

// TODO: deprecate as soon as scale sets are available.
func (a stubPublicIPAddressesAPI) Get(ctx context.Context, resourceGroupName string, publicIPAddressName string, options *armnetwork.PublicIPAddressesClientGetOptions) (
	armnetwork.PublicIPAddressesClientGetResponse, error,
) {
	return armnetwork.PublicIPAddressesClientGetResponse{
		PublicIPAddressesClientGetResult: armnetwork.PublicIPAddressesClientGetResult{
			PublicIPAddress: armnetwork.PublicIPAddress{
				Properties: &armnetwork.PublicIPAddressPropertiesFormat{
					IPAddress: to.StringPtr("192.0.2.1"),
				},
			},
		},
	}, a.getErr
}

type stubNetworkInterfacesAPI struct {
	getErr error
	// TODO: deprecate as soon as scale sets are available
	createErr error
	// TODO: deprecate as soon as scale sets are available
	stubResp stubInterfacesClientCreateOrUpdatePollerResponse
}

func (a stubNetworkInterfacesAPI) GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string,
	virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
	options *armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceOptions,
) (armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse, error) {
	if a.getErr != nil {
		return armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse{}, a.getErr
	}
	return armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse{
		InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResult: armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResult{
			Interface: armnetwork.Interface{
				Properties: &armnetwork.InterfacePropertiesFormat{
					IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
						{
							Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
								PrivateIPAddress: to.StringPtr("192.0.2.1"),
							},
						},
					},
				},
			},
		},
	}, nil
}

// TODO: deprecate as soon as scale sets are available.
type stubInterfacesClientCreateOrUpdatePollerResponse struct {
	pollErr error
}

// TODO: deprecate as soon as scale sets are available.
func (r stubInterfacesClientCreateOrUpdatePollerResponse) PollUntilDone(ctx context.Context, freq time.Duration) (
	armnetwork.InterfacesClientCreateOrUpdateResponse, error,
) {
	return armnetwork.InterfacesClientCreateOrUpdateResponse{
		InterfacesClientCreateOrUpdateResult: armnetwork.InterfacesClientCreateOrUpdateResult{
			Interface: armnetwork.Interface{
				ID: to.StringPtr("interface-id"),
				Properties: &armnetwork.InterfacePropertiesFormat{
					IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
						{
							Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
								PrivateIPAddress: to.StringPtr("192.0.2.1"),
							},
						},
					},
				},
			},
		},
	}, r.pollErr
}

// TODO: deprecate as soon as scale sets are available.
func (a stubNetworkInterfacesAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, networkInterfaceName string,
	parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (
	interfacesClientCreateOrUpdatePollerResponse, error,
) {
	return a.stubResp, a.createErr
}

// TODO: deprecate as soon as scale sets are available.
type stubVirtualMachinesAPI struct {
	stubResponse stubVirtualMachinesClientCreateOrUpdatePollerResponse
	createErr    error
}

// TODO: deprecate as soon as scale sets are available.
type stubVirtualMachinesClientCreateOrUpdatePollerResponse struct {
	pollResponse armcompute.VirtualMachinesClientCreateOrUpdateResponse
	pollErr      error
}

// TODO: deprecate as soon as scale sets are available.
func (r stubVirtualMachinesClientCreateOrUpdatePollerResponse) PollUntilDone(ctx context.Context, freq time.Duration) (
	armcompute.VirtualMachinesClientCreateOrUpdateResponse, error,
) {
	return r.pollResponse, r.pollErr
}

// TODO: deprecate as soon as scale sets are available.
func (a stubVirtualMachinesAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, vmName string, parameters armcompute.VirtualMachine,
	options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions,
) (virtualMachinesClientCreateOrUpdatePollerResponse, error) {
	return a.stubResponse, a.createErr
}

type stubApplicationsAPI struct {
	createErr            error
	deleteErr            error
	updateCredentialsErr error
	createApplication    *graphrbac.Application
}

func (a stubApplicationsAPI) Create(ctx context.Context, parameters graphrbac.ApplicationCreateParameters) (graphrbac.Application, error) {
	if a.createErr != nil {
		return graphrbac.Application{}, a.createErr
	}
	if a.createApplication != nil {
		return *a.createApplication, nil
	}
	return graphrbac.Application{
		AppID:    to.StringPtr("00000000-0000-0000-0000-000000000000"),
		ObjectID: to.StringPtr("00000000-0000-0000-0000-000000000001"),
	}, nil
}

func (a stubApplicationsAPI) Delete(ctx context.Context, applicationObjectID string) (autorest.Response, error) {
	if a.deleteErr != nil {
		return autorest.Response{}, a.deleteErr
	}
	return autorest.Response{}, nil
}

func (a stubApplicationsAPI) UpdatePasswordCredentials(ctx context.Context, objectID string, parameters graphrbac.PasswordCredentialsUpdateParameters) (autorest.Response, error) {
	if a.updateCredentialsErr != nil {
		return autorest.Response{}, a.updateCredentialsErr
	}
	return autorest.Response{}, nil
}

type stubServicePrincipalsAPI struct {
	createErr              error
	createServicePrincipal *graphrbac.ServicePrincipal
}

func (a stubServicePrincipalsAPI) Create(ctx context.Context, parameters graphrbac.ServicePrincipalCreateParameters) (graphrbac.ServicePrincipal, error) {
	if a.createErr != nil {
		return graphrbac.ServicePrincipal{}, a.createErr
	}
	if a.createServicePrincipal != nil {
		return *a.createServicePrincipal, nil
	}
	return graphrbac.ServicePrincipal{
		AppID:    to.StringPtr("00000000-0000-0000-0000-000000000000"),
		ObjectID: to.StringPtr("00000000-0000-0000-0000-000000000002"),
	}, nil
}

type stubRoleAssignmentsAPI struct {
	createCounter int
	createErrors  []error
}

func (a *stubRoleAssignmentsAPI) Create(ctx context.Context, scope string, roleAssignmentName string, parameters authorization.RoleAssignmentCreateParameters) (authorization.RoleAssignment, error) {
	a.createCounter++
	if len(a.createErrors) == 0 {
		return authorization.RoleAssignment{}, nil
	}
	return authorization.RoleAssignment{}, a.createErrors[(a.createCounter-1)%len(a.createErrors)]
}
