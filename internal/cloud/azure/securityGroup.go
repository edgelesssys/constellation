package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

// getNetworkSecurityGroup retrieves the list of security groups for the given resource group.
func (m *Metadata) getNetworkSecurityGroup(ctx context.Context, resourceGroup string) (*armnetwork.SecurityGroup, error) {
	pager := m.securityGroupsAPI.NewListPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving security groups: %w", err)
		}
		for _, securityGroup := range page.Value {
			return securityGroup, nil
		}
	}
	return nil, fmt.Errorf("no security group found for resource group %q", resourceGroup)
}
