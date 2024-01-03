//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// End-to-end test that is used by the e2e Terraform provider test.
package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/e2e/internal/kubectl"
	"github.com/edgelesssys/constellation/v2/e2e/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/stretchr/testify/require"
)

var (
	targetImage         = flag.String("target-image", "", "Image (shortversion) to upgrade to.")
	targetKubernetes    = flag.String("target-kubernetes", "", "Kubernetes version (MAJOR.MINOR.PATCH) to upgrade to. Defaults to default version of target CLI.")
	targetMicroservices = flag.String("target-microservices", "", "Microservice version (MAJOR.MINOR.PATCH) to upgrade to. Defaults to default version of target CLI.")
	// When executing the test as a bazel target the CLI path is supplied through an env variable that bazel sets.
	// When executing via `go test` extra care should be taken that the supplied CLI is built on the same commit as this test.
	// When executing the test as a bazel target the workspace path is supplied through an env variable that bazel sets.
	workspace   = flag.String("workspace", "", "Constellation workspace in which to run the tests.")
	cliPath     = flag.String("cli", "", "Constellation CLI to run the tests.")
	wantWorker  = flag.Int("want-worker", 0, "Number of wanted worker nodes.")
	wantControl = flag.Int("want-control", 0, "Number of wanted control nodes.")
	timeout     = flag.Duration("timeout", 90*time.Minute, "Timeout after which the cluster should have converged to the target version.")
)

// TestUpgradeSuccessful tests that the upgrade to the target version is successful.
func TestUpgradeSuccessful(t *testing.T) {
	require := require.New(t)
	kubeconfigPath := os.Getenv("KUBECONFIG")
	require.NotEmpty(kubeconfigPath, "KUBECONFIG environment variable must be set")
	dir := filepath.Dir(kubeconfigPath)
	configPath := filepath.Join(dir, constants.ConfigFilename)

	// only done here to construct the version struct
	require.NotEqual(*targetImage, "", "--target-image needs to be specified")
	v := upgrade.WriteUpgradeConfig(require, *targetImage, *targetKubernetes, *targetMicroservices, configPath)
	// ignore Kubernetes check if targetKubernetes is not set; Kubernetes is only explicitly upgraded
	if *targetKubernetes == "" {
		v.Kubernetes = ""
	}
	k, err := kubectl.New()
	require.NoError(err)

	err = upgrade.Setup(*workspace, *cliPath)
	require.NoError(err)
	upgrade.AssertUpgradeSuccessful(t, *cliPath, v, k, *wantControl, *wantWorker, *timeout)
}
