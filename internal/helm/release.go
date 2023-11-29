/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package helm provides types and functions shared across services.
package helm

import "helm.sh/helm/v3/pkg/chart"

// Release bundles all information necessary to create a helm release.
type release struct {
	chart       *chart.Chart
	values      map[string]any
	releaseName string
	waitMode    WaitMode
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
