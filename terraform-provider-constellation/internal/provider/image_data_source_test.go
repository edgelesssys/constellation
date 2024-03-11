/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccImageDataSource(t *testing.T) {
	// Set the path to the Terraform binary for acceptance testing when running under Bazel.
	bazelPreCheck := func() { bazelSetTerraformBinaryPath(t) }

	testCases := map[string]resource.TestCase{
		"no version succeeds": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						attestation_variant = "aws-sev-snp"
						csp                 = "aws"
						region              = "eu-west-1"
					}
				`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrSet("data.constellation_image.test", "image.reference"),
						resource.TestCheckResourceAttrSet("data.constellation_image.test", "image.version"),
						resource.TestCheckResourceAttrSet("data.constellation_image.test", "image.short_path"),
					),
				},
			},
		},
		"aws succcess": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.13.0"
						attestation_variant = "aws-sev-snp"
						csp                 = "aws"
						region              = "eu-west-1"
					}
				`,
					Check: resource.TestCheckResourceAttr("data.constellation_image.test", "image.reference", "ami-04f8d522b113b73bf"), // should be immutable

				},
			},
		},
		"aws without region fails": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.13.0"
						attestation_variant = "aws-sev-snp"
						csp                 = "aws"
					}
				`,
					ExpectError: regexp.MustCompile(".*Region must be set for AWS.*"),
				},
			},
		},
		"aws marketplace success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.13.0"
						attestation_variant = "aws-sev-snp"
						csp                 = "aws"
						marketplace_image   = true
						region              = "eu-west-1"
					}
				`,
					Check: resource.TestCheckResourceAttr("data.constellation_image.test", "image.reference", "resolve:ssm:/aws/service/marketplace/prod-77ylkenlkgufs/v2.13.0"), // should be immutable,
				},
			},
		},
		"azure success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.13.0"
						attestation_variant = "azure-sev-snp"
						csp                 = "azure"
					}
				`,
					Check: resource.TestCheckResourceAttr("data.constellation_image.test", "image.reference", "/communityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/images/constellation/versions/2.13.0"), // should be immutable

				},
			},
		},
		"azure marketplace success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.13.0"
						attestation_variant = "azure-sev-snp"
						csp                 = "azure"
						marketplace_image   = true
					}
				`,
					Check: resource.TestCheckResourceAttr("data.constellation_image.test", "image.reference", "constellation-marketplace-image://Azure?offer=constellation&publisher=edgelesssystems&sku=constellation&version=2.13.0"), // should be immutable

				},
			},
		},
		"gcp success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.13.0"
						attestation_variant = "gcp-sev-es"
						csp                 = "gcp"
					}
				`,
					Check: resource.TestCheckResourceAttr("data.constellation_image.test", "image.reference", "projects/constellation-images/global/images/v2-13-0-gcp-sev-es-stable"), // should be immutable,
				},
			},
		},
		"stackit success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.16.0"
						attestation_variant = "qemu-vtpm"
						csp                 = "stackit"
					}
				`,
					Check: resource.TestCheckResourceAttr("data.constellation_image.test", "image.reference", "8ffc1740-1e41-4281-b872-f8088ffd7692"), // should be immutable,
				},
			},
		},
		"openstack success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.16.0"
						attestation_variant = "qemu-vtpm"
						csp                 = "openstack"
					}
				`,
					Check: resource.TestCheckResourceAttr("data.constellation_image.test", "image.reference", "8ffc1740-1e41-4281-b872-f8088ffd7692"), // should be immutable,
				},
			},
		},
		"unknown attestation variant": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.13.0"
						attestation_variant = "unknown"
						csp                 = "azure"
					}
				`,
					ExpectError: regexp.MustCompile(".*Attribute attestation_variant value must be one of.*"),
				},
			},
		},
		"unknown csp": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.13.0"
						attestation_variant = "azure-sev-snp"
						csp                 = "unknown"
					}
				`,
					ExpectError: regexp.MustCompile(".*Attribute csp value must be one of.*"),
				},
			},
		},
		"invalid version": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version             = "xxx"
						attestation_variant = "azure-sev-snp"
						csp                 = "azure"
					}
				`,
					ExpectError: regexp.MustCompile(".*Invalid Version.*"),
				},
			},
		},
		"gcp marketplace success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					data "constellation_image" "test" {
						version       = "v2.13.0"
						attestation_variant = "gcp-sev-es"
						csp                 = "gcp"
						marketplace_image   = true
					}
				`,
					Check: resource.TestCheckResourceAttr("data.constellation_image.test", "image.reference", "projects/mpi-edgeless-systems-public/global/images/v2-13-0-gcp-sev-es-stable"), // should be immutable,
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
