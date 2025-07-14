/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinVMSSID(t *testing.T) {
	assert.Equal(t,
		"/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set",
		joinVMSSID("subscription-id", "resource-group", "scale-set"))
}

func TestSplitVMSSID(t *testing.T) {
	testCases := map[string]struct {
		vmssID             string
		wantSubscriptionID string
		wantResourceGroup  string
		wantScaleSet       string
		wantInstanceID     string
		wantErr            bool
	}{
		"vmssID can be splitted": {
			vmssID:             "/subscriptions/subscription-id/resourceGroups/resource-group/providers/Microsoft.Compute/virtualMachineScaleSets/scale-set-name",
			wantSubscriptionID: "subscription-id",
			wantResourceGroup:  "resource-group",
			wantScaleSet:       "scale-set-name",
			wantInstanceID:     "instance-id",
		},
		"vmssID is malformed": {
			vmssID:  "malformed-vmss-id",
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			subscriptionID, resourceGroup, scaleSet, err := splitVMSSID(tc.vmssID)

			if tc.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.wantSubscriptionID, subscriptionID)
			assert.Equal(tc.wantResourceGroup, resourceGroup)
			assert.Equal(tc.wantScaleSet, scaleSet)
		})
	}
}
