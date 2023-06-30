/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/terraform"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
)

// terraformMigrationAWSNodeGroups migrates the AWS node groups from the old state to the new state.
// TODO(AB#3248): Remove this migration after we can assume that all existing clusters have been migrated.
func terraformMigrationAWSNodeGroups(csp cloudprovider.Provider, zone string) []terraform.StateMigration {
	if csp != cloudprovider.AWS {
		return nil
	}
	return []terraform.StateMigration{
		{
			DisplayName: "AWS node groups",
			Hook: func(ctx context.Context, tfClient terraform.TFMigrator) error {
				fromTo := []struct {
					from string
					to   string
				}{
					{
						from: "aws_eip.lb",
						to:   fmt.Sprintf("aws_eip.lb[%q]", zone),
					},
					{
						from: "module.public_private_subnet.aws_eip.nat",
						to:   fmt.Sprintf("module.public_private_subnet.aws_eip.nat[%q]", zone),
					},
					{
						from: "module.public_private_subnet.aws_nat_gateway.gw",
						to:   fmt.Sprintf("module.public_private_subnet.aws_nat_gateway.gw[%q]", zone),
					},
					{
						from: "module.public_private_subnet.aws_route_table.private_nat",
						to:   fmt.Sprintf("module.public_private_subnet.aws_route_table.private_nat[%q]", zone),
					},
					{
						from: "module.public_private_subnet.aws_route_table.public_igw",
						to:   fmt.Sprintf("module.public_private_subnet.aws_route_table.public_igw[%q]", zone),
					},
					{
						from: "module.public_private_subnet.aws_route_table_association.private-nat",
						to:   fmt.Sprintf("module.public_private_subnet.aws_route_table_association.private_nat[%q]", zone),
					},
					{
						from: "module.public_private_subnet.aws_route_table_association.route_to_internet",
						to:   fmt.Sprintf("module.public_private_subnet.aws_route_table_association.route_to_internet[%q]", zone),
					},
					{
						from: "module.public_private_subnet.aws_subnet.private",
						to:   fmt.Sprintf("module.public_private_subnet.aws_subnet.private[%q]", zone),
					},
					{
						from: "module.public_private_subnet.aws_subnet.public",
						to:   fmt.Sprintf("module.public_private_subnet.aws_subnet.public[%q]", zone),
					},
				}

				for _, move := range fromTo {
					// we need to drop the error here, because the migration has to be idempotent
					// and state mv will fail if the state is already migrated
					_ = tfClient.StateMv(ctx, move.from, move.to)
				}
				return nil
			},
		},
	}
}
