/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/edgelesssys/constellation/v2/terraform-provider-constellation/internal/data"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const providerVersion string = "v2.14.0"

func TestMicroserviceConstraint(t *testing.T) {
	providerVersion := semver.NewFromInt(2, 15, 0, "")
	sut := &ClusterResource{
		providerData: data.ProviderData{
			Version: providerVersion,
		},
	}
	testCases := []struct {
		name               string
		version            string
		expectedErrorCount int
	}{
		{
			name:               "outdated by 2 minor  versions is invalid",
			version:            "v2.13.0",
			expectedErrorCount: 1,
		},
		{
			name:               "outdated by 1 minor is allowed for upgrade",
			version:            "v2.14.0",
			expectedErrorCount: 0,
		},
		{
			name:               "same version is valid",
			version:            "v2.15.0",
			expectedErrorCount: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, diags := sut.getMicroserviceVersion(&ClusterResourceModel{
				MicroserviceVersion: basetypes.NewStringValue(tc.version),
			})
			require.Equal(t, tc.expectedErrorCount, diags.ErrorsCount())
		})
	}
}

func TestViolatedImageConstraint(t *testing.T) {
	sut := &ClusterResource{
		providerData: data.ProviderData{
			Version: semver.NewFromInt(2, 15, 0, ""),
		},
	}
	testCases := []struct {
		name               string
		version            string
		expectedErrorCount int
	}{
		{
			name:               "outdated by 2 minor versions is invalid",
			version:            "v2.13.0",
			expectedErrorCount: 1,
		},
		{
			name:               "outdated by 1 minor is allowed for upgrade",
			version:            "v2.14.0",
			expectedErrorCount: 0,
		},
		{
			name:               "same version is valid",
			version:            "v2.15.0",
			expectedErrorCount: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img := imageAttribute{
				Version: tc.version,
			}

			input, diags := basetypes.NewObjectValueFrom(context.Background(), map[string]attr.Type{
				"version":           basetypes.StringType{},
				"reference":         basetypes.StringType{},
				"short_path":        basetypes.StringType{},
				"marketplace_image": basetypes.BoolType{},
			}, img)
			require.Equal(t, 0, diags.ErrorsCount())
			_, _, diags2 := sut.getImageVersion(context.Background(), &ClusterResourceModel{
				Image: input,
			})
			require.Equal(t, tc.expectedErrorCount, diags2.ErrorsCount())
		})
	}
}

func TestAccClusterResourceImports(t *testing.T) {
	// Set the path to the Terraform binary for acceptance testing when running under Bazel.
	bazelPreCheck := func() { bazelSetTerraformBinaryPath(t) }

	testCases := map[string]resource.TestCase{
		"import success": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
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
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
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
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
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
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
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
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + fmt.Sprintf(`
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
						kubernetes_version = "%s"
						constellation_microservice_version = "%s"
					  }
				`, versions.Default, providerVersion),
					ExpectError: regexp.MustCompile(".*Master secret must be a hex-encoded 32-byte.*"),
				},
			},
		},
		"master secret salt not hex": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + fmt.Sprintf(`
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
						kubernetes_version = "%s"
						constellation_microservice_version = "%s"
					  }
				`, versions.Default, providerVersion),
					ExpectError: regexp.MustCompile(".*Master secret salt must be a hex-encoded 32-byte.*"),
				},
			},
		},
		"measurement salt not hex": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + fmt.Sprintf(`
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
						kubernetes_version = "%s"
						constellation_microservice_version = "%s"
					  }
				`, versions.Default, providerVersion),
					ExpectError: regexp.MustCompile(".*Measurement salt must be a hex-encoded 32-byte.*"),
				},
			},
		},
		"invalid node ip cidr": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + fmt.Sprintf(`
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
						kubernetes_version = "%s"
						constellation_microservice_version = "%s"
					  }
				`, versions.Default, providerVersion),
					ExpectError: regexp.MustCompile(".*Node IP CIDR must be a valid CIDR.*"),
				},
			},
		},
		"invalid service ip cidr": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "aws") + fmt.Sprintf(`
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
						kubernetes_version = "%s"
						constellation_microservice_version = "%s"
					  }
				`, versions.Default, providerVersion),
					ExpectError: regexp.MustCompile(".*Service IP CIDR must be a valid CIDR.*"),
				},
			},
		},
		"azure config missing": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "azure") + fmt.Sprintf(`
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
						kubernetes_version = "%s"
						constellation_microservice_version = "%s"
					  }
				`, versions.Default, providerVersion),
					ExpectError: regexp.MustCompile(".*When csp is set to 'azure', the 'azure' configuration must be set.*"),
				},
			},
		},
		"gcp config missing": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "gcp") + fmt.Sprintf(`
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
						kubernetes_version = "%s"
						constellation_microservice_version = "%s"
					  }
				`, versions.Default, providerVersion),
					ExpectError: regexp.MustCompile(".*When csp is set to 'gcp', the 'gcp' configuration must be set.*"),
				},
			},
		},
		"gcp pod ip cidr missing": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "gcp") + fmt.Sprintf(`
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
							kubernetes_version = "%s"
							constellation_microservice_version = "%s"
					  }
				`, versions.Default, providerVersion),
					ExpectError: regexp.MustCompile(`.*When csp is set to 'gcp', 'ip_cidr_pod' must be set.*`),
				},
			},
		},
		"gcp pod ip cidr not a valid cidr": {
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithVersion(providerVersion),
			PreCheck:                 bazelPreCheck,
			Steps: []resource.TestStep{
				{
					Config: fullClusterTestingConfig(t, "gcp") + fmt.Sprintf(`
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
									ip_cidr_pod     = "0.0.0.0/xxxx"
							}
							gcp = {
									project_id = "test"
									service_account_key = "eyJ0ZXN0IjogInRlc3QifQ=="
							}
							kubernetes_version = "%s"
							constellation_microservice_version = "%s"
					  }
				`, versions.Default, providerVersion),
					ExpectError: regexp.MustCompile(`.*Pod IP CIDR must be a valid CIDR range.*`),
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

	image := providerVersion
	switch csp {
	case "aws":
		return providerConfig + fmt.Sprintf(`
		data "constellation_image" "bar" {
			version             = "%s"
			attestation_variant = "aws-sev-snp"
			csp                 = "aws"
			region			    = "us-east-2"
		}

		data "constellation_attestation" "foo" {
			csp                 = "aws"
			attestation_variant = "aws-sev-snp"
			image               = data.constellation_image.bar.image
		}`, image)
	case "azure":
		return providerConfig + fmt.Sprintf(`
		data "constellation_image" "bar" {
			version             = "%s"
			attestation_variant = "azure-sev-snp"
			csp                 = "azure"
		}

		data "constellation_attestation" "foo" {
			csp                 = "azure"
			attestation_variant = "azure-sev-snp"
			image               = data.constellation_image.bar.image
		}`, image)
	case "gcp":
		return providerConfig + fmt.Sprintf(`
		data "constellation_image" "bar" {
			version             = "%s"
			attestation_variant = "gcp-sev-es"
			csp                 = "gcp"
		}

		data "constellation_attestation" "foo" {
			csp                 = "gcp"
			attestation_variant = "gcp-sev-es"
			image               = data.constellation_image.bar.image
		}`, image)
	default:
		t.Fatal("unknown csp")
		return ""
	}
}
