package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// getVirtualNetwork return the first virtual network found in the resource group.
func (m *Metadata) getVirtualNetwork(ctx context.Context, resourceGroup string) (*armnetwork.VirtualNetwork, error) {
	pager := m.virtualNetworksAPI.List(resourceGroup, nil)
	for pager.NextPage(ctx) {
		for _, network := range pager.PageResponse().Value {
			if network != nil {
				return network, nil
			}
		}
	}
	return nil, fmt.Errorf("no virtual network found in resource group %s", resourceGroup)
}
