package client

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	armcomputev2 "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v2"
	"github.com/edgelesssys/constellation/operators/constellation-node-operator/internal/poller"
)

// Client is a client for the Azure Cloud.
type Client struct {
	scaleSetsAPI
	virtualMachineScaleSetVMsAPI
	capacityPollerGenerator func(resourceGroup, scaleSet string, wantedCapacity int64) capacityPoller
	pollerOptions           *poller.PollUntilDoneOptions
}

// NewFromDefault creates a client with initialized clients.
func NewFromDefault(subscriptionID, tenantID string) (*Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	scaleSetAPI, err := armcomputev2.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}
	virtualMachineScaleSetVMsAPI, err := armcomputev2.NewVirtualMachineScaleSetVMsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
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
