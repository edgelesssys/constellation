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
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/ignore"

	"github.com/edgelesssys/constellation/v2/internal/attestation/variant"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/constellation/helm/imageversion"
	"github.com/edgelesssys/constellation/v2/internal/constellation/state"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/semver"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

// Run `go generate` to download (and patch) upstream helm charts.
//go:generate ./generateCilium.sh
//go:generate go run ./corednsgen/
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
	coreDNSInfo         = chartInfo{releaseName: "coredns", chartName: "coredns", path: "charts/coredns"}
	ciliumInfo          = chartInfo{releaseName: "cilium", chartName: "cilium", path: "charts/cilium"}
	certManagerInfo     = chartInfo{releaseName: "cert-manager", chartName: "cert-manager", path: "charts/cert-manager"}
	awsLBControllerInfo = chartInfo{releaseName: "aws-load-balancer-controller", chartName: "aws-load-balancer-controller", path: "charts/aws-load-balancer-controller"}

	// Bundled charts with embedded with version 0.0.0.
	constellationOperatorsInfo = chartInfo{releaseName: "constellation-operators", chartName: "constellation-operators", path: "charts/edgeless/operators"}
	constellationServicesInfo  = chartInfo{releaseName: "constellation-services", chartName: "constellation-services", path: "charts/edgeless/constellation-services"}
	csiInfo                    = chartInfo{releaseName: "constellation-csi", chartName: "constellation-csi", path: "charts/edgeless/csi"}
	yawolLBControllerInfo      = chartInfo{releaseName: "yawol", chartName: "yawol", path: "charts/yawol"}
)

// chartLoader loads embedded helm charts.
type chartLoader struct {
	csp                          cloudprovider.Provider
	attestationVariant           variant.Variant
	joinServiceImage             string
	keyServiceImage              string
	ccmImage                     string // cloud controller manager image
	azureCNMImage                string // Azure cloud node manager image
	autoscalerImage              string
	verificationServiceImage     string
	gcpGuestAgentImage           string
	constellationOperatorImage   string
	nodeMaintenanceOperatorImage string
	clusterName                  string
	stateFile                    *state.State
	cliVersion                   semver.Semver
}

// newLoader creates a new ChartLoader.
func newLoader(csp cloudprovider.Provider, attestationVariant variant.Variant, k8sVersion versions.ValidK8sVersion, stateFile *state.State, cliVersion semver.Semver) *chartLoader {
	// TODO(malt3): Allow overriding container image registry + prefix for all images
	// (e.g. for air-gapped environments).
	var ccmImage, cnmImage string
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
		attestationVariant:           attestationVariant,
		stateFile:                    stateFile,
		ccmImage:                     ccmImage,
		azureCNMImage:                cnmImage,
		joinServiceImage:             imageversion.JoinService("", ""),
		keyServiceImage:              imageversion.KeyService("", ""),
		autoscalerImage:              versions.VersionConfigs[k8sVersion].ClusterAutoscalerImage,
		verificationServiceImage:     imageversion.VerificationService("", ""),
		gcpGuestAgentImage:           versions.GcpGuestImage,
		constellationOperatorImage:   imageversion.ConstellationNodeOperator("", ""),
		nodeMaintenanceOperatorImage: versions.NodeMaintenanceOperatorImage,
	}
}

// releaseApplyOrder is a list of releases in the order they should be applied.
// makes sure if a release was removed as a dependency from one chart,
// and then added as a new standalone chart (or as a dependency of another chart),
// that the new release is installed after the existing one to avoid name conflicts.
type releaseApplyOrder []release

// OpenStackValues are helm values for OpenStack.
type OpenStackValues struct {
	DeployYawolLoadBalancer bool
	FloatingIPPoolID        string
	YawolFlavorID           string
	YawolImageID            string
}

