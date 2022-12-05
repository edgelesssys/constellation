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

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/pkg/errors"
	"helm.sh/helm/pkg/ignore"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Run `go generate` to download (and patch) upstream helm charts.
//go:generate ./generateCilium.sh
//go:generate ./update-csi-charts.sh
//go:generate ./generateCertManager.sh

//go:embed all:charts/*
var helmFS embed.FS

const (
	ciliumReleaseName       = "cilium"
	conServicesReleaseName  = "constellation-services"
	conOperatorsReleaseName = "constellation-operators"
	certManagerReleaseName  = "cert-manager"

	conServicesPath  = "charts/edgeless/constellation-services"
	conOperatorsPath = "charts/edgeless/operators"
	certManagerPath  = "charts/cert-manager"
	ciliumPath       = "charts/cilium"
)

// ChartLoader loads embedded helm charts.
type ChartLoader struct {
	joinServiceImage             string
	kmsImage                     string
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
	}

	return &ChartLoader{
		joinServiceImage:             versions.JoinImage,
		kmsImage:                     versions.KmsImage,
		ccmImage:                     ccmImage,
		cnmImage:                     cnmImage,
		autoscalerImage:              versions.VersionConfigs[k8sVersion].ClusterAutoscalerImage,
		verificationServiceImage:     versions.VerificationImage,
		gcpGuestAgentImage:           versions.GcpGuestImage,
		konnectivityImage:            versions.KonnectivityAgentImage,
		constellationOperatorImage:   versions.ConstellationOperatorImage,
		nodeMaintenanceOperatorImage: versions.NodeMaintenanceOperatorImage,
	}
}

// Load the embedded helm charts.
func (i *ChartLoader) Load(config *config.Config, conformanceMode bool, masterSecret, salt []byte) ([]byte, error) {
	ciliumRelease, err := i.loadCilium(config.GetProvider(), conformanceMode)
	if err != nil {
		return nil, fmt.Errorf("loading cilium: %w", err)
	}

	certManagerRelease, err := i.loadCertManager()
	if err != nil {
		return nil, fmt.Errorf("loading cilium: %w", err)
	}

	operatorRelease, err := i.loadOperators(config.GetProvider())
	if err != nil {
		return nil, fmt.Errorf("loading operators: %w", err)
	}

	conServicesRelease, err := i.loadConstellationServices(config, masterSecret, salt)
	if err != nil {
		return nil, fmt.Errorf("loading constellation-services: %w", err)
	}
	releases := helm.Releases{Cilium: ciliumRelease, CertManager: certManagerRelease, Operators: operatorRelease, ConstellationServices: conServicesRelease}

	rel, err := json.Marshal(releases)
	if err != nil {
		return nil, err
	}
	return rel, nil
}

func (i *ChartLoader) loadCilium(csp cloudprovider.Provider, conformanceMode bool) (helm.Release, error) {
	chart, err := loadChartsDir(helmFS, ciliumPath)
	if err != nil {
		return helm.Release{}, fmt.Errorf("loading cilium chart: %w", err)
	}
	values, err := i.loadCiliumValues(csp, conformanceMode)
	if err != nil {
		return helm.Release{}, err
	}

	chartRaw, err := i.marshalChart(chart)
	if err != nil {
		return helm.Release{}, fmt.Errorf("packaging cilium chart: %w", err)
	}

	return helm.Release{Chart: chartRaw, Values: values, ReleaseName: ciliumReleaseName, Wait: false}, nil
}

// loadCiliumHelper is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *ChartLoader) loadCiliumValues(csp cloudprovider.Provider, conformanceMode bool) (map[string]any, error) {
	var values map[string]any
	switch csp {
	case cloudprovider.AWS:
		values = awsVals
	case cloudprovider.Azure:
		values = azureVals
	case cloudprovider.GCP:
		values = gcpVals
	case cloudprovider.QEMU:
		values = qemuVals
	default:
		return nil, fmt.Errorf("unknown csp: %s", csp)
	}
	if conformanceMode {
		values["kubeProxyReplacementHealthzBindAddr"] = ""
		values["kubeProxyReplacement"] = "partial"
		values["sessionAffinity"] = true
		values["cni"] = map[string]any{
			"chainingMode": "portmap",
		}

	}
	return values, nil
}

