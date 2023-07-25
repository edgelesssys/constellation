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
// with filebeat and logstash in case the flags are set.
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
			tmpl, err := getLogstashTemplate(ctx, logger)
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
				InfoMap:     infoMapM,
				Credentials: creds,
			}
			if err := writeTemplate("/run/filebeat/filebeat.yml", tmpl, pipelineConf); err != nil {
				logger.Errorf("Writing logstash pipeline: %v", err)
				return
			}

			logger.Infof("Getting logstash config template")
			tmpl, err = getFilebeatTemplate(ctx, logger)
			if err != nil {
				logger.Errorf("Getting filebeat config template: %v", err)
				return
			}
			filebeatConf := filebeatConfInput{
				LogstashHost: "localhost:5044",
			}
			if err := writeTemplate("/run/logstash/pipeline/pipeline.conf", tmpl, filebeatConf); err != nil {
				logger.Errorf("Writing filebeat config: %v", err)
				return
			}

			logger.Infof("Starting log collection pod")
			if err := startPod(ctx, logger); err != nil {
				logger.Errorf("Starting filebeat: %v", err)
			}
		}()
	}
}

func getLogstashTemplate(ctx context.Context, logger *logger.Logger) (*template.Template, error) {
	createContainerArgs := []string{
		"create",
		"--name=template",
		versions.LogstashImage,
	}
	createContainerCmd := exec.CommandContext(ctx, "podman", createContainerArgs...)
	logger.Infof("Creating logstash template container")
	if out, err := createContainerCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("creating logstash template container: %w\n%s", err, out)
	}

	if err := os.MkdirAll("/run/logstash", 0o777); err != nil {
		return nil, fmt.Errorf("creating logstash template dir: %w", err)
	}

	copyFromArgs := []string{
		"cp",
		"template:/usr/share/constellogs/templates/",
		"/run/logstash/",
	}
	copyFromCmd := exec.CommandContext(ctx, "podman", copyFromArgs...)
	logger.Infof("Copying logstash templates")
	if out, err := copyFromCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("copying logstash templates: %w\n%s", err, out)
	}

	removeContainerArgs := []string{
		"rm",
		"template",
	}
	removeContainerCmd := exec.CommandContext(ctx, "podman", removeContainerArgs...)
	logger.Infof("Removing logstash template container")
	if out, err := removeContainerCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("removing logstash template container: %w\n%s", err, out)
	}

	tmpl, err := template.ParseFiles("/run/logstash/templates/pipeline.conf")
	if err != nil {
		return nil, fmt.Errorf("parsing logstash template: %w", err)
	}

	return tmpl, nil
}

func getFilebeatTemplate(ctx context.Context, logger *logger.Logger) (*template.Template, error) {
	createContainerArgs := []string{
		"create",
		"--name=template",
		versions.FilebeatImage,
	}
	createContainerCmd := exec.CommandContext(ctx, "podman", createContainerArgs...)
	logger.Infof("Creating filebeat template container")
	if out, err := createContainerCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("creating filebeat template container: %w\n%s", err, out)
	}

	if err := os.MkdirAll("/run/filebeat", 0o777); err != nil {
		return nil, fmt.Errorf("creating filebeat template dir: %w", err)
	}

	copyFromArgs := []string{
		"cp",
		"template:/usr/share/constellogs/templates/",
		"/run/filebeat/",
	}
	copyFromCmd := exec.CommandContext(ctx, "podman", copyFromArgs...)
	logger.Infof("Copying filebeat templates")
	if out, err := copyFromCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("copying filebeat templates: %w\n%s", err, out)
	}

	removeContainerArgs := []string{
		"rm",
		"template",
	}
	removeContainerCmd := exec.CommandContext(ctx, "podman", removeContainerArgs...)
	logger.Infof("Removing filebeat template container")
	if out, err := removeContainerCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("removing filebeat template container: %w\n%s", err, out)
	}

	tmpl, err := template.ParseFiles("/run/filebeat/templates/filebeat.yml")
	if err != nil {
		return nil, fmt.Errorf("parsing filebeat template: %w", err)
	}

	return tmpl, nil
}

func startPod(ctx context.Context, logger *logger.Logger) error {
	// create a shared pod for filebeat and logstash
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

	return nil
}

type logstashConfInput struct {
	Port        int
	Host        string
	InfoMap     map[string]string
	Credentials credentials
}

type filebeatConfInput struct {
	LogstashHost string
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
