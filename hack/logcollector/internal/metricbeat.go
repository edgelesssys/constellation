/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package internal

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/debugd/metricbeat"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed templates/metricbeat/*
	metricbeatHelmAssets embed.FS

	metricbeatAssets = metricbeat.Assets
)

// MetricbeatPreparer prepares the Metricbeat Helm chart.
type MetricbeatPreparer struct {
	fh   file.Handler
	port int
	templatePreparer
}

// NewMetricbeatPreparer returns a new MetricbeatPreparer.
func NewMetricbeatPreparer(port int) *MetricbeatPreparer {
	return &MetricbeatPreparer{
		fh:   file.NewHandler(afero.NewOsFs()),
		port: port,
	}
}

// Prepare prepares the Filebeat Helm chart by templating the metricbeat.yml file and placing it
// in the specified directory.
func (p *MetricbeatPreparer) Prepare(dir string) error {
	templatedMetricbeatYaml, err := p.template(metricbeatAssets, "templates/metricbeat.yml", MetricbeatTemplateData{
		LogstashHost: fmt.Sprintf("logstash-logstash:%d", p.port),
	})
	if err != nil {
		return fmt.Errorf("template metricbeat.yml: %w", err)
	}

	rawHelmValues, err := metricbeatHelmAssets.ReadFile("templates/metricbeat/values.yml")
	if err != nil {
		return fmt.Errorf("read values.yml: %w", err)
	}

	helmValuesYaml := &MetricbeatHelmValues{}
	if err := yaml.Unmarshal(rawHelmValues, helmValuesYaml); err != nil {
		return fmt.Errorf("unmarshal values.yml: %w", err)
	}

	helmValuesYaml.Daemonset.MetricbeatConfig.MetricbeatYml = templatedMetricbeatYaml.String()
	helmValues, err := yaml.Marshal(helmValuesYaml)
	if err != nil {
		return fmt.Errorf("marshal values.yml: %w", err)
	}

	if err = p.fh.Write(filepath.Join(dir, "metricbeat", "values.yml"), helmValues, file.OptMkdirAll); err != nil {
		return fmt.Errorf("write values.yml: %w", err)
	}

	return nil
}

// MetricbeatTemplateData is template data.
type MetricbeatTemplateData struct {
	LogstashHost string
}

// MetricbeatHelmValues repesents the Helm values.yml.
type MetricbeatHelmValues struct {
	Image      string `yaml:"image"`
	ImageTag   string `yaml:"imageTag"`
	Deployment struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"deployment"`
	Daemonset struct {
		Enabled          bool `yaml:"enabled"`
		MetricbeatConfig struct {
			MetricbeatYml string `yaml:"metricbeat.yml"`
		} `yaml:"metricbeatConfig"`
		ExtraEnvs    []interface{} `yaml:"extraEnvs"`
		SecretMounts []interface{} `yaml:"secretMounts"`
		Tolerations  []struct {
			Key      string `yaml:"key"`
			Operator string `yaml:"operator"`
			Effect   string `yaml:"effect"`
		} `yaml:"tolerations"`
		SecurityContext struct {
			Privileged bool `yaml:"privileged"`
			RunAsUser  int  `yaml:"runAsUser"`
		} `yaml:"securityContext"`
		ExtraVolumeMounts []struct {
			Name      string `yaml:"name"`
			MountPath string `yaml:"mountPath"`
			ReadOnly  bool   `yaml:"readOnly"`
		} `yaml:"extraVolumeMounts"`
		ExtraVolumes []struct {
			Name     string `yaml:"name"`
			HostPath struct {
				Path string `yaml:"path"`
				Type string `yaml:"type"`
			} `yaml:"hostPath"`
		} `yaml:"extraVolumes"`
	} `yaml:"daemonset"`
}
