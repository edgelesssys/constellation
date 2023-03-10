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
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"

	"github.com/edgelesssys/constellation/v2/internal/attestation/idkeydigest"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
)

// TestLoad checks if the serialized format that Load returns correctly preserves the dependencies of the loaded chart.
func TestLoad(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	config := &config.Config{Provider: config.ProviderConfig{GCP: &config.GCPConfig{}}}
	chartLoader := ChartLoader{csp: config.GetProvider()}
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

// TestConstellationServices checks if the rendered constellation-services chart produces the expected yaml files.
func TestConstellationServices(t *testing.T) {
	testCases := map[string]struct {
		config             *config.Config
		enforceIDKeyDigest bool
		valuesModifier     func(map[string]any) error
		ccmImage           string
		cnmImage           string
	}{
		"AWS": {
			config: &config.Config{
				Provider: config.ProviderConfig{AWS: &config.AWSConfig{}},
				Attestation: config.AttestationConfig{AWSNitroTPM: &config.AWSNitroTPM{
					Measurements: measurements.M{1: measurements.WithAllBytes(0xAA, measurements.Enforce, measurements.PCRMeasurementLength)},
				}},
			},
			valuesModifier: prepareAWSValues,
			ccmImage:       "ccmImageForAWS",
		},
		"Azure": {
			config: &config.Config{
				Provider: config.ProviderConfig{Azure: &config.AzureConfig{
					DeployCSIDriver: toPtr(true),
				}},
				Attestation: config.AttestationConfig{AzureSEVSNP: &config.AzureSEVSNP{
					Measurements: measurements.M{1: measurements.WithAllBytes(0xAA, measurements.Enforce, measurements.PCRMeasurementLength)},
					FirmwareSignerConfig: config.SNPFirmwareSignerConfig{
						AcceptedKeyDigests: idkeydigest.List{bytes.Repeat([]byte{0xAA}, 32)},
						EnforcementPolicy:  idkeydigest.MAAFallback,
						MAAURL:             "https://192.0.2.1:8080/maa",
					},
				}},
			},
			enforceIDKeyDigest: true,
			valuesModifier:     prepareAzureValues,
			ccmImage:           "ccmImageForAzure",
			cnmImage:           "cnmImageForAzure",
		},
		"GCP": {
			config: &config.Config{
				Provider: config.ProviderConfig{GCP: &config.GCPConfig{
					DeployCSIDriver: toPtr(true),
				}},
				Attestation: config.AttestationConfig{GCPSEVES: &config.GCPSEVES{
					Measurements: measurements.M{1: measurements.WithAllBytes(0xAA, measurements.Enforce, measurements.PCRMeasurementLength)},
				}},
			},
			valuesModifier: prepareGCPValues,
			ccmImage:       "ccmImageForGCP",
		},
		"OpenStack": {
			config: &config.Config{
				Provider: config.ProviderConfig{OpenStack: &config.OpenStackConfig{}},
				Attestation: config.AttestationConfig{QEMUVTPM: &config.QEMUVTPM{
					Measurements: measurements.M{1: measurements.WithAllBytes(0xAA, measurements.Enforce, measurements.PCRMeasurementLength)},
				}},
			},
			valuesModifier: prepareOpenStackValues,
			ccmImage:       "ccmImageForOpenStack",
		},
		"QEMU": {
			config: &config.Config{
				Provider: config.ProviderConfig{QEMU: &config.QEMUConfig{}},
				Attestation: config.AttestationConfig{QEMUVTPM: &config.QEMUVTPM{
					Measurements: measurements.M{1: measurements.WithAllBytes(0xAA, measurements.Enforce, measurements.PCRMeasurementLength)},
				}},
			},
			valuesModifier: prepareQEMUValues,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			chartLoader := ChartLoader{
				csp:                      tc.config.GetProvider(),
				joinServiceImage:         "joinServiceImage",
				keyServiceImage:          "keyServiceImage",
				ccmImage:                 tc.ccmImage,
				cnmImage:                 tc.cnmImage,
				autoscalerImage:          "autoscalerImage",
				verificationServiceImage: "verificationImage",
				konnectivityImage:        "konnectivityImage",
				gcpGuestAgentImage:       "gcpGuestAgentImage",
			}
			chart, err := loadChartsDir(helmFS, constellationServicesInfo.path)
			require.NoError(err)
			values, err := chartLoader.loadConstellationServicesValues()
			require.NoError(err)
			err = extendConstellationServicesValues(values, tc.config, []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
			require.NoError(err)

			options := chartutil.ReleaseOptions{
				Name:      "testRelease",
				Namespace: "testNamespace",
				Revision:  1,
				IsInstall: true,
				IsUpgrade: false,
			}
			caps := &chartutil.Capabilities{}

			err = tc.valuesModifier(values)
			require.NoError(err)

			// This step is needed to enabled/disable subcharts according to their tags/conditions.
			err = chartutil.ProcessDependencies(chart, values)
			require.NoError(err)

			valuesToRender, err := chartutil.ToRenderValues(chart, values, options, caps)
			require.NoError(err)

			result, err := engine.Render(chart, valuesToRender)
			require.NoError(err)
			testDataPath := path.Join("testdata", tc.config.GetProvider().String(), "constellation-services")

			// Build a map with the same structe as result: filepaths -> rendered template.
			expectedData := map[string]string{}
			err = filepath.Walk(testDataPath, buildTestdataMap(tc.config.GetProvider().String(), expectedData, require))
			require.NoError(err)

			compareMaps(expectedData, result, assert, require, t)
		})
	}
}

// TestOperators checks if the rendered constellation-services chart produces the expected yaml files.
func TestOperators(t *testing.T) {
	testCases := map[string]struct {
		csp cloudprovider.Provider
	}{
		"GCP": {
			csp: cloudprovider.GCP,
		},
		"Azure": {
			csp: cloudprovider.Azure,
		},
		"QEMU": {
			csp: cloudprovider.QEMU,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			chartLoader := ChartLoader{
				csp:                          tc.csp,
				joinServiceImage:             "joinServiceImage",
				keyServiceImage:              "keyServiceImage",
				ccmImage:                     "ccmImage",
				cnmImage:                     "cnmImage",
				autoscalerImage:              "autoscalerImage",
				constellationOperatorImage:   "constellationOperatorImage",
				nodeMaintenanceOperatorImage: "nodeMaintenanceOperatorImage",
			}
			chart, err := loadChartsDir(helmFS, constellationOperatorsInfo.path)
			require.NoError(err)
			vals, err := chartLoader.loadOperatorsValues()
			require.NoError(err)

			options := chartutil.ReleaseOptions{
				Name:      "testRelease",
				Namespace: "testNamespace",
				Revision:  1,
				IsInstall: true,
				IsUpgrade: false,
			}
			caps := &chartutil.Capabilities{}

			conOpVals, ok := vals["constellation-operator"].(map[string]any)
			require.True(ok)
			conOpVals["constellationUID"] = "42424242424242"

			// This step is needed to enabled/disable subcharts according to their tags/conditions.
			err = chartutil.ProcessDependencies(chart, vals)
			require.NoError(err)

			valuesToRender, err := chartutil.ToRenderValues(chart, vals, options, caps)
			require.NoError(err)

			result, err := engine.Render(chart, valuesToRender)
			require.NoError(err)
			testDataPath := path.Join("testdata", tc.csp.String(), "constellation-operators")

			// Build a map with the same structe as result: filepaths -> rendered template.
			expectedData := map[string]string{}
			err = filepath.Walk(testDataPath, buildTestdataMap(tc.csp.String(), expectedData, require))
			require.NoError(err)

			compareMaps(expectedData, result, assert, require, t)
		})
	}
}

