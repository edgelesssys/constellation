package azure

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/issues/1262
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
	)
}

type stubIMDSAPI struct {
	providerIDErr     error
	providerID        string
	subscriptionIDErr error
	subscriptionID    string
	resourceGroupErr  error
	resourceGroup     string
	uidErr            error
	uid               string
}

func (a *stubIMDSAPI) ProviderID(ctx context.Context) (string, error) {
	return a.providerID, a.providerIDErr
}

func (a *stubIMDSAPI) SubscriptionID(ctx context.Context) (string, error) {
	return a.subscriptionID, a.subscriptionIDErr
}

func (a *stubIMDSAPI) ResourceGroup(ctx context.Context) (string, error) {
	return a.resourceGroup, a.resourceGroupErr
}

func (a *stubIMDSAPI) UID(ctx context.Context) (string, error) {
	return a.uid, a.uidErr
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
		Interface: a.getInterface,
	}, a.getErr
}

func (a *stubNetworkInterfacesAPI) Get(ctx context.Context, resourceGroupName string, networkInterfaceName string,
	options *armnetwork.InterfacesClientGetOptions,
) (armnetwork.InterfacesClientGetResponse, error) {
	return armnetwork.InterfacesClientGetResponse{
		Interface: a.getInterface,
	}, a.getErr
}

type stubVirtualMachineScaleSetVMsAPI struct {
	getVM  armcomputev2.VirtualMachineScaleSetVM
	getErr error
	pager  *stubVirtualMachineScaleSetVMPager
}

func (a *stubVirtualMachineScaleSetVMsAPI) Get(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string, options *armcomputev2.VirtualMachineScaleSetVMsClientGetOptions) (armcomputev2.VirtualMachineScaleSetVMsClientGetResponse, error) {
	return armcomputev2.VirtualMachineScaleSetVMsClientGetResponse{
		VirtualMachineScaleSetVM: a.getVM,
	}, a.getErr
}

func (a *stubVirtualMachineScaleSetVMsAPI) NewListPager(resourceGroupName string, virtualMachineScaleSetName string, options *armcomputev2.VirtualMachineScaleSetVMsClientListOptions) *runtime.Pager[armcomputev2.VirtualMachineScaleSetVMsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armcomputev2.VirtualMachineScaleSetVMsClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}

type stubVirtualMachineScaleSetsClientListPager struct {
	list     []armcomputev2.VirtualMachineScaleSet
	fetchErr error
	more     bool
}

func (p *stubVirtualMachineScaleSetsClientListPager) moreFunc() func(armcomputev2.VirtualMachineScaleSetsClientListResponse) bool {
	return func(armcomputev2.VirtualMachineScaleSetsClientListResponse) bool {
		return p.more
	}
}

