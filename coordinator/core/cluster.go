package core

import (
	"context"
	"strings"

	"github.com/edgelesssys/constellation/coordinator/kubernetes"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/coordinator/role"
	"go.uber.org/zap"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// GetK8sJoinArgs returns the args needed by a Node to join the cluster.
func (c *Core) GetK8sJoinArgs() (*kubeadm.BootstrapTokenDiscovery, error) {
	return c.data().GetKubernetesJoinArgs()
}

// InitCluster initializes the cluster, stores the join args, and returns the kubeconfig.
func (c *Core) InitCluster(autoscalingNodeGroups []string, cloudServiceAccountURI string) ([]byte, error) {
	var nodeName string
	var providerID string
	var instance Instance
	var ccmConfigMaps resources.ConfigMaps
	var ccmSecrets resources.Secrets
	var err error
	nodeIP := coordinatorVPNIP.String()
	var err error
	if c.metadata.Supported() {
		instance, err = c.metadata.Self(context.TODO())
		if err != nil {
			c.zaplogger.Error("Retrieving own instance metadata failed", zap.Error(err))
			return nil, err
		}
		nodeName = instance.Name
		providerID = instance.ProviderID
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

	c.zaplogger.Info("Initializing cluster")
	joinCommand, err := c.kube.InitCluster(kubernetes.InitClusterInput{
		APIServerAdvertiseIP:               coordinatorVPNIP.String(),
		NodeIP:                             nodeIP,
		NodeName:                           k8sCompliantHostname(nodeName),
		ProviderID:                         providerID,
		SupportClusterAutoscaler:           c.clusterAutoscaler.Supported(),
		AutoscalingCloudprovider:           c.clusterAutoscaler.Name(),
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
	})
	if err != nil {
		c.zaplogger.Error("Initializing cluster failed", zap.Error(err))
		return nil, err
	}

	if err := c.data().PutKubernetesJoinArgs(joinCommand); err != nil {
		c.zaplogger.Error("Storing kubernetes join command failed", zap.Error(err))
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
func (c *Core) JoinCluster(args kubeadm.BootstrapTokenDiscovery) error {
	c.zaplogger.Info("Joining kubernetes cluster")
	nodeVPNIP, err := c.vpn.GetInterfaceIP()
	if err != nil {
		c.zaplogger.Error("Retrieving vpn ip failed", zap.Error(err))
		return err
	}
	var nodeName string
	var providerID string
	if c.metadata.Supported() {
		instance, err := c.metadata.Self(context.TODO())
		if err != nil {
			c.zaplogger.Error("Retrieving own instance metadata failed", zap.Error(err))
			return err
		}
		providerID = instance.ProviderID
		nodeName = instance.Name
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

	if err := c.kube.JoinCluster(&args, k8sCompliantHostname(nodeName), nodeVPNIP, providerID); err != nil {
		c.zaplogger.Error("Joining kubernetes cluster failed", zap.Error(err))
		return err
	}
	c.zaplogger.Info("Joined kubernetes cluster")
	// set role in cloud provider metadata for autoconfiguration
	if c.metadata.Supported() {
		if err := c.metadata.SignalRole(context.TODO(), role.Node); err != nil {
			c.zaplogger.Info("unable to update role in cloud provider metadata", zap.Error(err))
		}
	}

	return nil
}

// Cluster manages the overall cluster lifecycle (init, join).
type Cluster interface {
	// InitCluster bootstraps a new cluster with the current node being the master, returning the arguments required to join the cluster.
	InitCluster(kubernetes.InitClusterInput) (*kubeadm.BootstrapTokenDiscovery, error)
	// JoinCluster will join the current node to an existing cluster.
	JoinCluster(args *kubeadm.BootstrapTokenDiscovery, nodeName, nodeIP, providerID string) error
	// GetKubeconfig reads the kubeconfig from the filesystem. Only succeeds after cluster is initialized.
	GetKubeconfig() ([]byte, error)
}

// ClusterFake behaves like a real cluster, but does not actually initialize or join kubernetes.
type ClusterFake struct{}

// InitCluster fakes bootstrapping a new cluster with the current node being the master, returning the arguments required to join the cluster.
func (c *ClusterFake) InitCluster(kubernetes.InitClusterInput) (*kubeadm.BootstrapTokenDiscovery, error) {
	return &kubeadm.BootstrapTokenDiscovery{
		APIServerEndpoint: "0.0.0.0",
		Token:             "kube-fake-token",
		CACertHashes:      []string{"sha256:a60ebe9b0879090edd83b40a4df4bebb20506bac1e51d518ff8f4505a721930f"},
	}, nil
}

// JoinCluster will fake joining the current node to an existing cluster.
func (c *ClusterFake) JoinCluster(args *kubeadm.BootstrapTokenDiscovery, nodeName, nodeIP, providerID string) error {
	return nil
}

// GetKubeconfig fakes reading the kubeconfig from the filesystem. Only succeeds after cluster is initialized.
func (c *ClusterFake) GetKubeconfig() ([]byte, error) {
	return []byte("kubeconfig"), nil
}

// k8sCompliantHostname transforms a hostname to an RFC 1123 compliant, lowercase subdomain as required by kubernetes node names.
// The following regex is used by k8s for validation: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/ .
// Only a simple heuristic is used for now (to lowercase, replace underscores).
func k8sCompliantHostname(in string) string {
	hostname := strings.ToLower(in)
	hostname = strings.ReplaceAll(hostname, "_", "-")
	return hostname
}
