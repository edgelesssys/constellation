/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

package helm

import (
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
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

func fakeServiceAccURI(provider cloudprovider.Provider) string {
	switch provider {
	case cloudprovider.GCP:
		cred := gcpshared.ServiceAccountKey{
			Type:                    "service_account",
			ProjectID:               "project_id",
			PrivateKeyID:            "key_id",
			PrivateKey:              "key",
			ClientEmail:             "client_email",
			ClientID:                "client_id",
			AuthURI:                 "auth_uri",
			TokenURI:                "token_uri",
			AuthProviderX509CertURL: "cert",
			ClientX509CertURL:       "client_cert",
		}
		return cred.ToCloudServiceAccountURI()
	case cloudprovider.Azure:
		creds := azureshared.ApplicationCredentials{
			TenantID:            "TenantID",
			Location:            "Location",
			PreferredAuthMethod: azureshared.AuthMethodUserAssignedIdentity,
			UamiResourceID:      "uid",
		}
		return creds.ToCloudServiceAccountURI()
	case cloudprovider.OpenStack:
		creds := openstack.AccountKey{
			AuthURL:           "authURL",
			Username:          "username",
			Password:          "password",
			ProjectID:         "projectID",
			ProjectName:       "projectName",
			UserDomainName:    "userDomainName",
			ProjectDomainName: "projectDomainName",
			RegionName:        "regionName",
		}
		return creds.ToCloudServiceAccountURI()
	default:
		return ""
	}
}

func TestLoadReleases(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	chartLoader := newLoader(
		cloudprovider.GCP, variant.GCPSEVES{}, versions.Default,
		state.New().
			SetInfrastructure(state.Infrastructure{
				GCP: &state.GCP{
					ProjectID: "test-project-id",
					IPCidrPod: "test-pod-cidr",
				},
			}).
			SetClusterValues(state.ClusterValues{MeasurementSalt: []byte{0x41}}),
		semver.NewFromInt(2, 10, 0, ""),
	)
	helmReleases, err := chartLoader.loadReleases(
		true, false, WaitModeAtomic,
		uri.MasterSecret{Key: []byte("secret"), Salt: []byte("masterSalt")},
		fakeServiceAccURI(cloudprovider.GCP), nil, "172.16.128.0/17",
	)
	require.NoError(err)
	for _, release := range helmReleases {
		if release.releaseName == constellationServicesInfo.releaseName {
			assert.NotNil(release.chart.Dependencies())
		}
	}
}

func TestLoadAWSLoadBalancerValues(t *testing.T) {
	sut := chartLoader{
		clusterName: "testCluster",
		stateFile:   state.New().SetInfrastructure(state.Infrastructure{UID: "testuid", Name: "testCluster-testuid"}),
	}
	val := sut.loadAWSLBControllerValues()
	assert.Equal(t, "testCluster-testuid", val["clusterName"])
	// needs to run on control-plane
	assert.Contains(t, val["nodeSelector"].(map[string]any), "node-role.kubernetes.io/control-plane")
	assert.Contains(t, val["tolerations"].([]map[string]any),
		map[string]any{"key": "node-role.kubernetes.io/control-plane", "operator": "Exists", "effect": "NoSchedule"})
}

// TestConstellationServices checks if the rendered constellation-services chart produces the expected yaml files.
func TestConstellationServices(t *testing.T) {
	testCases := map[string]struct {
		config             *config.Config
		enforceIDKeyDigest bool
		ccmImage           string
		cnmImage           string
	}{
		"AWS": {
			config: &config.Config{
				Provider: config.ProviderConfig{AWS: &config.AWSConfig{
					DeployCSIDriver: toPtr(false),
				}},
				Attestation: config.AttestationConfig{AWSNitroTPM: &config.AWSNitroTPM{
					Measurements: measurements.M{1: measurements.WithAllBytes(0xAA, measurements.Enforce, measurements.PCRMeasurementLength)},
				}},
			},
			ccmImage: "ccmImageForAWS",
		},
		"Azure": {
			config: &config.Config{
				Provider: config.ProviderConfig{Azure: &config.AzureConfig{
					DeployCSIDriver: toPtr(true),
				}},
				Attestation: config.AttestationConfig{AzureSEVSNP: &config.AzureSEVSNP{}},
			},
			enforceIDKeyDigest: true,
			ccmImage:           "ccmImageForAzure",
			cnmImage:           "cnmImageForAzure",
		},
		"GCP": {
			config: &config.Config{
				Provider: config.ProviderConfig{GCP: &config.GCPConfig{
					DeployCSIDriver: toPtr(true),
				}},
				Attestation: config.AttestationConfig{GCPSEVES: &config.GCPSEVES{}},
			},
			ccmImage: "ccmImageForGCP",
		},
		"OpenStack": {
			config: &config.Config{
				Provider:    config.ProviderConfig{OpenStack: &config.OpenStackConfig{}},
				Attestation: config.AttestationConfig{QEMUVTPM: &config.QEMUVTPM{}},
			},
			ccmImage: "ccmImageForOpenStack",
		},
		"QEMU": {
			config: &config.Config{
				Provider:    config.ProviderConfig{QEMU: &config.QEMUConfig{}},
				Attestation: config.AttestationConfig{QEMUVTPM: &config.QEMUVTPM{}},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)
			var openstackValues *OpenStackValues
			if tc.config.Provider.OpenStack != nil {
				var deploy bool
				if tc.config.Provider.OpenStack.DeployYawolLoadBalancer != nil {
					deploy = *tc.config.Provider.OpenStack.DeployYawolLoadBalancer
				}
				openstackValues = &OpenStackValues{
					DeployYawolLoadBalancer: deploy,
					FloatingIPPoolID:        tc.config.Provider.OpenStack.FloatingIPPoolID,
					YawolFlavorID:           tc.config.Provider.OpenStack.YawolFlavorID,
					YawolImageID:            tc.config.Provider.OpenStack.YawolImageID,
				}
			}

			chartLoader := chartLoader{
				csp:                      tc.config.GetProvider(),
				joinServiceImage:         "joinServiceImage",
				keyServiceImage:          "keyServiceImage",
				ccmImage:                 tc.ccmImage,
				azureCNMImage:            tc.cnmImage,
				autoscalerImage:          "autoscalerImage",
				verificationServiceImage: "verificationImage",
				gcpGuestAgentImage:       "gcpGuestAgentImage",
				clusterName:              "testCluster",
			}
			chart, err := loadChartsDir(helmFS, constellationServicesInfo.path)
			require.NoError(err)
			values := chartLoader.loadConstellationServicesValues()
			serviceAccURI := fakeServiceAccURI(tc.config.GetProvider())
			extraVals, err := extraConstellationServicesValues(
				tc.config.GetProvider(), tc.config.GetAttestationConfig().GetVariant(), uri.MasterSecret{
					Key:  []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
					Salt: []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				}, serviceAccURI, state.Infrastructure{
					UID:   "uid",
					Azure: &state.Azure{},
					GCP:   &state.GCP{},
				}, openstackValues)
			require.NoError(err)
			values = mergeMaps(values, extraVals)

			options := chartutil.ReleaseOptions{
				Name:      "testRelease",
				Namespace: "testNamespace",
				Revision:  1,
				IsInstall: true,
				IsUpgrade: false,
			}

			kubeVersion, err := chartutil.ParseKubeVersion("1.18.0")
			require.NoError(err)
			caps := &chartutil.Capabilities{
				KubeVersion: *kubeVersion,
			}

			// Add provider tag
			values["tags"] = map[string]any{
				tc.config.GetProvider().String(): true,
			}

			// Add values that are only known after the cluster is created.
			err = addInClusterValues(values, tc.config.GetProvider())
			require.NoError(err)

			// This step is needed to enabled/disable subcharts according to their tags/conditions.
			err = chartutil.ProcessDependencies(chart, values)
			require.NoError(err)

			valuesToRender, err := chartutil.ToRenderValues(chart, values, options, caps)
			require.NoError(err)

			result, err := engine.Render(chart, valuesToRender)
			require.NoError(err)
			testDataPath := path.Join("testdata", tc.config.GetProvider().String(), "constellation-services")

			// Build a map with the same structure as result: filepaths -> rendered template.
			expectedData := map[string]string{}
			err = filepath.Walk(testDataPath, buildTestdataMap(tc.config.GetProvider().String(), expectedData, require))
			require.NoError(err)

			compareMaps(expectedData, result, assert, require, t)
		})
	}
}

