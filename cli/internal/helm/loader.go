/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"helm.sh/helm/pkg/ignore"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/edgelesssys/constellation/v2/cli/internal/helm/imageversion"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

// Run `go generate` to download (and patch) upstream helm charts.
//go:generate ./generateCilium.sh
//go:generate ./update-csi-charts.sh
//go:generate ./generateCertManager.sh

//go:embed all:charts/*
var helmFS embed.FS

type chartInfo struct {
	releaseName string
	chartName   string
	path        string
}

var (
	ciliumInfo                 = chartInfo{releaseName: "cilium", chartName: "cilium", path: "charts/cilium"}
	certManagerInfo            = chartInfo{releaseName: "cert-manager", chartName: "cert-manager", path: "charts/cert-manager"}
	constellationOperatorsInfo = chartInfo{releaseName: "constellation-operators", chartName: "constellation-operators", path: "charts/edgeless/operators"}
	constellationServicesInfo  = chartInfo{releaseName: "constellation-services", chartName: "constellation-services", path: "charts/edgeless/constellation-services"}
)

// ChartLoader loads embedded helm charts.
type ChartLoader struct {
	csp                          cloudprovider.Provider
	joinServiceImage             string
	keyServiceImage              string
	ccmImage                     string
	cnmImage                     string
	autoscalerImage              string
	verificationServiceImage     string
	gcpGuestAgentImage           string
	konnectivityImage            string
	constellationOperatorImage   string
	nodeMaintenanceOperatorImage string
}

// NewLoader creates a new ChartLoader.
func NewLoader(csp cloudprovider.Provider, k8sVersion versions.ValidK8sVersion) *ChartLoader {
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

	// TODO(malt3): Allow overriding container image registry + prefix for all images
	// (e.g. for air-gapped environments).
	return &ChartLoader{
		csp:                          csp,
		joinServiceImage:             imageversion.JoinService("", ""),
		keyServiceImage:              imageversion.KeyService("", ""),
		ccmImage:                     ccmImage,
		cnmImage:                     cnmImage,
		autoscalerImage:              versions.VersionConfigs[k8sVersion].ClusterAutoscalerImage,
		verificationServiceImage:     imageversion.VerificationService("", ""),
		gcpGuestAgentImage:           versions.GcpGuestImage,
		konnectivityImage:            versions.KonnectivityAgentImage,
		constellationOperatorImage:   imageversion.ConstellationNodeOperator("", ""),
		nodeMaintenanceOperatorImage: versions.NodeMaintenanceOperatorImage,
	}
}

// AvailableServiceVersions returns the chart version number of the bundled service versions.
func AvailableServiceVersions() (string, error) {
	servicesChart, err := loadChartsDir(helmFS, constellationServicesInfo.path)
	if err != nil {
		return "", fmt.Errorf("loading constellation-services chart: %w", err)
	}

	return compatibility.EnsurePrefixV(servicesChart.Metadata.Version), nil
}

// Load the embedded helm charts.
func (i *ChartLoader) Load(config *config.Config, conformanceMode bool, masterSecret, salt []byte) ([]byte, error) {
	ciliumRelease, err := i.loadRelease(ciliumInfo)
	if err != nil {
		return nil, fmt.Errorf("loading cilium: %w", err)
	}
	extendCiliumValues(ciliumRelease.Values, conformanceMode)

	certManagerRelease, err := i.loadRelease(certManagerInfo)
	if err != nil {
		return nil, fmt.Errorf("loading cert-manager: %w", err)
	}

	operatorRelease, err := i.loadRelease(constellationOperatorsInfo)
	if err != nil {
		return nil, fmt.Errorf("loading operators: %w", err)
	}

	conServicesRelease, err := i.loadRelease(constellationServicesInfo)
	if err != nil {
		return nil, fmt.Errorf("loading constellation-services: %w", err)
	}
	if err := extendConstellationServicesValues(conServicesRelease.Values, config, masterSecret, salt); err != nil {
		return nil, fmt.Errorf("extending constellation-services values: %w", err)
	}

	releases := helm.Releases{Cilium: ciliumRelease, CertManager: certManagerRelease, Operators: operatorRelease, ConstellationServices: conServicesRelease}

	rel, err := json.Marshal(releases)
	if err != nil {
		return nil, err
	}
	return rel, nil
}

