package client

import (
	"context"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
)

type stubNetworksAPI struct {
	createErr error
	pollErr   error
}

func (a stubNetworksAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	virtualNetworkName string, parameters armnetwork.VirtualNetwork,
	options *armnetwork.VirtualNetworksClientBeginCreateOrUpdateOptions) (
	*runtime.Poller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse], error,
) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armnetwork.VirtualNetworksClientCreateOrUpdateResponse]{
		Handler: &stubPoller[armnetwork.VirtualNetworksClientCreateOrUpdateResponse]{
			result: armnetwork.VirtualNetworksClientCreateOrUpdateResponse{
				VirtualNetwork: armnetwork.VirtualNetwork{
					Properties: &armnetwork.VirtualNetworkPropertiesFormat{
						Subnets: []*armnetwork.Subnet{
							{
								ID: to.Ptr("subnet-id"),
							},
						},
					},
				},
			},
			resultErr: a.pollErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.createErr
}

type stubLoadBalancersAPI struct {
	createErr    error
	stubResponse armnetwork.LoadBalancersClientCreateOrUpdateResponse
	pollErr      error
}

func (a stubLoadBalancersAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	loadBalancerName string, parameters armnetwork.LoadBalancer,
	options *armnetwork.LoadBalancersClientBeginCreateOrUpdateOptions) (
	*runtime.Poller[armnetwork.LoadBalancersClientCreateOrUpdateResponse], error,
) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armnetwork.LoadBalancersClientCreateOrUpdateResponse]{
		Handler: &stubPoller[armnetwork.LoadBalancersClientCreateOrUpdateResponse]{
			result:    a.stubResponse,
			resultErr: a.pollErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.createErr
}

type stubNetworkSecurityGroupsAPI struct {
	createErr error
	pollErr   error
}

func (a stubNetworkSecurityGroupsAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	networkSecurityGroupName string, parameters armnetwork.SecurityGroup,
	options *armnetwork.SecurityGroupsClientBeginCreateOrUpdateOptions) (
	*runtime.Poller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse], error,
) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armnetwork.SecurityGroupsClientCreateOrUpdateResponse]{
		Handler: &stubPoller[armnetwork.SecurityGroupsClientCreateOrUpdateResponse]{
			result: armnetwork.SecurityGroupsClientCreateOrUpdateResponse{
				SecurityGroup: armnetwork.SecurityGroup{ID: to.Ptr("network-security-group-id")},
			},
			resultErr: a.pollErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.createErr
}

type stubResourceGroupAPI struct {
	terminateErr     error
	createErr        error
	getErr           error
	getResourceGroup armresources.ResourceGroup
	pollErr          error
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
		ResourceGroup: a.getResourceGroup,
	}, a.getErr
}

