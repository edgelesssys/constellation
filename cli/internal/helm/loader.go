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
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/pkg/errors"
	"helm.sh/helm/pkg/ignore"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
)

// Run `go generate` to deterministically create the patched Helm deployment for cilium
//go:generate ./generateCilium.sh

//go:embed all:charts/*
var HelmFS embed.FS

type ChartLoader struct{}

func (i *ChartLoader) Load(csp cloudprovider.Provider, conformanceMode bool, masterSecret []byte, salt []byte, enforcedPCRs []uint32, enforceIDKeyDigest bool) ([]byte, error) {
	ciliumRelease, err := i.loadCilium(csp, conformanceMode)
	if err != nil {
		return nil, err
	}

	conServicesRelease, err := i.loadConstellationServices(csp, masterSecret, salt, enforcedPCRs, enforceIDKeyDigest)
	if err != nil {
		return nil, err
	}
	releases := helm.Releases{Cilium: ciliumRelease, ConstellationServices: conServicesRelease}

	rel, err := json.Marshal(releases)
	if err != nil {
		return nil, err
	}
	return rel, nil
}

func (i *ChartLoader) loadCilium(csp cloudprovider.Provider, conformanceMode bool) (helm.Release, error) {
	chart, err := loadChartsDir(HelmFS, "charts/cilium")
	if err != nil {
		return helm.Release{}, fmt.Errorf("loading cilium chart: %w", err)
	}

	chartRaw, err := i.marshalChart(chart)
	if err != nil {
		return helm.Release{}, fmt.Errorf("packaging chart: %w", err)
	}

	var ciliumVals map[string]any
	switch csp {
	case cloudprovider.GCP:
		ciliumVals = gcpVals
	case cloudprovider.Azure:
		ciliumVals = azureVals
	case cloudprovider.QEMU:
		ciliumVals = qemuVals
	default:
		return helm.Release{}, fmt.Errorf("unknown csp: %s", csp)
	}
	if conformanceMode {
		ciliumVals["kubeProxyReplacementHealthzBindAddr"] = ""
		ciliumVals["kubeProxyReplacement"] = "partial"
		ciliumVals["sessionAffinity"] = true
		ciliumVals["cni"] = map[string]any{
			"chainingMode": "portmap",
		}

	}

	return helm.Release{Chart: chartRaw, Values: ciliumVals, ReleaseName: "cilium", Wait: true}, nil
}

// loadConstellationServices loads the constellation-services chart from the embed.FS, marshals it into a helm-package .tgz and sets the values that can be set in the CLI.
func (i *ChartLoader) loadConstellationServices(csp cloudprovider.Provider, masterSecret []byte, salt []byte, enforcedPCRs []uint32, enforceIDKeyDigest bool) (helm.Release, error) {
	chart, err := loadChartsDir(HelmFS, "charts/edgeless/constellation-services")
	if err != nil {
		return helm.Release{}, fmt.Errorf("loading constellation-services chart: %w", err)
	}

	chartRaw, err := i.marshalChart(chart)
	if err != nil {
		return helm.Release{}, fmt.Errorf("packaging chart: %w", err)
	}

	enforcedPCRsJSON, err := json.Marshal(enforcedPCRs)
	if err != nil {
		return helm.Release{}, fmt.Errorf("marshaling enforcedPCRs: %w", err)
	}

	vals := map[string]any{
		"global": map[string]any{
			"kmsPort":          constants.KMSPort,
			"serviceBasePath":  constants.ServiceBasePath,
			"joinConfigCMName": constants.JoinConfigMap,
			"k8sVersionCMName": constants.K8sVersion,
			"internalCMName":   constants.InternalConfigMap,
		},
		"kms": map[string]any{
			"image":                versions.KmsImage,
			"masterSecret":         base64.StdEncoding.EncodeToString(masterSecret),
			"salt":                 base64.StdEncoding.EncodeToString(salt),
			"namespace":            constants.ConstellationNamespace,
			"saltKeyName":          constants.ConstellationSaltKey,
			"masterSecretKeyName":  constants.ConstellationMasterSecretKey,
			"masterSecretName":     constants.ConstellationMasterSecretStoreName,
			"measurementsFilename": constants.MeasurementsFilename,
		},
		"join-service": map[string]any{
			"csp":          csp,
			"enforcedPCRs": string(enforcedPCRsJSON),
			"image":        versions.JoinImage,
			"namespace":    constants.ConstellationNamespace,
		},
	}

	if csp == cloudprovider.Azure {
		joinServiceVals := vals["join-service"].(map[string]interface{})
		joinServiceVals["enforceIDKeyDigest"] = enforceIDKeyDigest
	}

	return helm.Release{Chart: chartRaw, Values: vals, ReleaseName: "constellation-services", Wait: true}, nil
}

// marshalChart takes a Chart object, packages it to a temporary file and returns the content of that file.
// We currently need to take this approach of marshaling as dependencies are not marshaled correctly with json.Marshal.
// This stems from the fact that chart.Chart does not export the dependencies property.
// See: https://github.com/helm/helm/issues/11454
func (i *ChartLoader) marshalChart(chart *chart.Chart) ([]byte, error) {
	path, err := chartutil.Save(chart, os.TempDir())
	defer os.Remove(path)
	if err != nil {
		return nil, fmt.Errorf("packaging chart: %w", err)
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
