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
	templatedSystemMetricbeatYaml, err := p.template(metricbeatAssets, "templates/metricbeat.yml", MetricbeatTemplateData{
		LogstashHost:         fmt.Sprintf("logstash-logstash:%d", p.port),
		Port:                 5066,
		CollectSystemMetrics: true,
		AddCloudMetadata:     true,
	})
	if err != nil {
		return fmt.Errorf("template system metricbeat.yml: %w", err)
	}
	templatedK8sMetricbeatYaml, err := p.template(metricbeatAssets, "templates/metricbeat.yml", MetricbeatTemplateData{
		LogstashHost:       fmt.Sprintf("logstash-logstash:%d", p.port),
		Port:               5067,
		CollectEtcdMetrics: true,
		AddCloudMetadata:   true,
	})
	if err != nil {
		return fmt.Errorf("template k8s metricbeat.yml: %w", err)
	}

	rawAllNodesHelmValues, err := metricbeatHelmAssets.ReadFile("templates/metricbeat/values-all-nodes.yml")
	if err != nil {
		return fmt.Errorf("read values-all-nodes.yml: %w", err)
	}
	rawControlPlaneHelmValues, err := metricbeatHelmAssets.ReadFile("templates/metricbeat/values-control-plane.yml")
	if err != nil {
		return fmt.Errorf("read values-control-plane.yml: %w", err)
	}

	allNodesHelmValuesYaml := &MetricbeatHelmValues{}
	if err := yaml.Unmarshal(rawAllNodesHelmValues, allNodesHelmValuesYaml); err != nil {
		return fmt.Errorf("unmarshal values-all-nodes.yml: %w", err)
	}
	controlPlaneHelmValuesYaml := &MetricbeatHelmValues{}
	if err := yaml.Unmarshal(rawControlPlaneHelmValues, controlPlaneHelmValuesYaml); err != nil {
		return fmt.Errorf("unmarshal values-control-plane.yml: %w", err)
	}

	allNodesHelmValuesYaml.Daemonset.MetricbeatConfig.MetricbeatYml = templatedSystemMetricbeatYaml.String()
	controlPlaneHelmValuesYaml.Daemonset.MetricbeatConfig.MetricbeatYml = templatedK8sMetricbeatYaml.String()

	allNodesHelmValues, err := yaml.Marshal(allNodesHelmValuesYaml)
	if err != nil {
		return fmt.Errorf("marshal values-all-nodes.ym: %w", err)
	}
	controlPlaneHelmValues, err := yaml.Marshal(controlPlaneHelmValuesYaml)
	if err != nil {
		return fmt.Errorf("marshal values-control-plane.yml: %w", err)
	}

	if err = p.fh.Write(filepath.Join(dir, "metricbeat", "values-all-nodes.yml"), allNodesHelmValues, file.OptMkdirAll); err != nil {
		return fmt.Errorf("write values-all-nodes.yml: %w", err)
	}
	if err = p.fh.Write(filepath.Join(dir, "metricbeat", "values-control-plane.yml"), controlPlaneHelmValues, file.OptMkdirAll); err != nil {
		return fmt.Errorf("write values-control-plane.yml: %w", err)
	}

	return nil
}

// MetricbeatTemplateData is template data.
type MetricbeatTemplateData struct {
	LogstashHost         string
	Port                 int
	CollectEtcdMetrics   bool
	CollectSystemMetrics bool
	CollectK8sMetrics    bool
	AddK8sMetadata       bool
	AddCloudMetadata     bool
}

// MetricbeatHelmValues repesents the Helm values.yml.
type MetricbeatHelmValues struct {
	Image            string `yaml:"image"`
	ImageTag         string `yaml:"imageTag"`
	KubeStateMetrics struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"kube_state_metrics"`
	Deployment struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"deployment"`
	Daemonset        Daemonset `yaml:"daemonset"`
	ClusterRoleRules []struct {
		APIGroups       []string `yaml:"apiGroups,omitempty"`
		Resources       []string `yaml:"resources,omitempty"`
		Verbs           []string `yaml:"verbs"`
		NonResourceURLs []string `yaml:"nonResourceURLs,omitempty"`
	} `yaml:"clusterRoleRules"`
}

// Daemonset represents the nested daemonset for the Helm values.yml.
type Daemonset struct {
	Enabled          bool `yaml:"enabled"`
	HostNetworking   bool `yaml:"hostNetworking"`
	MetricbeatConfig struct {
		MetricbeatYml string `yaml:"metricbeat.yml"`
	} `yaml:"metricbeatConfig"`
	ExtraEnvs    []any `yaml:"extraEnvs"`
	SecretMounts []any `yaml:"secretMounts"`
	NodeSelector any   `yaml:"nodeSelector"`
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
}
