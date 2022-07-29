package client

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/edgelesssys/constellation/operators/constellation-node-operator/internal/poller"
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

	scaleSetAPI, err := armcomputev2.NewVirtualMachineScaleSetsClient(config.SubscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	virtualMachineScaleSetVMsAPI, err := armcomputev2.NewVirtualMachineScaleSetVMsClient(config.SubscriptionID, cred, nil)
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
