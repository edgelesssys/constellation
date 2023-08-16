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
)

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

// ApplyCharts applies the given releases in the given order.
func ApplyCharts(ctx context.Context, releases ReleaseApplyOrder, kubeConfigPath string, force, allowDestructive bool, log debugLog) error {
	factory, err := newActionFactory(kubeConfigPath, log)
	if err != nil {
		return fmt.Errorf("creating Helm action factory: %w", err)
	}
	actions, _, err := factory.newActions(releases, force, allowDestructive)
	if err != nil {
		return fmt.Errorf("creating Helm actions: %w", err)
	}
	for _, action := range actions {
		log.Debugf("Applying %q", action.ReleaseName())
		if err := action.Apply(ctx); err != nil {
			return fmt.Errorf("applying %s: %w", action.ReleaseName(), err)
		}
	}
	return nil
}
