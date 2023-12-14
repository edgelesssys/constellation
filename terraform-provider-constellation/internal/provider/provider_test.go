/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// testingConfig is a shared configuration to combine with the actual
	// test configuration so the constellation provider is properly configured
	// for acceptance testing.
	testingConfig = `provider "constellation" {}`
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach. It sets a pseudo version for the provider version.
var testAccProtoV6ProviderFactories = testAccProtoV6ProviderFactoriesWithVersion(constants.BinaryVersion().String())

// testAccProtoV6ProviderFactoriesWithVersion are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactoriesWithVersion = func(version string) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"constellation": providerserver.NewProtocol6WithError(New(version)()),
	}
}

// bazelSetTerraformBinaryPath sets the path to the Terraform binary for
// acceptance testing when running under Bazel.
func bazelSetTerraformBinaryPath(t *testing.T) {
	if v := os.Getenv("TF_ACC"); v != "1" {
		t.Fatal("TF_ACC must be set to \"1\" for acceptance tests")
	}

	// If we don't run under Bazel, we need to use the host Terraform binary.
	// Print a warning and return without setting the path.
	if !runsUnderBazel() {
		fmt.Println(
			"Test is not runding under Bazel.\n" +
				"Using Host Terraform binary for acceptance testing.\n" +
				"Tests results may vary from the defaults (which run under Bazel).",
		)
		return
	}

	tfPath := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if tfPath == "" {
		t.Fatal("TF_ACC_TERRAFORM_PATH must be set for acceptance tests")
	}

	absTfPath, err := runfiles.Rlocation(tfPath)
	if err != nil {
		panic("could not find path to artifact")
	}

	// Set the path to the absolute Terraform binary for acceptance testing.
	t.Setenv("TF_ACC_TERRAFORM_PATH", absTfPath)
}

// runsUnderBazel checks whether the test is ran under bazel.
func runsUnderBazel() bool {
	return runsUnder == "bazel"
}

// runsUnder is redefined only by the Bazel build at link time.
var runsUnder = "go"