// loadRelease loads the embedded chart and values depending on the given info argument.
func (i *ChartLoader) loadRelease(info chartInfo) (helm.Release, error) {
	chart, err := loadChartsDir(helmFS, info.path)
	if err != nil {
		return helm.Release{}, fmt.Errorf("loading %s chart: %w", info.releaseName, err)
	}

	var values map[string]any

	switch info.releaseName {
	case ciliumInfo.releaseName:
		values, err = i.loadCiliumValues()
	case certManagerInfo.releaseName:
		values = i.loadCertManagerValues()
	case constellationOperatorsInfo.releaseName:
		updateVersions(chart, compatibility.EnsurePrefixV(constants.VersionInfo()))

		values, err = i.loadOperatorsValues()
	case constellationServicesInfo.releaseName:
		updateVersions(chart, compatibility.EnsurePrefixV(constants.VersionInfo()))

		values, err = i.loadConstellationServicesValues()
	}

	if err != nil {
		return helm.Release{}, fmt.Errorf("loading %s values: %w", info.releaseName, err)
	}

	chartRaw, err := i.marshalChart(chart)
	if err != nil {
		return helm.Release{}, fmt.Errorf("packaging %s chart: %w", info.releaseName, err)
	}

	return helm.Release{Chart: chartRaw, Values: values, ReleaseName: info.releaseName, Wait: false}, nil
}

// loadCiliumValues is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *ChartLoader) loadCiliumValues() (map[string]any, error) {
	var values map[string]any
	switch i.csp {
	case cloudprovider.AWS:
		values = awsVals
	case cloudprovider.Azure:
		values = azureVals
	case cloudprovider.GCP:
		values = gcpVals
	case cloudprovider.OpenStack:
		values = openStackVals
	case cloudprovider.QEMU:
		values = qemuVals
	default:
		return nil, fmt.Errorf("unknown csp: %s", i.csp)
	}

	return values, nil
}

// extendCiliumValues extends the given values map by some values depending on user input.
// This extra step of separating the application of user input is necessary since service upgrades should
// reuse user input from the init step. However, we can't rely on reuse-values, because
// during upgrades we all values need to be set locally as they might have changed.
// Also, the charts are not rendered correctly without all of these values.
func extendCiliumValues(in map[string]any, conformanceMode bool) {
	if conformanceMode {
		in["kubeProxyReplacementHealthzBindAddr"] = ""
		in["kubeProxyReplacement"] = "partial"
		in["sessionAffinity"] = true
		in["cni"] = map[string]any{
			"chainingMode": "portmap",
		}
	}
}

// loadCertManagerHelper is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *ChartLoader) loadCertManagerValues() map[string]any {
	return map[string]any{
		"installCRDs": true,
		"prometheus": map[string]any{
			"enabled": false,
		},
		"tolerations": []map[string]any{
			{
				"key":      "node-role.kubernetes.io/control-plane",
				"effect":   "NoSchedule",
				"operator": "Exists",
			},
			{
				"key":      "node-role.kubernetes.io/master",
				"effect":   "NoSchedule",
				"operator": "Exists",
			},
		},
		"webhook": map[string]any{
			"tolerations": []map[string]any{
				{
					"key":      "node-role.kubernetes.io/control-plane",
					"effect":   "NoSchedule",
					"operator": "Exists",
				},
				{
					"key":      "node-role.kubernetes.io/master",
					"effect":   "NoSchedule",
					"operator": "Exists",
				},
			},
		},
		"cainjector": map[string]any{
			"tolerations": []map[string]any{
				{
					"key":      "node-role.kubernetes.io/control-plane",
					"effect":   "NoSchedule",
					"operator": "Exists",
				},
				{
					"key":      "node-role.kubernetes.io/master",
					"effect":   "NoSchedule",
					"operator": "Exists",
				},
			},
		},
		"startupapicheck": map[string]any{
			"timeout": "5m",
			"extraArgs": []string{
				"--verbose",
			},
			"tolerations": []map[string]any{
				{
					"key":      "node-role.kubernetes.io/control-plane",
					"effect":   "NoSchedule",
					"operator": "Exists",
				},
				{
					"key":      "node-role.kubernetes.io/master",
					"effect":   "NoSchedule",
					"operator": "Exists",
				},
			},
		},
	}
}

