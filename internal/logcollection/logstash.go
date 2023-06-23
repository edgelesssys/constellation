/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package logcollection uses logstash and filebeat to collect logs centrally for debugging purposes.
package logcollection

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

//go:embed logstash/*
var logstashFS embed.FS

// Logstash contains deployment information for Logstash.
type Logstash struct {
	helmValuesTemplate   *template.Template
	pipelineConfTemplate *template.Template
}

// NewLogstash returns a new Logstash deployment configuration for a given pipeline template.
// If helmValuesTemplateFile is set, it will be used to generate the Helm values.
func NewLogstash() (*Logstash, error) {
	pipelineConfTemplate, err := template.ParseFS(logstashFS, "logstash/tepmlates/pipeline.conf.template")
	if err != nil {
		return nil, fmt.Errorf("parsing logstash pipeline template: %w", err)
	}

	helmValuesTemplate, err := template.ParseFS(logstashFS, "logstash/tepmlates/values.yaml.template")
	if err != nil {
		return nil, fmt.Errorf("parsing logstash helm values template: %w", err)
	}

	return &Logstash{
		helmValuesTemplate:   helmValuesTemplate,
		pipelineConfTemplate: pipelineConfTemplate,
	}, nil
}

// WritePipelineConf fills the template with the given input and writes
// the resulting Logstash pipeline config to the specified output file.
func (l *Logstash) WritePipelineConf(in LogstashPipelineConfInput, outputFile string) error {
	filename := filepath.Base(outputFile)
	dir := filepath.Dir(outputFile)

	if err := os.MkdirAll(dir, 0o777); err != nil {
		return fmt.Errorf("creating logstash config dir: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o777)
	if err != nil {
		return fmt.Errorf("opening logstash config file: %w", err)
	}
	defer file.Close()

	if err := l.pipelineConfTemplate.Execute(file, in); err != nil {
		return fmt.Errorf("executing logstash pipeline template: %w", err)
	}

	return nil
}

// LogstashPipelineConfInput is the template input for the logstash pipeline config.
type LogstashPipelineConfInput struct {
	Host        string
	InfoMap     map[string]string
	Credentials Credentials
}

// NewLogstashPipelineConfInput returns a new LogstashPipelineConfInput for
// the given OpenSearch credentials, infoMap and logging metadata.
func NewLogstashPipelineConfInput(creds Credentials, infoMap map[string]string, metadata LogMetadata) LogstashPipelineConfInput {
	return LogstashPipelineConfInput{
		Host:        openSearchHost,
		InfoMap:     prepareInfoMap(infoMap, metadata),
		Credentials: creds,
	}
}

// WriteHelmValues fills the template with the given pipeline config and writes
// the resulting Logstash Helm values to the specified output file.
func (l *Logstash) WriteHelmValues(pipelineConfFile, outputFile string) error {
	in := LogstashHelmValuesInput{}

	pipelineConf, err := os.ReadFile(pipelineConfFile)
	if err != nil {
		return fmt.Errorf("reading logstash pipeline config: %w", err)
	}
	in.PipelineConf = string(pipelineConf)

	log4jProperties, err := logstashFS.ReadFile("logstash/config/log4j2.properties")
	if err != nil {
		return fmt.Errorf("reading logstash log4j properties: %w", err)
	}
	in.Log4jProperties = string(log4jProperties)

	logstashYaml, err := logstashFS.ReadFile("logstash/config/logstash.yaml")
	if err != nil {
		return fmt.Errorf("reading logstash yaml: %w", err)
	}
	in.LogstashYaml = string(logstashYaml)

	file, err := makeTemplateOutputFile(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := l.pipelineConfTemplate.Execute(file, in); err != nil {
		return fmt.Errorf("executing logstash pipeline template: %w", err)
	}

	return nil
}

// LogstashHelmValuesInput is the template input for the logstash Helm values.
type LogstashHelmValuesInput struct {
	PipelineConf    string
	Log4jProperties string
	LogstashYaml    string
}

// makeTemplateOutputFile creates the output file and its parent directory.
func makeTemplateOutputFile(f string) (*os.File, error) {
	filename := filepath.Base(f)
	dir := filepath.Dir(f)

	if err := os.MkdirAll(dir, 0o777); err != nil {
		return nil, fmt.Errorf("creating logstash config dir: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o777)
	if err != nil {
		return nil, fmt.Errorf("opening logstash config file: %w", err)
	}

	return file, nil
}
