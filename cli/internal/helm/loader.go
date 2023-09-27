/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/pkg/ignore"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm/imageversion"
	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

// Run `go generate` to download (and patch) upstream helm charts.
//go:generate ./generateCilium.sh
//go:generate ./update-csi-charts.sh
//go:generate ./generateCertManager.sh
//go:generate ./update-aws-load-balancer-chart.sh

//go:embed all:charts/*
var helmFS embed.FS

type chartInfo struct {
	releaseName string
	chartName   string
	path        string
}

var (
	// Charts we fetch from an upstream with real versions.
	ciliumInfo          = chartInfo{releaseName: "cilium", chartName: "cilium", path: "charts/cilium"}
	certManagerInfo     = chartInfo{releaseName: "cert-manager", chartName: "cert-manager", path: "charts/cert-manager"}
	awsLBControllerInfo = chartInfo{releaseName: "aws-load-balancer-controller", chartName: "aws-load-balancer-controller", path: "charts/aws-load-balancer-controller"}

	// Bundled charts with embedded with version 0.0.0.
	constellationOperatorsInfo = chartInfo{releaseName: "constellation-operators", chartName: "constellation-operators", path: "charts/edgeless/operators"}
	constellationServicesInfo  = chartInfo{releaseName: "constellation-services", chartName: "constellation-services", path: "charts/edgeless/constellation-services"}
	csiInfo                    = chartInfo{releaseName: "constellation-csi", chartName: "constellation-csi", path: "charts/edgeless/csi"}
)

// chartLoader loads embedded helm charts.
type chartLoader struct {
	csp                          cloudprovider.Provider
	config                       *config.Config
	joinServiceImage             string
	keyServiceImage              string
	ccmImage                     string // cloud controller manager image
	azureCNMImage                string // Azure cloud node manager image
	autoscalerImage              string
	verificationServiceImage     string
	gcpGuestAgentImage           string
	konnectivityImage            string
	constellationOperatorImage   string
	nodeMaintenanceOperatorImage string
	clusterName                  string
	stateFile                    *state.State
	cliVersion                   semver.Semver
}

// newLoader creates a new ChartLoader.
func newLoader(config *config.Config, stateFile *state.State, cliVersion semver.Semver) *chartLoader {
	// TODO(malt3): Allow overriding container image registry + prefix for all images
	// (e.g. for air-gapped environments).
	var ccmImage, cnmImage string
	csp := config.GetProvider()
	k8sVersion := config.KubernetesVersion
	switch csp {
	case cloudprovider.AWS:
		ccmImage = versions.VersionConfigs[k8sVersion].CloudControllerManagerImageAWS
	case cloudprovider.Azure:
		ccmImage = versions.VersionConfigs[k8sVersion].CloudControllerManagerImageAzure
		cnmImage = versions.VersionConfigs[k8sVersion].CloudNodeManagerImageAzure
	case cloudprovider.GCP:
		ccmImage = versions.VersionConfigs[k8sVersion].CloudControllerManagerImageGCP
	case cloudprovider.OpenStack:
		ccmImage = versions.VersionConfigs[k8sVersion].CloudControllerManagerImageOpenStack
	}
	return &chartLoader{
		cliVersion:                   cliVersion,
		csp:                          csp,
		stateFile:                    stateFile,
		ccmImage:                     ccmImage,
		azureCNMImage:                cnmImage,
		config:                       config,
		joinServiceImage:             imageversion.JoinService("", ""),
		keyServiceImage:              imageversion.KeyService("", ""),
		autoscalerImage:              versions.VersionConfigs[k8sVersion].ClusterAutoscalerImage,
		verificationServiceImage:     imageversion.VerificationService("", ""),
		gcpGuestAgentImage:           versions.GcpGuestImage,
		konnectivityImage:            versions.KonnectivityAgentImage,
		constellationOperatorImage:   imageversion.ConstellationNodeOperator("", ""),
		nodeMaintenanceOperatorImage: versions.NodeMaintenanceOperatorImage,
	}
}

// releaseApplyOrder is a list of releases in the order they should be applied.
// makes sure if a release was removed as a dependency from one chart,
// and then added as a new standalone chart (or as a dependency of another chart),
// that the new release is installed after the existing one to avoid name conflicts.
type releaseApplyOrder []Release

