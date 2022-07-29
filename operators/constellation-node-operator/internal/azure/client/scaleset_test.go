package client

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetScaleSets(t *testing.T) {
	testCases := map[string]struct {
		scaleSet      armcomputev2.VirtualMachineScaleSet
		fetchPageErr  error
		wantScaleSets []string
		wantErr       bool
	}{
		"fetching scale sets works": {
			scaleSet: armcomputev2.VirtualMachineScaleSet{
				ID: to.Ptr("/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name"),
			},
			wantScaleSets: []string{"/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name"},
		},
		"fetching scale sets fails": {
			fetchPageErr: errors.New("fetch page error"),
			wantErr:      true,
		},
		"scale set is invalid": {},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			client := Client{
				scaleSetsAPI: &stubScaleSetsAPI{
					pager: &stubVMSSPager{
						list:     []armcomputev2.VirtualMachineScaleSet{tc.scaleSet},
						fetchErr: tc.fetchPageErr,
					},
				},
			}
			gotScaleSets, err := client.getScaleSets(context.Background())
			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.ElementsMatch(tc.wantScaleSets, gotScaleSets)
		})
	}
}