func (i *ChartLoader) loadCertManager() (helm.Release, error) {
	chart, err := loadChartsDir(helmFS, certManagerPath)
	if err != nil {
		return helm.Release{}, fmt.Errorf("loading cert-manager chart: %w", err)
	}
	values := i.loadCertManagerValues()

	chartRaw, err := i.marshalChart(chart)
	if err != nil {
		return helm.Release{}, fmt.Errorf("packaging cert-manager chart: %w", err)
	}

	return helm.Release{Chart: chartRaw, Values: values, ReleaseName: certManagerReleaseName, Wait: false}, nil
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

func (i *ChartLoader) loadOperators(csp cloudprovider.Provider) (helm.Release, error) {
	chart, err := loadChartsDir(helmFS, conOperatorsPath)
	if err != nil {
		return helm.Release{}, fmt.Errorf("loading operators chart: %w", err)
	}
	values, err := i.loadOperatorsValues(csp)
	if err != nil {
		return helm.Release{}, err
	}

	chartRaw, err := i.marshalChart(chart)
	if err != nil {
		return helm.Release{}, fmt.Errorf("packaging operators chart: %w", err)
	}

	return helm.Release{Chart: chartRaw, Values: values, ReleaseName: conOperatorsReleaseName, Wait: false}, nil
}

// loadOperatorsHelper is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *ChartLoader) loadOperatorsValues(csp cloudprovider.Provider) (map[string]any, error) {
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
	switch csp {
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
	case cloudprovider.QEMU:
		conOpVals, ok := values["constellation-operator"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid constellation-operator values")
		}
		conOpVals["csp"] = "QEMU"

		values["tags"] = map[string]any{
			"QEMU": true,
		}
	case cloudprovider.AWS:
		conOpVals, ok := values["constellation-operator"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid constellation-operator values")
		}
		conOpVals["csp"] = "AWS"

		values["tags"] = map[string]any{
			"AWS": true,
		}
	}

	return values, nil
}

// loadConstellationServices loads the constellation-services chart from the embed.FS,
// marshals it into a helm-package .tgz and sets the values that can be set in the CLI.
func (i *ChartLoader) loadConstellationServices(config *config.Config, masterSecret, salt []byte) (helm.Release, error) {
	chart, err := loadChartsDir(helmFS, conServicesPath)
	if err != nil {
		return helm.Release{}, fmt.Errorf("loading constellation-services chart: %w", err)
	}
	values, err := i.loadConstellationServicesValues(config, masterSecret, salt)
	if err != nil {
		return helm.Release{}, err
	}

	chartRaw, err := i.marshalChart(chart)
	if err != nil {
		return helm.Release{}, fmt.Errorf("packaging constellation-services chart: %w", err)
	}

	return helm.Release{Chart: chartRaw, Values: values, ReleaseName: conServicesReleaseName, Wait: false}, nil
}

// loadConstellationServicesHelper is used to separate the marshalling step from the loading step.
// This reduces the time unit tests take to execute.
func (i *ChartLoader) loadConstellationServicesValues(config *config.Config, masterSecret, salt []byte) (map[string]any, error) {
	csp := config.GetProvider()
	values := map[string]any{
		"global": map[string]any{
			"kmsPort":          constants.KMSPort,
			"serviceBasePath":  constants.ServiceBasePath,
			"joinConfigCMName": constants.JoinConfigMap,
			"k8sVersionCMName": constants.K8sVersionConfigMapName,
			"internalCMName":   constants.InternalConfigMap,
		},
		"kms": map[string]any{
			"image":                i.kmsImage,
			"masterSecret":         base64.StdEncoding.EncodeToString(masterSecret),
			"salt":                 base64.StdEncoding.EncodeToString(salt),
			"saltKeyName":          constants.ConstellationSaltKey,
			"masterSecretKeyName":  constants.ConstellationMasterSecretKey,
			"masterSecretName":     constants.ConstellationMasterSecretStoreName,
			"measurementsFilename": constants.MeasurementsFilename,
		},
		"join-service": map[string]any{
			"csp":   csp.String(),
			"image": i.joinServiceImage,
		},
		"ccm": map[string]any{
			"csp": csp.String(),
		},
		"autoscaler": map[string]any{
			"csp":   csp.String(),
			"image": i.autoscalerImage,
		},
		"verification-service": map[string]any{
			"csp":   csp.String(),
			"image": i.verificationServiceImage,
		},
		"gcp-guest-agent": map[string]any{
			"image": i.gcpGuestAgentImage,
		},
		"konnectivity": map[string]any{
			"image": i.konnectivityImage,
		},
	}

	switch csp {
	case cloudprovider.Azure:
		joinServiceVals, ok := values["join-service"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid join-service values")
		}
		joinServiceVals["enforceIdKeyDigest"] = config.EnforcesIDKeyDigest()

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

		values["azure"] = map[string]any{
			"deployCSIDriver": config.DeployCSIDriver(),
		}

		values["azuredisk-csi-driver"] = map[string]any{
			"node": map[string]any{
				"kmsPort":      constants.KMSPort,
				"kmsNamespace": "", // empty namespace means we use the release namespace
			},
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

		values["gcp"] = map[string]any{
			"deployCSIDriver": config.DeployCSIDriver(),
		}

		values["gcp-compute-persistent-disk-csi-driver"] = map[string]any{
			"csiNode": map[string]any{
				"kmsPort":      constants.KMSPort,
				"kmsNamespace": "", // empty namespace means we use the release namespace
			},
		}

		values["tags"] = map[string]any{
			"GCP": true,
		}

	case cloudprovider.QEMU:
		values["tags"] = map[string]interface{}{
			"QEMU": true,
		}

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
	}
	return values, nil
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
