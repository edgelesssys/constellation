/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/config/instancetypes"
	"github.com/spf13/cobra"
)

func newConfigInstanceTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance-types",
		Short: "Print the supported instance types for all cloud providers",
		Long:  "Print the supported instance types for all cloud providers.",
		Args:  cobra.ArbitraryArgs,
		Run:   printSupportedInstanceTypes,
	}

	return cmd
}

func printSupportedInstanceTypes(cmd *cobra.Command, args []string) {
	cmd.Printf(`AWS instance families:
%v
Azure instance types:
%v
GCP instance types:
%v
`, formatInstanceTypes(instancetypes.AWSSupportedInstanceFamilies), formatInstanceTypes(instancetypes.AzureInstanceTypes), formatInstanceTypes(instancetypes.GCPInstanceTypes))
}

func formatInstanceTypes(types []string) string {
	return "\t" + strings.Join(types, "\n\t")
}
