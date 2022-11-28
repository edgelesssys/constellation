/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetCurrentImageVersion(t *testing.T) {
	testCases := map[string]struct {
		stubUpgradePlanner stubUpgradePlanner
		csp                cloudprovider.Provider
		wantErr            bool
	}{
		"valid Azure": {
			stubUpgradePlanner: stubUpgradePlanner{
				image: "/CommunityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/Images/constellation/Versions/0.0.0",
			},
			csp: cloudprovider.Azure,
		},
		"invalid Azure": {
			stubUpgradePlanner: stubUpgradePlanner{
				image: "/CommunityGalleries/someone-else/Images/constellation/Versions/0.0.1",
			},
			csp:     cloudprovider.Azure,
			wantErr: true,
		},
		"valid GCP": {
			stubUpgradePlanner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v0-0-0",
			},
			csp: cloudprovider.GCP,
		},
		"invalid GCP": {
			stubUpgradePlanner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-debug-image",
			},
			csp:     cloudprovider.GCP,
			wantErr: true,
		},
		"invalid CSP": {
			stubUpgradePlanner: stubUpgradePlanner{
				image: "some-image",
			},
			csp:     cloudprovider.Unknown,
			wantErr: true,
		},
		"GetCurrentImage error": {
			stubUpgradePlanner: stubUpgradePlanner{
				err: errors.New("error"),
			},
			csp:     cloudprovider.Azure,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			version, err := getCurrentImageVersion(context.Background(), tc.stubUpgradePlanner, tc.csp)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.True(semver.IsValid(version))
		})
	}
}

type stubUpgradePlanner struct {
	image string
	err   error
}

func (u stubUpgradePlanner) GetCurrentImage(context.Context) (*unstructured.Unstructured, string, error) {
	return nil, u.image, u.err
}

func TestFetchImages(t *testing.T) {
	testImages := map[string]imageManifest{
		"v0.0.0": {
			AzureImage: "azure-v0.0.0",
			GCPImage:   "gcp-v0.0.0",
		},
		"v999.999.999": {
			AzureImage: "azure-v999.999.999",
			GCPImage:   "gcp-v999.999.999",
		},
	}

	testCases := map[string]struct {
		client  *http.Client
		wantErr bool
	}{
		"success": {
			client: newTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(mustMarshal(t, testImages))),
					Header:     make(http.Header),
				}
			}),
		},
		"error": {
			client: newTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBuffer([]byte{})),
					Header:     make(http.Header),
				}
			}),
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			images, err := fetchImages(context.Background(), tc.client)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.NotNil(images)
		})
	}
}

func TestGetCompatibleImages(t *testing.T) {
	imageList := map[string]imageManifest{
		"v0.0.0": {
			AzureImage: "azure-v0.0.0",
			GCPImage:   "gcp-v0.0.0",
		},
		"v1.0.0": {
			AzureImage: "azure-v1.0.0",
			GCPImage:   "gcp-v1.0.0",
		},
		"v1.0.1": {
			AzureImage: "azure-v1.0.1",
			GCPImage:   "gcp-v1.0.1",
		},
		"v1.0.2": {
			AzureImage: "azure-v1.0.2",
			GCPImage:   "gcp-v1.0.2",
		},
		"v1.1.0": {
			AzureImage: "azure-v1.1.0",
			GCPImage:   "gcp-v1.1.0",
		},
	}

	testCases := map[string]struct {
		images     map[string]imageManifest
		csp        cloudprovider.Provider
		version    string
		wantImages map[string]config.UpgradeConfig
	}{
		"azure": {
			images:  imageList,
			csp:     cloudprovider.Azure,
			version: "v1.0.0",
			wantImages: map[string]config.UpgradeConfig{
				"v1.0.1": {
					Image: "azure-v1.0.1",
					CSP:   cloudprovider.Azure,
				},
				"v1.0.2": {
					Image: "azure-v1.0.2",
					CSP:   cloudprovider.Azure,
				},
				"v1.1.0": {
					Image: "azure-v1.1.0",
					CSP:   cloudprovider.Azure,
				},
			},
		},
		"gcp": {
			images:  imageList,
			csp:     cloudprovider.GCP,
			version: "v1.0.0",
			wantImages: map[string]config.UpgradeConfig{
				"v1.0.1": {
					Image: "gcp-v1.0.1",
					CSP:   cloudprovider.GCP,
				},
				"v1.0.2": {
					Image: "gcp-v1.0.2",
					CSP:   cloudprovider.GCP,
				},
				"v1.1.0": {
					Image: "gcp-v1.1.0",
					CSP:   cloudprovider.GCP,
				},
			},
		},
		"no compatible images": {
			images:     imageList,
			csp:        cloudprovider.Azure,
			version:    "v999.999.999",
			wantImages: map[string]config.UpgradeConfig{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			compatibleImages := getCompatibleImages(tc.csp, tc.version, tc.images)
			assert.Equal(tc.wantImages, compatibleImages)
		})
	}
}

