//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package upgrade

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/edgelesssys/constellation/v2/e2e/internal/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Flags are defined globally as `go test` implicitly calls flag.Parse() before executing a testcase.
// Thus defining and parsing flags inside a testcase would result in a panic.
// See https://groups.google.com/g/golang-nuts/c/P6EdEdgvDuc/m/5-Dg6bPxmvQJ.
var (
	targetImage         = flag.String("target-image", "", "Image (shortversion) to upgrade to.")
	targetKubernetes    = flag.String("target-kubernetes", "", "Kubernetes version (MAJOR.MINOR.PATCH) to upgrade to. Defaults to default version of target CLI.")
	targetMicroservices = flag.String("target-microservices", "", "Microservice version (MAJOR.MINOR.PATCH) to upgrade to. Defaults to default version of target CLI.")
	// When executing the test as a bazel target the workspace path is supplied through an env variable that bazel sets.
	workspace = flag.String("workspace", "", "Constellation workspace in which to run the tests.")
	// When executing the test as a bazel target the CLI path is supplied through an env variable that bazel sets.
	// When executing via `go test` extra care should be taken that the supplied CLI is built on the same commit as this test.
	cliPath     = flag.String("cli", "", "Constellation CLI to run the tests.")
	wantWorker  = flag.Int("want-worker", 0, "Number of wanted worker nodes.")
	wantControl = flag.Int("want-control", 0, "Number of wanted control nodes.")
	timeout     = flag.Duration("timeout", 3*time.Hour, "Timeout after which the cluster should have converged to the target version.")
)

// TestUpgrade checks that the workspace's kubeconfig points to a healthy cluster,
// we can write an upgrade config, we can trigger an upgrade
// and the cluster eventually upgrades to the target version.
func TestUpgrade(t *testing.T) {
	require := require.New(t)

	err := setup()
	require.NoError(err)

	k, err := kubectl.New()
	require.NoError(err)
	require.NotNil(k)

	require.NotEqual(*targetImage, "", "--target-image needs to be specified")

	log.Println("Waiting for nodes and pods to be ready.")
	testNodesEventuallyAvailable(t, k, *wantControl, *wantWorker)
	testPodsEventuallyReady(t, k, "kube-system")

	cli, err := getCLIPath(*cliPath)
	require.NoError(err)

	targetVersions := writeUpgradeConfig(require, *targetImage, *targetKubernetes, *targetMicroservices)

	log.Println("Fetching measurements for new image.")
	cmd := exec.CommandContext(context.Background(), cli, "config", "fetch-measurements", "--insecure", "--debug")
	stdout, stderr, err := runCommandWithSeparateOutputs(cmd)
	require.NoError(err, "Stdout: %s\nStderr: %s", string(stdout), string(stderr))
	log.Println(string(stdout))

	data, err := os.ReadFile("./constellation-conf.yaml")
	require.NoError(err)
	log.Println(string(data))

	log.Println("Checking upgrade.")
	runUpgradeCheck(require, cli, *targetKubernetes)

	log.Println("Triggering upgrade.")
	runUpgradeApply(require, cli)

	AssertUpgradeSuccessful(t, cli, targetVersions, k, *wantControl, *wantWorker, *timeout)
}

// setup checks that the prerequisites for the test are met:
// - a workspace is set
// - a CLI path is set
// - the constellation-upgrade folder does not exist.
func setup() error {
	workingDir, err := workingDir(*workspace)
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	if err := os.Chdir(workingDir); err != nil {
		return fmt.Errorf("changing working directory: %w", err)
	}

	if _, err := getCLIPath(*cliPath); err != nil {
		return fmt.Errorf("getting CLI path: %w", err)
	}
	return nil
}

// workingDir returns the path to the workspace.
func workingDir(workspace string) (string, error) {
	workingDir := os.Getenv("BUILD_WORKING_DIRECTORY")
	switch {
	case workingDir != "":
		return workingDir, nil
	case workspace != "":
		return workspace, nil
	default:
		return "", errors.New("neither 'BUILD_WORKING_DIRECTORY' nor 'workspace' flag set")
	}
}

