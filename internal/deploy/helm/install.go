/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/retry"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/util/wait"
)

// TODO clenaup once agreed on design.
// copied from bootstrapper

const (
	// timeout is the maximum time given to the helm Installer.
	timeout = 5 * time.Minute
	// maximumRetryAttempts is the maximum number of attempts to retry a helm install.
	maximumRetryAttempts = 3
)

// Installer is a wrapper for a helm install action.
type Installer struct {
	*action.Install
	log *logger.Logger
}

// NewInstaller creates a new Installer with the given logger.
func NewInstaller(log *logger.Logger, kubeconfig string) (*Installer, error) {
	settings := cli.New()
	settings.KubeConfig = kubeconfig

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), constants.HelmNamespace,
		"secret", log.Infof); err != nil {
		return nil, err
	}

	action := action.NewInstall(actionConfig)
	action.Namespace = constants.HelmNamespace
	action.Timeout = timeout

	return &Installer{
		action,
		log,
	}, nil
}

// InstallChart is the generic install function for helm charts.
// When timeout is nil, the default timeout is used.
func (h *Installer) InstallChart(ctx context.Context, release Release, timeout *time.Duration) error {
	return h.InstallChartWithValues(ctx, release, release.Values, timeout)
}

// InstallChartWithValues is the generic install function for helm charts with custom values.
func (h *Installer) InstallChartWithValues(ctx context.Context, release Release, values map[string]any, timeout *time.Duration) error {
	h.ReleaseName = release.ReleaseName
	if err := h.SetWaitMode(release.WaitMode); err != nil {
		return err
	}
	if timeout != nil {
		h.Timeout = *timeout
	}
	return h.install(ctx, release.Chart, values)
}

// install tries to install the given chart and aborts after ~5 tries.
// The function will wait 30 seconds before retrying a failed installation attempt.
// After 3 tries, the retrier will be canceled and the function returns with an error.
func (h *Installer) install(ctx context.Context, chartRaw []byte, values map[string]any) error {
	var retries int
	retriable := func(err error) bool {
		// abort after maximumRetryAttempts tries.
		if retries >= maximumRetryAttempts {
			return false
		}
		retries++
		// only retry if atomic is set
		// otherwise helm doesn't uninstall
		// the release on failure
		if !h.Atomic {
			return false
		}
		// check if error is retriable
		return wait.Interrupted(err) ||
			strings.Contains(err.Error(), "connection refused")
	}

	reader := bytes.NewReader(chartRaw)
	chart, err := loader.LoadArchive(reader)
	if err != nil {
		return fmt.Errorf("helm load archive: %w", err)
	}

	doer := installDoer{
		h,
		chart,
		values,
		h.log,
	}
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, retriable)

	retryLoopStartTime := time.Now()
	if err := retrier.Do(ctx); err != nil {
		return fmt.Errorf("helm install: %w", err)
	}
	retryLoopFinishDuration := time.Since(retryLoopStartTime)
	h.log.With(zap.String("chart", chart.Name()), zap.Duration("duration", retryLoopFinishDuration)).Infof("Helm chart installation finished")

	return nil
}

// SetWaitMode sets the wait mode of the installer.
func (h *Installer) SetWaitMode(waitMode WaitMode) error {
	switch waitMode {
	case WaitModeNone:
		h.Wait = false
		h.Atomic = false
	case WaitModeWait:
		h.Wait = true
		h.Atomic = false
	case WaitModeAtomic:
		h.Wait = true
		h.Atomic = true
	default:
		return fmt.Errorf("unknown wait mode %q", waitMode)
	}
	return nil
}

// installDoer is a help struct to enable retrying helm's install action.
type installDoer struct {
	Installer *Installer
	chart     *chart.Chart
	values    map[string]any
	log       *logger.Logger
}

// Do logs which chart is installed and tries to install it.
func (i installDoer) Do(ctx context.Context) error {
	i.log.With(zap.String("chart", i.chart.Name())).Infof("Trying to install Helm chart")

	if _, err := i.Installer.RunWithContext(ctx, i.chart, i.values); err != nil {
		i.log.With(zap.Error(err), zap.String("chart", i.chart.Name())).Errorf("Helm chart installation failed")
		return err
	}

	return nil
}
