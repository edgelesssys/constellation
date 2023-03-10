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

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/variant"
	"github.com/edgelesssys/constellation/v2/internal/versionsapi"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"
)

// TestBuildString checks that the resulting user output is as expected. Slow part is the Sscanf in parseCanonicalSemver().
func TestBuildString(t *testing.T) {
	testCases := map[string]struct {
		upgrade   versionUpgrade
		expected  string
		wantError bool
	}{
		"update everything": {
			upgrade: versionUpgrade{
				newServices: "v2.5.0",
				newImages: map[string]measurements.M{
					"v2.5.0": measurements.DefaultsFor(variant.QEMUVTPM{}),
				},
				newKubernetes:     []string{"v1.24.12", "v1.25.6"},
				newCLI:            []string{"v2.5.0", "v2.6.0"},
				currentServices:   "v2.4.0",
				currentImage:      "v2.4.0",
				currentKubernetes: "v1.24.5",
				currentCLI:        "v2.4.0",
			},
			expected: "The following updates are available with this CLI:\n  Kubernetes: v1.24.5 --> v1.24.12 v1.25.6\n  Images:\n    v2.4.0 --> v2.5.0\n      Includes these measurements:\n      4:\n          expected: \"1234123412341234123412341234123412341234123412341234123412341234\"\n          warnOnly: false\n      8:\n          expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n          warnOnly: false\n      9:\n          expected: \"1234123412341234123412341234123412341234123412341234123412341234\"\n          warnOnly: false\n      11:\n          expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n          warnOnly: false\n      12:\n          expected: \"1234123412341234123412341234123412341234123412341234123412341234\"\n          warnOnly: false\n      13:\n          expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n          warnOnly: false\n      15:\n          expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n          warnOnly: false\n      \n  Services: v2.4.0 --> v2.5.0\n",
		},
		"cli incompatible with K8s": {
			upgrade: versionUpgrade{
				newCLI:     []string{"v2.5.0", "v2.6.0"},
				currentCLI: "v2.4.0",
			},
			expected: "There are newer CLIs available (v2.5.0 v2.6.0), however, you need to upgrade your cluster's Kubernetes version first.\n",
		},
		"cli compatible with K8s": {
			upgrade: versionUpgrade{
				newCompatibleCLI: []string{"v2.5.0", "v2.6.0"},
				currentCLI:       "v2.4.0",
			},
			expected: "Newer CLI versions that are compatible with your cluster are: v2.5.0 v2.6.0\n",
		},
		"k8s only": {
			upgrade: versionUpgrade{
				newKubernetes:     []string{"v1.24.12", "v1.25.6"},
				currentKubernetes: "v1.24.5",
			},
			expected: "The following updates are available with this CLI:\n  Kubernetes: v1.24.5 --> v1.24.12 v1.25.6\n",
		},
		"no upgrades": {
			upgrade: versionUpgrade{
				newServices:       "",
				newImages:         map[string]measurements.M{},
				newKubernetes:     []string{},
				newCLI:            []string{},
				currentServices:   "v2.5.0",
				currentImage:      "v2.5.0",
				currentKubernetes: "v1.25.6",
				currentCLI:        "v2.5.0",
			},
			expected: "You are up to date.\n",
		},
		"no upgrades #2": {
			upgrade:  versionUpgrade{},
			expected: "You are up to date.\n",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			result, err := tc.upgrade.buildString()
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.Equal(tc.expected, result)
		})
	}
}

func TestGetCurrentImageVersion(t *testing.T) {
	testCases := map[string]struct {
		stubUpgradeChecker stubUpgradeChecker
		wantErr            bool
	}{
		"valid version": {
			stubUpgradeChecker: stubUpgradeChecker{
				image: "v1.0.0",
			},
		},
		"invalid version": {
			stubUpgradeChecker: stubUpgradeChecker{
				image: "invalid",
			},
			wantErr: true,
		},
		"GetCurrentImage error": {
			stubUpgradeChecker: stubUpgradeChecker{
				err: errors.New("error"),
			},
			wantErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			version, err := getCurrentImageVersion(context.Background(), tc.stubUpgradeChecker)
			if tc.wantErr {
				assert.Error(err)
				return
			}

			assert.NoError(err)
			assert.True(semver.IsValid(version))
		})
	}
}

