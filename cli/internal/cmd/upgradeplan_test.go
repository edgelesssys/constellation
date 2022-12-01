/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
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
		wantErr            bool
	}{
		"valid version": {
			stubUpgradePlanner: stubUpgradePlanner{
				image: "v1.0.0",
			},
		},
		"invalid version": {
			stubUpgradePlanner: stubUpgradePlanner{
				image: "invalid",
			},
			wantErr: true,
		},
		"GetCurrentImage error": {
			stubUpgradePlanner: stubUpgradePlanner{
				err: errors.New("error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			version, err := getCurrentImageVersion(context.Background(), tc.stubUpgradePlanner)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.True(semver.IsValid(version))
		})
	}
}

func TestGetCompatibleImages(t *testing.T) {
	imageList := []string{
		"v0.0.0",
		"v1.0.0",
		"v1.0.1",
		"v1.0.2",
		"v1.1.0",
	}

	testCases := map[string]struct {
		images     []string
		version    string
		wantImages []string
	}{
		"filters <= v1.0.0": {
			images:  imageList,
			version: "v1.0.0",
			wantImages: []string{
				"v1.0.1",
				"v1.0.2",
				"v1.1.0",
			},
		},
		"no compatible images": {
			images:  imageList,
			version: "v999.999.999",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			compatibleImages := getCompatibleImages(tc.version, tc.images)
			assert.EqualValues(tc.wantImages, compatibleImages)
		})
	}
}

func TestGetCompatibleImageMeasurements(t *testing.T) {
	assert := assert.New(t)

	csp := cloudprovider.Azure
	images := []string{"v0.0.0", "v1.0.0"}

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

	upgrades, err := getCompatibleImageMeasurements(context.Background(), &cobra.Command{}, client, singleUUIDVerifier(), pubK, csp, images)
	assert.NoError(err)

	for _, image := range upgrades {
		assert.NotEmpty(image.Measurements)
	}
}

