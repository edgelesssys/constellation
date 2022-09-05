/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package client

import (
	"context"
	"fmt"
)

// getScaleSets retrieves the IDs of all scale sets of a resource group.
func (c *Client) getScaleSets(ctx context.Context) ([]string, error) {
	pager := c.scaleSetsAPI.NewListPager(c.config.ResourceGroup, nil)
	var scaleSets []string

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("paging scale sets: %w", err)
		}
		for _, scaleSet := range page.Value {
			if scaleSet == nil || scaleSet.ID == nil {
				continue
			}
			scaleSets = append(scaleSets, *scaleSet.ID)
		}
	}
	return scaleSets, nil
}
