/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/
package internal

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/debugd/logstash"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var (
	//go:embed templates/logstash/*
	logstashHelmAssets embed.FS

	logstashAssets = logstash.Assets
)

const (
	openSearchHost = "https://search-e2e-logs-y46renozy42lcojbvrt3qq7csm.eu-central-1.es.amazonaws.com:443"
)

// LogstashPreparer prepares the Logstash Helm chart.
type LogstashPreparer struct {
	fh          file.Handler
	fields      map[string]string
	username    string
	password    string
	port        int
	templatePreparer
}

// NewLogstashPreparer returns a new LogstashPreparer.
func NewLogstashPreparer(fields map[string]string, username, password string, port int) *LogstashPreparer {
	return &LogstashPreparer{
		username:    username,
		password:    password,
		fields:      fields,
		fh:          file.NewHandler(afero.NewOsFs()),
		port:        port,
	}
}

// Prepare prepares the Logstash Helm chart by templating the required files and placing them in the specified directory.
func (p *LogstashPreparer) Prepare(dir string) error {
	templatedPipelineConf, err := p.template(logstashAssets, "templates/pipeline.conf", pipelineConfTemplate{
		InfoMap:     p.fields,
		Host:        openSearchHost,
		Credentials: Credentials{
			Username: p.username,
			Password: p.password,
		},
		Port: p.port,
	})
	if err != nil {
		return fmt.Errorf("template pipeline.conf: %w", err)
	}

	logstashYaml, err := logstashAssets.ReadFile("config/logstash.yml")
	if err != nil {
		return fmt.Errorf("read logstash.yml: %w", err)
	}

	log4jProperties, err := logstashAssets.ReadFile("config/log4j2.properties")
	if err != nil {
		return fmt.Errorf("read log4j2.properties: %w", err)
	}

	rawHelmValues, err := logstashHelmAssets.ReadFile("templates/logstash/values.yml")
	if err != nil {
		return fmt.Errorf("read values.yml: %w", err)
	}

	helmValuesYaml := &LogstashHelmValues{}
	if err := yaml.Unmarshal(rawHelmValues, helmValuesYaml); err != nil {
		return fmt.Errorf("unmarshal values.yml: %w", err)
	}

	helmValuesYaml.LogstashConfig.LogstashYml = helmValuesYaml.LogstashConfig.LogstashYml + string(logstashYaml)
	helmValuesYaml.LogstashConfig.Log4J2Properties = string(log4jProperties)
	helmValuesYaml.LogstashPipeline.LogstashConf = templatedPipelineConf.String()
	helmValuesYaml.Service.Ports[0].Port = p.port
	helmValuesYaml.Service.Ports[0].TargetPort = p.port

	helmValues, err := yaml.Marshal(helmValuesYaml)
	if err != nil {
		return fmt.Errorf("marshal values.yml: %w", err)
	}

	if err = p.fh.Write(filepath.Join(dir, "logstash", "values.yml"), helmValues, file.OptMkdirAll); err != nil {
		return fmt.Errorf("write values.yml: %w", err)
	}

	return nil
}

// LogstashHelmValues represents the values.yml file for the Logstash Helm chart.
type LogstashHelmValues struct {
	Image          string `yaml:"image"`
	ImageTag       string `yaml:"imageTag"`
	LogstashConfig struct {
		LogstashYml      string `yaml:"logstash.yml"`
		Log4J2Properties string `yaml:"log4j2.properties"`
	} `yaml:"logstashConfig"`
	LogstashPipeline struct {
		LogstashConf string `yaml:"logstash.conf"`
	} `yaml:"logstashPipeline"`
	Service struct {
		Ports []struct {
			Name       string `yaml:"name"`
			Port       int    `yaml:"port"`
			Protocol   string `yaml:"protocol"`
			TargetPort int    `yaml:"targetPort"`
		} `yaml:"ports"`
	} `yaml:"service"`
	Tolerations []struct {
		Key      string `yaml:"key"`
		Operator string `yaml:"operator"`
		Effect   string `yaml:"effect"`
	} `yaml:"tolerations"`
}

// pipelineConfTemplate is template Data.
type pipelineConfTemplate struct {
	InfoMap     map[string]string
	Host        string
	Credentials Credentials
	Port        int
}

// Credentials is template Data.
type Credentials struct {
	Username string
	Password string
}
