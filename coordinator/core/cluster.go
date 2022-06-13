package core

import (
	"context"
	"time"

	"github.com/edgelesssys/constellation/coordinator/pubapi/pubproto"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/constants"
	"go.uber.org/zap"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// GetK8sJoinArgs returns the args needed by a Node to join the cluster.
func (c *Core) GetK8sJoinArgs(ctx context.Context) (*kubeadm.BootstrapTokenDiscovery, error) {
	return c.kube.GetJoinToken(ctx, constants.KubernetesJoinTokenTTL)
}

// GetK8SCertificateKey returns the key needed by a Coordinator to join the cluster.
func (c *Core) GetK8SCertificateKey(ctx context.Context) (string, error) {
	return c.kube.GetKubeadmCertificateKey(ctx)
}

// InitCluster initializes the cluster, stores the join args, and returns the kubeconfig.
func (c *Core) InitCluster(ctx context.Context, autoscalingNodeGroups []string, cloudServiceAccountURI string, masterSecret []byte, sshUsers []*pubproto.SSHUserKey) ([]byte, error) {
	c.zaplogger.Info("Initializing cluster")
	vpnIP, err := c.GetVPNIP()
	if err != nil {
		c.zaplogger.Error("Retrieving vpn ip failed", zap.Error(err))
		return nil, err
	}

	// Convert SSH users map from protobuffer to map
	sshUsersMap := make(map[string]string)
	if len(sshUsers) > 0 {
		for _, value := range sshUsers {
			sshUsersMap[value.Username] = value.PublicKey
		}
	}
	if err := c.kube.InitCluster(ctx, autoscalingNodeGroups, cloudServiceAccountURI, vpnIP, masterSecret, sshUsersMap); err != nil {
		c.zaplogger.Error("Initializing cluster failed", zap.Error(err))
		return nil, err
	}

	kubeconfig, err := c.kube.GetKubeconfig()
	if err != nil {
		return nil, err
	}

	if err := c.data().PutKubernetesConfig(kubeconfig); err != nil {
		return nil, err
	}

	// set role in cloud provider metadata for autoconfiguration
	if c.metadata.Supported() {
		if err := c.metadata.SignalRole(context.TODO(), role.Coordinator); err != nil {
			c.zaplogger.Info("unable to update role in cloud provider metadata", zap.Error(err))
		}
	}

	return kubeconfig, nil
}

// JoinCluster lets a Node join the cluster.
func (c *Core) JoinCluster(ctx context.Context, args *kubeadm.BootstrapTokenDiscovery, certKey string, peerRole role.Role) error {
	c.zaplogger.Info("Joining Kubernetes cluster")
	nodeVPNIP, err := c.vpn.GetInterfaceIP()
	if err != nil {
		c.zaplogger.Error("Retrieving vpn ip failed", zap.Error(err))
		return err
	}

	// we need to pass the VPNIP for another control-plane, otherwise etcd will bind itself to the wrong IP address and fails
	if err := c.kube.JoinCluster(ctx, args, nodeVPNIP, certKey, peerRole); err != nil {
		c.zaplogger.Error("Joining Kubernetes cluster failed", zap.Error(err))
		return err
	}
	c.zaplogger.Info("Joined Kubernetes cluster")
	// set role in cloud provider metadata for autoconfiguration
	if c.metadata.Supported() {
		if err := c.metadata.SignalRole(context.TODO(), peerRole); err != nil {
			c.zaplogger.Info("unable to update role in cloud provider metadata", zap.Error(err))
		}
	}

	return nil
}

// Cluster manages the overall cluster lifecycle (init, join).
type Cluster interface {
	// InitCluster bootstraps a new cluster with the current node being the master, returning the arguments required to join the cluster.
	InitCluster(ctx context.Context, autoscalingNodeGroups []string, cloudServiceAccountURI, vpnIP string, masterSecret []byte, sshUsers map[string]string) error
	// JoinCluster will join the current node to an existing cluster.
	JoinCluster(ctx context.Context, args *kubeadm.BootstrapTokenDiscovery, nodeVPNIP, certKey string, peerRole role.Role) error
	// GetKubeconfig reads the kubeconfig from the filesystem. Only succeeds after cluster is initialized.
	GetKubeconfig() ([]byte, error)
	// GetKubeadmCertificateKey returns the 64-byte hex string key needed to join the cluster as control-plane. This function must be executed on a control-plane.
	GetKubeadmCertificateKey(ctx context.Context) (string, error)
	// GetJoinToken returns a bootstrap (join) token.
	GetJoinToken(ctx context.Context, ttl time.Duration) (*kubeadm.BootstrapTokenDiscovery, error)
	// StartKubelet starts the kubelet service.
	StartKubelet() error
}

// ClusterFake behaves like a real cluster, but does not actually initialize or join Kubernetes.
type ClusterFake struct{}

// InitCluster fakes bootstrapping a new cluster with the current node being the master, returning the arguments required to join the cluster.
func (c *ClusterFake) InitCluster(ctx context.Context, autoscalingNodeGroups []string, cloudServiceAccountURI, vpnIP string, masterSecret []byte, sshUsers map[string]string) error {
	return nil
}

// JoinCluster will fake joining the current node to an existing cluster.
func (c *ClusterFake) JoinCluster(ctx context.Context, args *kubeadm.BootstrapTokenDiscovery, nodeVPNIP, certKey string, peerRole role.Role) error {
	return nil
}

// GetKubeconfig fakes reading the kubeconfig from the filesystem. Only succeeds after cluster is initialized.
func (c *ClusterFake) GetKubeconfig() ([]byte, error) {
	return []byte("kubeconfig"), nil
}

// GetKubeadmCertificateKey fakes generating a certificateKey.
func (c *ClusterFake) GetKubeadmCertificateKey(context.Context) (string, error) {
	return "controlPlaneCertficateKey", nil
}

// GetJoinToken returns a bootstrap (join) token.
func (c *ClusterFake) GetJoinToken(ctx context.Context, _ time.Duration) (*kubeadm.BootstrapTokenDiscovery, error) {
	return &kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: "0.0.0.0",
		Token:             "kube-fake-token",
		CACertHashes:      []string{"sha256:a60ebe9b0879090edd83b40a4df4bebb20506bac1e51d518ff8f4505a721930f"},
	}, nil
}

// StartKubelet starts the kubelet service.
func (c *ClusterFake) StartKubelet() error {
	return nil
}
