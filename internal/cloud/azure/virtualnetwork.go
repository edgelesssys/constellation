/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// getVirtualNetwork return the first virtual network found in the resource group.
func (m *Metadata) getVirtualNetwork(ctx context.Context, resourceGroup string) (*armnetwork.VirtualNetwork, error) {
	pager := m.virtualNetworksAPI.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving virtual networks: %w", err)
		}
		for _, network := range page.Value {
			if network != nil {
				return network, nil
			}
		}
	}
	return nil, fmt.Errorf("no virtual network found in resource group %s", resourceGroup)
}
