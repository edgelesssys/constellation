package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type stubIMDSAPI struct {
	res         metadataResponse
	retrieveErr error
}

func (a *stubIMDSAPI) Retrieve(ctx context.Context) (metadataResponse, error) {
	return a.res, a.retrieveErr
}

type stubNetworkInterfacesAPI struct {
	getInterface armnetwork.Interface
	getErr       error
}

func (a *stubNetworkInterfacesAPI) GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string,
	virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string,
	options *armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceOptions,
) (armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse, error) {
	return armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResponse{
		InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResult: armnetwork.InterfacesClientGetVirtualMachineScaleSetNetworkInterfaceResult{
			Interface: a.getInterface,
		},
	}, a.getErr
}

func (a *stubNetworkInterfacesAPI) Get(ctx context.Context, resourceGroupName string, networkInterfaceName string,
	options *armnetwork.InterfacesClientGetOptions,
) (armnetwork.InterfacesClientGetResponse, error) {
	return armnetwork.InterfacesClientGetResponse{
		InterfacesClientGetResult: armnetwork.InterfacesClientGetResult{
			Interface: a.getInterface,
		},
	}, a.getErr
}

type stubVirtualMachinesClientListPager struct {
	pagesCounter int
	pages        [][]*armcompute.VirtualMachine
}

func (p *stubVirtualMachinesClientListPager) NextPage(ctx context.Context) bool {
	return p.pagesCounter < len(p.pages)
}

func (p *stubVirtualMachinesClientListPager) PageResponse() armcompute.VirtualMachinesClientListResponse {
	if p.pagesCounter >= len(p.pages) {
		return armcompute.VirtualMachinesClientListResponse{}
	}
	p.pagesCounter = p.pagesCounter + 1
	return armcompute.VirtualMachinesClientListResponse{
		VirtualMachinesClientListResult: armcompute.VirtualMachinesClientListResult{
			VirtualMachineListResult: armcompute.VirtualMachineListResult{
				Value: p.pages[p.pagesCounter-1],
			},
		},
	}
}

type stubVirtualMachinesAPI struct {
	getVM     armcompute.VirtualMachine
	getErr    error
	listPages [][]*armcompute.VirtualMachine
}

func (a *stubVirtualMachinesAPI) Get(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientGetOptions) (armcompute.VirtualMachinesClientGetResponse, error) {
	return armcompute.VirtualMachinesClientGetResponse{
		VirtualMachinesClientGetResult: armcompute.VirtualMachinesClientGetResult{
			VirtualMachine: a.getVM,
		},
	}, a.getErr
}

func (a *stubVirtualMachinesAPI) List(resourceGroupName string, options *armcompute.VirtualMachinesClientListOptions) virtualMachinesClientListPager {
	return &stubVirtualMachinesClientListPager{
		pages: a.listPages,
	}
}

type stubVirtualMachineScaleSetVMsClientListPager struct {
	pagesCounter int
	pages        [][]*armcompute.VirtualMachineScaleSetVM
}

func (p *stubVirtualMachineScaleSetVMsClientListPager) NextPage(ctx context.Context) bool {
	return p.pagesCounter < len(p.pages)
}

func (p *stubVirtualMachineScaleSetVMsClientListPager) PageResponse() armcompute.VirtualMachineScaleSetVMsClientListResponse {
	if p.pagesCounter >= len(p.pages) {
		return armcompute.VirtualMachineScaleSetVMsClientListResponse{}
	}
	p.pagesCounter = p.pagesCounter + 1
	return armcompute.VirtualMachineScaleSetVMsClientListResponse{
		VirtualMachineScaleSetVMsClientListResult: armcompute.VirtualMachineScaleSetVMsClientListResult{
			VirtualMachineScaleSetVMListResult: armcompute.VirtualMachineScaleSetVMListResult{
				Value: p.pages[p.pagesCounter-1],
			},
		},
	}
}

