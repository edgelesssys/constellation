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
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/info"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
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
	metadata providerMetadata, logger *slog.Logger,
) func(*info.Map) {
	return func(infoMap *info.Map) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			logger.Info("Start trigger running")

			if err := ctx.Err(); err != nil {
				logger.With("err", err).Error("Start trigger canceled")
				return
			}

			logger.Info("Get flags from infos")
			_, ok, err := infoMap.Get("logcollect")
			if err != nil {
				logger.Error(fmt.Sprintf("Getting infos: %v", err))
				return
			}
			if !ok {
				logger.Info("Flag 'logcollect' not set")
				return
			}

			cerdsGetter, err := newCloudCredentialGetter(ctx, provider, infoMap)
			if err != nil {
				logger.Error(fmt.Sprintf("Creating cloud credential getter: %v", err))
				return
			}

			logger.Info("Getting credentials")
			creds, err := cerdsGetter.GetOpensearchCredentials(ctx)
			if err != nil {
				logger.Error(fmt.Sprintf("Getting opensearch credentials: %v", err))
				return
			}

			logger.Info(fmt.Sprintf("Getting logstash pipeline template from image %s", versions.LogstashImage))
			tmpl, err := getTemplate(ctx, logger, versions.LogstashImage, "/run/logstash/templates/pipeline.conf", "/run/logstash")
			if err != nil {
				logger.Error(fmt.Sprintf("Getting logstash pipeline template: %v", err))
				return
			}

			infoMapM, err := infoMap.GetCopy()
			if err != nil {
				logger.Error(fmt.Sprintf("Getting copy of map from info: %v", err))
				return
			}
			infoMapM = filterInfoMap(infoMapM)
			setCloudMetadata(ctx, infoMapM, provider, metadata)

			logger.Info("Writing logstash pipeline")
			pipelineConf := logstashConfInput{
				Port:        5044,
				Host:        openSearchHost,
				InfoMap:     infoMapM,
				Credentials: creds,
			}
			if err := writeTemplate("/run/logstash/pipeline/pipeline.conf", tmpl, pipelineConf); err != nil {
				logger.Error(fmt.Sprintf("Writing logstash config: %v", err))
				return
			}

			logger.Info(fmt.Sprintf("Getting filebeat config template from image %s", versions.FilebeatImage))
			tmpl, err = getTemplate(ctx, logger, versions.FilebeatImage, "/run/filebeat/templates/filebeat.yml", "/run/filebeat")
			if err != nil {
				logger.Error(fmt.Sprintf("Getting filebeat config template: %v", err))
				return
			}
			filebeatConf := filebeatConfInput{
				LogstashHost:     "localhost:5044",
				AddCloudMetadata: true,
			}
			if err := writeTemplate("/run/filebeat/filebeat.yml", tmpl, filebeatConf); err != nil {
				logger.Error(fmt.Sprintf("Writing filebeat pipeline: %v", err))
				return
			}

			logger.Info("Starting log collection pod")
			if err := startPod(ctx, logger); err != nil {
				logger.Error(fmt.Sprintf("Starting log collection: %v", err))
			}
		}()
	}
}

func getTemplate(ctx context.Context, logger *slog.Logger, image, templateDir, destDir string) (*template.Template, error) {
	createContainerArgs := []string{
		"create",
		"--name=template",
		image,
	}
	createContainerCmd := exec.CommandContext(ctx, "podman", createContainerArgs...)
	logger.Info("Creating template container")
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
	logger.Info("Copying templates")
	if out, err := copyFromCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("copying templates: %w\n%s", err, out)
	}

	removeContainerArgs := []string{
		"rm",
		"template",
	}
	removeContainerCmd := exec.CommandContext(ctx, "podman", removeContainerArgs...)
	logger.Info("Removing template container")
	if out, err := removeContainerCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("removing template container: %w\n%s", err, out)
	}

	tmpl, err := template.ParseFiles(templateDir)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return tmpl, nil
}

func startPod(ctx context.Context, logger *slog.Logger) error {
	// create a shared pod for filebeat, metricbeat and logstash
	createPodArgs := []string{
		"pod",
		"create",
		"logcollection",
	}
	createPodCmd := exec.CommandContext(ctx, "podman", createPodArgs...)
	logger.Info(fmt.Sprintf("Create pod command: %v", createPodCmd.String()))
	if out, err := createPodCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create pod: %w; output: %s", err, out)
	}

	// start logstash container
	logstashLog := newCmdLogger(logger.WithGroup("logstash"))
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
	logger.Info(fmt.Sprintf("Run logstash command: %v", runLogstashCmd.String()))
	runLogstashCmd.Stdout = logstashLog
	runLogstashCmd.Stderr = logstashLog
	if err := runLogstashCmd.Start(); err != nil {
		return fmt.Errorf("failed to start logstash: %w", err)
	}
	if out, err := exec.CommandContext(ctx, "podman", "wait", "logstash", "--condition=running", "--interval=15s").CombinedOutput(); err != nil {
		logger.Error("Logstash container failed to reach healthy status", "err", err, "output", out)
		return fmt.Errorf("waiting for logstash container to reach healthy status: %w; output: %s", err, out)
	}

	// start filebeat container
	filebeatLog := newCmdLogger(logger.WithGroup("filebeat"))
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
		"--volume=/run/filebeat/filebeat.yml:/usr/share/filebeat/filebeat.yml:ro",
		versions.FilebeatImage,
	}
	runFilebeatCmd := exec.CommandContext(ctx, "podman", runFilebeatArgs...)
	logger.Info(fmt.Sprintf("Run filebeat command: %v", runFilebeatCmd.String()))
	runFilebeatCmd.Stdout = filebeatLog
	runFilebeatCmd.Stderr = filebeatLog
	if err := runFilebeatCmd.Start(); err != nil {
		return fmt.Errorf("failed to run filebeat: %w", err)
	}
	if out, err := exec.CommandContext(ctx, "podman", "wait", "filebeat", "--condition=running", "--interval=15s").CombinedOutput(); err != nil {
		logger.Error("Filebeat container failed to reach healthy status", "err", err, "output", out)
		return fmt.Errorf("waiting for filebeat container to reach healthy status: %w; output: %s", err, out)
	}

	return nil
}

type logstashConfInput struct {
	Port        int
	Host        string
	InfoMap     map[string]string
	Credentials credentials
}

type filebeatConfInput struct {
	LogstashHost     string
	AddCloudMetadata bool
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

func newCmdLogger(logger *slog.Logger) io.Writer {
	return &cmdLogger{logger: logger}
}

type cmdLogger struct {
	logger *slog.Logger
}

func (c *cmdLogger) Write(p []byte) (n int, err error) {
	c.logger.Info(string(p))
	return len(p), nil
}

type providerMetadata interface {
	// Self retrieves the current instance.
	Self(ctx context.Context) (metadata.InstanceMetadata, error)
	// UID returns the UID of the current instance.
	UID(ctx context.Context) (string, error)
}
