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
	AWSLoadBalancerController *Release
	CSI                       *Release
	Cilium                    Release
	CertManager               Release
	ConstellationOperators    Release
	ConstellationServices     Release
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
