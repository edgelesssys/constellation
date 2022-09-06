/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/edgelesssys/constellation/internal/cloud/cloudtypes"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstances(t *testing.T) {
	someErr := errors.New("failed")
	testCases := map[string]struct {
		publicIPAddressesAPI publicIPAddressesAPI
		networkInterfacesAPI networkInterfacesAPI
		scaleSetsAPI         scaleSetsAPI
		createInstancesInput CreateInstancesInput
		wantErr              bool
	}{
		"successful create": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			scaleSetsAPI: stubScaleSetsAPI{
				getResponse: armcomputev2.VirtualMachineScaleSet{
					Identity: &armcomputev2.VirtualMachineScaleSetIdentity{PrincipalID: to.Ptr("principal-id")}, SKU: &armcomputev2.SKU{Capacity: to.Ptr[int64](0)},
				},
			},
			createInstancesInput: CreateInstancesInput{
				CountControlPlanes:   3,
				CountWorkers:         3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
				ConfidentialVM:       true,
			},
		},
		"error when creating scale set": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			scaleSetsAPI:         stubScaleSetsAPI{createErr: someErr},
			createInstancesInput: CreateInstancesInput{
				CountControlPlanes:   3,
				CountWorkers:         3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
				ConfidentialVM:       true,
			},
			wantErr: true,
		},
		"error when polling create scale set response": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{},
			scaleSetsAPI:         stubScaleSetsAPI{getErr: someErr},
			createInstancesInput: CreateInstancesInput{
				CountControlPlanes:   3,
				CountWorkers:         3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
				ConfidentialVM:       true,
			},
			wantErr: true,
		},
		"error when retrieving private IPs": {
			publicIPAddressesAPI: stubPublicIPAddressesAPI{},
			networkInterfacesAPI: stubNetworkInterfacesAPI{getErr: someErr},
			scaleSetsAPI:         stubScaleSetsAPI{},
			createInstancesInput: CreateInstancesInput{
				CountWorkers:         3,
				InstanceType:         "type",
				Image:                "image",
				UserAssingedIdentity: "identity",
				ConfidentialVM:       true,
			},
			wantErr: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := context.Background()

			client := Client{
				location:             "location",
				name:                 "name",
				uid:                  "uid",
				resourceGroup:        "name",
				publicIPAddressesAPI: tc.publicIPAddressesAPI,
				networkInterfacesAPI: tc.networkInterfacesAPI,
				scaleSetsAPI:         tc.scaleSetsAPI,
				workers:              make(cloudtypes.Instances),
				controlPlanes:        make(cloudtypes.Instances),
				loadBalancerPubIP:    "lbip",
			}

			if tc.wantErr {
				assert.Error(client.CreateInstances(ctx, tc.createInstancesInput))
			} else {
				assert.NoError(client.CreateInstances(ctx, tc.createInstancesInput))
				assert.Equal(tc.createInstancesInput.CountControlPlanes, len(client.controlPlanes))
				assert.Equal(tc.createInstancesInput.CountWorkers, len(client.workers))
				assert.NotEmpty(client.workers["0"].PrivateIP)
				assert.NotEmpty(client.workers["0"].PublicIP)
				assert.NotEmpty(client.controlPlanes["0"].PrivateIP)
				assert.NotEmpty(client.controlPlanes["0"].PublicIP)
			}
		})
	}
}
