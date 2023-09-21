/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package internal

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/debugd/filebeat"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed templates/filebeat/*
	filebeatHelmAssets embed.FS

	filebeatAssets = filebeat.Assets
)

// FilebeatPreparer prepares the Filebeat Helm chart.
type FilebeatPreparer struct {
	fh   file.Handler
	port int
	templatePreparer
}

// NewFilebeatPreparer returns a new FilebeatPreparer.
func NewFilebeatPreparer(port int) *FilebeatPreparer {
	return &FilebeatPreparer{
		fh:   file.NewHandler(afero.NewOsFs()),
		port: port,
	}
}

// Prepare prepares the Filebeat Helm chart by templating the filebeat.yml and inputs.yml files and placing them in the specified directory.
func (p *FilebeatPreparer) Prepare(dir string) error {
	templatedFilebeatYaml, err := p.template(filebeatAssets, "templates/filebeat.yml", FilebeatTemplateData{
		LogstashHost:     fmt.Sprintf("logstash-logstash:%d", p.port),
		AddCloudMetadata: true,
	})
	if err != nil {
		return fmt.Errorf("template filebeat.yml: %w", err)
	}

	inputsYaml, err := filebeatAssets.ReadFile("inputs.yml")
	if err != nil {
		return fmt.Errorf("read log4j2.properties: %w", err)
	}

	rawHelmValues, err := filebeatHelmAssets.ReadFile("templates/filebeat/values.yml")
	if err != nil {
		return fmt.Errorf("read values.yml: %w", err)
	}

	helmValuesYaml := &FilebeatHelmValues{}
	if err := yaml.Unmarshal(rawHelmValues, helmValuesYaml); err != nil {
		return fmt.Errorf("unmarshal values.yml: %w", err)
	}

	helmValuesYaml.Daemonset.FilebeatConfig.FilebeatYml = templatedFilebeatYaml.String()
	helmValuesYaml.Daemonset.FilebeatConfig.InputsYml = string(inputsYaml)

	helmValues, err := yaml.Marshal(helmValuesYaml)
	if err != nil {
		return fmt.Errorf("marshal values.yml: %w", err)
	}

	if err = p.fh.Write(filepath.Join(dir, "filebeat", "values.yml"), helmValues, file.OptMkdirAll); err != nil {
		return fmt.Errorf("write values.yml: %w", err)
	}

	return nil
}

// FilebeatTemplateData is template data.
type FilebeatTemplateData struct {
	LogstashHost     string
	AddCloudMetadata bool
}

// FilebeatHelmValues repesents the Helm values.yml.
type FilebeatHelmValues struct {
	Image     string `yaml:"image"`
	ImageTag  string `yaml:"imageTag"`
	Daemonset struct {
		Enabled        bool `yaml:"enabled"`
		FilebeatConfig struct {
			FilebeatYml string `yaml:"filebeat.yml"`
			InputsYml   string `yaml:"inputs.yml"`
		} `yaml:"filebeatConfig"`
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
