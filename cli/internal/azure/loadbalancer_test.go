/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/stretchr/testify/assert"
)

func TestAppendDebugRules(t *testing.T) {
	assert := assert.New(t)

	// Test with empty rules
	emptyAzureLoadBalancer := armnetwork.LoadBalancer{}
	someLoadBalancer := LoadBalancer{
		Name:          "test",
		Subscription:  "00000000-0000-0000-0000-000000000000",
		Location:      "westeurope",
		ResourceGroup: "test-resource-group",
		PublicIPID:    "some-public-ip-id",
		UID:           "test-uid",
	}

	appendedEmptyAzureLoadBalancer := someLoadBalancer.AppendDebugRules(emptyAzureLoadBalancer)
	assert.Equal("debugdLoadBalancerRule", *(appendedEmptyAzureLoadBalancer.Properties.LoadBalancingRules[0]).Name, "Debug load balancer rule not found at index 0")

	// Test with existing rules
	defaultAzureLoadBalancer := someLoadBalancer.Azure()
	appendedDefaultAzureLoadBalancer := someLoadBalancer.AppendDebugRules(defaultAzureLoadBalancer)
	var foundDebugLoadBalancer bool
	for _, rule := range appendedDefaultAzureLoadBalancer.Properties.LoadBalancingRules {
		if *(rule).Name == "debugdLoadBalancerRule" {
			foundDebugLoadBalancer = true
		}
	}
	assert.True(foundDebugLoadBalancer, "Debug load balancer rule not found")
}
