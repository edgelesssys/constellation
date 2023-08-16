/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/constants"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type applyAction interface {
	Apply(context.Context) error
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

func newHelmInstallAction(config *action.Configuration, release Release) *action.Install {
	action := action.NewInstall(config)
	action.Namespace = constants.HelmNamespace
	action.Timeout = timeout
	action.ReleaseName = release.ReleaseName
	setWaitMode(action, release.WaitMode)
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

// newInstallAction creates a new InstallAction.
func newInstallAction(config *action.Configuration, release Release, log debugLog) *installAction {
	return &installAction{helmAction: newHelmInstallAction(config, release), release: release, log: log}
}

// installAction is an action that installs a helm chart.
type installAction struct {
	preInstall  func(context.Context) error
	release     Release
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
	if err := retryApply(ctx, a, a.log); err != nil {
		return err
	}

	if a.postInstall != nil {
		if err := a.postInstall(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *installAction) apply(ctx context.Context) error {
	_, err := a.helmAction.RunWithContext(ctx, a.release.Chart, a.release.Values)
	return err
}

// ReleaseName returns the release name.
func (a *installAction) ReleaseName() string {
	return a.release.ReleaseName
}

// IsAtomic returns true if the action is atomic.
func (a *installAction) IsAtomic() bool {
	return a.helmAction.Atomic
}

func newHelmUpgradeAction(config *action.Configuration) *action.Upgrade {
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
	release     Release
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
	if err := retryApply(ctx, a, a.log); err != nil {
		return err
	}
	if a.postUpgrade != nil {
		if err := a.postUpgrade(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (a *upgradeAction) apply(ctx context.Context) error {
	_, err := a.helmAction.RunWithContext(ctx, a.release.ReleaseName, a.release.Chart, a.release.Values)
	return err
}

// ReleaseName returns the release name.
func (a *upgradeAction) ReleaseName() string {
	return a.release.ReleaseName
}

// IsAtomic returns true if the action is atomic.
func (a *upgradeAction) IsAtomic() bool {
	return a.helmAction.Atomic
}
