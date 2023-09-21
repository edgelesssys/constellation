/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package logcollector uses podman to deploy logstash and filebeat containers
// in order to collect logs centrally for debugging purposes.
package logcollector

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/info"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

const (
	openSearchHost = "https://search-e2e-logs-y46renozy42lcojbvrt3qq7csm.eu-central-1.es.amazonaws.com:443"
)

// NewStartTrigger returns a trigger func can be registered with an infos instance.
// The trigger is called when infos changes to received state and starts a log collection pod
// with filebeat, metricbeat and logstash in case the flags are set.
//
// This requires podman to be installed.
func NewStartTrigger(ctx context.Context, wg *sync.WaitGroup, provider cloudprovider.Provider,
	metadata providerMetadata, logger *logger.Logger,
) func(*info.Map) {
	return func(infoMap *info.Map) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			logger.Infof("Start trigger running")

			if err := ctx.Err(); err != nil {
				logger.With("err", err).Errorf("Start trigger canceled")
				return
			}

			logger.Infof("Get flags from infos")
			_, ok, err := infoMap.Get("logcollect")
			if err != nil {
				logger.Errorf("Getting infos: %v", err)
				return
			}
			if !ok {
				logger.Infof("Flag 'logcollect' not set")
				return
			}

			cerdsGetter, err := newCloudCredentialGetter(ctx, provider, infoMap)
			if err != nil {
				logger.Errorf("Creating cloud credential getter: %v", err)
				return
			}

			logger.Infof("Getting credentials")
			creds, err := cerdsGetter.GetOpensearchCredentials(ctx)
			if err != nil {
				logger.Errorf("Getting opensearch credentials: %v", err)
				return
			}

			logger.Infof("Getting logstash pipeline template")
			tmpl, err := getTemplate(ctx, logger, versions.LogstashImage, "/run/logstash/templates/pipeline.conf", "/run/logstash")
			if err != nil {
				logger.Errorf("Getting logstash pipeline template: %v", err)
				return
			}

			infoMapM, err := infoMap.GetCopy()
			if err != nil {
				logger.Errorf("Getting copy of map from info: %v", err)
				return
			}
			infoMapM = filterInfoMap(infoMapM)
			setCloudMetadata(ctx, infoMapM, provider, metadata)

			logger.Infof("Writing logstash pipeline")
			pipelineConf := logstashConfInput{
				Port:        5044,
				Host:        openSearchHost,
				IndexPrefix: "systemd-logs",
				InfoMap:     infoMapM,
				Credentials: creds,
			}
			if err := writeTemplate("/run/logstash/pipeline/pipeline.conf", tmpl, pipelineConf); err != nil {
				logger.Errorf("Writing logstash config: %v", err)
				return
			}

			logger.Infof("Getting filebeat config template")
			tmpl, err = getTemplate(ctx, logger, versions.FilebeatImage, "/run/filebeat/templates/filebeat.yml", "/run/filebeat")
			if err != nil {
				logger.Errorf("Getting filebeat config template: %v", err)
				return
			}
			filebeatConf := filebeatConfInput{
				LogstashHost:     "localhost:5044",
				AddCloudMetadata: true,
			}
			if err := writeTemplate("/run/filebeat/filebeat.yml", tmpl, filebeatConf); err != nil {
				logger.Errorf("Writing filebeat pipeline: %v", err)
				return
			}

			logger.Infof("Getting metricbeat config template")
			tmpl, err = getTemplate(ctx, logger, versions.MetricbeatImage, "/run/metricbeat/templates/metricbeat.yml", "/run/metricbeat")
			if err != nil {
				logger.Errorf("Getting metricbeat config template: %v", err)
				return
			}
			metricbeatConf := metricbeatConfInput{
				LogstashHost:         "localhost:5044",
				Port:                 5066,
				CollectSystemMetrics: true,
				AddCloudMetadata:     true,
			}
			if err := writeTemplate("/run/metricbeat/metricbeat.yml", tmpl, metricbeatConf); err != nil {
				logger.Errorf("Writing metricbeat pipeline: %v", err)
				return
			}

			logger.Infof("Starting log collection pod")
			if err := startPod(ctx, logger); err != nil {
				logger.Errorf("Starting log collection: %v", err)
			}
		}()
	}
}

func getTemplate(ctx context.Context, logger *logger.Logger, image, templateDir, destDir string) (*template.Template, error) {
	createContainerArgs := []string{
		"create",
		"--name=template",
		image,
	}
	createContainerCmd := exec.CommandContext(ctx, "podman", createContainerArgs...)
	logger.Infof("Creating template container")
	if out, err := createContainerCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("creating template container: %w\n%s", err, out)
	}

	if err := os.MkdirAll(destDir, 0o777); err != nil {
		return nil, fmt.Errorf("creating template dir: %w", err)
	}

	copyFromArgs := []string{
		"cp",
		"template:/usr/share/constellogs/templates/",
		destDir,
	}
	copyFromCmd := exec.CommandContext(ctx, "podman", copyFromArgs...)
	logger.Infof("Copying templates")
	if out, err := copyFromCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("copying templates: %w\n%s", err, out)
	}

	removeContainerArgs := []string{
		"rm",
		"template",
	}
	removeContainerCmd := exec.CommandContext(ctx, "podman", removeContainerArgs...)
	logger.Infof("Removing template container")
	if out, err := removeContainerCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("removing template container: %w\n%s", err, out)
	}

	tmpl, err := template.ParseFiles(templateDir)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return tmpl, nil
}

