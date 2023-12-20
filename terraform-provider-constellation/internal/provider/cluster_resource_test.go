/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccClusterResourceImports(t *testing.T) {
	// Set the path to the Terraform binary for acceptance testing when running under Bazel.
	bazelPreCheck := func() { bazelSetTerraformBinaryPath(t) }

	testCases := map[string]resource.TestCase{
		"import success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					resource "constellation_cluster" "test" {}
				`,
					ResourceName: "constellation_cluster.test",
					ImportState:  true,
					ImportStateId: "constellation-cluster://?kubeConfig=YWJjZGU=&" + // valid base64 of "abcde"
						"clusterEndpoint=b&" +
						"masterSecret=de&" +
						"masterSecretSalt=ad",
					ImportStateCheck: func(states []*terraform.InstanceState) error {
						state := states[0]
						assert := assert.New(t)
						assert.Equal("abcde", state.Attributes["kubeconfig"])
						assert.Equal("b", state.Attributes["out_of_cluster_endpoint"])
						assert.Equal("de", state.Attributes["master_secret"])
						assert.Equal("ad", state.Attributes["master_secret_salt"])
						return nil
					},
				},
			},
		},
		"kubeconfig not base64": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					resource "constellation_cluster" "test" {}
				`,
					ResourceName: "constellation_cluster.test",
					ImportState:  true,
					ImportStateId: "constellation-cluster://?kubeConfig=a&" +
						"clusterEndpoint=b&" +
						"masterSecret=de&" +
						"masterSecretSalt=ad",
					ExpectError: regexp.MustCompile(".*illegal base64 data.*"),
				},
			},
		},
		"mastersecret not hex": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					resource "constellation_cluster" "test" {}
				`,
					ResourceName: "constellation_cluster.test",
					ImportState:  true,
					ImportStateId: "constellation-cluster://?kubeConfig=test&" +
						"clusterEndpoint=b&" +
						"masterSecret=xx&" +
						"masterSecretSalt=ad",
					ExpectError: regexp.MustCompile(".*Decoding hex-encoded master secret.*"),
				},
			},
		},
		"parameter missing": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: testingConfig + `
					resource "constellation_cluster" "test" {}
				`,
					ResourceName: "constellation_cluster.test",
					ImportState:  true,
					ImportStateId: "constellation-cluster://?kubeConfig=test&" +
						"clusterEndpoint=b&" +
						"masterSecret=xx&",
					ExpectError: regexp.MustCompile(".*Missing query parameter.*"),
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

func TestAccClusterResource(t *testing.T) {
	// Set the path to the Terraform binary for acceptance testing when running under Bazel.
	bazelPreCheck := func() { bazelSetTerraformBinaryPath(t) }

	testCases := map[string]resource.TestCase{
		"master secret not hex": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + `
					resource "constellation_cluster" "test" {
						csp                     = "aws"
						name                    = "constell"
						uid                     = "test"
						image                   = data.constellation_image.bar.image
						attestation             = data.constellation_attestation.foo.attestation
						init_secret             = "deadbeef"
						master_secret           = "xxx"
						master_secret_salt      = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						measurement_salt        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						out_of_cluster_endpoint = "192.0.2.1"
						in_cluster_endpoint     = "192.0.2.1"
						network_config = {
						  ip_cidr_node    = "0.0.0.0/24"
						  ip_cidr_service = "0.0.0.0/24"
						}
					  }
				`,
					ExpectError: regexp.MustCompile(".*Master secret must be a hex-encoded 32-byte.*"),
				},
			},
		},
		"master secret salt not hex": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + `
					resource "constellation_cluster" "test" {
						csp                     = "aws"
						name                    = "constell"
						uid                     = "test"
						image                   = data.constellation_image.bar.image
						attestation             = data.constellation_attestation.foo.attestation
						init_secret             = "deadbeef"
						master_secret           = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						master_secret_salt      = "xxx"
						measurement_salt        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						out_of_cluster_endpoint = "192.0.2.1"
						in_cluster_endpoint     = "192.0.2.1"
						network_config = {
						  ip_cidr_node    = "0.0.0.0/24"
						  ip_cidr_service = "0.0.0.0/24"
						}
					  }
				`,
					ExpectError: regexp.MustCompile(".*Master secret salt must be a hex-encoded 32-byte.*"),
				},
			},
		},
		"measurement salt not hex": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + `
					resource "constellation_cluster" "test" {
						csp                     = "aws"
						name                    = "constell"
						uid                     = "test"
						image                   = data.constellation_image.bar.image
						attestation             = data.constellation_attestation.foo.attestation
						init_secret             = "deadbeef"
						master_secret           = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						master_secret_salt      = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						measurement_salt        = "xxx"
						out_of_cluster_endpoint = "192.0.2.1"
						in_cluster_endpoint     = "192.0.2.1"
						network_config = {
						  ip_cidr_node    = "0.0.0.0/24"
						  ip_cidr_service = "0.0.0.0/24"
						}
					  }
				`,
					ExpectError: regexp.MustCompile(".*Measurement salt must be a hex-encoded 32-byte.*"),
				},
			},
		},
		"invalid node ip cidr": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + `
					resource "constellation_cluster" "test" {
						csp                     = "aws"
						name                    = "constell"
						uid                     = "test"
						image                   = data.constellation_image.bar.image
						attestation             = data.constellation_attestation.foo.attestation
						init_secret             = "deadbeef"
						master_secret           = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						master_secret_salt      = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						measurement_salt        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						out_of_cluster_endpoint = "192.0.2.1"
						in_cluster_endpoint     = "192.0.2.1"
						network_config = {
						  ip_cidr_node    = "0.0.0x.0/xxx"
						  ip_cidr_service = "0.0.0.0/24"
						}
					  }
				`,
					ExpectError: regexp.MustCompile(".*Node IP CIDR must be a valid CIDR.*"),
				},
			},
		},
		"invalid service ip cidr": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + `
					resource "constellation_cluster" "test" {
						csp                     = "aws"
						name                    = "constell"
						uid                     = "test"
						image                   = data.constellation_image.bar.image
						attestation             = data.constellation_attestation.foo.attestation
						init_secret             = "deadbeef"
						master_secret           = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						master_secret_salt      = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						measurement_salt        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						out_of_cluster_endpoint = "192.0.2.1"
						in_cluster_endpoint     = "192.0.2.1"
						network_config = {
						  ip_cidr_node    = "0.0.0.0/24"
						  ip_cidr_service = "0.0.0x.0/xxx"
						}
					  }
				`,
					ExpectError: regexp.MustCompile(".*Service IP CIDR must be a valid CIDR.*"),
				},
			},
		},
		"azure config missing": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "azure") + `
					resource "constellation_cluster" "test" {
						csp                     = "azure"
						name                    = "constell"
						uid                     = "test"
						image                   = data.constellation_image.bar.image
						attestation             = data.constellation_attestation.foo.attestation
						init_secret             = "deadbeef"
						master_secret           = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						master_secret_salt      = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						measurement_salt        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						out_of_cluster_endpoint = "192.0.2.1"
						in_cluster_endpoint     = "192.0.2.1"
						network_config = {
						  ip_cidr_node    = "0.0.0.0/24"
						  ip_cidr_service = "0.0.0.0/24"
						}
					  }
				`,
					ExpectError: regexp.MustCompile(".*When csp is set to 'azure', the 'azure' configuration must be set.*"),
				},
			},
		},
		"gcp config missing": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "gcp") + `
					resource "constellation_cluster" "test" {
						csp                     = "gcp"
						name                    = "constell"
						uid                     = "test"
						image                   = data.constellation_image.bar.image
						attestation             = data.constellation_attestation.foo.attestation
						init_secret             = "deadbeef"
						master_secret           = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						master_secret_salt      = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						measurement_salt        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						out_of_cluster_endpoint = "192.0.2.1"
						in_cluster_endpoint     = "192.0.2.1"
						network_config = {
						  ip_cidr_node    = "0.0.0.0/24"
						  ip_cidr_service = "0.0.0.0/24"
						  ip_cidr_pod    = "0.0.0.0/24"
						}
					  }
				`,
					ExpectError: regexp.MustCompile(".*When csp is set to 'gcp', the 'gcp' configuration must be set.*"),
				},
			},
		},
		"gcp pod ip cidr missing": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion("v2.13.0"),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "gcp") + `
					resource "constellation_cluster" "test" {
						csp                     = "gcp"
						name                    = "constell"
						uid                     = "test"
						image                   = data.constellation_image.bar.image
						attestation             = data.constellation_attestation.foo.attestation
						init_secret             = "deadbeef"
						master_secret           = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						master_secret_salt      = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						measurement_salt        = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
						out_of_cluster_endpoint = "192.0.2.1"
						in_cluster_endpoint     = "192.0.2.1"
						network_config = {
						  ip_cidr_node    = "0.0.0.0/24"
						  ip_cidr_service = "0.0.0.0/24"
						}
						gcp = {
							project_id = "test"
							service_account_key = "eyJ0ZXN0IjogInRlc3QifQ=="
						}
					  }
				`,
					ExpectError: regexp.MustCompile(".*When csp is set to 'gcp', 'ip_cidr_pod' must be set.*"),
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

func fullClusterTestingConfig(t *testing.T, csp string) string {
	t.Helper()

	providerConfig := `
	provider "constellation" {}
	`

	switch csp {
	case "aws":
		return providerConfig + `
		data "constellation_image" "bar" {
			version             = "v2.13.0"
			attestation_variant = "aws-sev-snp"
			csp                 = "aws"
			region			    = "us-east-2"
		}

		data "constellation_attestation" "foo" {
			csp                 = "aws"
			attestation_variant = "aws-sev-snp"
			image               = data.constellation_image.bar.image
		}`
	case "azure":
		return providerConfig + `
		data "constellation_image" "bar" {
			version             = "v2.13.0"
			attestation_variant = "azure-sev-snp"
			csp                 = "azure"
		}

		data "constellation_attestation" "foo" {
			csp                 = "azure"
			attestation_variant = "azure-sev-snp"
			image               = data.constellation_image.bar.image
		}`
	case "gcp":
		return providerConfig + `
		data "constellation_image" "bar" {
			version             = "v2.13.0"
			attestation_variant = "gcp-sev-es"
			csp                 = "gcp"
		}

		data "constellation_attestation" "foo" {
			csp                 = "gcp"
			attestation_variant = "gcp-sev-es"
			image               = data.constellation_image.bar.image
		}`
	default:
		t.Fatal("unknown csp")
		return ""
	}
}
