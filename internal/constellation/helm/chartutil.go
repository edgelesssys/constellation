/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/internal/file"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

// saveChartToDisk saves a chart as files in a directory.
//
// This takes the chart name, and creates a new subdirectory inside of the given dest
// directory, writing the chart's contents to that subdirectory.
// Dependencies are written using the same format, instead of writing them as tar files
//
// View the SaveDir implementation in chartutil as reference: https://github.com/helm/helm/blob/3a31588ad33fe3b89af5a2a54ee1d25bfe6eaa5e/pkg/chartutil/save.go#L40
func saveChartToDisk(c *chart.Chart, dest string, fileHandler file.Handler) error {
	// Create the chart directory
	outdir := filepath.Join(dest, c.Name())
	if fi, err := fileHandler.Stat(outdir); err == nil && !fi.IsDir() {
		return fmt.Errorf("file %s already exists and is not a directory", outdir)
	}
	if err := fileHandler.MkdirAll(outdir); err != nil {
		return err
	}

	// Save the chart file.
	if err := chartutil.SaveChartfile(filepath.Join(outdir, chartutil.ChartfileName), c.Metadata); err != nil {
		return err
	}

	// Save values.yaml
	for _, f := range c.Raw {
		if f.Name == chartutil.ValuesfileName {
			vf := filepath.Join(outdir, chartutil.ValuesfileName)
			if err := writeFile(vf, f.Data, fileHandler); err != nil {
				return err
			}
		}
	}

	// Save values.schema.json if it exists
	if c.Schema != nil {
		filename := filepath.Join(outdir, chartutil.SchemafileName)
		if err := writeFile(filename, c.Schema, fileHandler); err != nil {
			return err
		}
	}

	// Save templates and files
	for _, o := range [][]*chart.File{c.Templates, c.Files} {
		for _, f := range o {
			n := filepath.Join(outdir, f.Name)
			if err := writeFile(n, f.Data, fileHandler); err != nil {
				return err
			}
		}
	}

	// Save dependencies
	base := filepath.Join(outdir, chartutil.ChartsDir)
	for _, dep := range c.Dependencies() {
		// Don't write dependencies as tar files
		// Instead recursively use saveChartToDisk
		if err := saveChartToDisk(dep, base, fileHandler); err != nil {
			return fmt.Errorf("saving %s: %w", dep.ChartFullPath(), err)
		}
	}

	return nil
}

func writeFile(name string, content []byte, fileHandler file.Handler) error {
	if err := fileHandler.MkdirAll(filepath.Dir(name)); err != nil {
		return err
	}
	return fileHandler.Write(name, content)
}
