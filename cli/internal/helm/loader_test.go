/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

// TestLoad checks if the serialized format that Load returns correctly preserves the dependencies of the loaded chart.
func TestLoad(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	chartLoader := ChartLoader{}
	config := &config.Config{Provider: config.ProviderConfig{GCP: &config.GCPConfig{}}}
	release, err := chartLoader.Load(config, true, []byte("secret"), []byte("salt"))
	require.NoError(err)

	var helmReleases helm.Releases
	err = json.Unmarshal(release, &helmReleases)
	require.NoError(err)
	reader := bytes.NewReader(helmReleases.ConstellationServices.Chart)
	chart, err := loader.LoadArchive(reader)
	require.NoError(err)
	assert.NotNil(chart.Dependencies())
}

// TestTemplate checks if the rendered constellation-services chart produces the expected yaml files.
func TestTemplate(t *testing.T) {
	testCases := map[string]struct {
		config             *config.Config
		enforceIDKeyDigest bool
		valuesModifier     func(map[string]any) error
		ccmImage           string
		cnmImage           string
	}{
		"GCP": {
			config: &config.Config{Provider: config.ProviderConfig{GCP: &config.GCPConfig{
				DeployCSIDriver:      func() *bool { b := true; return &b }(),
				EnforcedMeasurements: []uint32{1, 11},
			}}},
			enforceIDKeyDigest: false,
			valuesModifier:     prepareGCPValues,
			ccmImage:           "ccmImageForGCP",
		},
		"Azure": {
			config: &config.Config{Provider: config.ProviderConfig{Azure: &config.AzureConfig{
				EnforcedMeasurements: []uint32{1, 11},
				EnforceIDKeyDigest:   func() *bool { b := true; return &b }(),
			}}},
			enforceIDKeyDigest: true,
			valuesModifier:     prepareAzureValues,
			ccmImage:           "ccmImageForAzure",
			cnmImage:           "cnmImageForAzure",
		},
		"QEMU": {
			config: &config.Config{Provider: config.ProviderConfig{QEMU: &config.QEMUConfig{
				EnforcedMeasurements: []uint32{1, 11},
			}}},
			enforceIDKeyDigest: false,
			valuesModifier:     prepareQEMUValues,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			chartLoader := ChartLoader{joinServiceImage: "joinServiceImage", kmsImage: "kmsImage", ccmImage: tc.ccmImage, cnmImage: tc.cnmImage, autoscalerImage: "autoscalerImage"}
			release, err := chartLoader.Load(tc.config, true, []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
			require.NoError(err)

			var helmReleases helm.Releases
			err = json.Unmarshal(release, &helmReleases)
			require.NoError(err)
			reader := bytes.NewReader(helmReleases.ConstellationServices.Chart)
			chart, err := loader.LoadArchive(reader)
			require.NoError(err)

			options := chartutil.ReleaseOptions{
				Name:      "testRelease",
				Namespace: "testNamespace",
				Revision:  1,
				IsInstall: true,
				IsUpgrade: false,
			}
			caps := &chartutil.Capabilities{}

			err = tc.valuesModifier(helmReleases.ConstellationServices.Values)
			require.NoError(err)

			// This step is needed to enabled/disable subcharts according to their tags/conditions.
			err = chartutil.ProcessDependencies(chart, helmReleases.ConstellationServices.Values)
			require.NoError(err)

			valuesToRender, err := chartutil.ToRenderValues(chart, helmReleases.ConstellationServices.Values, options, caps)
			require.NoError(err)

			result, err := engine.Render(chart, valuesToRender)
			require.NoError(err)
			for k, v := range result {
				currentFile := path.Join("testdata", tc.config.GetProvider().String(), k)
				content, err := os.ReadFile(currentFile)

				// If a file does not exist, we expect the render for that path to be empty.
				if errors.Is(err, fs.ErrNotExist) {
					assert.YAMLEq("", v, fmt.Sprintf("current file: %s", currentFile))
					continue
				}
				assert.NoError(err)
				assert.YAMLEq(string(content), v, fmt.Sprintf("current file: %s", currentFile))
			}
		})
	}
}

func prepareGCPValues(values map[string]any) error {
	joinVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'join-service' key")
	}
	joinVals["measurements"] = "{'1':'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA','15':'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA='}"
	joinVals["measurementSalt"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	ccmVals, ok := values["ccm"].(map[string]any)
	if !ok {
		return errors.New("missing 'ccm' key")
	}
	ccmVals["GCP"].(map[string]any)["subnetworkPodCIDR"] = "192.0.2.0/24"
	ccmVals["GCP"].(map[string]any)["projectID"] = "42424242424242"
	ccmVals["GCP"].(map[string]any)["uid"] = "242424242424"
	ccmVals["GCP"].(map[string]any)["secretData"] = "baaaaaad"

	testTag := "v0.0.0"
	pullPolicy := "IfNotPresent"
	csiVals, ok := values["csi-gcp-pd"].(map[string]any)
	if !ok {
		return errors.New("missing 'csi-gcp-pd' key")
	}
	csiVals["image"] = map[string]any{
		"csiProvisioner": map[string]any{
			"repo":       "csi-provisioner",
			"tag":        testTag,
			"pullPolicy": pullPolicy,
		},
		"csiAttacher": map[string]any{
			"repo":       "csi-attacher",
			"tag":        testTag,
			"pullPolicy": pullPolicy,
		},
		"csiResizer": map[string]any{
			"repo":       "csi-resizer",
			"tag":        testTag,
			"pullPolicy": pullPolicy,
		},
		"csiSnapshotter": map[string]any{
			"repo":       "csi-snapshotter",
			"tag":        testTag,
			"pullPolicy": pullPolicy,
		},
		"csiNodeRegistrar": map[string]any{
			"repo":       "csi-registrar",
			"tag":        testTag,
			"pullPolicy": pullPolicy,
		},
		"gcepdDriver": map[string]any{
			"repo":       "csi-driver",
			"tag":        testTag,
			"pullPolicy": pullPolicy,
		},
	}

	return nil
}

func prepareAzureValues(values map[string]any) error {
	joinVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'join-service' key")
	}
	joinVals["idkeydigest"] = "baaaaaadbaaaaaadbaaaaaadbaaaaaadbaaaaaadbaaaaaadbaaaaaadbaaaaaad"
	joinVals["measurements"] = "{'1':'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA','15':'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA='}"
	joinVals["measurementSalt"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	ccmVals, ok := values["ccm"].(map[string]any)
	if !ok {
		return errors.New("missing 'ccm' key")
	}
	ccmVals["Azure"].(map[string]any)["subnetworkPodCIDR"] = "192.0.2.0/24"
	ccmVals["Azure"].(map[string]any)["azureConfig"] = "baaaaaad"

	autoscalerVals, ok := values["autoscaler"].(map[string]any)
	if !ok {
		return errors.New("missing 'autoscaler' key")
	}
	autoscalerVals["Azure"] = map[string]any{
		"clientID":       "AppClientID",
		"clientSecret":   "ClientSecretValue",
		"resourceGroup":  "resourceGroup",
		"subscriptionID": "subscriptionID",
		"tenantID":       "TenantID",
	}
	return nil
}

func prepareQEMUValues(values map[string]any) error {
	joinVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'join-service' key")
	}
	joinVals["measurements"] = "{'1':'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA','15':'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA='}"
	joinVals["measurementSalt"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	return nil
}
