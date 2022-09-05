/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import "helm.sh/helm/v3/pkg/chart"

type Deployment struct {
	Chart  *chart.Chart
	Values map[string]interface{}
}

type Deployments struct {
	Cilium Deployment
}