// getCLIPath returns the path to the CLI.
func getCLIPath(cliPathFlag string) (string, error) {
	pathCLI := os.Getenv("PATH_CLI")
	var relCLIPath string
	switch {
	case pathCLI != "":
		relCLIPath = pathCLI
	case cliPathFlag != "":
		relCLIPath = cliPathFlag
	default:
		return "", errors.New("neither 'PATH_CLI' nor 'cli' flag set")
	}

	// try to find the CLI in the working directory
	// (e.g. when running via `go test` or when specifying a path manually)
	workdir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	absCLIPath := relCLIPath
	if !filepath.IsAbs(relCLIPath) {
		absCLIPath = filepath.Join(workdir, relCLIPath)
	}
	if _, err := os.Stat(absCLIPath); err == nil {
		return absCLIPath, nil
	}

	// fall back to runfiles (e.g. when running via bazel)
	return runfiles.Rlocation(pathCLI)
}

// testPodsEventuallyReady checks that:
// 1) all pods are running.
// 2) all pods have good status conditions.
func testPodsEventuallyReady(t *testing.T, k *kubernetes.Clientset, namespace string) {
	require.Eventually(t, func() bool {
		pods, err := k.CoreV1().Pods(namespace).List(context.Background(), metaV1.ListOptions{})
		if err != nil {
			log.Println(err)
			return false
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != coreV1.PodRunning {
				log.Printf("Pod %s is not running, but %s\n", pod.Name, pod.Status.Phase)
				return false
			}

			for _, condition := range pod.Status.Conditions {
				switch condition.Type {
				case coreV1.ContainersReady, coreV1.PodInitialized, coreV1.PodReady, coreV1.PodScheduled:
					if condition.Status != coreV1.ConditionTrue {
						log.Printf("Pod %s's status %s is false\n", pod.Name, coreV1.ContainersReady)
						return false
					}
				}
			}
		}
		return true
	}, time.Minute*30, time.Minute)
}

// testNodesEventuallyAvailable checks that:
// 1) all nodes only have good status conditions.
// 2) the expected number of nodes have joined the cluster.
func testNodesEventuallyAvailable(t *testing.T, k *kubernetes.Clientset, wantControlNodeCount, wantWorkerNodeCount int) {
	require.Eventually(t, func() bool {
		nodes, err := k.CoreV1().Nodes().List(context.Background(), metaV1.ListOptions{})
		if err != nil {
			log.Println(err)
			return false
		}

		var controlNodeCount, workerNodeCount int
		for _, node := range nodes.Items {
			if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
				controlNodeCount++
			} else {
				workerNodeCount++
			}

			for _, condition := range node.Status.Conditions {
				switch condition.Type {
				case coreV1.NodeReady:
					if condition.Status != coreV1.ConditionTrue {
						fmt.Printf("Status %s for node %s is %s\n", condition.Type, node.Name, condition.Status)
						return false
					}
				case coreV1.NodeMemoryPressure, coreV1.NodeDiskPressure, coreV1.NodePIDPressure, coreV1.NodeNetworkUnavailable:
					if condition.Status != coreV1.ConditionFalse {
						fmt.Printf("Status %s for node %s is %s\n", condition.Type, node.Name, condition.Status)
						return false
					}
				}
			}
		}

		if controlNodeCount != wantControlNodeCount {
			log.Printf("Want %d control nodes but got %d\n", wantControlNodeCount, controlNodeCount)
			return false
		}
		if workerNodeCount != wantWorkerNodeCount {
			log.Printf("Want %d worker nodes but got %d\n", wantWorkerNodeCount, workerNodeCount)
			return false
		}

		return true
	}, time.Minute*30, time.Minute)
}

