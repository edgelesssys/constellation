/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

// Release bundles all information necessary to create a helm release.
type Release struct {
	Chart       []byte
	Values      map[string]interface{}
	ReleaseName string
	Wait        bool
}

type Releases struct {
	Cilium                Release
	ConstellationServices Release
}
