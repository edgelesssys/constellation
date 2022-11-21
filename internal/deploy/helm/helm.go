/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

// Release bundles all information necessary to create a helm release.
type Release struct {
	Chart       []byte
	Values      map[string]any
	ReleaseName string
	Wait        bool
}

// Releases bundles all helm releases to be deployed to Constellation.
type Releases struct {
	Cilium                Release
	CertManager           Release
	Operators             Release
	ConstellationServices Release
}