func TestGetCompatibleImageMeasurements(t *testing.T) {
	assert := assert.New(t)

	testImages := map[string]config.UpgradeConfig{
		"v0.0.0": {
			Image: "v0.0.0",
			CSP:   cloudprovider.Azure,
		},
		"v1.0.0": {
			Image: "v1.0.0",
			CSP:   cloudprovider.Azure,
		},
	}

	client := newTestClient(func(req *http.Request) *http.Response {
		if strings.HasSuffix(req.URL.String(), "v0.0.0/azure/measurements.json") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"csp":"azure","image":"v0.0.0","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}`)),
				Header:     make(http.Header),
			}
		}
		if strings.HasSuffix(req.URL.String(), "v0.0.0/azure/measurements.json.sig") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("MEQCIGRR7RaSMs892Ta06/Tz7LqPUxI05X4wQcP+nFFmZtmaAiBNl9X8mUKmUBfxg13LQBfmmpw6JwYQor5hOwM3NFVPAg==")),
				Header:     make(http.Header),
			}
		}

		if strings.HasSuffix(req.URL.String(), "v1.0.0/azure/measurements.json") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"csp":"azure","image":"v1.0.0","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}`)),
				Header:     make(http.Header),
			}
		}
		if strings.HasSuffix(req.URL.String(), "v1.0.0/azure/measurements.json.sig") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("MEQCIFh8CVELp/Da2U2Jt404OXsUeDfqtrf3pqGRuvxnxhI8AiBTHF9tHEPwFedYG3Jgn2ELOxss+Ybc6135vEtClBrbpg==")),
				Header:     make(http.Header),
			}
		}

		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("Not found.")),
			Header:     make(http.Header),
		}
	})

	pubK := []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----")

	err := getCompatibleImageMeasurements(context.Background(), &cobra.Command{}, client, singleUUIDVerifier(), pubK, testImages)
	assert.NoError(err)

	for _, image := range testImages {
		assert.NotEmpty(image.Measurements)
	}
}

