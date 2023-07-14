/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"net"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
)

type clusterUtil interface {
	InstallComponents(ctx context.Context, kubernetesComponents components.Components) error
	InitCluster(ctx context.Context, initConfig []byte, nodeName, clusterName string, ips []net.IP, controlPlaneEndpoint string, conformanceMode bool, log *logger.Logger) ([]byte, error)
	JoinCluster(ctx context.Context, joinConfig []byte, peerRole role.Role, controlPlaneEndpoint string, log *logger.Logger) error
	WaitForCilium(ctx context.Context, log *logger.Logger) error
	FixCilium(ctx context.Context) error
	StartKubelet() error
}

// helmClient bundles functions related to microservice deployment. Only microservices that can be deployed purely via Helm are deployed with this interface.
// Currently only a subset of microservices is deployed via Helm.
// Naming is inspired by Helm.
type helmClient interface {
	InstallCilium(context.Context, k8sapi.Client, helm.Release, k8sapi.SetupPodNetworkInput) error
	InstallCertManager(ctx context.Context, release helm.Release) error
	InstallOperators(ctx context.Context, release helm.Release, extraVals map[string]any) error
	InstallConstellationServices(ctx context.Context, release helm.Release, extraVals map[string]any) error
	InstallAWSLoadBalancerController(context.Context, k8sapi.Client, helm.Release) error
}
