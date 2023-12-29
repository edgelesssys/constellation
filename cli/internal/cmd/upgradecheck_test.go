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
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/api/versionsapi"
	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	consemver "github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				newServices: consemver.NewFromInt(2, 5, 0, ""),
				newImages: map[string]measurements.M{
					"v2.5.0": measurements.DefaultsFor(cloudprovider.QEMU, variant.QEMUVTPM{}),
				},
				newKubernetes:     []string{"v1.24.12", "v1.25.6"},
				newCLI:            []consemver.Semver{consemver.NewFromInt(2, 5, 0, ""), consemver.NewFromInt(2, 6, 0, "")},
				currentServices:   consemver.NewFromInt(2, 4, 0, ""),
				currentImage:      consemver.NewFromInt(2, 4, 0, ""),
				currentKubernetes: consemver.NewFromInt(1, 24, 5, ""),
				currentCLI:        consemver.NewFromInt(2, 4, 0, ""),
			},
			expected: "The following updates are available with this CLI:\n  Kubernetes: v1.24.5 --> v1.24.12 v1.25.6\n  Images:\n    v2.4.0 --> v2.5.0\n      Includes these measurements:\n      4:\n          expected: \"1234123412341234123412341234123412341234123412341234123412341234\"\n          warnOnly: false\n      8:\n          expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n          warnOnly: false\n      9:\n          expected: \"1234123412341234123412341234123412341234123412341234123412341234\"\n          warnOnly: false\n      11:\n          expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n          warnOnly: false\n      12:\n          expected: \"1234123412341234123412341234123412341234123412341234123412341234\"\n          warnOnly: false\n      13:\n          expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n          warnOnly: false\n      15:\n          expected: \"0000000000000000000000000000000000000000000000000000000000000000\"\n          warnOnly: false\n      \n  Services: v2.4.0 --> v2.5.0\n",
		},
		"cli incompatible with K8s": {
			upgrade: versionUpgrade{
				newCLI:     []consemver.Semver{consemver.NewFromInt(2, 5, 0, ""), consemver.NewFromInt(2, 6, 0, "")},
				currentCLI: consemver.NewFromInt(2, 4, 0, ""),
			},
			expected: "There are newer CLIs available (v2.5.0 v2.6.0), however, you need to upgrade your cluster's Kubernetes version first.\n",
		},
		"cli compatible with K8s": {
			upgrade: versionUpgrade{
				newCompatibleCLI: []consemver.Semver{consemver.NewFromInt(2, 5, 0, ""), consemver.NewFromInt(2, 6, 0, "")},
				currentCLI:       consemver.NewFromInt(2, 4, 0, ""),
			},
			expected: "Newer CLI versions that are compatible with your cluster are: v2.5.0 v2.6.0\n",
		},
		"k8s only": {
			upgrade: versionUpgrade{
				newKubernetes:     []string{"v1.24.12", "v1.25.6"},
				currentKubernetes: consemver.NewFromInt(1, 24, 5, ""),
			},
			expected: "The following updates are available with this CLI:\n  Kubernetes: v1.24.5 --> v1.24.12 v1.25.6\n",
		},
		"no upgrades": {
			upgrade: versionUpgrade{
				newServices:       consemver.Semver{},
				newImages:         map[string]measurements.M{},
				newKubernetes:     []string{},
				newCLI:            []consemver.Semver{},
				currentServices:   consemver.NewFromInt(2, 5, 0, ""),
				currentImage:      consemver.NewFromInt(2, 5, 0, ""),
				currentKubernetes: consemver.NewFromInt(1, 25, 6, ""),
				currentCLI:        consemver.NewFromInt(2, 5, 0, ""),
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

func TestGetCompatibleImageMeasurements(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	csp := cloudprovider.Azure
	attestationVariant := variant.AzureSEVSNP{}

	versionZero, err := versionsapi.NewVersion("-", "stable", "v0.0.0", versionsapi.VersionKindImage)
	require.NoError(err)

	client := newTestClient(func(req *http.Request) *http.Response {
		if strings.HasSuffix(req.URL.String(), "v0.0.0/image/measurements.json") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"version": "v0.0.0","ref": "-","stream": "stable","list": [{"csp": "Azure","attestationVariant": "azure-sev-snp","measurements": {"0": {"expected": "0000000000000000000000000000000000000000000000000000000000000000","warnOnly": false}}}]}`)),
				Header:     make(http.Header),
			}
		}
		if strings.HasSuffix(req.URL.String(), "v0.0.0/image/measurements.json.sig") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("MEQCIGRR7RaSMs892Ta06/Tz7LqPUxI05X4wQcP+nFFmZtmaAiBNl9X8mUKmUBfxg13LQBfmmpw6JwYQor5hOwM3NFVPAg==")),
				Header:     make(http.Header),
			}
		}

		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("Not found.")),
			Header:     make(http.Header),
		}
	})

  upgrades, err := getCompatibleImageMeasurements(context.Background(), &bytes.Buffer{}, client, &stubCosignVerifier{}, singleUUIDVerifier(), csp, attestationVariant, versionZero, slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)))
	assert.NoError(err)

	for _, measurement := range upgrades {
		assert.NotEmpty(measurement)
	}
}

func TestUpgradeCheck(t *testing.T) {
	require := require.New(t)
	v2_3, err := versionsapi.NewVersion("-", "stable", "v2.3.0", versionsapi.VersionKindImage)
	require.NoError(err)

	v2_5, err := versionsapi.NewVersion("-", "stable", "v2.5.0", versionsapi.VersionKindImage)
	require.NoError(err)

	collector := stubVersionCollector{
		supportedServicesVersions: consemver.NewFromInt(2, 5, 0, ""),
		supportedImages:           []versionsapi.Version{v2_3},
		supportedImageVersions: map[string]measurements.M{
			"v2.3.0": measurements.DefaultsFor(cloudprovider.GCP, variant.GCPSEVES{}),
		},
		supportedK8sVersions:    []string{"v1.24.5", "v1.24.12", "v1.25.6"},
		currentServicesVersions: consemver.NewFromInt(2, 4, 0, ""),
		currentImageVersion:     consemver.NewFromInt(2, 4, 0, ""),
		currentK8sVersion:       consemver.NewFromInt(1, 24, 5, ""),
		currentCLIVersion:       consemver.NewFromInt(2, 4, 0, ""),
		images:                  []versionsapi.Version{v2_5},
		newCLIVersionsList:      []consemver.Semver{consemver.NewFromInt(2, 5, 0, ""), consemver.NewFromInt(2, 6, 0, "")},
	}

	testCases := map[string]struct {
		collector  stubVersionCollector
		csp        cloudprovider.Provider
		checker    stubTerraformChecker
		cliVersion string
		wantError  bool
	}{
		"upgrades gcp": {
			collector:  collector,
			checker:    stubTerraformChecker{},
			csp:        cloudprovider.GCP,
			cliVersion: "v1.0.0",
		},
		"terraform plan err": {
			collector: collector,
			checker: stubTerraformChecker{
				planErr: assert.AnError,
			},
			csp:        cloudprovider.GCP,
			cliVersion: "v1.0.0",
			wantError:  true,
		},
		"terraform rollback err, log only": {
			collector: collector,
			checker: stubTerraformChecker{
				rollbackErr: assert.AnError,
			},
			csp:        cloudprovider.GCP,
			cliVersion: "v1.0.0",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			fileHandler := file.NewHandler(afero.NewMemMapFs())
			cfg := defaultConfigWithExpectedMeasurements(t, config.Default(), tc.csp)
			require.NoError(fileHandler.WriteYAML(constants.ConfigFilename, cfg))

			checkCmd := upgradeCheckCmd{
				canUpgradeCheck:  true,
				collect:          &tc.collector,
				terraformChecker: tc.checker,
				fileHandler:      fileHandler,
        log:              slog.New(slog.NewTextHandler(logger.TestWriter{T: t}, nil)),
			}

			cmd := newUpgradeCheckCmd()

			err := checkCmd.upgradeCheck(cmd, stubAttestationFetcher{})
			if tc.wantError {
				assert.Error(err)
				return
			}
			assert.NoError(err)
		})
	}
}

type stubVersionCollector struct {
	supportedServicesVersions    consemver.Semver
	supportedImages              []versionsapi.Version
	supportedImageVersions       map[string]measurements.M
	supportedK8sVersions         []string
	supportedCLIVersions         []consemver.Semver
	currentServicesVersions      consemver.Semver
	currentImageVersion          consemver.Semver
	currentK8sVersion            consemver.Semver
	currentCLIVersion            consemver.Semver
	images                       []versionsapi.Version
	newCLIVersionsList           []consemver.Semver
	newCompatibleCLIVersionsList []consemver.Semver
	someErr                      error
}

func (s *stubVersionCollector) newMeasurements(_ context.Context, _ cloudprovider.Provider, _ variant.Variant, _ []versionsapi.Version) (map[string]measurements.M, error) {
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

func (s *stubVersionCollector) supportedVersions(_ context.Context, _, _ consemver.Semver) (supportedVersionInfo, error) {
	return supportedVersionInfo{
		service: s.supportedServicesVersions,
		image:   s.supportedImages,
		k8s:     s.supportedK8sVersions,
		cli:     s.supportedCLIVersions,
	}, s.someErr
}

func (s *stubVersionCollector) newImages(_ context.Context, _ consemver.Semver) ([]versionsapi.Version, error) {
	return s.images, nil
}

func (s *stubVersionCollector) newerVersions(_ context.Context, _ []string) ([]versionsapi.Version, error) {
	return s.images, nil
}

func (s *stubVersionCollector) newCLIVersions(_ context.Context) ([]consemver.Semver, error) {
	return s.newCLIVersionsList, nil
}

func (s *stubVersionCollector) filterCompatibleCLIVersions(_ context.Context, _ []consemver.Semver, _ consemver.Semver) ([]consemver.Semver, error) {
	return s.newCompatibleCLIVersionsList, nil
}

type stubTerraformChecker struct {
	tfDiff      bool
	planErr     error
	rollbackErr error
}

func (s stubTerraformChecker) Plan(_ context.Context, _ *config.Config) (bool, error) {
	return s.tfDiff, s.planErr
}

func (s stubTerraformChecker) RestoreWorkspace() error {
	return s.rollbackErr
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
			cliVersion: consemver.NewFromInt(0, 1, 0, ""),
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
			cliVersion: consemver.NewFromInt(0, 1, 0, ""),
			versionsapi: stubVersionFetcher{
				cliInfoErr: cliInfoErr,
			},
		}
	}

	testCases := map[string]struct {
		verCollector     versionCollector
		cliPatchVersions []consemver.Semver
		wantErr          bool
	}{
		"works": {
			verCollector:     verCollector(nil),
			cliPatchVersions: []consemver.Semver{consemver.NewFromInt(0, 1, 1, "")},
		},
		"cli info error": {
			verCollector:     verCollector(someErr),
			cliPatchVersions: []consemver.Semver{consemver.NewFromInt(0, 1, 1, "")},
			wantErr:          true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			_, err := tc.verCollector.filterCompatibleCLIVersions(context.Background(), tc.cliPatchVersions, consemver.NewFromInt(1, 24, 5, ""))
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
