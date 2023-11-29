/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

const applyRetryInterval = 30 * time.Second

type applyAction interface {
	Apply(context.Context) error
	SaveChart(chartsDir string, fileHandler file.Handler) error
	ReleaseName() string
	IsAtomic() bool
}

// newActionConfig creates a new action configuration for helm actions.
func newActionConfig(kubeconfig string, logger debugLog) (*action.Configuration, error) {
	settings := cli.New()
	settings.KubeConfig = kubeconfig

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), constants.HelmNamespace,
		"secret", logger.Debugf); err != nil {
		return nil, err
	}
	return actionConfig, nil
}

func newHelmInstallAction(config *action.Configuration, release release, timeout time.Duration) *action.Install {
	action := action.NewInstall(config)
	action.Namespace = constants.HelmNamespace
	action.Timeout = timeout
	action.ReleaseName = release.releaseName
	setWaitMode(action, release.waitMode)
	return action
}

func setWaitMode(a *action.Install, waitMode WaitMode) {
	switch waitMode {
	case WaitModeNone:
		a.Wait = false
		a.Atomic = false
	case WaitModeWait:
		a.Wait = true
		a.Atomic = false
	case WaitModeAtomic:
		a.Wait = true
		a.Atomic = true
	default:
		panic(fmt.Errorf("unknown wait mode %q", waitMode))
	}
}

// installAction is an action that installs a helm chart.
type installAction struct {
	preInstall  func(context.Context) error
	release     release
	helmAction  *action.Install
	postInstall func(context.Context) error
	log         debugLog
}

// Apply installs the chart.
func (a *installAction) Apply(ctx context.Context) error {
	if a.preInstall != nil {
		if err := a.preInstall(ctx); err != nil {
			return err
		}
	}
	if err := retryApply(ctx, a, applyRetryInterval, a.log); err != nil {
		return err
	}

	if a.postInstall != nil {
		if err := a.postInstall(ctx); err != nil {
			return err
		}
	}
	return nil
}

// SaveChart saves the chart to the given directory under the `install/<chart-name>` subdirectory.
func (a *installAction) SaveChart(chartsDir string, fileHandler file.Handler) error {
	return saveChart(a.release, chartsDir, fileHandler)
}

func (a *installAction) apply(ctx context.Context) error {
	_, err := a.helmAction.RunWithContext(ctx, a.release.chart, a.release.values)
	return err
}

// ReleaseName returns the release name.
func (a *installAction) ReleaseName() string {
	return a.release.releaseName
}

// IsAtomic returns true if the action is atomic.
func (a *installAction) IsAtomic() bool {
	return a.helmAction.Atomic
}

func newHelmUpgradeAction(config *action.Configuration, timeout time.Duration) *action.Upgrade {
	action := action.NewUpgrade(config)
	action.Namespace = constants.HelmNamespace
	action.Timeout = timeout
	action.ReuseValues = false
	action.Atomic = true
	return action
}

// upgradeAction is an action that upgrades a helm chart.
type upgradeAction struct {
	preUpgrade  func(context.Context) error
	postUpgrade func(context.Context) error
	release     release
	helmAction  *action.Upgrade
	log         debugLog
}

// Apply installs the chart.
func (a *upgradeAction) Apply(ctx context.Context) error {
	if a.preUpgrade != nil {
		if err := a.preUpgrade(ctx); err != nil {
			return err
		}
	}
	if err := retryApply(ctx, a, applyRetryInterval, a.log); err != nil {
		return err
	}
	if a.postUpgrade != nil {
		if err := a.postUpgrade(ctx); err != nil {
			return err
		}
	}
	return nil
}

// SaveChart saves the chart to the given directory under the `upgrade/<chart-name>` subdirectory.
func (a *upgradeAction) SaveChart(chartsDir string, fileHandler file.Handler) error {
	return saveChart(a.release, chartsDir, fileHandler)
}

func (a *upgradeAction) apply(ctx context.Context) error {
	_, err := a.helmAction.RunWithContext(ctx, a.release.releaseName, a.release.chart, a.release.values)
	return err
}

// ReleaseName returns the release name.
func (a *upgradeAction) ReleaseName() string {
	return a.release.releaseName
}

// IsAtomic returns true if the action is atomic.
func (a *upgradeAction) IsAtomic() bool {
	return a.helmAction.Atomic
}

func saveChart(release release, chartsDir string, fileHandler file.Handler) error {
	if err := saveChartToDisk(release.chart, chartsDir, fileHandler); err != nil {
		return fmt.Errorf("saving chart %s to %q: %w", release.releaseName, chartsDir, err)
	}
	if err := fileHandler.WriteYAML(filepath.Join(chartsDir, release.chart.Metadata.Name, "overrides.yaml"), release.values); err != nil {
		return fmt.Errorf("saving override values for chart %s to %q: %w", release.releaseName, filepath.Join(chartsDir, release.chart.Metadata.Name), err)
	}

	return nil
}
