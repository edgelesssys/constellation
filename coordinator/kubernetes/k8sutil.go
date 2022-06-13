package kubernetes

import (
	"context"
	"time"

	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

type clusterUtil interface {
	InstallComponents(ctx context.Context, version string) error
	InitCluster(ctx context.Context, initConfig []byte) error
	JoinCluster(ctx context.Context, joinConfig []byte) error
	SetupPodNetwork(context.Context, k8sapi.SetupPodNetworkInput) error
	SetupAccessManager(kubectl k8sapi.Client, sshUsers resources.Marshaler) error
	SetupAutoscaling(kubectl k8sapi.Client, clusterAutoscalerConfiguration resources.Marshaler, secrets resources.Marshaler) error
	SetupCloudControllerManager(kubectl k8sapi.Client, cloudControllerManagerConfiguration resources.Marshaler, configMaps resources.Marshaler, secrets resources.Marshaler) error
	SetupCloudNodeManager(kubectl k8sapi.Client, cloudNodeManagerConfiguration resources.Marshaler) error
	SetupKMS(kubectl k8sapi.Client, kmsConfiguration resources.Marshaler) error
	StartKubelet() error
	RestartKubelet() error
	GetControlPlaneJoinCertificateKey(ctx context.Context) (string, error)
	CreateJoinToken(ctx context.Context, ttl time.Duration) (*kubeadm.BootstrapTokenDiscovery, error)
}
