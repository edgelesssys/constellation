package client

import (
	"context"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
)

type stubScaleSetsAPI struct {
	scaleSet       armcomputev2.VirtualMachineScaleSetsClientGetResponse
	getErr         error
	updateResponse armcomputev2.VirtualMachineScaleSetsClientUpdateResponse
	updateErr      error
	deleteResponse armcomputev2.VirtualMachineScaleSetsClientDeleteInstancesResponse
	deleteErr      error
	resultErr      error
}

func (a *stubScaleSetsAPI) Get(ctx context.Context, resourceGroupName string, vmScaleSetName string,
	options *armcomputev2.VirtualMachineScaleSetsClientGetOptions,
) (armcomputev2.VirtualMachineScaleSetsClientGetResponse, error) {
	return a.scaleSet, a.getErr
}

func (a *stubScaleSetsAPI) BeginUpdate(ctx context.Context, resourceGroupName string, vmScaleSetName string, parameters armcomputev2.VirtualMachineScaleSetUpdate,
	options *armcomputev2.VirtualMachineScaleSetsClientBeginUpdateOptions,
) (*runtime.Poller[armcomputev2.VirtualMachineScaleSetsClientUpdateResponse], error) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armcomputev2.VirtualMachineScaleSetsClientUpdateResponse]{
		Handler: &stubPoller[armcomputev2.VirtualMachineScaleSetsClientUpdateResponse]{
			result:    a.updateResponse,
			resultErr: a.resultErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.updateErr
}

func (a *stubScaleSetsAPI) BeginDeleteInstances(ctx context.Context, resourceGroupName string, vmScaleSetName string, vmInstanceIDs armcomputev2.VirtualMachineScaleSetVMInstanceRequiredIDs,
	options *armcomputev2.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions,
) (*runtime.Poller[armcomputev2.VirtualMachineScaleSetsClientDeleteInstancesResponse], error) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armcomputev2.VirtualMachineScaleSetsClientDeleteInstancesResponse]{
		Handler: &stubPoller[armcomputev2.VirtualMachineScaleSetsClientDeleteInstancesResponse]{
			result:    a.deleteResponse,
			resultErr: a.resultErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.deleteErr
}

type stubvirtualMachineScaleSetVMsAPI struct {
	scaleSetVM      armcomputev2.VirtualMachineScaleSetVMsClientGetResponse
	getErr          error
	instanceView    armcomputev2.VirtualMachineScaleSetVMsClientGetInstanceViewResponse
	instanceViewErr error
	pager           *stubPager
}

func (a *stubvirtualMachineScaleSetVMsAPI) Get(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string,
	options *armcomputev2.VirtualMachineScaleSetVMsClientGetOptions,
) (armcomputev2.VirtualMachineScaleSetVMsClientGetResponse, error) {
	return a.scaleSetVM, a.getErr
}

func (a *stubvirtualMachineScaleSetVMsAPI) GetInstanceView(ctx context.Context, resourceGroupName string, vmScaleSetName string, instanceID string,
	options *armcomputev2.VirtualMachineScaleSetVMsClientGetInstanceViewOptions,
) (armcomputev2.VirtualMachineScaleSetVMsClientGetInstanceViewResponse, error) {
	return a.instanceView, a.instanceViewErr
}

func (a *stubvirtualMachineScaleSetVMsAPI) NewListPager(resourceGroupName string, virtualMachineScaleSetName string,
	options *armcomputev2.VirtualMachineScaleSetVMsClientListOptions,
) *runtime.Pager[armcomputev2.VirtualMachineScaleSetVMsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armcomputev2.VirtualMachineScaleSetVMsClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
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

type stubPager struct {
	list     []armcomputev2.VirtualMachineScaleSetVM
	fetchErr error
	more     bool
}

func (p *stubPager) moreFunc() func(armcomputev2.VirtualMachineScaleSetVMsClientListResponse) bool {
	return func(armcomputev2.VirtualMachineScaleSetVMsClientListResponse) bool {
		return p.more
	}
}

func (p *stubPager) fetcherFunc() func(context.Context, *armcomputev2.VirtualMachineScaleSetVMsClientListResponse) (armcomputev2.VirtualMachineScaleSetVMsClientListResponse, error) {
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