// loadOperatorsHelper is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *ChartLoader) loadOperatorsValues() (map[string]any, error) {
	values := map[string]any{
		"constellation-operator": map[string]any{
			"controllerManager": map[string]any{
				"manager": map[string]any{
					"image": i.constellationOperatorImage,
				},
			},
		},
		"node-maintenance-operator": map[string]any{
			"controllerManager": map[string]any{
				"manager": map[string]any{
					"image": i.nodeMaintenanceOperatorImage,
				},
			},
		},
	}
	switch i.csp {
	case cloudprovider.AWS:
		conOpVals, ok := values["constellation-operator"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid constellation-operator values")
		}
		conOpVals["csp"] = "AWS"

		values["tags"] = map[string]any{
			"AWS": true,
		}
	case cloudprovider.Azure:
		conOpVals, ok := values["constellation-operator"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid constellation-operator values")
		}
		conOpVals["csp"] = "Azure"

		values["tags"] = map[string]any{
			"Azure": true,
		}
	case cloudprovider.GCP:
		conOpVals, ok := values["constellation-operator"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid constellation-operator values")
		}
		conOpVals["csp"] = "GCP"

		values["tags"] = map[string]any{
			"GCP": true,
		}
	case cloudprovider.OpenStack:
		conOpVals, ok := values["constellation-operator"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid constellation-operator values")
		}
		conOpVals["csp"] = "OpenStack"

		values["tags"] = map[string]any{
			"OpenStack": true,
		}
	case cloudprovider.QEMU:
		conOpVals, ok := values["constellation-operator"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid constellation-operator values")
		}
		conOpVals["csp"] = "QEMU"

		values["tags"] = map[string]any{
			"QEMU": true,
		}
	}

	return values, nil
}

// loadConstellationServicesHelper is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *ChartLoader) loadConstellationServicesValues() (map[string]any, error) {
	values := map[string]any{
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
			"csp": i.csp.String(),
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
	}

	switch i.csp {
	case cloudprovider.AWS:
		ccmVals, ok := values["ccm"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid ccm values")
		}
		ccmVals["AWS"] = map[string]any{
			"image": i.ccmImage,
		}

		values["tags"] = map[string]any{
			"AWS": true,
		}
	case cloudprovider.Azure:
		ccmVals, ok := values["ccm"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid ccm values")
		}
		ccmVals["Azure"] = map[string]any{
			"image": i.ccmImage,
		}

		values["cnm"] = map[string]any{
			"image": i.cnmImage,
		}

		values["tags"] = map[string]any{
			"Azure": true,
		}

	case cloudprovider.GCP:
		ccmVals, ok := values["ccm"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid ccm values")
		}
		ccmVals["GCP"] = map[string]any{
			"image": i.ccmImage,
		}

		values["tags"] = map[string]any{
			"GCP": true,
		}
	case cloudprovider.OpenStack:
		ccmVals, ok := values["ccm"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid ccm values")
		}
		ccmVals["OpenStack"] = map[string]any{
			"image": i.ccmImage,
		}

		values["tags"] = map[string]any{
			"OpenStack": true,
		}
	case cloudprovider.QEMU:
		values["tags"] = map[string]any{
			"QEMU": true,
		}

	}
	return values, nil
}

