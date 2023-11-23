/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccImageDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { bazelSetTerraformBinaryPath(t) },
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testingConfig + `
					data "constellation_image" "test" {
						image_version       = "v2.13.0"
						attestation_variant = "aws-sev-snp"
						csp                 = "aws"
						region              = "eu-west-1"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.constellation_image.test", "reference", "ami-04f8d522b113b73bf"), // should be immutable
					resource.TestCheckResourceAttr("data.constellation_image.test", "id", "placeholder"),
				),
			},
		},
	})
}