func startPod(ctx context.Context, logger *logger.Logger) error {
	// create a shared pod for filebeat, metricbeat and logstash
	createPodArgs := []string{
		"pod",
		"create",
		"logcollection",
	}
	createPodCmd := exec.CommandContext(ctx, "podman", createPodArgs...)
	logger.Infof("Create pod command: %v", createPodCmd.String())
	if out, err := createPodCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create pod: %w; output: %s", err, out)
	}

	// start logstash container
	logstashLog := newCmdLogger(logger.Named("logstash"))
	runLogstashArgs := []string{
		"run",
		"--rm",
		"--name=logstash",
		"--pod=logcollection",
		"--log-driver=none",
		"--volume=/run/logstash/pipeline:/usr/share/logstash/pipeline/:ro",
		versions.LogstashImage,
	}
	runLogstashCmd := exec.CommandContext(ctx, "podman", runLogstashArgs...)
	logger.Infof("Run logstash command: %v", runLogstashCmd.String())
	runLogstashCmd.Stdout = logstashLog
	runLogstashCmd.Stderr = logstashLog
	if err := runLogstashCmd.Start(); err != nil {
		return fmt.Errorf("failed to start logstash: %w", err)
	}

	// start filebeat container
	filebeatLog := newCmdLogger(logger.Named("filebeat"))
	runFilebeatArgs := []string{
		"run",
		"--rm",
		"--name=filebeat",
		"--pod=logcollection",
		"--privileged",
		"--log-driver=none",
		"--volume=/run/log/journal:/run/log/journal:ro",
		"--volume=/etc/machine-id:/etc/machine-id:ro",
		"--volume=/run/systemd:/run/systemd:ro",
		"--volume=/run/systemd/journal/socket:/run/systemd/journal/socket:rw",
		"--volume=/run/state/var/log:/var/log:ro",
		"--volume=/run/filebeat:/usr/share/filebeat/:ro",
		versions.FilebeatImage,
	}
	runFilebeatCmd := exec.CommandContext(ctx, "podman", runFilebeatArgs...)
	logger.Infof("Run filebeat command: %v", runFilebeatCmd.String())
	runFilebeatCmd.Stdout = filebeatLog
	runFilebeatCmd.Stderr = filebeatLog
	if err := runFilebeatCmd.Start(); err != nil {
		return fmt.Errorf("failed to run filebeat: %w", err)
	}

	// start metricbeat container
	metricbeatLog := newCmdLogger(logger.Named("metricbeat"))
	runMetricbeatArgs := []string{
		"run",
		"--rm",
		"--name=metricbeat",
		"--pod=logcollection",
		"--privileged",
		"--log-driver=none",
		"--volume=/proc:/hostfs/proc:ro",
		"--volume=/sys/fs/cgroup:/hostfs/sys/fs/cgroup:ro",
		"--volume=/run/metricbeat:/usr/share/metricbeat/:ro",
		versions.MetricbeatImage,
	}
	runMetricbeatCmd := exec.CommandContext(ctx, "podman", runMetricbeatArgs...)
	logger.Infof("Run metricbeat command: %v", runMetricbeatCmd.String())
	runMetricbeatCmd.Stdout = metricbeatLog
	runMetricbeatCmd.Stderr = metricbeatLog
	if err := runMetricbeatCmd.Start(); err != nil {
		return fmt.Errorf("failed to run metricbeat: %w", err)
	}

	return nil
}

type logstashConfInput struct {
	Port        int
	Host        string
	IndexPrefix string
	InfoMap     map[string]string
	Credentials credentials
}

type filebeatConfInput struct {
	LogstashHost     string
	AddCloudMetadata bool
}

type metricbeatConfInput struct {
	Port                 int
	LogstashHost         string
	CollectEtcdMetrics   bool
	CollectSystemMetrics bool
	AddCloudMetadata     bool
}

func writeTemplate(path string, templ *template.Template, in any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return fmt.Errorf("creating template dir: %w", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o777)
	if err != nil {
		return fmt.Errorf("opening template file: %w", err)
	}
	defer file.Close()

	if err := templ.Execute(file, in); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}

func filterInfoMap(in map[string]string) map[string]string {
	out := make(map[string]string)

	for k, v := range in {
		if strings.HasPrefix(k, "logcollect.") {
			out[strings.TrimPrefix(k, "logcollect.")] = v
		}
	}

	delete(out, "logcollect")

	return out
}

func setCloudMetadata(ctx context.Context, m map[string]string, provider cloudprovider.Provider, metadata providerMetadata) {
	m["provider"] = provider.String()

	self, err := metadata.Self(ctx)
	if err != nil {
		m["name"] = "unknown"
		m["role"] = "unknown"
		m["vpcip"] = "unknown"
	} else {
		m["name"] = self.Name
		m["role"] = self.Role.String()
		m["vpcip"] = self.VPCIP
	}

	uid, err := metadata.UID(ctx)
	if err != nil {
		m["uid"] = "unknown"
	} else {
		m["uid"] = uid
	}
}

func newCmdLogger(logger *logger.Logger) io.Writer {
	return &cmdLogger{logger: logger}
}

type cmdLogger struct {
	logger *logger.Logger
}

func (c *cmdLogger) Write(p []byte) (n int, err error) {
	c.logger.Infof("%s", p)
	return len(p), nil
}

type providerMetadata interface {
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
	// UID returns the UID of the current instance.
	UID(ctx context.Context) (string, error)
}