type stubVirtualMachineScaleSetVMsAPI struct {
	getVM     armcompute.VirtualMachineScaleSetVM
	getErr    error
	listPages [][]*armcompute.VirtualMachineScaleSetVM
}

func (a *stubVirtualMachineScaleSetVMsAPI) Get(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string, options *armcompute.VirtualMachineScaleSetVMsClientGetOptions) (armcompute.VirtualMachineScaleSetVMsClientGetResponse, error) {
	return armcompute.VirtualMachineScaleSetVMsClientGetResponse{
		VirtualMachineScaleSetVMsClientGetResult: armcompute.VirtualMachineScaleSetVMsClientGetResult{
			VirtualMachineScaleSetVM: a.getVM,
		},
	}, a.getErr
}

func (a *stubVirtualMachineScaleSetVMsAPI) List(resourceGroupName string, virtualMachineScaleSetName string, options *armcompute.VirtualMachineScaleSetVMsClientListOptions) virtualMachineScaleSetVMsClientListPager {
	return &stubVirtualMachineScaleSetVMsClientListPager{
		pages: a.listPages,
	}
}

type stubVirtualMachineScaleSetsClientListPager struct {
	pagesCounter int
	pages        [][]*armcompute.VirtualMachineScaleSet
}

func (p *stubVirtualMachineScaleSetsClientListPager) NextPage(ctx context.Context) bool {
	return p.pagesCounter < len(p.pages)
}

func (p *stubVirtualMachineScaleSetsClientListPager) PageResponse() armcompute.VirtualMachineScaleSetsClientListResponse {
	if p.pagesCounter >= len(p.pages) {
		return armcompute.VirtualMachineScaleSetsClientListResponse{}
	}
	p.pagesCounter = p.pagesCounter + 1
	return armcompute.VirtualMachineScaleSetsClientListResponse{
		VirtualMachineScaleSetsClientListResult: armcompute.VirtualMachineScaleSetsClientListResult{
			VirtualMachineScaleSetListResult: armcompute.VirtualMachineScaleSetListResult{
				Value: p.pages[p.pagesCounter-1],
			},
		},
	}
}

type stubScaleSetsAPI struct {
	listPages [][]*armcompute.VirtualMachineScaleSet
}

func (a *stubScaleSetsAPI) List(resourceGroupName string, options *armcompute.VirtualMachineScaleSetsClientListOptions) virtualMachineScaleSetsClientListPager {
	return &stubVirtualMachineScaleSetsClientListPager{
		pages: a.listPages,
	}
}

type stubTagsAPI struct {
	createOrUpdateAtScopeErr error
	updateAtScopeErr         error
}

func (a *stubTagsAPI) CreateOrUpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsResource, options *armresources.TagsClientCreateOrUpdateAtScopeOptions) (armresources.TagsClientCreateOrUpdateAtScopeResponse, error) {
	return armresources.TagsClientCreateOrUpdateAtScopeResponse{}, a.createOrUpdateAtScopeErr
}

func (a *stubTagsAPI) UpdateAtScope(ctx context.Context, scope string, parameters armresources.TagsPatchResource, options *armresources.TagsClientUpdateAtScopeOptions) (armresources.TagsClientUpdateAtScopeResponse, error) {
	return armresources.TagsClientUpdateAtScopeResponse{}, a.updateAtScopeErr
}

type stubSecurityGroupsClientListPager struct {
	pagesCounter int
	pages        [][]*armnetwork.SecurityGroup
}

func (p *stubSecurityGroupsClientListPager) NextPage(ctx context.Context) bool {
	return p.pagesCounter < len(p.pages)
}

func (p *stubSecurityGroupsClientListPager) PageResponse() armnetwork.SecurityGroupsClientListResponse {
	if p.pagesCounter >= len(p.pages) {
		return armnetwork.SecurityGroupsClientListResponse{}
	}
	p.pagesCounter = p.pagesCounter + 1
	return armnetwork.SecurityGroupsClientListResponse{
		SecurityGroupsClientListResult: armnetwork.SecurityGroupsClientListResult{
			SecurityGroupListResult: armnetwork.SecurityGroupListResult{
				Value: p.pages[p.pagesCounter-1],
			},
		},
	}
}