func TestGetCompatibleImageMeasurements(t *testing.T) {
	assert := assert.New(t)

	csp := cloudprovider.Azure
	zero := versionsapi.Version{
		Ref:     "-",
		Stream:  "stable",
		Version: "v0.0.0",
		Kind:    versionsapi.VersionKindImage,
	}
	one := versionsapi.Version{
		Ref:     "-",
		Stream:  "stable",
		Version: "v1.0.0",
		Kind:    versionsapi.VersionKindImage,
	}
	images := []versionsapi.Version{zero, one}

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

	upgrades, err := getCompatibleImageMeasurements(context.Background(), &bytes.Buffer{}, client, singleUUIDVerifier(), pubK, csp, images, logger.NewTest(t))
	assert.NoError(err)

	for _, measurement := range upgrades {
		assert.NotEmpty(measurement)
	}
}

func TestUpgradeCheck(t *testing.T) {
	v2_3 := versionsapi.Version{
		Ref:     "-",
		Stream:  "stable",
		Version: "v2.3.0",
		Kind:    versionsapi.VersionKindImage,
	}
	v2_5 := versionsapi.Version{
		Ref:     "-",
		Stream:  "stable",
		Version: "v2.5.0",
		Kind:    versionsapi.VersionKindImage,
	}
	testCases := map[string]struct {
		collector  stubVersionCollector
		flags      upgradeCheckFlags
		csp        cloudprovider.Provider
		cliVersion string
		wantError  bool
	}{
		"upgrades gcp": {
			collector: stubVersionCollector{
				supportedServicesVersions: "v2.5.0",
				supportedImages:           []versionsapi.Version{v2_3},
				supportedImageVersions: map[string]measurements.M{
					"v2.3.0": measurements.DefaultsFor(variant.QEMUVTPM{}),
				},
				supportedK8sVersions:    []string{"v1.24.5", "v1.24.12", "v1.25.6"},
				currentServicesVersions: "v2.4.0",
				currentImageVersion:     "v2.4.0",
				currentK8sVersion:       "v1.24.5",
				currentCLIVersion:       "v2.4.0",
				images:                  []versionsapi.Version{v2_5},
				newCLIVersionsList:      []string{"v2.5.0", "v2.6.0"},
			},
			flags: upgradeCheckFlags{
				configPath: constants.ConfigFilename,
			},
			csp:        cloudprovider.GCP,
			cliVersion: "v1.0.0",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), tc.csp)
			require.NoError(fileHandler.WriteYAML(tc.flags.configPath, cfg))

			checkCmd := upgradeCheckCmd{
				collect: &tc.collector,
				log:     logger.NewTest(t),
			}

			cmd := newUpgradeCheckCmd()

			err := checkCmd.upgradeCheck(cmd, fileHandler, tc.flags)
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

type stubVersionCollector struct {
	supportedServicesVersions    string
	supportedImages              []versionsapi.Version
	supportedImageVersions       map[string]measurements.M
	supportedK8sVersions         []string
	supportedCLIVersions         []string
	currentServicesVersions      string
	currentImageVersion          string
	currentK8sVersion            string
	currentCLIVersion            string
	images                       []versionsapi.Version
	newCLIVersionsList           []string
	newCompatibleCLIVersionsList []string
	someErr                      error
}

func (s *stubVersionCollector) newMeasurements(_ context.Context, _ cloudprovider.Provider, _ []versionsapi.Version) (map[string]measurements.M, error) {
	return s.supportedImageVersions, nil
}

func (s *stubVersionCollector) currentVersions(_ context.Context) (currentVersionInfo, error) {
	return currentVersionInfo{
		service: s.currentServicesVersions,
		image:   s.currentImageVersion,
		k8s:     s.currentK8sVersion,
		cli:     s.currentCLIVersion,
	}, s.someErr
}

func (s *stubVersionCollector) supportedVersions(_ context.Context, _, _ string) (supportedVersionInfo, error) {
	return supportedVersionInfo{
		service: s.supportedServicesVersions,
		image:   s.supportedImages,
		k8s:     s.supportedK8sVersions,
		cli:     s.supportedCLIVersions,
	}, s.someErr
}

func (s *stubVersionCollector) newImages(_ context.Context, _ string) ([]versionsapi.Version, error) {
	return s.images, nil
}

func (s *stubVersionCollector) newerVersions(_ context.Context, _ []string) ([]versionsapi.Version, error) {
	return s.images, nil
}

func (s *stubVersionCollector) newCLIVersions(_ context.Context) ([]string, error) {
	return s.newCLIVersionsList, nil
}

func (s *stubVersionCollector) filterCompatibleCLIVersions(_ context.Context, _ []string, _ string) ([]string, error) {
	return s.newCompatibleCLIVersionsList, nil
}

type stubUpgradeChecker struct {
	image      string
	k8sVersion string
	err        error
}

func (u stubUpgradeChecker) CurrentImage(context.Context) (string, error) {
	return u.image, u.err
}

func (u stubUpgradeChecker) CurrentKubernetesVersion(_ context.Context) (string, error) {
	return u.k8sVersion, u.err
}

func TestNewCLIVersions(t *testing.T) {
	someErr := errors.New("some error")
	minorList := func() versionsapi.List {
		return versionsapi.List{
			Versions: []string{"v0.2.0"},
		}
	}
	patchList := func() versionsapi.List {
		return versionsapi.List{
			Versions: []string{"v0.2.1"},
		}
	}
	emptyVerList := func() versionsapi.List {
		return versionsapi.List{}
	}
	verCollector := func(minorList, patchList versionsapi.List, verListErr error) versionCollector {
		return versionCollector{
			cliVersion: "v0.1.0",
			versionsapi: stubVersionFetcher{
				minorList:      minorList,
				patchList:      patchList,
				versionListErr: verListErr,
			},
		}
	}

	testCases := map[string]struct {
		verCollector versionCollector
		wantErr      bool
	}{
		"works": {
			verCollector: verCollector(minorList(), patchList(), nil),
		},
		"empty versions list": {
			verCollector: verCollector(emptyVerList(), emptyVerList(), nil),
		},
		"version list error": {
			verCollector: verCollector(minorList(), patchList(), someErr),
			wantErr:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			_, err := tc.verCollector.newCLIVersions(context.Background())
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

func TestFilterCompatibleCLIVersions(t *testing.T) {
	someErr := errors.New("some error")
	verCollector := func(cliInfoErr error) versionCollector {
		return versionCollector{
			cliVersion: "v0.1.0",
			versionsapi: stubVersionFetcher{
				cliInfoErr: cliInfoErr,
			},
		}
	}

	testCases := map[string]struct {
		verCollector     versionCollector
		cliPatchVersions []string
		wantErr          bool
	}{
		"works": {
			verCollector:     verCollector(nil),
			cliPatchVersions: []string{"v0.1.1"},
		},
		"cli info error": {
			verCollector:     verCollector(someErr),
			cliPatchVersions: []string{"v0.1.1"},
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			_, err := tc.verCollector.filterCompatibleCLIVersions(context.Background(), tc.cliPatchVersions, "v1.24.5")
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
		})
	}
}

type stubVersionFetcher struct {
	minorList      versionsapi.List
	patchList      versionsapi.List
	versionListErr error
	cliInfoErr     error
}

func (f stubVersionFetcher) FetchVersionList(_ context.Context, list versionsapi.List) (versionsapi.List, error) {
	switch list.Granularity {
	case versionsapi.GranularityMajor:
		return f.minorList, f.versionListErr
	case versionsapi.GranularityMinor:
		return f.patchList, f.versionListErr
	}
	return versionsapi.List{}, f.versionListErr
}

func (f stubVersionFetcher) FetchCLIInfo(_ context.Context, _ versionsapi.CLIInfo) (versionsapi.CLIInfo, error) {
	return versionsapi.CLIInfo{}, f.cliInfoErr
}
