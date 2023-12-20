/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// End-to-end test that is used by the e2e Terraform provider test.
package main

import (
	"flag"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/e2e/internal/kubectl"
	"github.com/edgelesssys/constellation/v2/e2e/internal/upgrade"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/stretchr/testify/require"
)

var (
	targetImage         = flag.String("target-image", "", "Image (shortversion) to upgrade to.")
	targetKubernetes    = flag.String("target-kubernetes", "", "Kubernetes version (MAJOR.MINOR.PATCH) to upgrade to. Defaults to default version of target CLI.")
	targetMicroservices = flag.String("target-microservices", "", "Microservice version (MAJOR.MINOR.PATCH) to upgrade to. Defaults to default version of target CLI.")
	// When executing the test as a bazel target the CLI path is supplied through an env variable that bazel sets.
	// When executing via `go test` extra care should be taken that the supplied CLI is built on the same commit as this test.
	cliPath     = flag.String("cli", "", "Constellation CLI to run the tests.")
	wantWorker  = flag.Int("want-worker", 0, "Number of wanted worker nodes.")
	wantControl = flag.Int("want-control", 0, "Number of wanted control nodes.")
	timeout     = flag.Duration("timeout", 3*time.Hour, "Timeout after which the cluster should have converged to the target version.")
)

// TestUpgradeSuccessful tests that the upgrade to the target version is successful.
func TestUpgradeSuccessful(t *testing.T) {
	defaultConfig := config.Default()
	var k8sV versions.ValidK8sVersion
	if *targetKubernetes == "" {
		k8sV = defaultConfig.KubernetesVersion
	} else {
		k8sV = versions.ValidK8sVersion(*targetKubernetes)
	}
	var microV semver.Semver
	if *targetMicroservices == "" {
		microV = defaultConfig.MicroserviceVersion
	} else {
		var err error
		microV, err = semver.New(*targetMicroservices)
		require.NoError(t, err)
	}
	v := upgrade.VersionContainer{
		ImageRef:      *targetImage,
		Kubernetes:    k8sV,
		Microservices: microV,
	}
	k, err := kubectl.New()
	require.NoError(t, err)
	upgrade.AssertUpgradeSuccessful(t, *cliPath, v, k, *wantControl, *wantWorker, *timeout)
}
