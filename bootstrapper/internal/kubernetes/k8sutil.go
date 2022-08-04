package kubernetes

import (
	"context"
	"net"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/versions"
)

type clusterUtil interface {
	InstallComponents(ctx context.Context, version versions.ValidK8sVersion) error
	InitCluster(ctx context.Context, initConfig []byte, nodeName string, ips []net.IP, log *logger.Logger) error
	JoinCluster(ctx context.Context, joinConfig []byte, log *logger.Logger) error
	SetupPodNetwork(context.Context, k8sapi.SetupPodNetworkInput, k8sapi.Client) error
	SetupAccessManager(kubectl k8sapi.Client, sshUsers resources.Marshaler) error
	SetupAutoscaling(kubectl k8sapi.Client, clusterAutoscalerConfiguration resources.Marshaler, secrets resources.Marshaler) error
	SetupJoinService(kubectl k8sapi.Client, joinServiceConfiguration resources.Marshaler) error
	SetupCloudControllerManager(kubectl k8sapi.Client, cloudControllerManagerConfiguration resources.Marshaler, configMaps resources.Marshaler, secrets resources.Marshaler) error
	SetupCloudNodeManager(kubectl k8sapi.Client, cloudNodeManagerConfiguration resources.Marshaler) error
	SetupKMS(kubectl k8sapi.Client, kmsConfiguration resources.Marshaler) error
	SetupVerificationService(kubectl k8sapi.Client, verificationServiceConfiguration resources.Marshaler) error
	SetupGCPGuestAgent(kubectl k8sapi.Client, gcpGuestAgentConfiguration resources.Marshaler) error
	SetupOperatorLifecycleManager(ctx context.Context, kubectl k8sapi.Client, olmCRDs, olmConfiguration resources.Marshaler, crdNames []string) error
	SetupNodeMaintenanceOperator(kubectl k8sapi.Client, nodeMaintenanceOperatorConfiguration resources.Marshaler) error
	SetupNodeOperator(ctx context.Context, kubectl k8sapi.Client, nodeOperatorConfiguration resources.Marshaler) error
	StartKubelet() error
	RestartKubelet() error
	FixCilium(nodeNameK8s string, log *logger.Logger)
}