func TestUpgradePlan(t *testing.T) {
	availablePatches := versionsapi.List{
		Versions: []string{"v1.0.0", "v1.0.1"},
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
		patchLister             stubPatchLister
		planner                 stubUpgradePlanner
		flags                   upgradePlanFlags
		cliVersion              string
		csp                     cloudprovider.Provider
		verifier                rekorVerifier
		measurementsFetchStatus int
		wantUpgrade             bool
		wantErr                 bool
	}{
		"upgrades gcp": {
			patchLister: stubPatchLister{list: availablePatches},
			planner: stubUpgradePlanner{
				image: "v1.0.0",
			},
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			cliVersion:  "v1.0.0",
			csp:         cloudprovider.GCP,
			verifier:    singleUUIDVerifier(),
			wantUpgrade: true,
		},
		"upgrades azure": {
			patchLister: stubPatchLister{list: availablePatches},
			planner: stubUpgradePlanner{
				image: "v1.0.0",
			},
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:         cloudprovider.Azure,
			cliVersion:  "v999.999.999",
			verifier:    singleUUIDVerifier(),
			wantUpgrade: true,
		},
		"current image newer than updates": {
			patchLister: stubPatchLister{list: availablePatches},
			planner: stubUpgradePlanner{
				image: "v999.999.999",
			},
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
		"current image newer than cli": {
			patchLister: stubPatchLister{list: availablePatches},
			planner: stubUpgradePlanner{
				image: "v999.999.999",
			},
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:         cloudprovider.GCP,
			cliVersion:  "v1.0.0",
			verifier:    singleUUIDVerifier(),
			wantUpgrade: false,
		},
		"upgrade to stdout": {
			patchLister: stubPatchLister{list: availablePatches},
			planner: stubUpgradePlanner{
				image: "v1.0.0",
			},
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "-",
				cosignPubKey: pubK,
			},
			csp:         cloudprovider.GCP,
			cliVersion:  "v1.0.0",
			verifier:    singleUUIDVerifier(),
			wantUpgrade: true,
		},
		"current image not valid": {
			patchLister: stubPatchLister{list: availablePatches},
			planner: stubUpgradePlanner{
				image: "not-valid",
			},
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:        cloudprovider.GCP,
			cliVersion: "v1.0.0",
			verifier:   singleUUIDVerifier(),
			wantErr:    true,
		},
		"image fetch error": {
			patchLister: stubPatchLister{err: errors.New("error")},
			planner: stubUpgradePlanner{
				image: "v1.0.0",
			},
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:        cloudprovider.GCP,
			cliVersion: "v1.0.0",
			verifier:   singleUUIDVerifier(),
		},
		"measurements fetch error": {
			patchLister: stubPatchLister{list: availablePatches},
			planner: stubUpgradePlanner{
				image: "v1.0.0",
			},
			measurementsFetchStatus: http.StatusInternalServerError,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:        cloudprovider.GCP,
			cliVersion: "v1.0.0",
			verifier:   singleUUIDVerifier(),
		},
		"failing search should not result in error": {
			patchLister: stubPatchLister{list: availablePatches},
			planner: stubUpgradePlanner{
				image: "v1.0.0",
			},
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:        cloudprovider.GCP,
			cliVersion: "v1.0.0",
			verifier: &stubRekorVerifier{
				SearchByHashUUIDs: []string{},
				SearchByHashError: errors.New("some error"),
			},
			wantUpgrade: true,
		},
		"failing verify should not result in error": {
			patchLister: stubPatchLister{list: availablePatches},
			planner: stubUpgradePlanner{
				image: "v1.0.0",
			},
			measurementsFetchStatus: http.StatusOK,
			flags: upgradePlanFlags{
				configPath:   constants.ConfigFilename,
				filePath:     "upgrade-plan.yaml",
				cosignPubKey: pubK,
			},
			csp:        cloudprovider.GCP,
			cliVersion: "v1.0.0",
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
				if strings.HasSuffix(req.URL.String(), "azure/measurements.json") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader(`{"csp":"azure","image":"v1.0.1","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}`)),
						Header:     make(http.Header),
					}
				}
				if strings.HasSuffix(req.URL.String(), "azure/measurements.json.sig") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader("MEYCIQDu2Sft91FjN278uP+r/HFMms6IH/tRtaHzYvIN0xPgdwIhAJhiFxVsHCa0NK6bZOGLE9c4miZHIqFTKvgpTf3rJ9dW")),
						Header:     make(http.Header),
					}
				}

				if strings.HasSuffix(req.URL.String(), "gcp/measurements.json") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader(`{"csp":"gcp","image":"v1.0.1","measurements":{"0":{"expected":"0000000000000000000000000000000000000000000000000000000000000000","warnOnly":false}}}`)),
						Header:     make(http.Header),
					}
				}
				if strings.HasSuffix(req.URL.String(), "gcp/measurements.json.sig") {
					return &http.Response{
						StatusCode: tc.measurementsFetchStatus,
						Body:       io.NopCloser(strings.NewReader("MEQCIBUssv92LpSMiXE1UAVf2fW8J9pZHiLseo2tdZjxv2OMAiB6K8e8yL0768jWjlFnRe3Rc2x/dX34uzX3h0XUrlYt1A==")),
						Header:     make(http.Header),
					}
				}

				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(strings.NewReader("Not found.")),
					Header:     make(http.Header),
				}
			})

			err := upgradePlan(cmd, tc.planner, tc.patchLister, fileHandler, client, tc.verifier, tc.flags, tc.cliVersion)
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

func TestNextMinorVersion(t *testing.T) {
	testCases := map[string]struct {
		version              string
		wantNextMinorVersion string
		wantErr              bool
	}{
		"gets next": {
			version:              "v1.0.0",
			wantNextMinorVersion: "v1.1",
		},
		"gets next from minor version": {
			version:              "v1.0",
			wantNextMinorVersion: "v1.1",
		},
		"empty version": {
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			gotNext, err := nextMinorVersion(tc.version)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.Equal(tc.wantNextMinorVersion, gotNext)
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

type stubPatchLister struct {
	list versionsapi.List
	err  error
}

func (s stubPatchLister) PatchVersionsOf(ctx context.Context, stream, minor, kind string) (*versionsapi.List, error) {
	return &s.list, s.err
}
