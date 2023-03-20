/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/applicationinsights/armapplicationinsights"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/edgelesssys/constellation/v2/internal/cloud"
	"github.com/stretchr/testify/assert"
)

func TestGetAppInsightsKey(t *testing.T) {
	someErr := errors.New("failed")
	goodAppInsights := armapplicationinsights.Component{
		Tags: map[string]*string{
			cloud.TagUID: to.StringPtr("uid"),
		},
		Properties: &armapplicationinsights.ComponentProperties{
			InstrumentationKey: to.StringPtr("key"),
		},
	}

	testCases := map[string]struct {
		imds        *stubIMDSAPI
		appInsights *stubApplicationsInsightsAPI
		wantKey     string
		wantErr     bool
	}{
		"success": {
			imds: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			appInsights: &stubApplicationsInsightsAPI{
				pager: &stubApplicationKeyPager{list: []armapplicationinsights.Component{goodAppInsights}},
			},
			wantKey: "key",
		},
		"multiple apps": {
			imds: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			appInsights: &stubApplicationsInsightsAPI{
				pager: &stubApplicationKeyPager{list: []armapplicationinsights.Component{
					{
						Tags: map[string]*string{
							cloud.TagUID: to.StringPtr("different-uid"),
						},
						Properties: &armapplicationinsights.ComponentProperties{
							InstrumentationKey: to.StringPtr("different-key"),
						},
					},
					goodAppInsights,
				}},
			},
			wantKey: "key",
		},
		"missing properties": {
			imds: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			appInsights: &stubApplicationsInsightsAPI{
				pager: &stubApplicationKeyPager{list: []armapplicationinsights.Component{
					{
						Tags: map[string]*string{
							cloud.TagUID: to.StringPtr("uid"),
						},
					},
				}},
			},
			wantErr: true,
		},
		"no app with matching uid": {
			imds: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			appInsights: &stubApplicationsInsightsAPI{
				pager: &stubApplicationKeyPager{list: []armapplicationinsights.Component{
					{
						Tags: map[string]*string{
							cloud.TagUID: to.StringPtr("different-uid"),
						},
						Properties: &armapplicationinsights.ComponentProperties{
							InstrumentationKey: to.StringPtr("different-key"),
						},
					},
				}},
			},
			wantErr: true,
		},
		"imds resource group error": {
			imds: &stubIMDSAPI{
				resourceGroupErr: someErr,
				uidVal:           "uid",
			},
			appInsights: &stubApplicationsInsightsAPI{
				pager: &stubApplicationKeyPager{list: []armapplicationinsights.Component{goodAppInsights}},
			},
			wantErr: true,
		},
		"imds uid error": {
			imds: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidErr:           someErr,
			},
			appInsights: &stubApplicationsInsightsAPI{
				pager: &stubApplicationKeyPager{list: []armapplicationinsights.Component{goodAppInsights}},
			},
			wantErr: true,
		},
		"app insights list error": {
			imds: &stubIMDSAPI{
				resourceGroupVal: "resource-group",
				uidVal:           "uid",
			},
			appInsights: &stubApplicationsInsightsAPI{
				pager: &stubApplicationKeyPager{fetchErr: someErr},
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			key, err := getAppInsightsKey(context.Background(), tc.imds, tc.appInsights)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantKey, key)
			}
		})
	}
}

type stubApplicationKeyPager struct {
	list     []armapplicationinsights.Component
	fetchErr error
	more     bool
}

func (p *stubApplicationKeyPager) moreFunc() func(armapplicationinsights.ComponentsClientListByResourceGroupResponse) bool {
	return func(armapplicationinsights.ComponentsClientListByResourceGroupResponse) bool {
		return p.more
	}
}

func (p *stubApplicationKeyPager) fetcherFunc() func(context.Context, *armapplicationinsights.ComponentsClientListByResourceGroupResponse,
) (armapplicationinsights.ComponentsClientListByResourceGroupResponse, error) {
	return func(context.Context, *armapplicationinsights.ComponentsClientListByResourceGroupResponse) (armapplicationinsights.ComponentsClientListByResourceGroupResponse, error) {
		page := make([]*armapplicationinsights.Component, len(p.list))
		for i := range p.list {
			page[i] = &p.list[i]
		}
		return armapplicationinsights.ComponentsClientListByResourceGroupResponse{
			ComponentListResult: armapplicationinsights.ComponentListResult{
				Value: page,
			},
		}, p.fetchErr
	}
}

type stubApplicationsInsightsAPI struct {
	pager *stubApplicationKeyPager
}

func (a *stubApplicationsInsightsAPI) NewListByResourceGroupPager(_ string, _ *armapplicationinsights.ComponentsClientListByResourceGroupOptions,
) *runtime.Pager[armapplicationinsights.ComponentsClientListByResourceGroupResponse] {
	return runtime.NewPager(runtime.PagingHandler[armapplicationinsights.ComponentsClientListByResourceGroupResponse]{
		More:    a.pager.moreFunc(),
		Fetcher: a.pager.fetcherFunc(),
	})
}
