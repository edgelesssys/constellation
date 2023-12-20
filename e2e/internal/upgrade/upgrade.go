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
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type VersionContainer struct {
	ImageRef      string
	Kubernetes    versions.ValidK8sVersion
	Microservices semver.Semver
}

func AssertUpgradeSuccessful(t *testing.T, cli string, targetVersions VersionContainer, k *kubernetes.Clientset, wantControl, wantWorker int, timeout time.Duration) {
	wg := queryStatusAsync(t, cli)
	require.NotNil(t, k)
	testMicroservicesEventuallyHaveVersion(t, targetVersions.Microservices, timeout)
	testNodesEventuallyHaveVersion(t, k, targetVersions, wantControl+wantWorker, timeout)

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
	}, timeout, time.Minute)
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
	}, timeout, time.Minute)
}

func testNodesEventuallyHaveVersion(t *testing.T, k *kubernetes.Clientset, targetVersions VersionContainer, totalNodeCount int, timeout time.Duration) {
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
					if !strings.EqualFold(value, targetVersions.ImageRef) {
						log.Printf("\t%s: Image %s, want %s\n", node.Name, value, targetVersions.ImageRef)
						allUpdated = false
					}
				}
			}

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

		return allUpdated
	}, timeout, time.Minute)
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
