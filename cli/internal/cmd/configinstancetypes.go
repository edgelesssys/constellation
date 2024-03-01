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

func printSupportedInstanceTypes(cmd *cobra.Command, _ []string) {
	cmd.Printf(`AWS SNP-enabled instance types:
%v
AWS NitroTPM-enabled instance types:
%v
Azure Intel TDX instance types:
%v
Azure AMD SEV-SNP instance types:
%v
Azure Trusted Launch instance types:
%v
GCP instance types:
%v
STACKIT instance types:
%v
`,
		formatInstanceTypes(instancetypes.AWSSNPSupportedInstanceFamilies),
		formatInstanceTypes(instancetypes.AWSSupportedInstanceFamilies),
		formatInstanceTypes(instancetypes.AzureTDXInstanceTypes),
		formatInstanceTypes(instancetypes.AzureSNPInstanceTypes),
		formatInstanceTypes(instancetypes.AzureTrustedLaunchInstanceTypes),
		formatInstanceTypes(instancetypes.GCPInstanceTypes),
		formatInstanceTypes(instancetypes.STACKITInstanceTypes),
	)
}

func formatInstanceTypes(types []string) string {
	return "\t" + strings.Join(types, "\n\t")
}
