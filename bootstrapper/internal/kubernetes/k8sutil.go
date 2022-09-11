/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"net"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/internal/kubernetes"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/role"
	"github.com/edgelesssys/constellation/internal/versions"
)

type clusterUtil interface {
	InstallComponents(ctx context.Context, version versions.ValidK8sVersion) error
	InitCluster(ctx context.Context, initConfig []byte, nodeName string, ips []net.IP, controlPlaneEndpoint string, nodeCIDR string, log *logger.Logger) error
	JoinCluster(ctx context.Context, joinConfig []byte, peerRole role.Role, controlPlaneEndpoint string, nodeCIDR string, log *logger.Logger) error
	SetupHelmDeployments(ctx context.Context, client k8sapi.Client, helmDeployments []byte, in k8sapi.SetupPodNetworkInput, log *logger.Logger) error
	SetupAccessManager(kubectl k8sapi.Client, sshUsers kubernetes.Marshaler) error
	SetupAutoscaling(kubectl k8sapi.Client, clusterAutoscalerConfiguration kubernetes.Marshaler, secrets kubernetes.Marshaler) error
	SetupJoinService(kubectl k8sapi.Client, joinServiceConfiguration kubernetes.Marshaler) error
	SetupCloudControllerManager(kubectl k8sapi.Client, cloudControllerManagerConfiguration kubernetes.Marshaler, configMaps kubernetes.Marshaler, secrets kubernetes.Marshaler) error
	SetupCloudNodeManager(kubectl k8sapi.Client, cloudNodeManagerConfiguration kubernetes.Marshaler) error
	SetupKonnectivity(kubectl k8sapi.Client, konnectivityAgentsDaemonSet kubernetes.Marshaler) error
	SetupKMS(kubectl k8sapi.Client, kmsConfiguration kubernetes.Marshaler) error
	SetupVerificationService(kubectl k8sapi.Client, verificationServiceConfiguration kubernetes.Marshaler) error
	SetupGCPGuestAgent(kubectl k8sapi.Client, gcpGuestAgentConfiguration kubernetes.Marshaler) error
	SetupOperatorLifecycleManager(ctx context.Context, kubectl k8sapi.Client, olmCRDs, olmConfiguration kubernetes.Marshaler, crdNames []string) error
	SetupNodeMaintenanceOperator(kubectl k8sapi.Client, nodeMaintenanceOperatorConfiguration kubernetes.Marshaler) error
	SetupNodeOperator(ctx context.Context, kubectl k8sapi.Client, nodeOperatorConfiguration kubernetes.Marshaler) error
	StartKubelet() error
	FixCilium(log *logger.Logger)
}
