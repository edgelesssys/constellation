// TODO //go:build e2eupgrade

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package test

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/e2e/internal/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	constellationCLIVersion = "2.3.0"
	wantWorkerNodeCount     = 2
	wantControlNodeCount    = 3
	wantNodeCount           = wantControlNodeCount + wantWorkerNodeCount
	upgradeToImage          = "ref/main/stream/nightly/v2.4.0-pre.0.20221219160053-43123e36f931"
	// TODO: Get this value by fetching LUT of upgradeToImage, then read value for current CSP
	wantGCPImage = "projects/constellation-images/global/images/constellation-main-nightly-20221219170053"
)

var (
	cliDownloadURL        = fmt.Sprintf("https://github.com/edgelesssys/constellation/releases/download/v%s/constellation-linux-amd64", constellationCLIVersion)
	upgradeToMeasurements = measurements.M{
		4: {
			Expected: mustDecodeHex("3297593d02932ce71cd33ddff07f092ba7431b6f74fe81db2ddc45ecef446bea"),
			WarnOnly: false,
		},
		9: {
			Expected: mustDecodeHex("574edd9a13bc28463bec03c86d414f3d2cbbe6da00966428ed1fb640efcbf585"),
			WarnOnly: false,
		},
		12: {
			Expected: mustDecodeHex("39e4b328e30c7a990ed1057d7e3d0f6a046ea6817ecd79c1f55197030b54086d"),
			WarnOnly: false,
		},
	}
)