// compareMaps ensures that both maps specify the same templates.
func compareMaps(expectedData map[string]string, result map[string]string, assert *assert.Assertions, require *require.Assertions, t *testing.T) {
	// This whole block is only to produce useful error messages.
	// It should allow a developer to see the missing template from just the error message.
	if len(expectedData) > len(result) {
		keys := getKeys(expectedData)
		sort.Strings(keys)
		t.Logf("expected these templates:\n%s", strings.Join(keys, "\n"))

		keys = getKeys(result)
		sort.Strings(keys)
		t.Logf("got these templates:\n%s", strings.Join(keys, "\n"))

		require.FailNow("missing templates in results.")
	}

	// Walk the map and compare each result with it's expected render.
	// Results where the expected-file is missing are errors.
	for k, actualTemplates := range result {
		if len(strings.TrimSpace(actualTemplates)) == 0 {
			continue
		}
		// testify has an issue where when multiple documents are contained in one YAML string,
		// only the first document is parsed [1]. For this reason we split the YAML string
		// into a slice of strings, each entry containing one document.
		// [1] https://github.com/stretchr/testify/issues/1281
		renderedTemplates, ok := expectedData[k]
		require.True(ok, fmt.Sprintf("unexpected render in results, missing file with expected data: %s len: %d", k, len(actualTemplates)))
		expectedSplit := strings.Split(renderedTemplates, "\n---\n")
		sort.Strings(expectedSplit)
		actualSplit := strings.Split(actualTemplates, "\n---\n")
		sort.Strings(actualSplit)
		require.Equal(len(expectedSplit), len(actualSplit))

		for i := range expectedSplit {
			assert.YAMLEq(expectedSplit[i], actualSplit[i], fmt.Sprintf("current file: %s", k))
		}
	}
}