func writeUpgradeConfig(require *require.Assertions, image string, kubernetes string, microservices string) VersionContainer {
	fileHandler := file.NewHandler(afero.NewOsFs())
	attestationFetcher := attestationconfigapi.NewFetcher()
	cfg, err := config.New(fileHandler, constants.ConfigFilename, attestationFetcher, true)
	var cfgErr *config.ValidationError
	var longMsg string
	if errors.As(err, &cfgErr) {
		longMsg = cfgErr.LongMessage()
	}
	require.NoError(err, longMsg)

	imageFetcher := imagefetcher.New()
	imageRef, err := imageFetcher.FetchReference(
		context.Background(),
		cfg.GetProvider(),
		cfg.GetAttestationConfig().GetVariant(),
		image,
		cfg.GetRegion(), cfg.UseMarketplaceImage(),
	)
	require.NoError(err)

	log.Printf("Setting image version: %s\n", image)
	cfg.Image = image

	defaultConfig := config.Default()
	var kubernetesVersion versions.ValidK8sVersion
	if kubernetes == "" {
		kubernetesVersion = defaultConfig.KubernetesVersion
	} else {
		kubernetesVersion = versions.ValidK8sVersion(kubernetes) // ignore validation because the config is only written to file
	}

	var microserviceVersion semver.Semver
	if microservices == "" {
		microserviceVersion = defaultConfig.MicroserviceVersion
	} else {
		version, err := semver.New(microservices)
		require.NoError(err)
		microserviceVersion = version
	}

	log.Printf("Setting K8s version: %s\n", kubernetesVersion)
	cfg.KubernetesVersion = kubernetesVersion
	log.Printf("Setting microservice version: %s\n", microserviceVersion)
	cfg.MicroserviceVersion = microserviceVersion

	err = fileHandler.WriteYAML(constants.ConfigFilename, cfg, file.OptOverwrite)
	require.NoError(err)

	return VersionContainer{ImageRef: imageRef, Kubernetes: kubernetesVersion, Microservices: microserviceVersion}
}

// runUpgradeCheck executes 'upgrade check' and does basic checks on the output.
// We can not check images upgrades because we might use unpublished images. CLI uses public CDN to check for available images.
func runUpgradeCheck(require *require.Assertions, cli, targetKubernetes string) {
	cmd := exec.CommandContext(context.Background(), cli, "upgrade", "check", "--debug")
	stdout, stderr, err := runCommandWithSeparateOutputs(cmd)
	require.NoError(err, "Stdout: %s\nStderr: %s", string(stdout), string(stderr))

	require.Contains(string(stdout), "The following updates are available with this CLI:")
	require.Contains(string(stdout), "Kubernetes:")
	log.Printf("targetKubernetes: %s\n", targetKubernetes)

	if targetKubernetes == "" {
		log.Printf("true\n")
		require.True(containsAny(string(stdout), versions.SupportedK8sVersions()))
	} else {
		log.Printf("false. targetKubernetes: %s\n", targetKubernetes)
		require.Contains(string(stdout), targetKubernetes, fmt.Sprintf("Expected Kubernetes version %s in output.", targetKubernetes))
	}

	require.Contains(string(stdout), "Services:")
	require.Contains(string(stdout), fmt.Sprintf("--> %s", constants.BinaryVersion().String()))

	log.Println(string(stdout))
}

func containsAny(text string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(text, substr) {
			return true
		}
	}
	return false
}

func runUpgradeApply(require *require.Assertions, cli string) {
	tfLogFlag := ""
	cmd := exec.CommandContext(context.Background(), cli, "--help")
	stdout, stderr, err := runCommandWithSeparateOutputs(cmd)
	require.NoError(err, "Stdout: %s\nStderr: %s", string(stdout), string(stderr))
	if strings.Contains(string(stdout), "--tf-log") {
		tfLogFlag = "--tf-log=DEBUG"
	}

	cmd = exec.CommandContext(context.Background(), cli, "apply", "--debug", "--yes", tfLogFlag)
	stdout, stderr, err = runCommandWithSeparateOutputs(cmd)
	require.NoError(err, "Stdout: %s\nStderr: %s", string(stdout), string(stderr))
	require.NoError(containsUnexepectedMsg(string(stdout)))
	log.Println(string(stdout))
	log.Println(string(stderr)) // also print debug logs.
}

// containsUnexepectedMsg checks if the given input contains any unexpected messages.
// unexepcted messages are:
// "Skipping image & Kubernetes upgrades. Another upgrade is in progress".
func containsUnexepectedMsg(input string) error {
	if strings.Contains(input, "Skipping image & Kubernetes upgrades. Another upgrade is in progress") {
		return errors.New("unexpected upgrade in progress")
	}
	return nil
}
