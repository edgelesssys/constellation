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

func TestAccClusteResourceImports(t *testing.T) {
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
