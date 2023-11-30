/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"errors"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/semver"
	"helm.sh/helm/v3/pkg/action"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/util/retry"
)

// ReleaseVersionClient is a client that can retrieve the version of a helm release.
type ReleaseVersionClient struct {
	config *action.Configuration
}

// NewReleaseVersionClient creates a new ReleaseVersionClient.
func NewReleaseVersionClient(kubeConfig []byte, log debugLog) (*ReleaseVersionClient, error) {
	config, err := newActionConfig(kubeConfig, log)
	if err != nil {
		return nil, err
	}
	return &ReleaseVersionClient{
		config: config,
	}, nil
}

// listAction execute a List action by wrapping helm's action package.
// It creates the action, runs it at returns results and errors.
func (c ReleaseVersionClient) listAction(release string) (res []*helmrelease.Release, err error) {
	action := action.NewList(c.config)
	action.Filter = release
	// during init, the kube API might not yet be reachable, so we retry
	err = retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return err != nil
	}, func() error {
		res, err = action.Run()
		return err
	})
	return
}

// Versions queries the cluster for running versions and returns a map of releaseName -> version.
func (c ReleaseVersionClient) Versions() (ServiceVersions, error) {
	ciliumVersion, err := c.currentVersion(ciliumInfo.releaseName)
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", ciliumInfo.releaseName, err)
	}
	certManagerVersion, err := c.currentVersion(certManagerInfo.releaseName)
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", certManagerInfo.releaseName, err)
	}
	operatorsVersion, err := c.currentVersion(constellationOperatorsInfo.releaseName)
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", constellationOperatorsInfo.releaseName, err)
	}
	servicesVersion, err := c.currentVersion(constellationServicesInfo.releaseName)
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", constellationServicesInfo.releaseName, err)
	}
	csiVersions, err := c.csiVersions()
	if err != nil {
		return ServiceVersions{}, fmt.Errorf("getting CSI versions: %w", err)
	}

	serviceVersions := ServiceVersions{
		cilium:                 ciliumVersion,
		certManager:            certManagerVersion,
		constellationOperators: operatorsVersion,
		constellationServices:  servicesVersion,
		csiVersions:            csiVersions,
	}

	if awsLBVersion, err := c.currentVersion(awsLBControllerInfo.releaseName); err == nil {
		serviceVersions.awsLBController = awsLBVersion
	} else if !errors.Is(err, errReleaseNotFound) {
		return ServiceVersions{}, fmt.Errorf("getting %s version: %w", awsLBControllerInfo.releaseName, err)
	}

	return serviceVersions, nil
}

// currentVersion returns the version of the currently installed helm release.
// If the CSI chart is not installed, no error is returned because the user can configure if the chart should be installed.
func (c ReleaseVersionClient) currentVersion(release string) (semver.Semver, error) {
	rel, err := c.listAction(release)
	if err != nil {
		return semver.Semver{}, err
	}

	if len(rel) == 0 {
		return semver.Semver{}, errReleaseNotFound
	}
	if len(rel) > 1 {
		return semver.Semver{}, fmt.Errorf("multiple releases found for %s", release)
	}

	if rel[0] == nil || rel[0].Chart == nil || rel[0].Chart.Metadata == nil {
		return semver.Semver{}, fmt.Errorf("received invalid release %s", release)
	}
	return semver.New(rel[0].Chart.Metadata.Version)
}

// csi versions needs special handling because all versions of its subcharts should be gathered.
func (c ReleaseVersionClient) csiVersions() (map[string]semver.Semver, error) {
	packedChartRelease, err := c.listAction(csiInfo.releaseName)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", csiInfo.releaseName, err)
	}

	csiVersions := make(map[string]semver.Semver)

	// No CSI driver installed
	if len(packedChartRelease) == 0 {
		return csiVersions, nil
	}

	if len(packedChartRelease) > 1 {
		return nil, fmt.Errorf("multiple releases found for %s", csiInfo.releaseName)
	}

	if packedChartRelease[0] == nil || packedChartRelease[0].Chart == nil {
		return nil, fmt.Errorf("received invalid release %s", csiInfo.releaseName)
	}

	dependencies := packedChartRelease[0].Chart.Metadata.Dependencies
	for _, dep := range dependencies {
		var err error
		csiVersions[dep.Name], err = semver.New(dep.Version)
		if err != nil {
			return nil, fmt.Errorf("parsing CSI version %q: %w", dep.Name, err)
		}
	}
	return csiVersions, nil
}