func (a stubResourceGroupAPI) BeginDelete(ctx context.Context, resourceGroupName string,
	options *armresources.ResourceGroupsClientBeginDeleteOptions) (
	*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error,
) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armresources.ResourceGroupsClientDeleteResponse]{
		Handler: &stubPoller[armresources.ResourceGroupsClientDeleteResponse]{
			result:    armresources.ResourceGroupsClientDeleteResponse{},
			resultErr: a.pollErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.terminateErr
}

type stubScaleSetsAPI struct {
	createErr    error
	stubResponse armcomputev2.VirtualMachineScaleSetsClientCreateOrUpdateResponse
	pollErr      error
}

func (a stubScaleSetsAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string,
	vmScaleSetName string, parameters armcomputev2.VirtualMachineScaleSet,
	options *armcomputev2.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions) (
	*runtime.Poller[armcomputev2.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error,
) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armcomputev2.VirtualMachineScaleSetsClientCreateOrUpdateResponse]{
		Handler: &stubPoller[armcomputev2.VirtualMachineScaleSetsClientCreateOrUpdateResponse]{
			result:    a.stubResponse,
			resultErr: a.pollErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.createErr
}

type stubPublicIPAddressesAPI struct {
	createErr error
	getErr    error
	pollErr   error
}

type stubPublicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager struct {
	pages    int
	fetchErr error
	more     bool
}

func (p *stubPublicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager) moreFunc() func(
	armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse) bool {
	return func(armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse) bool {
		return p.more
	}
}

func (p *stubPublicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager) fetcherFunc() func(
	context.Context, *armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse) (
	armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse, error) {
	return func(context.Context, *armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse) (
		armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse, error,
	) {
		page := make([]*armnetwork.PublicIPAddress, p.pages)
		for i := 0; i < p.pages; i++ {
			page[i] = &armnetwork.PublicIPAddress{
				Properties: &armnetwork.PublicIPAddressPropertiesFormat{
					IPAddress: to.Ptr("192.0.2.1"),
				},
			}
		}
		return armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse{
			PublicIPAddressListResult: armnetwork.PublicIPAddressListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

func (a stubPublicIPAddressesAPI) NewListVirtualMachineScaleSetVMPublicIPAddressesPager(
	resourceGroupName string, virtualMachineScaleSetName string,
	virtualmachineIndex string, networkInterfaceName string,
	ipConfigurationName string,
	options *armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesOptions,
) *runtime.Pager[armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse] {
	pager := &stubPublicIPAddressesListVirtualMachineScaleSetVMPublicIPAddressesPager{
		pages: 1,
	}
	return runtime.NewPager(runtime.PagingHandler[armnetwork.PublicIPAddressesClientListVirtualMachineScaleSetVMPublicIPAddressesResponse]{
		More:    pager.moreFunc(),
		Fetcher: pager.fetcherFunc(),
	})
}

func (a stubPublicIPAddressesAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, publicIPAddressName string,
	parameters armnetwork.PublicIPAddress, options *armnetwork.PublicIPAddressesClientBeginCreateOrUpdateOptions) (
	*runtime.Poller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse], error,
) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse]{
		Handler: &stubPoller[armnetwork.PublicIPAddressesClientCreateOrUpdateResponse]{
			result: armnetwork.PublicIPAddressesClientCreateOrUpdateResponse{
				PublicIPAddress: armnetwork.PublicIPAddress{
					ID: to.Ptr("ip-address-id"),
					Properties: &armnetwork.PublicIPAddressPropertiesFormat{
						IPAddress: to.Ptr("192.0.2.1"),
					},
				},
			},
			resultErr: a.pollErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.createErr
}

func (a stubPublicIPAddressesAPI) Get(ctx context.Context, resourceGroupName string, publicIPAddressName string, options *armnetwork.PublicIPAddressesClientGetOptions) (
	armnetwork.PublicIPAddressesClientGetResponse, error,
) {
	return armnetwork.PublicIPAddressesClientGetResponse{
		PublicIPAddress: armnetwork.PublicIPAddress{
			Properties: &armnetwork.PublicIPAddressPropertiesFormat{
				IPAddress: to.Ptr("192.0.2.1"),
			},
		},
	}, a.getErr
}

type stubNetworkInterfacesAPI struct {
	getErr    error
	createErr error
	pollErr   error
}

func (a stubNetworkInterfacesAPI) GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string,
	virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
	options *armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceOptions,
) (armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse, error) {
	if a.getErr != nil {
		return armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse{}, a.getErr
	}
	return armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse{
		Interface: armnetwork.Interface{
			Properties: &armnetwork.InterfacePropertiesFormat{
				IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
					{
						Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
							PrivateIPAddress: to.Ptr("192.0.2.1"),
						},
					},
				},
			},
		},
	}, nil
}

// TODO: deprecate as soon as scale sets are available.
func (a stubNetworkInterfacesAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, networkInterfaceName string,
	parameters armnetwork.Interface, options *armnetwork.InterfacesClientBeginCreateOrUpdateOptions) (
	*runtime.Poller[armnetwork.InterfacesClientCreateOrUpdateResponse], error,
) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armnetwork.InterfacesClientCreateOrUpdateResponse]{
		Handler: &stubPoller[armnetwork.InterfacesClientCreateOrUpdateResponse]{
			result: armnetwork.InterfacesClientCreateOrUpdateResponse{
				Interface: armnetwork.Interface{
					ID: to.Ptr("interface-id"),
					Properties: &armnetwork.InterfacePropertiesFormat{
						IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
							{
								Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
									PrivateIPAddress: to.Ptr("192.0.2.1"),
								},
							},
						},
					},
				},
			},
			resultErr: a.pollErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.createErr
}

// TODO: deprecate as soon as scale sets are available.
type stubVirtualMachinesAPI struct {
	stubResponse armcomputev2.VirtualMachinesClientCreateOrUpdateResponse
	pollErr      error
	createErr    error
}

// TODO: deprecate as soon as scale sets are available.
func (a stubVirtualMachinesAPI) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, vmName string, parameters armcomputev2.VirtualMachine,
	options *armcomputev2.VirtualMachinesClientBeginCreateOrUpdateOptions,
) (*runtime.Poller[armcomputev2.VirtualMachinesClientCreateOrUpdateResponse], error) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armcomputev2.VirtualMachinesClientCreateOrUpdateResponse]{
		Handler: &stubPoller[armcomputev2.VirtualMachinesClientCreateOrUpdateResponse]{
			result:    a.stubResponse,
			resultErr: a.pollErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.createErr
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
		AppID:    to.Ptr("00000000-0000-0000-0000-000000000000"),
		ObjectID: to.Ptr("00000000-0000-0000-0000-000000000001"),
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
		AppID:    to.Ptr("00000000-0000-0000-0000-000000000000"),
		ObjectID: to.Ptr("00000000-0000-0000-0000-000000000002"),
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

type stubApplicationInsightsAPI struct {
	err error
}

func (a *stubApplicationInsightsAPI) CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, insightProperties armapplicationinsights.Component, options *armapplicationinsights.ComponentsClientCreateOrUpdateOptions) (armapplicationinsights.ComponentsClientCreateOrUpdateResponse, error) {
	resp := armapplicationinsights.ComponentsClientCreateOrUpdateResponse{}
	return resp, a.err
}

type stubPoller[T any] struct {
	result    T
	pollErr   error
	resultErr error
}

func (p *stubPoller[T]) Done() bool {
	return true
}

func (p *stubPoller[T]) Poll(context.Context) (*http.Response, error) {
	return nil, p.pollErr
}

func (p *stubPoller[T]) Result(ctx context.Context, out *T) error {
	*out = p.result
	return p.resultErr
}
