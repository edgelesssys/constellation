/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azureshared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicsFromProviderID(t *testing.T) {
	testCases := map[string]struct {
		providerID         string
		wantErr            bool
		wantSubscriptionID string
		wantResourceGroup  string
	}{
		"providerID for scale set instance works": {
			providerID:         "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			wantSubscriptionID: "subscription-id",
			wantResourceGroup:  "resource-group",
		},
		"providerID is malformed": {
			providerID: "malformed-provider-id",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			subscriptionID, resourceGroup, err := BasicsFromProviderID(tc.providerID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			require.NoError(err)
			assert.Equal(tc.wantSubscriptionID, subscriptionID)
			assert.Equal(tc.wantResourceGroup, resourceGroup)
		})
	}
}

func TestScaleSetInformationFromProviderID(t *testing.T) {
	testCases := map[string]struct {
		providerID         string
		wantSubscriptionID string
		wantResourceGroup  string
		wantScaleSet       string
		wantInstanceID     string
		wantErr            bool
	}{
		"providerID for scale set instance works": {
			providerID:         "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name/virtualMachines/instance-id",
			wantSubscriptionID: "subscription-id",
			wantResourceGroup:  "resource-group",
			wantScaleSet:       "scale-set-name",
			wantInstanceID:     "instance-id",
		},
		"providerID for individual instance must fail": {
			providerID: "azure:///subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachines/instance-name",
			wantErr:    true,
		},
		"providerID is malformed": {
			providerID: "malformed-provider-id",
			wantErr:    true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			subscriptionID, resourceGroup, scaleSet, instanceName, err := ScaleSetInformationFromProviderID(tc.providerID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantSubscriptionID, subscriptionID)
			assert.Equal(tc.wantResourceGroup, resourceGroup)
			assert.Equal(tc.wantScaleSet, scaleSet)
			assert.Equal(tc.wantInstanceID, instanceName)
		})
	}
}
