//go:build e2e

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package upgrade tests that the CLI's apply command works as expected and
// the operators eventually upgrade all nodes inside the cluster.
// The test is written as a go test because:
// 1. the helm cli does not directly provide the chart version of a release
//
// 2. the image patch needs to be parsed from the image-api's info.json
//
// 3. there is some logic required to setup the test correctly:
//
//   - set microservice, k8s version correctly depending on input
//
//   - set or fetch measurements depending on target image
package upgrade

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/imagefetcher"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// tickDuration is the duration between two checks to see if the upgrade is successful.
var tickDuration = 10 * time.Second // small tick duration to speed up tests

// VersionContainer contains the versions that the cluster should be upgraded to.
type VersionContainer struct {
	ImageRef      string
	Kubernetes    versions.ValidK8sVersion
	Microservices semver.Semver
}

// AssertUpgradeSuccessful tests that the upgrade to the target version is successful.
func AssertUpgradeSuccessful(t *testing.T, cli string, targetVersions VersionContainer, k *kubernetes.Clientset, wantControl, wantWorker int, timeout time.Duration) {
	wg := queryStatusAsync(t, cli)
	require.NotNil(t, k)

	testMicroservicesEventuallyHaveVersion(t, targetVersions.Microservices, timeout)
	log.Println("Microservices are upgraded.")

	testNodesEventuallyHaveVersion(t, k, targetVersions, wantControl+wantWorker, timeout)
	log.Println("Nodes are upgraded.")
	wg.Wait()
}

func queryStatusAsync(t *testing.T, cli string) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// The first control plane node should finish upgrading after 20 minutes. If it does not, something is fishy.
		// Nodes can upgrade in <5mins.
		testStatusEventuallyWorks(t, cli, 20*time.Minute)
	}()

	return &wg
}

func testStatusEventuallyWorks(t *testing.T, cli string, timeout time.Duration) {
	require.Eventually(t, func() bool {
		// Show versions set in cluster.
		// The string after "Cluster status:" in the output might not be updated yet.
		// This is only updated after the operator finishes one reconcile loop.
		cmd := exec.CommandContext(context.Background(), cli, "status")
		stdout, stderr, err := runCommandWithSeparateOutputs(cmd)
		if err != nil {
			log.Printf("Stdout: %s\nStderr: %s", string(stdout), string(stderr))
			return false
		}

		log.Println(string(stdout))
		return true
	}, timeout, tickDuration)
}

func testMicroservicesEventuallyHaveVersion(t *testing.T, wantMicroserviceVersion semver.Semver, timeout time.Duration) {
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
	}, timeout, tickDuration)
}

func testNodesEventuallyHaveVersion(t *testing.T, k *kubernetes.Clientset, targetVersions VersionContainer, totalNodeCount int, timeout time.Duration) {
	require.Eventually(t, func() bool {
		nodes, err := k.CoreV1().Nodes().List(context.Background(), metaV1.ListOptions{})
		if err != nil {
			log.Println(err)
			return false
		}

		// require is not printed in the logs, so we use fmt
		tooSmallNodeCount := len(nodes.Items) < totalNodeCount
		if tooSmallNodeCount {
			log.Printf("expected at least %v nodes, got %v", totalNodeCount, len(nodes.Items))
			return false
		}

		allUpdated := true
		log.Printf("Node status (%v):", time.Now())
		for _, node := range nodes.Items {
			for key, value := range node.Annotations {
				if targetVersions.ImageRef != "" {
					if key == "constellation.edgeless.systems/node-image" {
						if !strings.EqualFold(value, targetVersions.ImageRef) {
							log.Printf("\t%s: Image %s, want %s\n", node.Name, value, targetVersions.ImageRef)
							allUpdated = false
						}
					}
				}
			}
			if targetVersions.Kubernetes != "" {
				kubeletVersion := node.Status.NodeInfo.KubeletVersion
				if kubeletVersion != string(targetVersions.Kubernetes) {
					log.Printf("\t%s: K8s (Kubelet) %s, want %s\n", node.Name, kubeletVersion, targetVersions.Kubernetes)
					allUpdated = false
				}
				kubeProxyVersion := node.Status.NodeInfo.KubeProxyVersion
				if kubeProxyVersion != string(targetVersions.Kubernetes) {
					log.Printf("\t%s: K8s (Proxy) %s, want %s\n", node.Name, kubeProxyVersion, targetVersions.Kubernetes)
					allUpdated = false
				}
			}
		}
		return allUpdated
	}, timeout, tickDuration)
}

// runCommandWithSeparateOutputs runs the given command while separating buffers for
// stdout and stderr.
func runCommandWithSeparateOutputs(cmd *exec.Cmd) (stdout, stderr []byte, err error) {
	stdout = []byte{}
	stderr = []byte{}

	stdoutIn, err := cmd.StdoutPipe()
	if err != nil {
		err = fmt.Errorf("create stdout pipe: %w", err)
		return
	}
	stderrIn, err := cmd.StderrPipe()
	if err != nil {
		err = fmt.Errorf("create stderr pipe: %w", err)
		return
	}

	err = cmd.Start()
	if err != nil {
		err = fmt.Errorf("start command: %w", err)
		return
	}

	continuouslyPrintOutput := func(r io.Reader, prefix string) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			output := scanner.Text()
			fmt.Printf("%s: %s\n", prefix, output)
			switch prefix {
			case "stdout":
				stdout = append(stdout, output...)
			case "stderr":
				stderr = append(stderr, output...)
			}
		}
	}

	go continuouslyPrintOutput(stdoutIn, "stdout")
	go continuouslyPrintOutput(stderrIn, "stderr")

	if err = cmd.Wait(); err != nil {
		err = fmt.Errorf("wait for command to finish: %w", err)
	}

	return stdout, stderr, err
}

// Setup checks that the prerequisites for the test are met:
// - a workspace is set
// - a CLI path is set
// - the constellation-upgrade folder does not exist.
func Setup(workspace, cliPath string) error {
	workingDir, err := workingDir(workspace)
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	if err := os.Chdir(workingDir); err != nil {
		return fmt.Errorf("changing working directory: %w", err)
	}

	if _, err := getCLIPath(cliPath); err != nil {
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

// WriteUpgradeConfig writes the target versions to the config file.
func WriteUpgradeConfig(require *require.Assertions, image string, kubernetes string, microservices string, configPath string) VersionContainer {
	fileHandler := file.NewHandler(afero.NewOsFs())
	attestationFetcher := attestationconfigapi.NewFetcher()
	cfg, err := config.New(fileHandler, configPath, attestationFetcher, true)
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
