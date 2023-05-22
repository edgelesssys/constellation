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
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/e2e/internal/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/semver"
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

// setup checks that the prerequisites for the test are met:
// - a workspace is set
// - a CLI path is set
// - the upgrade folder does not exist.
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
	if _, err := os.Stat(constants.UpgradeDir); err == nil {
		return fmt.Errorf("please remove the existing %s folder", constants.UpgradeDir)
	}

	return nil
}

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

	testNodesEventuallyAvailable(t, k, *wantControl, *wantWorker)
	testPodsEventuallyReady(t, k, "kube-system")

	cli, err := getCLIPath(*cliPath)
	require.NoError(err)

	// Migrate config if necessary.
	cmd := exec.CommandContext(context.Background(), cli, "config", "migrate", "--config", constants.ConfigFilename, "--force", "--debug")
	msg, err := cmd.CombinedOutput()
	require.NoErrorf(err, "%s", string(msg))
	log.Println(string(msg))

	targetVersions := writeUpgradeConfig(require, *targetImage, *targetKubernetes, *targetMicroservices)

	data, err := os.ReadFile("./constellation-conf.yaml")
	require.NoError(err)
	log.Println(string(data))

	log.Println("Triggering upgrade.")

	tfLogFlag := ""
	cmd = exec.CommandContext(context.Background(), cli, "--help")
	msg, err = cmd.CombinedOutput()
	require.NoErrorf(err, "%s", string(msg))
	if strings.Contains(string(msg), "--tf-log") {
		tfLogFlag = "--tf-log=DEBUG"
	}

	cmd = exec.CommandContext(context.Background(), cli, "upgrade", "apply", "--force", "--debug", "--yes", tfLogFlag)
	msg, err = cmd.CombinedOutput()
	require.NoErrorf(err, "%s", string(msg))
	require.NoError(containsUnexepectedMsg(string(msg)))
	log.Println(string(msg))

	// Show versions set in cluster.
	// The string after "Cluster status:" in the output might not be updated yet.
	// This is only updated after the operator finishes one reconcile loop.
	cmd = exec.CommandContext(context.Background(), cli, "status")
	msg, err = cmd.CombinedOutput()
	require.NoError(err, string(msg))
	log.Println(string(msg))

	testMicroservicesEventuallyHaveVersion(t, targetVersions.microservices, *timeout)
	testNodesEventuallyHaveVersion(t, k, targetVersions, *wantControl+*wantWorker, *timeout)
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
func getCLIPath(cliPath string) (string, error) {
	pathCLI := os.Getenv("PATH_CLI")
	switch {
	case pathCLI != "":
		return pathCLI, nil
	case cliPath != "":
		return cliPath, nil
	default:
		return "", errors.New("neither 'PATH_CLI' nor 'cli' flag set")
	}
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

func writeUpgradeConfig(require *require.Assertions, image string, kubernetes string, microservices string) versionContainer {
	fileHandler := file.NewHandler(afero.NewOsFs())
	cfg, err := config.New(fileHandler, constants.ConfigFilename, true)
	var cfgErr *config.ValidationError
	var longMsg string
	if errors.As(err, &cfgErr) {
		longMsg = cfgErr.LongMessage()
	}
	require.NoError(err, longMsg)

	info, err := fetchUpgradeInfo(context.Background(), cfg.GetProvider(), image)
	require.NoError(err)

	log.Printf("Setting image version: %s\n", info.shortPath)
	cfg.Image = info.shortPath
	cfg.UpdateMeasurements(info.measurements)

	defaultConfig := config.Default()
	var kubernetesVersion semver.Semver
	if kubernetes == "" {
		kubernetesVersion, err = semver.New(defaultConfig.KubernetesVersion)
		require.NoError(err)
	} else {
		kubernetesVersion, err = semver.New(kubernetes)
		require.NoError(err)
	}

	var microserviceVersion string
	if microservices == "" {
		microserviceVersion = defaultConfig.MicroserviceVersion
	} else {
		microserviceVersion = microservices
	}

	log.Printf("Setting K8s version: %s\n", kubernetesVersion.String())
	cfg.KubernetesVersion = kubernetesVersion.String()
	log.Printf("Setting microservice version: %s\n", microserviceVersion)
	cfg.MicroserviceVersion = microserviceVersion

	err = fileHandler.WriteYAML(constants.ConfigFilename, cfg, file.OptOverwrite)
	require.NoError(err)

	return versionContainer{imageRef: info.imageRef, kubernetes: kubernetesVersion, microservices: microserviceVersion}
}

func testMicroservicesEventuallyHaveVersion(t *testing.T, wantMicroserviceVersion string, timeout time.Duration) {
	require.Eventually(t, func() bool {
		version, err := servicesVersion(t)
		if err != nil {
			log.Printf("Unable to fetch microservice version: %v\n", err)
			return false
		}

		if version != wantMicroserviceVersion {
			log.Printf("Microservices still at version %v, want %v\n", version, wantMicroserviceVersion)
			return false
		}

		return true
	}, timeout, time.Minute)
}

func testNodesEventuallyHaveVersion(t *testing.T, k *kubernetes.Clientset, targetVersions versionContainer, totalNodeCount int, timeout time.Duration) {
	require.Eventually(t, func() bool {
		nodes, err := k.CoreV1().Nodes().List(context.Background(), metaV1.ListOptions{})
		if err != nil {
			log.Println(err)
			return false
		}
		require.False(t, len(nodes.Items) < totalNodeCount, "expected at least %v nodes, got %v", totalNodeCount, len(nodes.Items))

		allUpdated := true
		log.Printf("Node status (%v):", time.Now())
		for _, node := range nodes.Items {
			for key, value := range node.Annotations {
				if key == "constellation.edgeless.systems/node-image" {
					if !strings.EqualFold(value, targetVersions.imageRef) {
						log.Printf("\t%s: Image %s, want %s\n", node.Name, value, targetVersions.imageRef)
						allUpdated = false
					}
				}
			}

			kubeletVersion := node.Status.NodeInfo.KubeletVersion
			if kubeletVersion != targetVersions.kubernetes.String() {
				log.Printf("\t%s: K8s (Kubelet) %s, want %s\n", node.Name, kubeletVersion, targetVersions.kubernetes.String())
				allUpdated = false
			}
			kubeProxyVersion := node.Status.NodeInfo.KubeProxyVersion
			if kubeProxyVersion != targetVersions.kubernetes.String() {
				log.Printf("\t%s: K8s (Proxy) %s, want %s\n", node.Name, kubeProxyVersion, targetVersions.kubernetes.String())
				allUpdated = false
			}
		}

		return allUpdated
	}, timeout, time.Minute)
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

type versionContainer struct {
	imageRef      string
	kubernetes    semver.Semver
	microservices string
}