// E2E is creating cluster from root of Constellation repository, we need to
// execute the upgrade test in that context.
func TestMain(m *testing.M) {
	if err := os.Chdir("../../.."); err != nil {
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func mustDecodeHex(s string) [32]byte {
	val, _ := hex.DecodeString(s)
	return *(*[32]byte)(val)
}

func TestUpgrade(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	k, err := kubectl.New()
	require.NoError(err)
	assert.NotNil(k)

	testNodesEventuallyAvailable(t, k, wantControlNodeCount, wantWorkerNodeCount)
	testPodsEventuallyReady(t, k, "kube-system")

	downloadFile(t, cliDownloadURL, "constellation")
	testCLIHasVersion(t, constellationCLIVersion)

	// TODO: replace with ./constellation upgrade plan --image=...
	fileHandler := file.NewHandler(afero.NewOsFs())
	cfg, err := config.New(fileHandler, constants.ConfigFilename)
	require.NoError(err)
	cfg.Upgrade.Image = upgradeToImage
	cfg.Upgrade.Measurements = upgradeToMeasurements
	err = fileHandler.WriteYAML(constants.ConfigFilename, cfg, file.OptOverwrite)
	require.NoError(err)

	cmd := exec.CommandContext(context.Background(), "./constellation", "upgrade", "execute")
	_, err = cmd.CombinedOutput()
	require.NoError(err)

	testNodesEventuallyHaveImage(t, k, wantGCPImage)
}

func downloadFile(t *testing.T, url, outFile string) {
	t.Helper()

	require := require.New(t)

	out, err := os.Create(outFile)
	require.NoError(err)
	defer func() {
		err := out.Close()
		require.NoError(err)
	}()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	require.NoError(err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(err)
	defer func() {
		err := resp.Body.Close()
		require.NoError(err)
	}()
	require.Equal(http.StatusOK, resp.StatusCode)

	_, err = io.Copy(out, resp.Body)
	require.NoError(err)

	err = out.Chmod(0o700)
	require.NoError(err)
}

func testNodesEventuallyHaveImage(t *testing.T, k *kubernetes.Clientset, wantImage string) {
	require.Eventually(t, func() bool {
		nodes, err := k.CoreV1().Nodes().List(context.Background(), metaV1.ListOptions{})
		if err != nil {
			fmt.Println(err)
			return false
		}

		allUpdated := true
		fmt.Printf("Node status (%v):\n", time.Now())
		for _, node := range nodes.Items {
			for key, value := range node.Annotations {
				if key == "constellation.edgeless.systems/node-image" {
					fmt.Printf("\t%s: %s\n", node.Name, value)
					if value != wantImage {
						allUpdated = false
					}
				}
			}
		}

		return allUpdated
	}, time.Hour*3, time.Minute)
}

// testCLIHasVersion checks that `constellation version` states the expected version.
func testCLIHasVersion(t *testing.T, wantVersion string) {
	require := require.New(t)
	cmd := exec.CommandContext(context.Background(), "./constellation", "version")
	output, err := cmd.CombinedOutput()
	require.NoError(err)
	require.Contains(string(output), wantVersion)
}

// testPodsEventuallyReady checks that:
// 1) all pods are running.
// 2) all pods have good status conditions.
func testPodsEventuallyReady(t *testing.T, k *kubernetes.Clientset, namespace string) {
	require.Eventually(t, func() bool {
		pods, err := k.CoreV1().Pods(namespace).List(context.Background(), metaV1.ListOptions{})
		if err != nil {
			fmt.Println(err)
			return false
		}

		for _, pod := range pods.Items {
			if pod.Status.Phase != coreV1.PodRunning {
				fmt.Printf("Pod %s is not running, but %s\n", pod.Name, pod.Status.Phase)
				return false
			}

			for _, condition := range pod.Status.Conditions {
				switch condition.Type {
				case coreV1.ContainersReady:
					if condition.Status != coreV1.ConditionTrue {
						fmt.Printf("Pod %s's status %s is false\n", pod.Name, coreV1.ContainersReady)
						return false
					}
				case coreV1.PodInitialized:
					if condition.Status != coreV1.ConditionTrue {
						fmt.Printf("Pod %s's status %s is false\n", pod.Name, coreV1.PodInitialized)
						return false
					}
				case coreV1.PodReady:
					if condition.Status != coreV1.ConditionTrue {
						fmt.Printf("Pod %s's status %s is false\n", pod.Name, coreV1.PodReady)
						return false
					}
				case coreV1.PodScheduled:
					if condition.Status != coreV1.ConditionTrue {
						fmt.Printf("Pod %s's status %s is false\n", pod.Name, coreV1.PodScheduled)
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
			fmt.Println(err)
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
						fmt.Printf("Status %s for node %s is %s\n", coreV1.NodeReady, node.Name, condition.Status)
						return false
					}
				case coreV1.NodeMemoryPressure:
					if condition.Status != coreV1.ConditionFalse {
						fmt.Printf("Status %s for node %s is %s\n", coreV1.NodeMemoryPressure, node.Name, condition.Status)
						return false
					}
				case coreV1.NodeDiskPressure:
					if condition.Status != coreV1.ConditionFalse {
						fmt.Printf("Status %s for node %s is %s\n", coreV1.NodeDiskPressure, node.Name, condition.Status)
						return false
					}
				case coreV1.NodePIDPressure:
					if condition.Status != coreV1.ConditionFalse {
						fmt.Printf("Status %s for node %s is %s\n", coreV1.NodePIDPressure, node.Name, condition.Status)
						return false
					}
				case coreV1.NodeNetworkUnavailable:
					if condition.Status != coreV1.ConditionFalse {
						fmt.Printf("Status %s for node %s is %s\n", coreV1.NodeNetworkUnavailable, node.Name, condition.Status)
						return false
					}
				}
			}
		}

		if controlNodeCount != wantControlNodeCount {
			fmt.Printf("Want %d control nodes but got %d\n", wantControlNodeCount, controlNodeCount)
			return false
		}
		if workerNodeCount != wantWorkerNodeCount {
			fmt.Printf("Want %d worker nodes but got %d\n", wantWorkerNodeCount, workerNodeCount)
			return false
		}

		return true
	}, time.Minute*30, time.Minute)
}
