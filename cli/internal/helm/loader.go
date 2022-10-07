/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/pkg/errors"
	"helm.sh/helm/pkg/ignore"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// Run `go generate` to deterministically create the patched Helm deployment for cilium
//go:generate ./generateCilium.sh

//go:embed all:charts/cilium/*
var HelmFS embed.FS

type ChartLoader struct{}

func (i *ChartLoader) Load(csp cloudprovider.Provider, conformanceMode bool) ([]byte, error) {
	ciliumDeployment, err := i.loadCilium(csp, conformanceMode)
	if err != nil {
		return nil, err
	}
	deployments := helm.Deployments{Cilium: ciliumDeployment}
	depl, err := json.Marshal(deployments)
	if err != nil {
		return nil, err
	}
	return depl, nil
}

func (i *ChartLoader) loadCilium(csp cloudprovider.Provider, conformanceMode bool) (helm.Deployment, error) {
	chart, err := loadChartsDir(HelmFS, "charts/cilium")
	if err != nil {
		return helm.Deployment{}, err
	}
	var ciliumVals map[string]interface{}
	switch csp {
	case cloudprovider.GCP:
		ciliumVals = gcpVals
	case cloudprovider.Azure:
		ciliumVals = azureVals
	case cloudprovider.QEMU:
		ciliumVals = qemuVals
	default:
		return helm.Deployment{}, fmt.Errorf("unknown csp: %s", csp)
	}
	if conformanceMode {
		ciliumVals["kubeProxyReplacementHealthzBindAddr"] = ""
		ciliumVals["kubeProxyReplacement"] = "partial"
		ciliumVals["sessionAffinity"] = true
		ciliumVals["cni"] = map[string]interface{}{
			"chainingMode": "portmap",
		}

	}

	return helm.Deployment{Chart: chart, Values: ciliumVals}, nil
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