type stubSecurityGroupsAPI struct {
	listPages [][]*armnetwork.SecurityGroup
}

func (a *stubSecurityGroupsAPI) List(resourceGroupName string, options *armnetwork.SecurityGroupsClientListOptions) securityGroupsClientListPager {
	return &stubSecurityGroupsClientListPager{
		pages: a.listPages,
	}
}

type stubVirtualNetworksClientListPager struct {
	pagesCounter int
	pages        [][]*armnetwork.VirtualNetwork
}

func (p *stubVirtualNetworksClientListPager) NextPage(ctx context.Context) bool {
	return p.pagesCounter < len(p.pages)
}

func (p *stubVirtualNetworksClientListPager) PageResponse() armnetwork.VirtualNetworksClientListResponse {
	if p.pagesCounter >= len(p.pages) {
		return armnetwork.VirtualNetworksClientListResponse{}
	}
	p.pagesCounter = p.pagesCounter + 1
	return armnetwork.VirtualNetworksClientListResponse{
		VirtualNetworksClientListResult: armnetwork.VirtualNetworksClientListResult{
			VirtualNetworkListResult: armnetwork.VirtualNetworkListResult{
				Value: p.pages[p.pagesCounter-1],
			},
		},
	}
}

type stubVirtualNetworksAPI struct {
	listPages [][]*armnetwork.VirtualNetwork
}

func (a *stubVirtualNetworksAPI) List(resourceGroupName string, options *armnetwork.VirtualNetworksClientListOptions) virtualNetworksClientListPager {
	return &stubVirtualNetworksClientListPager{
		pages: a.listPages,
	}
}

type stubLoadBalancersClientListPager struct {
	pagesCounter int
	pages        [][]*armnetwork.LoadBalancer
}

func (p *stubLoadBalancersClientListPager) NextPage(ctx context.Context) bool {
	return p.pagesCounter < len(p.pages)
}

func (p *stubLoadBalancersClientListPager) PageResponse() armnetwork.LoadBalancersClientListResponse {
	if p.pagesCounter >= len(p.pages) {
		return armnetwork.LoadBalancersClientListResponse{}
	}
	p.pagesCounter = p.pagesCounter + 1
	return armnetwork.LoadBalancersClientListResponse{
		LoadBalancersClientListResult: armnetwork.LoadBalancersClientListResult{
			LoadBalancerListResult: armnetwork.LoadBalancerListResult{
				Value: p.pages[p.pagesCounter-1],
			},
		},
	}
}

type stubLoadBalancersAPI struct {
	listPages [][]*armnetwork.LoadBalancer
}

func (a *stubLoadBalancersAPI) List(resourceGroupName string, options *armnetwork.LoadBalancersClientListOptions) loadBalancersClientListPager {
	return &stubLoadBalancersClientListPager{
		pages: a.listPages,
	}
}

type stubPublicIPAddressesAPI struct {
	getResponse                                      armnetwork.PublicIPAddressesClientGetResponse
	getVirtualMachineScaleSetPublicIPAddressResponse armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse
	getErr                                           error
}

func (a *stubPublicIPAddressesAPI) Get(ctx context.Context, resourceGroupName string, publicIPAddressName string,
	options *armnetwork.PublicIPAddressesClientGetOptions,
) (armnetwork.PublicIPAddressesClientGetResponse, error) {
	return a.getResponse, a.getErr
}

func (a *stubPublicIPAddressesAPI) GetVirtualMachineScaleSetPublicIPAddress(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string,
	virtualmachineIndex string, networkInterfaceName string, IPConfigurationName string, publicIPAddressName string,
	options *armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressOptions,
) (armnetwork.PublicIPAddressesClientGetVirtualMachineScaleSetPublicIPAddressResponse, error) {
	return a.getVirtualMachineScaleSetPublicIPAddressResponse, a.getErr
}