// loadReleases loads the embedded helm charts and returns them as a HelmReleases object.
func (i *chartLoader) loadReleases(conformanceMode, deployCSIDriver bool, helmWaitMode WaitMode, masterSecret uri.MasterSecret,
	serviceAccURI string, openStackValues *OpenStackValues, serviceCIDR string,
) (releaseApplyOrder, error) {
	ciliumRelease, err := i.loadRelease(ciliumInfo, helmWaitMode)
	if err != nil {
		return nil, fmt.Errorf("loading cilium: %w", err)
	}
	ciliumVals := extraCiliumValues(i.csp, conformanceMode, i.stateFile.Infrastructure)
	ciliumRelease.values = mergeMaps(ciliumRelease.values, ciliumVals)

	coreDNSRelease, err := i.loadRelease(coreDNSInfo, helmWaitMode)
	if err != nil {
		return nil, fmt.Errorf("loading coredns: %w", err)
	}
	coreDNSVals, err := extraCoreDNSValues(serviceCIDR)
	if err != nil {
		return nil, fmt.Errorf("loading coredns values: %w", err)
	}
	coreDNSRelease.values = mergeMaps(coreDNSRelease.values, coreDNSVals)

	certManagerRelease, err := i.loadRelease(certManagerInfo, helmWaitMode)
	if err != nil {
		return nil, fmt.Errorf("loading cert-manager: %w", err)
	}

	operatorRelease, err := i.loadRelease(constellationOperatorsInfo, helmWaitMode)
	if err != nil {
		return nil, fmt.Errorf("loading operators: %w", err)
	}
	operatorRelease.values = mergeMaps(operatorRelease.values, extraOperatorValues(i.stateFile.Infrastructure.UID))

	conServicesRelease, err := i.loadRelease(constellationServicesInfo, helmWaitMode)
	if err != nil {
		return nil, fmt.Errorf("loading constellation-services: %w", err)
	}

	svcVals, err := extraConstellationServicesValues(i.csp, i.attestationVariant, masterSecret,
		serviceAccURI, i.stateFile.Infrastructure, openStackValues)
	if err != nil {
		return nil, fmt.Errorf("extending constellation-services values: %w", err)
	}
	conServicesRelease.values = mergeMaps(conServicesRelease.values, svcVals)

	releases := releaseApplyOrder{ciliumRelease, coreDNSRelease, conServicesRelease, certManagerRelease, operatorRelease}
	if deployCSIDriver {
		csiRelease, err := i.loadRelease(csiInfo, WaitModeNone)
		if err != nil {
			return nil, fmt.Errorf("loading snapshot CRDs: %w", err)
		}
		extraCSIvals, err := extraCSIValues(i.csp, serviceAccURI)
		if err != nil {
			return nil, fmt.Errorf("extending CSI values: %w", err)
		}
		csiRelease.values = mergeMaps(csiRelease.values, extraCSIvals)
		releases = append(releases, csiRelease)
	}
	if i.csp == cloudprovider.AWS {
		awsRelease, err := i.loadRelease(awsLBControllerInfo, WaitModeNone)
		if err != nil {
			return nil, fmt.Errorf("loading aws-services: %w", err)
		}
		releases = append(releases, awsRelease)
	}
	if i.csp == cloudprovider.OpenStack {
		if openStackValues == nil {
			return nil, errors.New("provider is OpenStack but OpenStack config is missing")
		}
		if openStackValues.DeployYawolLoadBalancer {
			yawolRelease, err := i.loadRelease(yawolLBControllerInfo, WaitModeNone)
			if err != nil {
				return nil, fmt.Errorf("loading yawol chart: %w", err)
			}

			yawolVals, err := extraYawolValues(serviceAccURI, i.stateFile.Infrastructure, openStackValues)
			if err != nil {
				return nil, fmt.Errorf("extending yawol chart values: %w", err)
			}
			yawolRelease.values = mergeMaps(yawolRelease.values, yawolVals)
			releases = append(releases, yawolRelease)
		}
	}

	return releases, nil
}

