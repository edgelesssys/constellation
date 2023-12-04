/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccClusterResource(t *testing.T) {
	// Set the path to the Terraform binary for acceptance testing when running under Bazel.
	bazelPreCheck := func() { bazelSetTerraformBinaryPath(t) }

	testCases := map[string]resource.TestCase{
		"azure sev-snp success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_attestation" "att" {
						csp = "azure"
						attestation_variant = "azure-sev-snp"
						image_version = "v2.13.0"
					}

					resource "constellation_cluster" "test" {
						uid = "test"
						master_secret = "secret"
						init_secret = "secret"
						attestation = data.constellation_attestation.att.attestation
						extra_microservices = {
							csi_driver = true
						}
					}
					`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("constellation_cluster.test", "attestation.bootloader_version", "3"),
						// TODO(elchead): check output in follow up PR
					),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			resource.Test(t, tc)
		})
	}
}
