/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package helm provides a higher level interface to the Helm Go SDK.

It is used by the CLI to:

  - load embedded charts
  - install charts
  - update helm releases
  - get versions for installed helm releases
  - create local backups before running service upgrades

The charts themselves are embedded in the CLI binary, and values are dynamically updated depending on configuration.
The charts can be found in “./charts/“.
Values should be added in the chart's "values.yaml“ file if they are static i.e. don't depend on user input,
otherwise they need to be dynamically created depending on a user's configuration.

Helm logic should not be implemented outside this package.
All values loading, parsing, installing, uninstalling, and updating of charts should be implemented here.
As such, the helm package requires to implement some CSP specific logic.
However, exported functions should be CSP agnostic and take a cloudprovider.Provider as argument.
As such, the number of exported functions should be kept minimal.
*/
package helm

import (
	"context"
	"fmt"

	"github.com/edgelesssys/constellation/v2/cli/internal/state"
	"github.com/edgelesssys/constellation/v2/internal/config"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/kms/uri"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
	"github.com/edgelesssys/constellation/v2/internal/semver"
)

const (
	// AllowDestructive is a named bool to signal that destructive actions have been confirmed by the user.
	AllowDestructive = true
	// DenyDestructive is a named bool to signal that destructive actions have not been confirmed by the user yet.
	DenyDestructive = false
)

type debugLog interface {
	Debugf(format string, args ...any)
}

// Client is a Helm client to apply charts.
type Client struct {
	factory    *actionFactory
	cliVersion semver.Semver
	log        debugLog
}

// NewClient returns a new Helm client.
func NewClient(kubeConfigPath string, log debugLog) (*Client, error) {
	kubeClient, err := kubectl.NewFromConfig(kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("initializing kubectl: %w", err)
	}
	actionConfig, err := newActionConfig(kubeConfigPath, log)
	if err != nil {
		return nil, fmt.Errorf("creating action config: %w", err)
	}
	lister := ReleaseVersionClient{actionConfig}
	cliVersion := constants.BinaryVersion()
	factory := newActionFactory(kubeClient, lister, actionConfig, log)
	return &Client{factory, cliVersion, log}, nil
}

// Options are options for loading charts.
type Options struct {
	Conformance      bool
	HelmWaitMode     WaitMode
	AllowDestructive bool
	Force            bool
}

// PrepareApply loads the charts and returns the executor to apply them.
// TODO(elchead): remove validK8sVersion by putting ValidK8sVersion into config.Config, see AB#3374.
func (h Client) PrepareApply(
	conf *config.Config, stateFile *state.State,
	flags Options, serviceAccURI string, masterSecret uri.MasterSecret,
) (Applier, bool, error) {
	releases, err := h.loadReleases(conf, masterSecret, stateFile, flags, serviceAccURI)
	if err != nil {
		return nil, false, fmt.Errorf("loading Helm releases: %w", err)
	}
	h.log.Debugf("Loaded Helm releases")
	actions, includesUpgrades, err := h.factory.GetActions(releases, conf.MicroserviceVersion, flags.Force, flags.AllowDestructive)
	return &ChartApplyExecutor{actions: actions, log: h.log}, includesUpgrades, err
}

func (h Client) loadReleases(
	conf *config.Config, secret uri.MasterSecret,
	stateFile *state.State, flags Options, serviceAccURI string,
) ([]Release, error) {
	helmLoader := newLoader(conf, stateFile, h.cliVersion)
	h.log.Debugf("Created new Helm loader")
	return helmLoader.loadReleases(flags.Conformance, flags.HelmWaitMode, secret, serviceAccURI)
}

// Applier runs the Helm actions.
type Applier interface {
	Apply(ctx context.Context) error
	SaveCharts(chartsDir string, fileHandler file.Handler) error
}

// ChartApplyExecutor is a Helm action executor that applies all actions.
type ChartApplyExecutor struct {
	actions []applyAction
	log     debugLog
}

// Apply applies the charts in order.
func (c ChartApplyExecutor) Apply(ctx context.Context) error {
	for _, action := range c.actions {
		c.log.Debugf("Applying %q", action.ReleaseName())
		if err := action.Apply(ctx); err != nil {
			return fmt.Errorf("applying %s: %w", action.ReleaseName(), err)
		}
	}
	return nil
}

// SaveCharts saves all Helm charts and their values to the given directory.
func (c ChartApplyExecutor) SaveCharts(chartsDir string, fileHandler file.Handler) error {
	for _, action := range c.actions {
		if err := action.SaveChart(chartsDir, fileHandler); err != nil {
			return fmt.Errorf("saving chart %s: %w", action.ReleaseName(), err)
		}
	}
	return nil
}

// mergeMaps returns a new map that is the merger of it's inputs.
// Key collisions are resolved by taking the value of the second argument (map b).
// Taken from: https://github.com/helm/helm/blob/dbc6d8e20fe1d58d50e6ed30f09a04a77e4c68db/pkg/cli/values/options.go#L91-L108.
func mergeMaps(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]any); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]any); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