func TestExtraCoreDNSValues(t *testing.T) {
	testCases := map[string]struct {
		cidr      string
		wantIP    string
		wantUnset bool
		wantErr   bool
	}{
		"default": {
			cidr:   "10.96.0.0/12",
			wantIP: "10.96.0.10",
		},
		"custom": {
			cidr:   "172.16.128.0/17",
			wantIP: "172.16.128.10",
		},
		"too small": {
			cidr:    "172.16.0.0/30",
			wantErr: true,
		},
		"bad ip": {
			cidr:    "cluster.local",
			wantErr: true,
		},
		"v6": {
			cidr:   "fd12:3456:789a:100::/56",
			wantIP: "fd12:3456:789a:100::a",
		},
		"no ip": {
			wantUnset: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			values, err := extraCoreDNSValues(tc.cidr)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			ip, ok := values["clusterIP"]
			if tc.wantUnset {
				assert.False(t, ok)
				return
			}
			assert.Equal(t, tc.wantIP, ip)
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

			chartLoader := chartLoader{
				csp:                          tc.csp,
				joinServiceImage:             "joinServiceImage",
				keyServiceImage:              "keyServiceImage",
				ccmImage:                     "ccmImage",
				azureCNMImage:                "cnmImage",
				autoscalerImage:              "autoscalerImage",
				constellationOperatorImage:   "constellationOperatorImage",
				nodeMaintenanceOperatorImage: "nodeMaintenanceOperatorImage",
			}
			chart, err := loadChartsDir(helmFS, constellationOperatorsInfo.path)
			require.NoError(err)
			vals := chartLoader.loadOperatorsValues()

			options := chartutil.ReleaseOptions{
				Name:      "testRelease",
				Namespace: "testNamespace",
				Revision:  1,
				IsInstall: true,
				IsUpgrade: false,
			}
			caps := &chartutil.Capabilities{}

			vals["tags"] = map[string]any{
				tc.csp.String(): true,
			}
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

// addInClusterValues adds values that are only known after the cluster is created.
func addInClusterValues(values map[string]any, csp cloudprovider.Provider) error {
	verificationVals, ok := values["verification-service"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing 'verification-service' key %v", values)
	}
	verificationVals["loadBalancerIP"] = "127.0.0.1"

	ccmVals, ok := values["ccm"].(map[string]any)
	if !ok {
		return errors.New("missing 'ccm' key")
	}

	switch csp {
	case cloudprovider.Azure:
		ccmVals[cloudprovider.Azure.String()] = map[string]any{
			"azureConfig": "baaaaaad",
		}

		autoscalerVals, ok := values["autoscaler"].(map[string]any)
		if !ok {
			return errors.New("missing 'autoscaler' key")
		}
		autoscalerVals["Azure"] = map[string]any{
			"resourceGroup":  "resourceGroup",
			"subscriptionID": "subscriptionID",
			"tenantID":       "TenantID",
		}

	case cloudprovider.GCP:
		ccmVals[cloudprovider.GCP.String()] = map[string]any{
			"subnetworkPodCIDR": "192.0.2.0/24",
			"projectID":         "42424242424242",
			"uid":               "242424242424",
			"secretData":        "baaaaaad",
		}

	case cloudprovider.OpenStack:
		ccmVals["OpenStack"] = map[string]any{
			"secretData": "baaaaaad",
		}
	}

	return nil
}

func toPtr[T any](v T) *T {
	return &v
}
