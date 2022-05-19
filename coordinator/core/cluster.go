package core

import (
	"context"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/coordinator/kubernetes"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/coordinator/role"
	"github.com/edgelesssys/constellation/internal/constants"
	"go.uber.org/zap"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// GetK8sJoinArgs returns the args needed by a Node to join the cluster.
func (c *Core) GetK8sJoinArgs() (*kubeadm.BootstrapTokenDiscovery, error) {
	return c.kube.GetJoinToken(constants.KubernetesJoinTokenTTL)
}

// GetK8SCertificateKey returns the key needed by a Coordinator to join the cluster.
func (c *Core) GetK8SCertificateKey() (string, error) {
	return c.kube.GetKubeadmCertificateKey()
}

// InitCluster initializes the cluster, stores the join args, and returns the kubeconfig.
func (c *Core) InitCluster(autoscalingNodeGroups []string, cloudServiceAccountURI string) ([]byte, error) {
	var nodeName string
	var providerID string
	var instance Instance
	var ccmConfigMaps resources.ConfigMaps
	var ccmSecrets resources.Secrets
	var caSecrets resources.Secrets
	var err error
	nodeIP := coordinatorVPNIP.String()
	if c.metadata.Supported() {
		instance, err = c.metadata.Self(context.TODO())
		if err != nil {
			c.zaplogger.Error("Retrieving own instance metadata failed", zap.Error(err))
			return nil, err
		}
		nodeName = instance.Name
		providerID = instance.ProviderID
		if len(instance.IPs) > 0 {
			nodeIP = instance.IPs[0]
		}
	} else {
		nodeName = coordinatorVPNIP.String()
	}
	if c.cloudControllerManager.Supported() && c.metadata.Supported() {
		c.zaplogger.Info("Preparing node for cloud-controller-manager")
		if err := PrepareInstanceForCCM(context.TODO(), c.metadata, c.cloudControllerManager, coordinatorVPNIP.String()); err != nil {
			c.zaplogger.Error("Preparing node for CCM failed", zap.Error(err))
			return nil, err
		}
		ccmConfigMaps, err = c.cloudControllerManager.ConfigMaps(instance)
		if err != nil {
			c.zaplogger.Error("Defining ConfigMaps for CCM failed", zap.Error(err))
			return nil, err
		}
		ccmSecrets, err = c.cloudControllerManager.Secrets(instance, cloudServiceAccountURI)
		if err != nil {
			c.zaplogger.Error("Defining Secrets for CCM failed", zap.Error(err))
			return nil, err
		}
	}
	if c.clusterAutoscaler.Supported() {
		caSecrets, err = c.clusterAutoscaler.Secrets(instance, cloudServiceAccountURI)
		if err != nil {
			c.zaplogger.Error("Defining Secrets for cluster-autoscaler failed", zap.Error(err))
			return nil, err
		}
	}

	c.zaplogger.Info("Initializing cluster")
	if err := c.kube.InitCluster(kubernetes.InitClusterInput{
		APIServerAdvertiseIP:               coordinatorVPNIP.String(),
		NodeIP:                             nodeIP,
		NodeName:                           k8sCompliantHostname(nodeName),
		ProviderID:                         providerID,
		SupportClusterAutoscaler:           c.clusterAutoscaler.Supported(),
		AutoscalingCloudprovider:           c.clusterAutoscaler.Name(),
		AutoscalingSecrets:                 caSecrets,
		AutoscalingVolumes:                 c.clusterAutoscaler.Volumes(),
		AutoscalingVolumeMounts:            c.clusterAutoscaler.VolumeMounts(),
		AutoscalingEnv:                     c.clusterAutoscaler.Env(),
		AutoscalingNodeGroups:              autoscalingNodeGroups,
		SupportsCloudControllerManager:     c.cloudControllerManager.Supported(),
		CloudControllerManagerName:         c.cloudControllerManager.Name(),
		CloudControllerManagerImage:        c.cloudControllerManager.Image(),
		CloudControllerManagerPath:         c.cloudControllerManager.Path(),
		CloudControllerManagerExtraArgs:    c.cloudControllerManager.ExtraArgs(),
		CloudControllerManagerConfigMaps:   ccmConfigMaps,
		CloudControllerManagerSecrets:      ccmSecrets,
		CloudControllerManagerVolumes:      c.cloudControllerManager.Volumes(),
		CloudControllerManagerVolumeMounts: c.cloudControllerManager.VolumeMounts(),
		CloudControllerManagerEnv:          c.cloudControllerManager.Env(),
		SupportsCloudNodeManager:           c.cloudNodeManager.Supported(),
		CloudNodeManagerImage:              c.cloudNodeManager.Image(),
		CloudNodeManagerPath:               c.cloudNodeManager.Path(),
		CloudNodeManagerExtraArgs:          c.cloudNodeManager.ExtraArgs(),
	}); err != nil {
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
func (c *Core) JoinCluster(args *kubeadm.BootstrapTokenDiscovery, certKey string, peerRole role.Role) error {
	c.zaplogger.Info("Joining Kubernetes cluster")
	nodeVPNIP, err := c.vpn.GetInterfaceIP()
	if err != nil {
		c.zaplogger.Error("Retrieving vpn ip failed", zap.Error(err))
		return err
	}
	var nodeName string
	var providerID string
	nodeIP := nodeVPNIP
	if c.metadata.Supported() {
		instance, err := c.metadata.Self(context.TODO())
		if err != nil {
			c.zaplogger.Error("Retrieving own instance metadata failed", zap.Error(err))
			return err
		}
		providerID = instance.ProviderID
		nodeName = instance.Name
		if len(instance.IPs) > 0 {
			nodeIP = instance.IPs[0]
		}
	} else {
		nodeName = nodeVPNIP
	}
	if c.cloudControllerManager.Supported() && c.metadata.Supported() {
		c.zaplogger.Info("Preparing node for cloud-controller-manager")
		if err := PrepareInstanceForCCM(context.TODO(), c.metadata, c.cloudControllerManager, nodeVPNIP); err != nil {
			c.zaplogger.Error("Preparing node for CCM failed", zap.Error(err))
			return err
		}
	}

	c.zaplogger.Info("k8s Join data", zap.String("nodename", nodeName), zap.String("nodeIP", nodeIP), zap.String("nodeVPNIP", nodeVPNIP), zap.String("provid", providerID))
	// we need to pass the VPNIP for another control-plane, otherwise etcd will bind itself to the wrong IP address and fails
	if err := c.kube.JoinCluster(args, k8sCompliantHostname(nodeName), nodeIP, nodeVPNIP, providerID, certKey, c.cloudControllerManager.Supported(), peerRole); err != nil {
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
	InitCluster(kubernetes.InitClusterInput) error
	// JoinCluster will join the current node to an existing cluster.
	JoinCluster(args *kubeadm.BootstrapTokenDiscovery, nodeName, nodeIP, nodeVPNIP, providerID, certKey string, ccmSupported bool, peerRole role.Role) error
	// GetKubeconfig reads the kubeconfig from the filesystem. Only succeeds after cluster is initialized.
	GetKubeconfig() ([]byte, error)
	// GetKubeadmCertificateKey returns the 64-byte hex string key needed to join the cluster as control-plane. This function must be executed on a control-plane.
	GetKubeadmCertificateKey() (string, error)
	// GetJoinToken returns a bootstrap (join) token.
	GetJoinToken(ttl time.Duration) (*kubeadm.BootstrapTokenDiscovery, error)
	// StartKubelet starts the kubelet service.
	StartKubelet() error
}

// ClusterFake behaves like a real cluster, but does not actually initialize or join Kubernetes.
type ClusterFake struct{}

// InitCluster fakes bootstrapping a new cluster with the current node being the master, returning the arguments required to join the cluster.
func (c *ClusterFake) InitCluster(kubernetes.InitClusterInput) error {
	return nil
}

// JoinCluster will fake joining the current node to an existing cluster.
func (c *ClusterFake) JoinCluster(args *kubeadm.BootstrapTokenDiscovery, nodeName, nodeIP, nodeVPNIP, providerID, certKey string, _ bool, _ role.Role) error {
	return nil
}

// GetKubeconfig fakes reading the kubeconfig from the filesystem. Only succeeds after cluster is initialized.
func (c *ClusterFake) GetKubeconfig() ([]byte, error) {
	return []byte("kubeconfig"), nil
}

// GetKubeadmCertificateKey fakes generating a certificateKey.
func (c *ClusterFake) GetKubeadmCertificateKey() (string, error) {
	return "controlPlaneCertficateKey", nil
}

// GetJoinToken returns a bootstrap (join) token.
func (c *ClusterFake) GetJoinToken(_ time.Duration) (*kubeadm.BootstrapTokenDiscovery, error) {
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

// k8sCompliantHostname transforms a hostname to an RFC 1123 compliant, lowercase subdomain as required by Kubernetes node names.
// The following regex is used by k8s for validation: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/ .
// Only a simple heuristic is used for now (to lowercase, replace underscores).
func k8sCompliantHostname(in string) string {
	hostname := strings.ToLower(in)
	hostname = strings.ReplaceAll(hostname, "_", "-")
	return hostname
}
