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
	"os/exec"
	"sync"

	"github.com/edgelesssys/constellation/v2/debugd/internal/debugd/info"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/logcollection"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

// NewStartTrigger returns a trigger func can be registered with an infos instance.
// The trigger is called when infos changes to received state and starts a log collection pod
// with filebeat and logstash in case the flags are set.
//
// This requires podman to be installed.
func NewStartTrigger(ctx context.Context, wg *sync.WaitGroup, provider cloudprovider.Provider,
	meta providerMetadata, logger *logger.Logger,
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

			qemuPassword := ""
			if provider == cloudprovider.QEMU {
				qemuPassword, ok, err = infoMap.Get("qemu.opensearch-pw")
				if err != nil {
					logger.Errorf("getting qemu.opensearch-pw from info: %w", err)
					return
				}
				if !ok {
					logger.Errorf("qemu.opensearch-pw not found in info")
					return
				}
			}
			credsGetter, err := logcollection.NewCloudCredentialGetter(ctx, provider, qemuPassword)
			if err != nil {
				logger.Errorf("Creating cloud credential getter: %v", err)
				return
			}

			logger.Infof("Getting credentials")
			creds, err := credsGetter.GetOpensearchCredentials(ctx)
			if err != nil {
				logger.Errorf("Getting opensearch credentials: %v", err)
				return
			}

			infoMapM, err := infoMap.GetCopy()
			if err != nil {
				logger.Errorf("Getting copy of map from info: %v", err)
				return
			}

			logger.Infof("Writing logstash pipeline")
			ls, err := logcollection.NewLogstash()
			if err != nil {
				logger.Errorf("Creating logstash: %v", err)
				return
			}

			self := metadata.InstanceMetadata{}
			self, err = meta.Self(ctx)
			if err != nil {
				logger.Errorf("Getting self metadata: %v", err)
			}

			uid := ""
			uid, err = meta.UID(ctx)
			if err != nil {
				logger.Errorf("Getting UID: %v", err)
			}

			tmplInput := logcollection.NewLogstashPipelineConfInput(
				creds,
				infoMapM,
				logcollection.LogMetadata{
					Provider: provider.String(),
					Name:     self.Name,
					Role:     self.Role,
					VPCIP:    self.VPCIP,
					UID:      uid,
				},
			)
			if err := ls.WritePipelineConf(tmplInput, "pipeline.conf"); err != nil {
				logger.Errorf("Writing logstash pipeline: %v", err)
				return
			}

			logger.Infof("Starting log collection pod")
			if err := startPod(ctx, logger); err != nil {
				logger.Errorf("Starting filebeat: %v", err)
			}
		}()
	}
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
