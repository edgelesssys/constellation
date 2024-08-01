/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

type stubScaleSetsAPI struct {
	scaleSet       armcompute.VirtualMachineScaleSetsClientGetResponse
	getErr         error
	updateResponse armcompute.VirtualMachineScaleSetsClientUpdateResponse
	updateErr      error
	deleteResponse armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse
	deleteErr      error
	resultErr      error
	pager          *stubVMSSPager
}

func (a *stubScaleSetsAPI) Get(_ context.Context, _, _ string,
	_ *armcompute.VirtualMachineScaleSetsClientGetOptions,
) (armcompute.VirtualMachineScaleSetsClientGetResponse, error) {
	return a.scaleSet, a.getErr
}

func (a *stubScaleSetsAPI) BeginUpdate(_ context.Context, _, _ string, _ armcompute.VirtualMachineScaleSetUpdate,
	_ *armcompute.VirtualMachineScaleSetsClientBeginUpdateOptions,
) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientUpdateResponse], error) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armcompute.VirtualMachineScaleSetsClientUpdateResponse]{
		Handler: &stubPoller[armcompute.VirtualMachineScaleSetsClientUpdateResponse]{
			result:    a.updateResponse,
			resultErr: a.resultErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.updateErr
}

func (a *stubScaleSetsAPI) BeginDeleteInstances(_ context.Context, _, _ string, _ armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs,
	_ *armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions,
) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse], error) {
	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse]{
		Handler: &stubPoller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse]{
			result:    a.deleteResponse,
			resultErr: a.resultErr,
		},
	})
	if err != nil {
		panic(err)
	}
	return poller, a.deleteErr
}

func (a *stubScaleSetsAPI) NewListPager(_ string, _ *armcompute.VirtualMachineScaleSetsClientListOptions,
) *runtime.Pager[armcompute.VirtualMachineScaleSetsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armcompute.VirtualMachineScaleSetsClientListResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}

type stubvirtualMachineScaleSetVMsAPI struct {
	scaleSetVM      armcompute.VirtualMachineScaleSetVMsClientGetResponse
	getErr          error
	instanceView    armcompute.VirtualMachineScaleSetVMsClientGetInstanceViewResponse
	instanceViewErr error
	pager           *stubVMSSVMPager
}

func (a *stubvirtualMachineScaleSetVMsAPI) Get(_ context.Context, _, _, _ string,
	_ *armcompute.VirtualMachineScaleSetVMsClientGetOptions,
) (armcompute.VirtualMachineScaleSetVMsClientGetResponse, error) {
	return a.scaleSetVM, a.getErr
}

func (a *stubvirtualMachineScaleSetVMsAPI) GetInstanceView(_ context.Context, _, _, _ string,
	_ *armcompute.VirtualMachineScaleSetVMsClientGetInstanceViewOptions,
) (armcompute.VirtualMachineScaleSetVMsClientGetInstanceViewResponse, error) {
	return a.instanceView, a.instanceViewErr
}

func (a *stubvirtualMachineScaleSetVMsAPI) NewListPager(_, _ string,
	_ *armcompute.VirtualMachineScaleSetVMsClientListOptions,
) *runtime.Pager[armcompute.VirtualMachineScaleSetVMsClientListResponse] {
	return runtime.NewPager(runtime.PagingHandler[armcompute.VirtualMachineScaleSetVMsClientListResponse]{
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

func (p *stubPoller[T]) Result(_ context.Context, out *T) error {
	*out = p.result
	return p.resultErr
}

type stubVMSSVMPager struct {
	list     []armcompute.VirtualMachineScaleSetVM
	fetchErr error
	more     bool
}

func (p *stubVMSSVMPager) moreFunc() func(armcompute.VirtualMachineScaleSetVMsClientListResponse) bool {
	return func(armcompute.VirtualMachineScaleSetVMsClientListResponse) bool {
		return p.more
	}
}

func (p *stubVMSSVMPager) fetcherFunc() func(context.Context, *armcompute.VirtualMachineScaleSetVMsClientListResponse) (armcompute.VirtualMachineScaleSetVMsClientListResponse, error) {
	return func(context.Context, *armcompute.VirtualMachineScaleSetVMsClientListResponse) (armcompute.VirtualMachineScaleSetVMsClientListResponse, error) {
		page := make([]*armcompute.VirtualMachineScaleSetVM, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armcompute.VirtualMachineScaleSetVMsClientListResponse{
			VirtualMachineScaleSetVMListResult: armcompute.VirtualMachineScaleSetVMListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubVMSSPager struct {
	list     []armcompute.VirtualMachineScaleSet
	fetchErr error
	more     bool
}

func (p *stubVMSSPager) moreFunc() func(armcompute.VirtualMachineScaleSetsClientListResponse) bool {
	return func(armcompute.VirtualMachineScaleSetsClientListResponse) bool {
		return p.more
	}
}

func (p *stubVMSSPager) fetcherFunc() func(context.Context, *armcompute.VirtualMachineScaleSetsClientListResponse) (armcompute.VirtualMachineScaleSetsClientListResponse, error) {
	return func(context.Context, *armcompute.VirtualMachineScaleSetsClientListResponse) (armcompute.VirtualMachineScaleSetsClientListResponse, error) {
		page := make([]*armcompute.VirtualMachineScaleSet, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armcompute.VirtualMachineScaleSetsClientListResponse{
			VirtualMachineScaleSetListResult: armcompute.VirtualMachineScaleSetListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}