func TestUpgradePlan(t *testing.T) {
	testImages := map[string]imageManifest{
		"v1.0.0": {
			AzureImage: "v1.0.0",
			GCPImage:   "v1.0.0",
		},
	}

	// Cosign private key used to sign the measurements.
	// Generated with: cosign generate-key-pair
	// Password left empty.
	//
	// -----BEGIN ENCRYPTED COSIGN PRIVATE KEY-----
	// eyJrZGYiOnsibmFtZSI6InNjcnlwdCIsInBhcmFtcyI6eyJOIjozMjc2OCwiciI6
	// OCwicCI6MX0sInNhbHQiOiJlRHVYMWRQMGtIWVRnK0xkbjcxM0tjbFVJaU92eFVX
	// VXgvNi9BbitFVk5BPSJ9LCJjaXBoZXIiOnsibmFtZSI6Im5hY2wvc2VjcmV0Ym94
	// Iiwibm9uY2UiOiJwaWhLL2txNmFXa2hqSVVHR3RVUzhTVkdHTDNIWWp4TCJ9LCJj
	// aXBoZXJ0ZXh0Ijoidm81SHVWRVFWcUZ2WFlQTTVPaTVaWHM5a255bndZU2dvcyth
	// VklIeHcrOGFPamNZNEtvVjVmL3lHRHR0K3BHV2toanJPR1FLOWdBbmtsazFpQ0c5
	// a2czUXpPQTZsU2JRaHgvZlowRVRZQ0hLeElncEdPRVRyTDlDenZDemhPZXVSOXJ6
	// TDcvRjBBVy9vUDVqZXR3dmJMNmQxOEhjck9kWE8yVmYxY2w0YzNLZjVRcnFSZzlN
	// dlRxQWFsNXJCNHNpY1JaMVhpUUJjb0YwNHc9PSJ9
	// -----END ENCRYPTED COSIGN PRIVATE KEY-----
	pubK := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEu78QgxOOcao6U91CSzEXxrKhvFTt\nJHNy+eX6EMePtDm8CnDF9HSwnTlD0itGJ/XHPQA5YX10fJAqI1y+ehlFMw==\n-----END PUBLIC KEY-----"

	testCases := map[string]struct {
		planner                 stubUpgradePlanner
		flags                   upgradePlanFlags
		csp                     cloudprovider.Provider
		verifier                rekorVerifier
		imageFetchStatus        int
		measurementsFetchStatus int
		wantUpgrade             bool
		wantErr                 bool
	}{
		"no compatible images": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v999-999-999",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:         cloudprovider.GCP,
			verifier:    singleUUIDVerifier(),
			wantUpgrade: false,
		},
		"upgrades gcp": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v0-0-0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:         cloudprovider.GCP,
			verifier:    singleUUIDVerifier(),
			wantUpgrade: true,
		},
		"upgrades azure": {
			planner: stubUpgradePlanner{
				image: "/CommunityGalleries/ConstellationCVM-b3782fa0-0df7-4f2f-963e-fc7fc42663df/Images/constellation/Versions/0.0.0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:         cloudprovider.Azure,
			verifier:    singleUUIDVerifier(),
			wantUpgrade: true,
		},
		"upgrade to stdout": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v0-0-0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "-",
				cosignPubKey: pubK,
			},
			csp:         cloudprovider.GCP,
			verifier:    singleUUIDVerifier(),
			wantUpgrade: true,
		},
		"current image not valid": {
			planner: stubUpgradePlanner{
				image: "not-valid",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:      cloudprovider.GCP,
			verifier: singleUUIDVerifier(),
			wantErr:  true,
		},
		"image fetch error": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v0-0-0",
			},
			imageFetchStatus:        http.StatusInternalServerError,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:      cloudprovider.GCP,
			verifier: singleUUIDVerifier(),
			wantErr:  true,
		},
		"measurements fetch error": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v0-0-0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusInternalServerError,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:      cloudprovider.GCP,
			verifier: singleUUIDVerifier(),
			wantErr:  true,
		},
		"failing search should not result in error": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v0-0-0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp: cloudprovider.GCP,
			verifier: &stubRekorVerifier{
				SearchByHashUUIDs: []string{},
				SearchByHashError: errors.New("some error"),
			},
			wantUpgrade: true,
		},
		"failing verify should not result in error": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v0-0-0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp: cloudprovider.GCP,
			verifier: &stubRekorVerifier{
				SearchByHashUUIDs: []string{"11111111111111111111111111111111111111111111111111111111111111111111111111111111"},
				VerifyEntryError:  errors.New("some error"),
			},
			wantUpgrade: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), tc.csp)

			require.NoError(fileHandler.WriteYAML(tc.flags.configPath, cfg))

			cmd := newUpgradePlanCmd()
			cmd.SetContext(context.Background())
			var outTarget bytes.Buffer
			cmd.SetOut(&outTarget)
			var errTarget bytes.Buffer
			cmd.SetErr(&errTarget)

			client := newTestClient(func(req *http.Request) *http.Response {
				if req.URL.String() == imageReleaseURL {
					return &http.Response{
						StatusCode: tc.imageFetchStatus,
						Body:       io.NopCloser(bytes.NewBuffer(mustMarshal(t, testImages))),
						Header:     make(http.Header),
					}
				}
				if strings.HasSuffix(req.URL.String(), "azure/measurements.json") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader(`{"csp":"azure","image":"v1.0.0","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}`)),
						Header:     make(http.Header),
					}
				}
				if strings.HasSuffix(req.URL.String(), "azure/measurements.json.sig") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader("MEQCIFh8CVELp/Da2U2Jt404OXsUeDfqtrf3pqGRuvxnxhI8AiBTHF9tHEPwFedYG3Jgn2ELOxss+Ybc6135vEtClBrbpg==")),
						Header:     make(http.Header),
					}
				}

				if strings.HasSuffix(req.URL.String(), "gcp/measurements.json") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader(`{"csp":"gcp","image":"v1.0.0","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}`)),
						Header:     make(http.Header),
					}
				}
				if strings.HasSuffix(req.URL.String(), "gcp/measurements.json.sig") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader("MEYCIQCr/gDGjj11mR5OeImwOLjxnBqMbBmqoK7yXqy0cXR3HQIhALpVDdYwR9VNJnWwtl8bTfrezyJbc7UNZJO4PJe+stFP")),
						Header:     make(http.Header),
					}
				}

				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("Not found.")),
					Header:     make(http.Header),
				}
			})

			err := upgradePlan(cmd, tc.planner, fileHandler, client, tc.verifier, tc.flags)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if !tc.wantUpgrade {
				assert.Contains(errTarget.String(), "No compatible images")
				return
			}

			var availableUpgrades map[string]config.UpgradeConfig
			if tc.flags.filePath == "-" {
				require.NoError(yaml.Unmarshal(outTarget.Bytes(), &availableUpgrades))
			} else {
				require.NoError(fileHandler.ReadYAMLStrict(tc.flags.filePath, &availableUpgrades))
			}

			assert.GreaterOrEqual(len(availableUpgrades), 1)
			for _, upgrade := range availableUpgrades {
				assert.NotEmpty(upgrade.Image)
				assert.NotEmpty(upgrade.Measurements)
			}
		})
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}
	return b
}