// loadReleases loads the embedded helm charts and returns them as a HelmReleases object.
func (i *chartLoader) loadReleases(conformanceMode bool, helmWaitMode WaitMode, masterSecret uri.MasterSecret,
	serviceAccURI string,
) (releaseApplyOrder, error) {
	ciliumRelease, err := i.loadRelease(ciliumInfo, helmWaitMode)
	if err != nil {
		return nil, fmt.Errorf("loading cilium: %w", err)
	}
	ciliumVals := extraCiliumValues(i.config.GetProvider(), conformanceMode, i.stateFile.Infrastructure)
	ciliumRelease.Values = mergeMaps(ciliumRelease.Values, ciliumVals)

	certManagerRelease, err := i.loadRelease(certManagerInfo, helmWaitMode)
	if err != nil {
		return nil, fmt.Errorf("loading cert-manager: %w", err)
	}

	operatorRelease, err := i.loadRelease(constellationOperatorsInfo, helmWaitMode)
	if err != nil {
		return nil, fmt.Errorf("loading operators: %w", err)
	}
	operatorRelease.Values = mergeMaps(operatorRelease.Values, extraOperatorValues(i.stateFile.Infrastructure.UID))

	conServicesRelease, err := i.loadRelease(constellationServicesInfo, helmWaitMode)
	if err != nil {
		return nil, fmt.Errorf("loading constellation-services: %w", err)
	}

	svcVals, err := extraConstellationServicesValues(i.config, masterSecret, serviceAccURI, i.stateFile.Infrastructure)
	if err != nil {
		return nil, fmt.Errorf("extending constellation-services values: %w", err)
	}
	conServicesRelease.Values = mergeMaps(conServicesRelease.Values, svcVals)

	releases := releaseApplyOrder{ciliumRelease, conServicesRelease, certManagerRelease}
	if i.config.DeployCSIDriver() {
		csiRelease, err := i.loadRelease(csiInfo, helmWaitMode)
		if err != nil {
			return nil, fmt.Errorf("loading snapshot CRDs: %w", err)
		}
		extraCSIvals, err := extraCSIValues(i.config.GetProvider(), serviceAccURI)
		if err != nil {
			return nil, fmt.Errorf("extending CSI values: %w", err)
		}
		csiRelease.Values = mergeMaps(csiRelease.Values, extraCSIvals)
		releases = append(releases, csiRelease)
	}
	if i.config.HasProvider(cloudprovider.AWS) {
		awsRelease, err := i.loadRelease(awsLBControllerInfo, helmWaitMode)
		if err != nil {
			return nil, fmt.Errorf("loading aws-services: %w", err)
		}
		releases = append(releases, awsRelease)
	}
	releases = append(releases, operatorRelease)

	return releases, nil
}

// loadRelease loads the embedded chart and values depending on the given info argument.
// IMPORTANT: .helmignore rules specifying files in subdirectories are not applied (e.g. crds/kustomization.yaml).
func (i *chartLoader) loadRelease(info chartInfo, helmWaitMode WaitMode) (Release, error) {
	chart, err := loadChartsDir(helmFS, info.path)
	if err != nil {
		return Release{}, fmt.Errorf("loading %s chart: %w", info.releaseName, err)
	}

	var values map[string]any

	switch info.releaseName {
	case ciliumInfo.releaseName:
		var ok bool
		values, ok = ciliumVals[i.csp.String()]
		if !ok {
			return Release{}, fmt.Errorf("cilium values for csp %q not found", i.csp.String())
		}
	case certManagerInfo.releaseName:
		values = i.loadCertManagerValues()
	case constellationOperatorsInfo.releaseName:
		values = i.loadOperatorsValues()
	case constellationServicesInfo.releaseName:
		values = i.loadConstellationServicesValues()
	case awsLBControllerInfo.releaseName:
		values = i.loadAWSLBControllerValues()
	case csiInfo.releaseName:
		values = i.loadCSIValues()
	}

	// Charts we package ourselves have version 0.0.0.
	// Before use, we need to update the version of the chart to the current CLI version.
	if isCLIVersionedRelease(info.releaseName) {
		updateVersions(chart, i.cliVersion)
	}

	return Release{Chart: chart, Values: values, ReleaseName: info.releaseName, WaitMode: helmWaitMode}, nil
}

func (i *chartLoader) loadAWSLBControllerValues() map[string]any {
	return map[string]any{
		"clusterName":  i.stateFile.ClusterName(i.config),
		"tolerations":  controlPlaneTolerations,
		"nodeSelector": controlPlaneNodeSelector,
	}
}

// loadCertManagerHelper is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *chartLoader) loadCertManagerValues() map[string]any {
	return map[string]any{
		"installCRDs": true,
		"prometheus": map[string]any{
			"enabled": false,
		},
		"tolerations": controlPlaneTolerations,
		"webhook": map[string]any{
			"tolerations": controlPlaneTolerations,
		},
		"cainjector": map[string]any{
			"tolerations": controlPlaneTolerations,
		},
		"startupapicheck": map[string]any{
			"timeout": "5m",
			"extraArgs": []string{
				"--verbose",
			},
			"tolerations": controlPlaneTolerations,
		},
	}
}

