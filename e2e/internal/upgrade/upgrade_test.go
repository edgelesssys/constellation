// TODO: //go:build e2eupgrade

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/edgelesssys/constellation/v2/e2e/internal/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi/fetcher"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	toImage     = flag.String("to-image", "", "Image (shortversion) to upgrade to.")
	toVersion   = flag.String("to-version", "", "CLI version to use.")
	wantWorker  = flag.Int("want-worker", 0, "Number of wanted worker nodes.")
	wantControl = flag.Int("want-control", 0, "Number of wanted control nodes.")
)

// E2E is creating cluster from root of Constellation repository, we need to
// execute the upgrade test in that Constellation workspace.
func TestMain(m *testing.M) {
	if err := os.Chdir("../../.."); err != nil {
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestUpgrade(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	k, err := kubectl.New()
	require.NoError(err)
	assert.NotNil(k)

	testNodesEventuallyAvailable(t, k, *wantControl, *wantWorker)
	testPodsEventuallyReady(t, k, "kube-system")

	testCLIHasVersion(t, *toVersion)
	wantImage := writeUpgradeConfig(t, *toImage)
	cmd := exec.CommandContext(context.Background(), "constellation", "upgrade", "execute")
	msg, err := cmd.CombinedOutput()
	require.NoErrorf(err, "%s", string(msg))

	testNodesEventuallyHaveImage(t, k, wantImage)
}

func writeUpgradeConfig(t *testing.T, toImage string) string {
	fileHandler := file.NewHandler(afero.NewOsFs())
	cfg, err := config.New(fileHandler, constants.ConfigFilename)
	require.NoError(t, err)

	info, err := fetchUpgradeInfo(cfg.GetProvider(), toImage)
	require.NoError(t, err)

	cfg.Upgrade.Image = info.shortPath
	cfg.Upgrade.Measurements = make(measurements.M)
	for key, value := range info.measurements {
		cfg.Upgrade.Measurements[key] = measurements.Measurement{
			Expected: mustDecodeHex(value),
			WarnOnly: false,
		}
	}
	err = fileHandler.WriteYAML(constants.ConfigFilename, cfg, file.OptOverwrite)
	require.NoError(t, err)

	return info.wantImage
}

func testNodesEventuallyHaveImage(t *testing.T, k *kubernetes.Clientset, wantImage string) {
	require.Eventually(t, func() bool {
		nodes, err := k.CoreV1().Nodes().List(context.Background(), metaV1.ListOptions{})
		if err != nil {
			fmt.Println(err)
			return false
		}

		allUpdated := true
		t.Logf("Node status (%v):", time.Now())
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
	cmd := exec.CommandContext(context.Background(), "constellation", "version")
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

type imageMeasurements struct {
	Cmdline       string `json:"cmdline"`
	CmdlineSHA256 string `json:"cmdline-sha256"`
	EFIstages     []struct {
		Name   string `json:"name"`
		SHA256 string `json:"sha256"`
	} `json:"efistages"`
	InitrdSHA256 string                       `json:"initrd-sha256"`
	Measurements map[string]map[string]string `json:"measurements"`
}

type upgradeInfo struct {
	measurements map[uint32]string
	shortPath    string
	wantImage    string
}

func fetchUpgradeInfo(csp cloudprovider.Provider, toImage string) (upgradeInfo, error) {
	info := upgradeInfo{
		measurements: make(map[uint32]string),
		shortPath:    toImage,
	}
	versionsClient := fetcher.NewFetcher()

	ver, err := versionsapi.NewVersionFromShortPath(toImage, versionsapi.VersionKindImage)
	if err != nil {
		return upgradeInfo{}, err
	}

	artifactURL := ver.ArtifactURL()
	measurementsURL, err := url.JoinPath(artifactURL, "image/csp", strings.ToLower(csp.String()), "measurements.image.json")
	if err != nil {
		return upgradeInfo{}, err
	}

	imageMeasurements, err := fetchMeasurements(measurementsURL)
	if err != nil {
		return upgradeInfo{}, err
	}

	for key, value := range imageMeasurements.Measurements {
		idx, err := strconv.Atoi(key)
		if err != nil {
			return upgradeInfo{}, err
		}
		info.measurements[uint32(idx)] = value["expected"]
	}

	wantImage, err := fetchWantImage(versionsClient, csp, versionsapi.ImageInfo{
		Ref:     ver.Ref,
		Stream:  ver.Stream,
		Version: ver.Version,
	})
	if err != nil {
		return upgradeInfo{}, err
	}
	info.wantImage = wantImage

	return info, nil
}

func fetchMeasurements(url string) (imageMeasurements, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return imageMeasurements{}, err
	}
	fmt.Printf("Fetching measurements from: %v\n", req.URL)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return imageMeasurements{}, err
	}
	defer resp.Body.Close()

	var imageMeasurements imageMeasurements
	err = json.NewDecoder(resp.Body).Decode(&imageMeasurements)
	return imageMeasurements, err
}

func fetchWantImage(client *fetcher.Fetcher, csp cloudprovider.Provider, imageInfo versionsapi.ImageInfo) (string, error) {
	imageInfo, err := client.FetchImageInfo(context.Background(), imageInfo)
	if err != nil {
		return "", err
	}

	switch csp {
	case cloudprovider.GCP:
		return imageInfo.GCP["sev-es"], nil
	case cloudprovider.Azure:
		return imageInfo.Azure["cvm"], nil
	case cloudprovider.AWS:
		return imageInfo.AWS["eu-central-1"], nil
	default:
		return "", errors.New("finding wanted image")
	}
}

func mustDecodeHex(s string) [32]byte {
	val, _ := hex.DecodeString(s)
	return *(*[32]byte)(val)
}
