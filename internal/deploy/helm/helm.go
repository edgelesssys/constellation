/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package helm provides types and functions shared across services.
package helm

// Release bundles all information necessary to create a helm release.
type Release struct {
	Chart       []byte
	Values      map[string]any
	ReleaseName string
	WaitMode    WaitMode
}

// Releases bundles all helm releases to be deployed to Constellation.
type Releases struct {
	Cilium                Release
	CertManager           Release
	Operators             Release
	ConstellationServices Release
}

// MergeMaps returns a new map that is the merger of it's inputs.
// Key collisions are resolved by taking the value of the second argument (map b).
// Taken from: https://github.com/helm/helm/blob/dbc6d8e20fe1d58d50e6ed30f09a04a77e4c68db/pkg/cli/values/options.go#L91-L108.
func MergeMaps(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]any); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]any); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// WaitMode specifies the wait mode for a helm release.
type WaitMode string

const (
	// WaitModeNone specifies that the helm release should not wait for the resources to be ready.
	WaitModeNone WaitMode = ""
	// WaitModeWait specifies that the helm release should wait for the resources to be ready.
	WaitModeWait WaitMode = "wait"
	// WaitModeAtomic specifies that the helm release should
	// wait for the resources to be ready and roll back atomically on failure.
	WaitModeAtomic WaitMode = "atomic"
)
