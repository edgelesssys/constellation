/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"errors"
	"testing"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"helm.sh/helm/v3/pkg/action"
)

func TestMergeMaps(t *testing.T) {
	testCases := map[string]struct {
		vals      map[string]any
		extraVals map[string]any
		expected  map[string]any
	}{
		"equal": {
			vals: map[string]any{
				"join-service": map[string]any{
					"key1": "foo",
					"key2": "bar",
				},
			},
			extraVals: map[string]any{
				"join-service": map[string]any{
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
			expected: map[string]any{
				"join-service": map[string]any{
					"key1":      "foo",
					"key2":      "bar",
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
		},
		"missing join-service extraVals": {
			vals: map[string]any{
				"join-service": map[string]any{
					"key1": "foo",
					"key2": "bar",
				},
			},
			extraVals: map[string]any{
				"extraKey1": "extraFoo",
				"extraKey2": "extraBar",
			},
			expected: map[string]any{
				"join-service": map[string]any{
					"key1": "foo",
					"key2": "bar",
				},
				"extraKey1": "extraFoo",
				"extraKey2": "extraBar",
			},
		},
		"missing join-service vals": {
			vals: map[string]any{
				"key1": "foo",
				"key2": "bar",
			},
			extraVals: map[string]any{
				"join-service": map[string]any{
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
			expected: map[string]any{
				"key1": "foo",
				"key2": "bar",
				"join-service": map[string]any{
					"extraKey1": "extraFoo",
					"extraKey2": "extraBar",
				},
			},
		},
		"key collision": {
			vals: map[string]any{
				"join-service": map[string]any{
					"key1": "foo",
				},
			},
			extraVals: map[string]any{
				"join-service": map[string]any{
					"key1": "bar",
				},
			},
			expected: map[string]any{
				"join-service": map[string]any{
					"key1": "bar",
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			newVals := mergeMaps(tc.vals, tc.extraVals)
			assert.Equal(tc.expected, newVals)
		})
	}
}

func TestHelmApply(t *testing.T) {
	cliVersion := semver.NewFromInt(1, 99, 0, "")
	csp := cloudprovider.AWS // using AWS since it has an additional chart: aws-load-balancer-controller
	attestationVariant := variant.AWSSEVSNP{}
	microserviceCharts := []string{
		"constellation-services",
		"constellation-operators",
		"constellation-csi",
	}
	testCases := map[string]struct {
		clusterMicroServiceVersion string
		expectedActions            []string
		expectUpgrade              bool
		clusterCertManagerVersion  *string
		clusterAWSLBVersion        *string
		allowDestructive           bool
		expectError                bool
	}{
		"CLI microservices are 1 minor version newer than cluster ones": {
			clusterMicroServiceVersion: "v1.98.1",
			expectedActions:            microserviceCharts,
			expectUpgrade:              true,
		},
		"CLI microservices are 2 minor versions newer than cluster ones": {
			clusterMicroServiceVersion: "v1.97.0",
			expectedActions:            []string{},
		},
		"cluster microservices are newer than CLI": {
			clusterMicroServiceVersion: "v1.100.0",
		},
		"cluster and CLI microservices have the same version": {
			clusterMicroServiceVersion: "v1.99.0",
			expectedActions:            []string{},
		},
		"cert-manager upgrade is ignored when denying destructive upgrades": {
			clusterMicroServiceVersion: "v1.99.0",
			clusterCertManagerVersion:  toPtr("v1.9.0"),
			allowDestructive:           false,
			expectError:                true,
		},
		"both microservices and cert-manager are upgraded in destructive mode": {
			clusterMicroServiceVersion: "v1.98.1",
			clusterCertManagerVersion:  toPtr("v1.9.0"),
			expectedActions:            append(microserviceCharts, "cert-manager"),
			expectUpgrade:              true,
			allowDestructive:           true,
		},
		"only missing aws-load-balancer-controller is installed": {
			clusterMicroServiceVersion: "v1.99.0",
			clusterAWSLBVersion:        toPtr(""),
			expectedActions:            []string{"aws-load-balancer-controller"},
		},
	}

	log := logger.NewTest(t)
	options := Options{
		DeployCSIDriver:  true,
		Conformance:      false,
		HelmWaitMode:     WaitModeWait,
		AllowDestructive: true,
		Force:            false,
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lister := &releaseVersionMock{}
			sut := Client{
				factory:    newActionFactory(nil, lister, &action.Configuration{}, log),
				log:        log,
				cliVersion: cliVersion,
			}
			awsLbVersion := "v1.5.4" // current version
			if tc.clusterAWSLBVersion != nil {
				awsLbVersion = *tc.clusterAWSLBVersion
			}

			certManagerVersion := "v1.15.0" // current version
			if tc.clusterCertManagerVersion != nil {
				certManagerVersion = *tc.clusterCertManagerVersion
			}
			helmListVersion(lister, "cilium", "v1.15.8-edg.0")
			helmListVersion(lister, "coredns", "v0.0.0")
			helmListVersion(lister, "cert-manager", certManagerVersion)
			helmListVersion(lister, "constellation-services", tc.clusterMicroServiceVersion)
			helmListVersion(lister, "constellation-operators", tc.clusterMicroServiceVersion)
			helmListVersion(lister, "constellation-csi", tc.clusterMicroServiceVersion)
			helmListVersion(lister, "aws-load-balancer-controller", awsLbVersion)

			options.AllowDestructive = tc.allowDestructive
			options.CSP = csp
			options.AttestationVariant = attestationVariant
			options.K8sVersion = versions.Default
			options.MicroserviceVersion = cliVersion

			ex, includesUpgrade, err := sut.PrepareApply(
				options,
				state.New().
					SetInfrastructure(state.Infrastructure{UID: "testuid"}).
					SetClusterValues(state.ClusterValues{MeasurementSalt: []byte{0x41}}),
				fakeServiceAccURI(csp),
				uri.MasterSecret{Key: []byte("secret"), Salt: []byte("masterSalt")})
			var upgradeErr *compatibility.InvalidUpgradeError
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.True(t, err == nil || errors.As(err, &upgradeErr))
			}
			assert.Equal(t, tc.expectUpgrade, includesUpgrade)
			chartExecutor, ok := ex.(*ChartApplyExecutor)
			assert.True(t, ok)
			assert.ElementsMatch(t, tc.expectedActions, getActionReleaseNames(chartExecutor.actions))
		})
	}
}

func getActionReleaseNames(actions []applyAction) []string {
	releaseActionNames := []string{}
	for _, action := range actions {
		releaseActionNames = append(releaseActionNames, action.ReleaseName())
	}
	return releaseActionNames
}

func helmListVersion(l *releaseVersionMock, releaseName string, installedVersion string) {
	if installedVersion == "" {
		l.On("currentVersion", releaseName).Return(semver.Semver{}, errReleaseNotFound)
		return
	}
	v, _ := semver.New(installedVersion)
	l.On("currentVersion", releaseName).Return(v, nil)
}

type releaseVersionMock struct {
	mock.Mock
}

func (s *releaseVersionMock) currentVersion(release string) (semver.Semver, error) {
	args := s.Called(release)
	return args.Get(0).(semver.Semver), args.Error(1)
}
