package helm

import "helm.sh/helm/v3/pkg/chart"

type Deployment struct {
	Chart  *chart.Chart
	Values map[string]interface{}
}

type Deployments struct {
	Cilium Deployment
}