// extendConstellationServicesValues extends the given values map by some values depending on user input.
// Values set inside this function are only applied during init, not during upgrade.
func extendConstellationServicesValues(
	in map[string]any, cfg *config.Config, masterSecret, salt []byte,
) error {
	keyServiceValues, ok := in["key-service"].(map[string]any)
	if !ok {
		return errors.New("missing 'key-service' key")
	}
	keyServiceValues["masterSecret"] = base64.StdEncoding.EncodeToString(masterSecret)
	keyServiceValues["salt"] = base64.StdEncoding.EncodeToString(salt)

	joinServiceVals, ok := in["join-service"].(map[string]any)
	if !ok {
		return errors.New("invalid join-service values")
	}
	joinServiceVals["attestationVariant"] = cfg.GetAttestationConfig().GetVariant().String()

	// attestation config is updated separately during upgrade,
	// so we only set them in Helm during init.
	attestationConfigJSON, err := json.Marshal(cfg.GetAttestationConfig())
	if err != nil {
		return fmt.Errorf("marshalling measurements: %w", err)
	}
	joinServiceVals["attestationConfig"] = string(attestationConfigJSON)

	verifyServiceVals, ok := in["verification-service"].(map[string]any)
	if !ok {
		return errors.New("invalid verification-service values")
	}
	verifyServiceVals["attestationVariant"] = cfg.GetAttestationConfig().GetVariant().String()

	csp := cfg.GetProvider()
	switch csp {
	case cloudprovider.Azure:
		in["azure"] = map[string]any{
			"deployCSIDriver": cfg.DeployCSIDriver(),
		}

	case cloudprovider.GCP:
		in["gcp"] = map[string]any{
			"deployCSIDriver": cfg.DeployCSIDriver(),
		}

	case cloudprovider.OpenStack:
		in["openstack"] = map[string]any{
			"deployYawolLoadBalancer": cfg.DeployYawolLoadBalancer(),
			"deployCSIDriver":         cfg.DeployCSIDriver(),
		}
		if cfg.DeployYawolLoadBalancer() {
			in["yawol-controller"] = map[string]any{
				"yawolOSSecretName": "yawolkey",
				// has to be larger than ~30s to account for slow OpenStack API calls.
				"openstackTimeout": "1m",
				"yawolFloatingID":  cfg.Provider.OpenStack.FloatingIPPoolID,
				"yawolFlavorID":    cfg.Provider.OpenStack.YawolFlavorID,
				"yawolImageID":     cfg.Provider.OpenStack.YawolImageID,
			}
		}
	}

	return nil
}

// updateVersions changes all versions of direct dependencies that are set to "0.0.0" to newVersion.
func updateVersions(chart *chart.Chart, newVersion string) {
	chart.Metadata.Version = newVersion
	selectedDeps := chart.Metadata.Dependencies
	for i := range selectedDeps {
		if selectedDeps[i].Version == "0.0.0" {
			selectedDeps[i].Version = newVersion
		}
	}

	deps := chart.Dependencies()
	for i := range deps {
		if deps[i].Metadata.Version == "0.0.0" {
			deps[i].Metadata.Version = newVersion
		}
	}
}

// marshalChart takes a Chart object, packages it to a temporary file and returns the content of that file.
// We currently need to take this approach of marshaling as dependencies are not marshaled correctly with json.Marshal.
// This stems from the fact that chart.Chart does not export the dependencies property.
func (i *ChartLoader) marshalChart(chart *chart.Chart) ([]byte, error) {
	// A separate tmpdir path is necessary since during unit testing multiple go routines are accessing the same path, possibly deleting files for other routines.
	tmpDirPath, err := os.MkdirTemp("", "*")
	defer os.Remove(tmpDirPath)
	if err != nil {
		return nil, fmt.Errorf("creating tmp dir: %w", err)
	}

	path, err := chartutil.Save(chart, tmpDirPath)
	defer os.Remove(path)
	if err != nil {
		return nil, fmt.Errorf("chartutil save: %w", err)
	}
	chartRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading packaged chart: %w", err)
	}
	return chartRaw, nil
}

// taken from loader.LoadDir from the helm go module
// loadChartsDir loads from a directory.
//
// This loads charts only from directories.
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
