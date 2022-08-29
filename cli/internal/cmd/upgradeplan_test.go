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

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/config"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
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
				image: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/0.0.0",
			},
			csp: cloudprovider.Azure,
		},
		"invalid Azure": {
			stubUpgradePlanner: stubUpgradePlanner{
				image: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation_Debug/images/constellation/versions/0.0.0",
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
				image: "azure",
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
				},
				"v1.0.2": {
					Image: "azure-v1.0.2",
				},
				"v1.1.0": {
					Image: "azure-v1.1.0",
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
				},
				"v1.0.2": {
					Image: "gcp-v1.0.2",
				},
				"v1.1.0": {
					Image: "gcp-v1.1.0",
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
			Image: "azure-v0.0.0",
		},
		"v1.0.0": {
			Image: "azure-v1.0.0",
		},
	}

	client := newTestClient(func(req *http.Request) *http.Response {
		if strings.HasSuffix(req.URL.String(), "/measurements.yaml") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("0: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n")),
				Header:     make(http.Header),
			}
		}
		if strings.HasSuffix(req.URL.String(), "/measurements.yaml.sig") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("MEUCIBs1g2/n0FsgPfJ+0uLD5TaunGhxwDcQcUGBroejKvg3AiEAzZtcLU9O6IiVhxB8tBS+ty6MXoPNwL8WRWMzyr35eKI=")),
				Header:     make(http.Header),
			}
		}

		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("Not found.")),
			Header:     make(http.Header),
		}
	})

	pubK := []byte("-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----")

	err := getCompatibleImageMeasurements(context.Background(), client, pubK, testImages)
	assert.NoError(err)

	for _, image := range testImages {
		assert.NotEmpty(image.Measurements)
	}
}

func TestUpgradePlan(t *testing.T) {
	testImages := map[string]imageManifest{
		"v1.0.0": {
			AzureImage: "azure-v1.0.0",
			GCPImage:   "gcp-v1.0.0",
		},
		"v2.0.0": {
			AzureImage: "azure-v2.0.0",
			GCPImage:   "gcp-v2.0.0",
		},
	}

	testCases := map[string]struct {
		planner                 stubUpgradePlanner
		flags                   upgradePlanFlags
		csp                     cloudprovider.Provider
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
				cosignPubKey: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----",
			},
			csp:         cloudprovider.GCP,
			wantUpgrade: false,
		},
		"upgrades gcp": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v1-0-0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----",
			},
			csp:         cloudprovider.GCP,
			wantUpgrade: true,
		},
		"upgrades azure": {
			planner: stubUpgradePlanner{
				image: "/subscriptions/0d202bbb-4fa7-4af8-8125-58c269a05435/resourceGroups/CONSTELLATION-IMAGES/providers/Microsoft.Compute/galleries/Constellation/images/constellation/versions/1.0.0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----",
			},
			csp:         cloudprovider.Azure,
			wantUpgrade: true,
		},
		"upgrade to stdout": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v1-0-0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "-",
				cosignPubKey: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----",
			},
			csp:         cloudprovider.GCP,
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
				cosignPubKey: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----",
			},
			csp:     cloudprovider.GCP,
			wantErr: true,
		},
		"image fetch error": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v1-0-0",
			},
			imageFetchStatus:        http.StatusInternalServerError,
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----",
			},
			csp:     cloudprovider.GCP,
			wantErr: true,
		},
		"measurements fetch error": {
			planner: stubUpgradePlanner{
				image: "projects/constellation-images/global/images/constellation-v1-0-0",
			},
			imageFetchStatus:        http.StatusOK,
			measurementsFetchStatus: http.StatusInternalServerError,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEUs5fDUIz9aiwrfr8BK4VjN7jE6sl\ngz7UuXsOin8+dB0SGrbNHy7TJToa2fAiIKPVLTOfvY75DqRAtffhO1fpBA==\n-----END PUBLIC KEY-----",
			},
			csp:     cloudprovider.GCP,
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			cfg := config.Default()
			cfg.RemoveProviderExcept(tc.csp)
			require.NoError(fileHandler.WriteYAML(tc.flags.configPath, cfg))

			cmd := newUpgradePlanCmd()
			cmd.SetContext(context.Background())
			var out bytes.Buffer
			cmd.SetOut(&out)

			client := newTestClient(func(req *http.Request) *http.Response {
				if req.URL.String() == imageReleaseURL {
					return &http.Response{
						StatusCode: tc.imageFetchStatus,
						Body:       io.NopCloser(bytes.NewBuffer(mustMarshal(t, testImages))),
						Header:     make(http.Header),
					}
				}
				if strings.HasSuffix(req.URL.String(), "/measurements.yaml") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader("0: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n")),
						Header:     make(http.Header),
					}
				}
				if strings.HasSuffix(req.URL.String(), "/measurements.yaml.sig") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader("MEUCIBs1g2/n0FsgPfJ+0uLD5TaunGhxwDcQcUGBroejKvg3AiEAzZtcLU9O6IiVhxB8tBS+ty6MXoPNwL8WRWMzyr35eKI=")),
						Header:     make(http.Header),
					}
				}

				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("Not found.")),
					Header:     make(http.Header),
				}
			})

			err := upgradePlan(cmd, tc.planner, fileHandler, client, tc.flags)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			if !tc.wantUpgrade {
				assert.Contains(out.String(), "No compatible images")
				return
			}

			var availableUpgrades map[string]config.UpgradeConfig
			if tc.flags.filePath == "-" {
				require.NoError(yaml.Unmarshal(out.Bytes(), &availableUpgrades))
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

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %s", err)
	}
	return b
}
