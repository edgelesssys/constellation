/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package provider

import (
	"os"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
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
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"constellation": providerserver.NewProtocol6WithError(New("test")()),
}

// bazelSetTerraformBinaryPath sets the path to the Terraform binary for
// acceptance testing when running under Bazel.
func bazelSetTerraformBinaryPath(t *testing.T) {
	if !runsUnderBazel() {
		t.Skip("Skipping test as not running under Bazel.")
	}

	if v := os.Getenv("TF_ACC"); v != "1" {
		t.Fatal("TF_ACC must be set to \"1\" for acceptance tests")
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
	os.Setenv("TF_ACC_TERRAFORM_PATH", absTfPath)
}

// runsUnderBazel checks whether the test is ran under bazel
func runsUnderBazel() bool {
	return runsUnder == "bazel"
}

// runsUnder is redefined only by the Bazel build at link time.
var runsUnder = "go"
