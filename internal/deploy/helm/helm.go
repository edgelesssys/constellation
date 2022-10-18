/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import "helm.sh/helm/v3/pkg/chart"

// Release bundles all information necessary to create a helm release.
type Release struct {
	Chart       *chart.Chart
	Values      map[string]interface{}
	ReleaseName string
	Wait        bool
}

type Releases struct {
	Cilium Release
	KMS    Release
}
