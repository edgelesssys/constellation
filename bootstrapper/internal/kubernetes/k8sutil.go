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
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions"
)

type clusterUtil interface {
	InstallComponents(ctx context.Context, version versions.ValidK8sVersion) error
	InitCluster(ctx context.Context, initConfig []byte, nodeName string, ips []net.IP, controlPlaneEndpoint string, conformanceMode bool, log *logger.Logger) error
	JoinCluster(ctx context.Context, joinConfig []byte, peerRole role.Role, controlPlaneEndpoint string, log *logger.Logger) error
	SetupAccessManager(kubectl k8sapi.Client, sshUsers kubernetes.Marshaler) error
	SetupAutoscaling(kubectl k8sapi.Client, clusterAutoscalerConfiguration kubernetes.Marshaler, secrets kubernetes.Marshaler) error
	SetupJoinService(kubectl k8sapi.Client, joinServiceConfiguration kubernetes.Marshaler) error
	SetupCloudControllerManager(kubectl k8sapi.Client, cloudControllerManagerConfiguration kubernetes.Marshaler, configMaps kubernetes.Marshaler, secrets kubernetes.Marshaler) error
	SetupCloudNodeManager(kubectl k8sapi.Client, cloudNodeManagerConfiguration kubernetes.Marshaler) error
	SetupKonnectivity(kubectl k8sapi.Client, konnectivityAgentsDaemonSet kubernetes.Marshaler) error
	SetupVerificationService(kubectl k8sapi.Client, verificationServiceConfiguration kubernetes.Marshaler) error
	SetupGCPGuestAgent(kubectl k8sapi.Client, gcpGuestAgentConfiguration kubernetes.Marshaler) error
	SetupOperatorLifecycleManager(ctx context.Context, kubectl k8sapi.Client, olmCRDs, olmConfiguration kubernetes.Marshaler, crdNames []string) error
	SetupNodeMaintenanceOperator(kubectl k8sapi.Client, nodeMaintenanceOperatorConfiguration kubernetes.Marshaler) error
	SetupNodeOperator(ctx context.Context, kubectl k8sapi.Client, nodeOperatorConfiguration kubernetes.Marshaler) error
	FixCilium(log *logger.Logger)
	StartKubelet() error
}

// helmClient bundles functions related to microservice deployment. Only microservices that can be deployed purely via Helm are deployed with this interface.
// Currently only a subset of microservices is deployed via Helm.
// Naming is inspired by Helm.
type helmClient interface {
	InstallCilium(context.Context, k8sapi.Client, helm.Release, k8sapi.SetupPodNetworkInput) error
	InstallConstellationServices(ctx context.Context, release helm.Release) error
}
