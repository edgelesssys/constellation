/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/edgelesssys/constellation/v2/operators/constellation-node-operator/internal/poller"
	"github.com/spf13/afero"
)

// Client is a client for the Azure Cloud.
type Client struct {
	config cloudConfig
	scaleSetsAPI
	virtualMachineScaleSetVMsAPI
	capacityPollerGenerator func(resourceGroup, scaleSet string, wantedCapacity int64) capacityPoller
	pollerOptions           *poller.PollUntilDoneOptions
}

// NewFromDefault creates a client with initialized clients.
func NewFromDefault(configPath string) (*Client, error) {
	config, err := loadConfig(afero.NewOsFs(), configPath)
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	scaleSetAPI, err := armcompute.NewVirtualMachineScaleSetsClient(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	virtualMachineScaleSetVMsAPI, err := armcompute.NewVirtualMachineScaleSetVMsClient(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		config:                       *config,
		scaleSetsAPI:                 scaleSetAPI,
		virtualMachineScaleSetVMsAPI: virtualMachineScaleSetVMsAPI,
		capacityPollerGenerator: func(resourceGroup, scaleSet string, wantedCapacity int64) capacityPoller {
			return poller.New[int64](&capacityPollingHandler{
				resourceGroup:  resourceGroup,
				scaleSet:       scaleSet,
				wantedCapacity: wantedCapacity,
				scaleSetsAPI:   scaleSetAPI,
			})
		},
	}, nil
}
