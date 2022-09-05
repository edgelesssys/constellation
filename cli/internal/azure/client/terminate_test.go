/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestTerminateResourceGroupResources(t *testing.T) {
	someErr := errors.New("failed")
	apiVersionErr := errors.New("NoRegisteredProviderFound, The supported api-versions are: 2015-01-01'")

	testCases := map[string]struct {
		resourceAPI resourceAPI
	}{
		"no resources": {
			resourceAPI: &fakeResourceAPI{},
		},
		"some resources": {
			resourceAPI: &fakeResourceAPI{
				resources: map[string]fakeResource{
					"id-0": {beginDeleteByIDErr: apiVersionErr, pollErr: someErr},
					"id-1": {beginDeleteByIDErr: apiVersionErr},
					"id-2": {},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			client := &Client{
				resourceAPI: tc.resourceAPI,
			}

			ctx := context.Background()
			err := client.TerminateResourceGroupResources(ctx)
			assert.NoError(err)
		})
	}
}

type fakeResourceAPI struct {
	resources map[string]fakeResource
	fetchErr  error
}

type fakeResource struct {
	beginDeleteByIDErr error
	pollErr            error
}

func (a fakeResourceAPI) NewListByResourceGroupPager(resourceGroupName string,
	options *armresources.ClientListByResourceGroupOptions,
) *runtime.Pager[armresources.ClientListByResourceGroupResponse] {
	pager := &stubClientListByResourceGroupResponsePager{
		resources: a.resources,
		fetchErr:  a.fetchErr,
	}
	return runtime.NewPager(runtime.PagingHandler[armresources.ClientListByResourceGroupResponse]{
		More:    pager.moreFunc(),
		Fetcher: pager.fetcherFunc(),
	})
}

func (a fakeResourceAPI) BeginDeleteByID(ctx context.Context, resourceID string, apiVersion string,
	options *armresources.ClientBeginDeleteByIDOptions,
) (*runtime.Poller[armresources.ClientDeleteByIDResponse], error) {
	res := a.resources[resourceID]

	pollErr := res.pollErr
	if pollErr != nil {
		res.pollErr = nil
	}

	poller, err := runtime.NewPoller(nil, runtime.NewPipeline("", "", runtime.PipelineOptions{}, nil), &runtime.NewPollerOptions[armresources.ClientDeleteByIDResponse]{
		Handler: &stubPoller[armresources.ClientDeleteByIDResponse]{
			result:    armresources.ClientDeleteByIDResponse{},
			resultErr: pollErr,
		},
	})
	if err != nil {
		panic(err)
	}

	beginDeleteByIDErr := res.beginDeleteByIDErr
	if beginDeleteByIDErr != nil {
		res.beginDeleteByIDErr = nil
	}

	if res.beginDeleteByIDErr == nil && res.pollErr == nil {
		delete(a.resources, resourceID)
		fmt.Printf("fake delete %s\n", resourceID)
	} else {
		a.resources[resourceID] = res
	}

	return poller, beginDeleteByIDErr
}

type stubClientListByResourceGroupResponsePager struct {
	resources map[string]fakeResource
	fetchErr  error
	more      bool
}

func (p *stubClientListByResourceGroupResponsePager) moreFunc() func(
	armresources.ClientListByResourceGroupResponse) bool {
	return func(armresources.ClientListByResourceGroupResponse) bool {
		return p.more
	}
}

func (p *stubClientListByResourceGroupResponsePager) fetcherFunc() func(
	context.Context, *armresources.ClientListByResourceGroupResponse) (
	armresources.ClientListByResourceGroupResponse, error) {
	return func(context.Context, *armresources.ClientListByResourceGroupResponse) (
		armresources.ClientListByResourceGroupResponse, error,
	) {
		var resources []*armresources.GenericResourceExpanded
		for id := range p.resources {
			resources = append(resources, &armresources.GenericResourceExpanded{ID: proto.String(id)})
		}
		p.more = false
		return armresources.ClientListByResourceGroupResponse{
			ResourceListResult: armresources.ResourceListResult{
				Value: resources,
			},
		}, p.fetchErr
	}
}