// loadRelease loads the embedded chart and values depending on the given info argument.
// IMPORTANT: .helmignore rules specifying files in subdirectories are not applied (e.g. crds/kustomization.yaml).
func (i *chartLoader) loadRelease(info chartInfo, helmWaitMode WaitMode) (release, error) {
	chart, err := loadChartsDir(helmFS, info.path)
	if err != nil {
		return release{}, fmt.Errorf("loading %s chart: %w", info.releaseName, err)
	}

	var values map[string]any

	switch info.releaseName {
	case ciliumInfo.releaseName:
		values, err = i.loadCiliumValues(i.csp)
		if err != nil {
			return release{}, fmt.Errorf("loading cilium values: %w", err)
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
	default:
		values = map[string]any{}
	}

	// Charts we package ourselves have version 0.0.0.
	// Before use, we need to update the version of the chart to the current CLI version.
	if isCLIVersionedRelease(info.releaseName) {
		updateVersions(chart, i.cliVersion)
	}

	return release{chart: chart, values: values, releaseName: info.releaseName, waitMode: helmWaitMode}, nil
}

func (i *chartLoader) loadAWSLBControllerValues() map[string]any {
	return map[string]any{
		"clusterName":  i.stateFile.Infrastructure.Name,
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
			"podDisruptionBudget": map[string]any{
				"enabled": true,
			},
			"replicaCount": 2,
		},
		"cainjector": map[string]any{
			"tolerations": controlPlaneTolerations,
			"podDisruptionBudget": map[string]any{
				"enabled": true,
			},
			"replicaCount": 2,
		},
		"startupapicheck": map[string]any{
			"timeout": "5m",
			"extraArgs": []string{
				"-v",
			},
			"tolerations": controlPlaneTolerations,
		},
		"podDisruptionBudget": map[string]any{
			"enabled": true,
		},
		"replicaCount": 2,
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

func (i *chartLoader) loadCiliumValues(cloudprovider.Provider) (map[string]any, error) {
	sharedConfig := map[string]any{
		"extraArgs": []string{"--node-encryption-opt-out-labels=invalid.label", "--bpf-filter-priority=128"},
		"endpointRoutes": map[string]any{
			"enabled": true,
		},
		"l7Proxy": false,
		"image": map[string]any{
			"repository": "ghcr.io/edgelesssys/cilium/cilium",
			"suffix":     "",
			"tag":        "v1.15.8-edg.0",
			"digest":     "sha256:67aedd821a732e9ba3e34d200c389122384b70c05ba9a5ffb6ad813a53f2d4db",
			"useDigest":  true,
		},
		"operator": map[string]any{
			"image": map[string]any{
				"repository": "ghcr.io/edgelesssys/cilium/operator",
				"suffix":     "",
				"tag":        "v1.15.8-edg.0",
				// Careful: this is the digest of ghcr.io/.../operator-generic!
				// See magic image manipulation in ./helm/charts/cilium/templates/cilium-operator/_helpers.tpl.
				"genericDigest": "sha256:dd41e2a65c607ac929d872f10b9d0c3eff88aafa99e7c062e9c240b14943dd2e",
				"useDigest":     true,
			},
			"podDisruptionBudget": map[string]any{
				"enabled": true,
			},
		},
		"encryption": map[string]any{
			"enabled":        true,
			"type":           "wireguard",
			"nodeEncryption": true,
			"strictMode": map[string]any{
				"enabled":                   true,
				"podCIDRList":               []string{"10.244.0.0/16"},
				"allowRemoteNodeIdentities": false,
			},
		},
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": []string{
					"10.244.0.0/16",
				},
			},
		},
		"bpf": map[string]any{
			"masquerade": true,
		},
		"ipMasqAgent": map[string]any{
			"enabled": true,
			"config": map[string]any{
				"masqLinkLocal": true,
			},
		},
		"kubeProxyReplacement":                "strict",
		"enableCiliumEndpointSlice":           true,
		"kubeProxyReplacementHealthzBindAddr": "0.0.0.0:10256",
		"cleanBpfState":                       true,
	}
	cspOverrideConfigs := map[string]map[string]any{
		cloudprovider.AWS.String():   {},
		cloudprovider.Azure.String(): {},
		cloudprovider.GCP.String(): {
			"routingMode": "native",
			"encryption": map[string]any{
				"strictMode": map[string]any{
					"podCIDRList": []string{""},
				},
			},
			"ipam": map[string]any{
				"mode": "kubernetes",
			},
		},
		cloudprovider.OpenStack.String(): {},
		cloudprovider.QEMU.String(): {
			"extraArgs": []string{""},
		},
	}

	cspValues, ok := cspOverrideConfigs[i.csp.String()]
	if !ok {
		return nil, fmt.Errorf("cilium values for csp %q not found", i.csp.String())
	}
	return mergeMaps(sharedConfig, cspValues), nil
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