func getKeys(input map[string]string) []string {
	keys := []string{}
	for k := range input {
		keys = append(keys, k)
	}
	return keys
}

func buildTestdataMap(csp string, expectedData map[string]string, require *require.Assertions) func(path string, info fs.FileInfo, err error) error {
	return func(currentPath string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(currentPath, ".yaml") {
			return nil
		}

		_, after, _ := strings.Cut(currentPath, "testdata/"+csp+"/")

		data, err := os.ReadFile(currentPath)
		require.NoError(err)
		_, ok := expectedData[after]
		require.False(ok, "read same path twice during expected data collection.")
		expectedData[after] = string(data)

		return nil
	}
}

func prepareAWSValues(values map[string]any) error {
	joinVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'join-service' key")
	}

	joinVals["measurementSalt"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	ccmVals, ok := values["ccm"].(map[string]any)
	if !ok {
		return errors.New("missing 'ccm' key")
	}
	ccmVals["AWS"].(map[string]any)["subnetworkPodCIDR"] = "192.0.2.0/24"

	verificationVals, ok := values["verification-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'verification-service' key")
	}
	verificationVals["loadBalancerIP"] = "127.0.0.1"

	konnectivityVals, ok := values["konnectivity"].(map[string]any)
	if !ok {
		return errors.New("missing 'konnectivity' key")
	}
	konnectivityVals["loadBalancerIP"] = "127.0.0.1"

	return nil
}

