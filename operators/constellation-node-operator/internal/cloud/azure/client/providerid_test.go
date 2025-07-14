/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

			subscriptionID, resourceGroup, scaleSet, instanceName, err := scaleSetInformationFromProviderID(tc.providerID)

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