// loadOperatorsHelper is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *chartLoader) loadOperatorsValues() map[string]any {
	return map[string]any{
		"constellation-operator": map[string]any{
			"controllerManager": map[string]any{
				"manager": map[string]any{
					"image": i.constellationOperatorImage,
				},
			},
			"csp": i.csp.String(),
		},
		"node-maintenance-operator": map[string]any{
			"controllerManager": map[string]any{
				"manager": map[string]any{
					"image": i.nodeMaintenanceOperatorImage,
				},
			},
		},
		"tags": i.cspTags(),
	}
}

// loadConstellationServicesHelper is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *chartLoader) loadConstellationServicesValues() map[string]any {
	return map[string]any{
		"global": map[string]any{
			"keyServicePort":      constants.KeyServicePort,
			"keyServiceNamespace": "", // empty namespace means we use the release namespace
			"serviceBasePath":     constants.ServiceBasePath,
			"joinConfigCMName":    constants.JoinConfigMap,
			"internalCMName":      constants.InternalConfigMap,
		},
		"key-service": map[string]any{
			"image":               i.keyServiceImage,
			"saltKeyName":         constants.ConstellationSaltKey,
			"masterSecretKeyName": constants.ConstellationMasterSecretKey,
			"masterSecretName":    constants.ConstellationMasterSecretStoreName,
		},
		"join-service": map[string]any{
			"csp":   i.csp.String(),
			"image": i.joinServiceImage,
		},
		"ccm": map[string]any{
			"csp":   i.csp.String(),
			"image": i.ccmImage,
		},
		"cnm": map[string]any{
			"image": i.azureCNMImage,
		},
		"autoscaler": map[string]any{
			"csp":   i.csp.String(),
			"image": i.autoscalerImage,
		},
		"verification-service": map[string]any{
			"image": i.verificationServiceImage,
		},
		"gcp-guest-agent": map[string]any{
			"image": i.gcpGuestAgentImage,
		},
		"konnectivity": map[string]any{
			"image": i.konnectivityImage,
		},
		"tags": i.cspTags(),
	}
}

func (i *chartLoader) loadCSIValues() map[string]any {
	return map[string]any{
		"tags": i.cspTags(),
	}
}

func (i *chartLoader) cspTags() map[string]any {
	return map[string]any{
		i.csp.String(): true,
	}
}

// updateVersions changes all versions of direct dependencies that are set to "0.0.0" to newVersion.
func updateVersions(chart *chart.Chart, newVersion semver.Semver) {
	chart.Metadata.Version = newVersion.String()
	selectedDeps := chart.Metadata.Dependencies
	for i := range selectedDeps {
		if selectedDeps[i].Version == "0.0.0" {
			selectedDeps[i].Version = newVersion.String()
		}
	}

	deps := chart.Dependencies()
	for i := range deps {
		if deps[i].Metadata.Version == "0.0.0" {
			deps[i].Metadata.Version = newVersion.String()
		}
	}
}

// taken from loader.LoadDir from the helm go module
// loadChartsDir loads from a directory.
//
// This loads charts only from directories.
// IMPORTANT: .helmignore rules specifying files in subdirectories are not applied (e.g. crds/kustomization.yaml).
func loadChartsDir(efs embed.FS, dir string) (*chart.Chart, error) {
	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	// Just used for errors.
	c := &chart.Chart{}

	rules := ignore.Empty()
	ifile, err := efs.ReadFile(filepath.Join(dir, ignore.HelmIgnore))
	if err == nil {
		r, err := ignore.Parse(bytes.NewReader(ifile))
		if err != nil {
			return c, err
		}
		rules = r
	}
	rules.AddDefaults()

	files := []*loader.BufferedFile{}

	walk := func(path string, d fs.DirEntry, err error) error {
		n := strings.TrimPrefix(path, dir)
		if n == "" {
			// No need to process top level. Avoid bug with helmignore .* matching
			// empty names. See issue https://github.com/kubernetes/helm/issues/1776.
			return nil
		}

		// Normalize to / since it will also work on Windows
		n = filepath.ToSlash(n)

		// Check input err
		if err != nil {
			return err
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Directory-based ignore rules should involve skipping the entire
			// contents of that directory.
			if rules.Ignore(n, fi) {
				return filepath.SkipDir
			}
			return nil
		}

		// If a .helmignore file matches, skip this file.
		if rules.Ignore(n, fi) {
			return nil
		}

		// Irregular files include devices, sockets, and other uses of files that
		// are not regular files. In Go they have a file mode type bit set.
		// See https://golang.org/pkg/os/#FileMode for examples.
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("cannot load irregular file %s as it has file mode type bits set", path)
		}

		data, err := efs.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "error reading %s", n)
		}

		data = bytes.TrimPrefix(data, utf8bom)
		n = strings.TrimPrefix(n, "/")

		files = append(files, &loader.BufferedFile{Name: n, Data: data})
		return nil
	}

	if err := fs.WalkDir(efs, dir, walk); err != nil {
		return c, err
	}

	return loader.LoadFiles(files)
}