func (p *stubVirtualMachineScaleSetsClientListPager) fetcherFunc() func(context.Context, *armcomputev2.VirtualMachineScaleSetsClientListResponse) (armcomputev2.VirtualMachineScaleSetsClientListResponse, error) {
	return func(context.Context, *armcomputev2.VirtualMachineScaleSetsClientListResponse) (armcomputev2.VirtualMachineScaleSetsClientListResponse, error) {
		page := make([]*armcomputev2.VirtualMachineScaleSet, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armcomputev2.VirtualMachineScaleSetsClientListResponse{
			VirtualMachineScaleSetListResult: armcomputev2.VirtualMachineScaleSetListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubScaleSetsAPI struct {
	pager *stubVirtualMachineScaleSetsClientListPager
}

func (a *stubScaleSetsAPI) NewListPager(resourceGroupName string, options *armcomputev2.VirtualMachineScaleSetsClientListOptions) *runtime.Pager[armcomputev2.VirtualMachineScaleSetsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armcomputev2.VirtualMachineScaleSetsClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
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
	list     []armnetwork.SecurityGroup
	fetchErr error
	more     bool
}

func (p *stubSecurityGroupsClientListPager) moreFunc() func(armnetwork.SecurityGroupsClientListResponse) bool {
	return func(armnetwork.SecurityGroupsClientListResponse) bool {
		return p.more
	}
}

func (p *stubSecurityGroupsClientListPager) fetcherFunc() func(context.Context, *armnetwork.SecurityGroupsClientListResponse) (armnetwork.SecurityGroupsClientListResponse, error) {
	return func(context.Context, *armnetwork.SecurityGroupsClientListResponse) (armnetwork.SecurityGroupsClientListResponse, error) {
		page := make([]*armnetwork.SecurityGroup, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armnetwork.SecurityGroupsClientListResponse{
			SecurityGroupListResult: armnetwork.SecurityGroupListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubSecurityGroupsAPI struct {
	pager *stubSecurityGroupsClientListPager
}

func (a *stubSecurityGroupsAPI) NewListPager(resourceGroupName string, options *armnetwork.SecurityGroupsClientListOptions) *runtime.Pager[armnetwork.SecurityGroupsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armnetwork.SecurityGroupsClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}

type stubVirtualNetworksClientListPager struct {
	list     []armnetwork.VirtualNetwork
	fetchErr error
	more     bool
}

func (p *stubVirtualNetworksClientListPager) moreFunc() func(armnetwork.VirtualNetworksClientListResponse) bool {
	return func(armnetwork.VirtualNetworksClientListResponse) bool {
		return p.more
	}
}

func (p *stubVirtualNetworksClientListPager) fetcherFunc() func(context.Context, *armnetwork.VirtualNetworksClientListResponse) (armnetwork.VirtualNetworksClientListResponse, error) {
	return func(context.Context, *armnetwork.VirtualNetworksClientListResponse) (armnetwork.VirtualNetworksClientListResponse, error) {
		page := make([]*armnetwork.VirtualNetwork, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armnetwork.VirtualNetworksClientListResponse{
			VirtualNetworkListResult: armnetwork.VirtualNetworkListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubVirtualNetworksAPI struct {
	pager *stubVirtualNetworksClientListPager
}

func (a *stubVirtualNetworksAPI) NewListPager(resourceGroupName string, options *armnetwork.VirtualNetworksClientListOptions) *runtime.Pager[armnetwork.VirtualNetworksClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armnetwork.VirtualNetworksClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}

type stubLoadBalancersAPI struct {
	pager *stubLoadBalancersClientListPager
}

func (a *stubLoadBalancersAPI) NewListPager(resourceGroupName string, options *armnetwork.LoadBalancersClientListOptions,
) *runtime.Pager[armnetwork.LoadBalancersClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armnetwork.LoadBalancersClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
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

type stubVirtualMachineScaleSetVMPager struct {
	list     []armcomputev2.VirtualMachineScaleSetVM
	fetchErr error
	more     bool
}

func (p *stubVirtualMachineScaleSetVMPager) moreFunc() func(armcomputev2.VirtualMachineScaleSetVMsClientListResponse) bool {
	return func(armcomputev2.VirtualMachineScaleSetVMsClientListResponse) bool {
		return p.more
	}
}

func (p *stubVirtualMachineScaleSetVMPager) fetcherFunc() func(context.Context, *armcomputev2.VirtualMachineScaleSetVMsClientListResponse) (armcomputev2.VirtualMachineScaleSetVMsClientListResponse, error) {
	return func(context.Context, *armcomputev2.VirtualMachineScaleSetVMsClientListResponse) (armcomputev2.VirtualMachineScaleSetVMsClientListResponse, error) {
		page := make([]*armcomputev2.VirtualMachineScaleSetVM, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armcomputev2.VirtualMachineScaleSetVMsClientListResponse{
			VirtualMachineScaleSetVMListResult: armcomputev2.VirtualMachineScaleSetVMListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubLoadBalancersClientListPager struct {
	list     []armnetwork.LoadBalancer
	fetchErr error
	more     bool
}

func (p *stubLoadBalancersClientListPager) moreFunc() func(armnetwork.LoadBalancersClientListResponse) bool {
	return func(armnetwork.LoadBalancersClientListResponse) bool {
		return p.more
	}
}

func (p *stubLoadBalancersClientListPager) fetcherFunc() func(context.Context, *armnetwork.LoadBalancersClientListResponse) (armnetwork.LoadBalancersClientListResponse, error) {
	return func(context.Context, *armnetwork.LoadBalancersClientListResponse) (armnetwork.LoadBalancersClientListResponse, error) {
		page := make([]*armnetwork.LoadBalancer, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armnetwork.LoadBalancersClientListResponse{
			LoadBalancerListResult: armnetwork.LoadBalancerListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}