func prepareAzureValues(values map[string]any) error {
	joinVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'join-service' key")
	}

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

	testTag := "v0.0.0"
	pullPolicy := "IfNotPresent"
	verificationVals, ok := values["verification-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'verification-service' key")
	}
	verificationVals["loadBalancerIP"] = "127.0.0.1"

	konnectivityVals, ok := values["konnectivity"].(map[string]any)
	if !ok {
		return errors.New("missing 'konnectivity' key")
	}
	konnectivityVals["loadBalancerIP"] = "127.0.0.1"

	csiVals, ok := values["azuredisk-csi-driver"].(map[string]any)
	if !ok {
		csiVals = map[string]any{}
		values["azuredisk-csi-driver"] = csiVals
	}
	csiImages, ok := csiVals["image"].(map[string]any)
	if !ok {
		csiImages = map[string]any{}
		csiVals["image"] = csiImages
	}
	csiImages["azuredisk"] = map[string]any{
		"repository": "azure-csi-driver",
		"tag":        testTag,
		"pullPolicy": pullPolicy,
	}
	csiImages["csiProvisioner"] = map[string]any{
		"repository": "csi-provisioner",
		"tag":        testTag,
		"pullPolicy": pullPolicy,
	}
	csiImages["csiAttacher"] = map[string]any{
		"repository": "csi-attacher",
		"tag":        testTag,
		"pullPolicy": pullPolicy,
	}
	csiImages["csiResizer"] = map[string]any{
		"repository": "csi-resizer",
		"tag":        testTag,
		"pullPolicy": pullPolicy,
	}
	csiImages["livenessProbe"] = map[string]any{
		"repository": "livenessprobe",
		"tag":        testTag,
		"pullPolicy": pullPolicy,
	}
	csiImages["nodeDriverRegistrar"] = map[string]any{
		"repository": "csi-node-driver-registrar",
		"tag":        testTag,
		"pullPolicy": pullPolicy,
	}
	csiSnapshot, ok := csiVals["snapshot"].(map[string]any)
	if !ok {
		csiSnapshot = map[string]any{}
		csiVals["snapshot"] = csiSnapshot
	}
	csiSnapshotImage, ok := csiSnapshot["image"].(map[string]any)
	if !ok {
		csiSnapshotImage = map[string]any{}
		csiSnapshot["image"] = csiSnapshotImage
	}
	csiSnapshotImage["csiSnapshotter"] = map[string]any{
		"repository": "csi-snapshotter",
		"tag":        testTag,
		"pullPolicy": pullPolicy,
	}
	csiSnapshotImage["snapshotController"] = map[string]any{
		"repository": "snapshot-controller",
		"tag":        testTag,
		"pullPolicy": pullPolicy,
	}

	return nil
}

func prepareGCPValues(values map[string]any) error {
	joinVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'join-service' key")
	}

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
	values["gcp-compute-persistent-disk-csi-driver"] = map[string]any{
		"image": map[string]any{
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
		},
	}

	verificationVals, ok := values["verification-service"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing 'verification-service' key %v", values)
	}
	verificationVals["loadBalancerIP"] = "127.0.0.1"

	konnectivityVals, ok := values["konnectivity"].(map[string]any)
	if !ok {
		return errors.New("missing 'konnectivity' key")
	}
	konnectivityVals["loadBalancerIP"] = "127.0.0.1"

	return nil
}

func prepareOpenStackValues(values map[string]any) error {
	joinVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'join-service' key")
	}
	joinVals["measurementSalt"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	ccmVals, ok := values["ccm"].(map[string]any)
	if !ok {
		return errors.New("missing 'ccm' key")
	}
	ccmVals["OpenStack"].(map[string]any)["subnetworkPodCIDR"] = "192.0.2.0/24"
	ccmVals["OpenStack"].(map[string]any)["secretData"] = "baaaaaad"

	verificationVals, ok := values["verification-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'verification-service' key")
	}
	verificationVals["loadBalancerIP"] = "127.0.0.1"

	konnectivityVals, ok := values["konnectivity"].(map[string]any)
	if !ok {
		return errors.New("missing 'konnectivity' key")
	}
	konnectivityVals["loadBalancerIP"] = "127.0.0.1"

	return nil
}

func prepareQEMUValues(values map[string]any) error {
	joinVals, ok := values["join-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'join-service' key")
	}
	joinVals["measurementSalt"] = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	verificationVals, ok := values["verification-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'verification-service' key")
	}
	verificationVals["loadBalancerIP"] = "127.0.0.1"

	konnectivityVals, ok := values["konnectivity"].(map[string]any)
	if !ok {
		return errors.New("missing 'konnectivity' key")
	}
	konnectivityVals["loadBalancerIP"] = "127.0.0.1"

	return nil
}

func toPtr[T any](v T) *T {
	return &v
}
