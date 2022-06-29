package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/bootstrapper/util"
	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/spf13/afero"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// configReader provides kubeconfig as []byte.
type configReader interface {
	ReadKubeconfig() ([]byte, error)
}

// configurationProvider provides kubeadm init and join configuration.
type configurationProvider interface {
	InitConfiguration(externalCloudProvider bool) k8sapi.KubeadmInitYAML
	JoinConfiguration(externalCloudProvider bool) k8sapi.KubeadmJoinYAML
}

// KubeWrapper implements Cluster interface.
type KubeWrapper struct {
	cloudProvider           string
	clusterUtil             clusterUtil
	configProvider          configurationProvider
	client                  k8sapi.Client
	kubeconfigReader        configReader
	cloudControllerManager  CloudControllerManager
	cloudNodeManager        CloudNodeManager
	clusterAutoscaler       ClusterAutoscaler
	providerMetadata        ProviderMetadata
	initialMeasurementsJSON []byte
	getIPAddr               func() (string, error)
}

// New creates a new KubeWrapper with real values.
func New(cloudProvider string, clusterUtil clusterUtil, configProvider configurationProvider, client k8sapi.Client, cloudControllerManager CloudControllerManager,
	cloudNodeManager CloudNodeManager, clusterAutoscaler ClusterAutoscaler, providerMetadata ProviderMetadata, initialMeasurementsJSON []byte,
) *KubeWrapper {
	return &KubeWrapper{
		cloudProvider:           cloudProvider,
		clusterUtil:             clusterUtil,
		configProvider:          configProvider,
		client:                  client,
		kubeconfigReader:        &KubeconfigReader{fs: afero.Afero{Fs: afero.NewOsFs()}},
		cloudControllerManager:  cloudControllerManager,
		cloudNodeManager:        cloudNodeManager,
		clusterAutoscaler:       clusterAutoscaler,
		providerMetadata:        providerMetadata,
		initialMeasurementsJSON: initialMeasurementsJSON,
		getIPAddr:               util.GetIPAddr,
	}
}

type KMSConfig struct {
	MasterSecret       []byte
	KMSURI             string
	StorageURI         string
	KeyEncryptionKeyID string
	UseExistingKEK     bool
}

// InitCluster initializes a new Kubernetes cluster and applies pod network provider.
func (k *KubeWrapper) InitCluster(
	ctx context.Context, autoscalingNodeGroups []string, cloudServiceAccountURI, k8sVersion string,
	id attestationtypes.ID, kmsConfig KMSConfig, sshUsers map[string]string,
) ([]byte, error) {
	// TODO: k8s version should be user input
	if err := k.clusterUtil.InstallComponents(ctx, k8sVersion); err != nil {
		return nil, err
	}

	ip, err := k.getIPAddr()
	if err != nil {
		return nil, err
	}
	nodeName := ip
	var providerID string
	var instance metadata.InstanceMetadata
	var publicIP string
	var nodePodCIDR string
	var subnetworkPodCIDR string
	// this is the IP in "kubeadm init --control-plane-endpoint=<IP/DNS>:<port>" hence the unfortunate name
	var controlPlaneEndpointIP string
	var nodeIP string

	// Step 1: retrieve cloud metadata for Kubernetes configuration
	if k.providerMetadata.Supported() {
		instance, err = k.providerMetadata.Self(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving own instance metadata failed: %w", err)
		}
		nodeName = k8sCompliantHostname(instance.Name)
		providerID = instance.ProviderID
		if len(instance.PrivateIPs) > 0 {
			nodeIP = instance.PrivateIPs[0]
		}
		if len(instance.PublicIPs) > 0 {
			publicIP = instance.PublicIPs[0]
		}
		if len(instance.AliasIPRanges) > 0 {
			nodePodCIDR = instance.AliasIPRanges[0]
		}
		subnetworkPodCIDR, err = k.providerMetadata.GetSubnetworkCIDR(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving subnetwork CIDR failed: %w", err)
		}
		controlPlaneEndpointIP = publicIP
		if k.providerMetadata.SupportsLoadBalancer() {
			controlPlaneEndpointIP, err = k.providerMetadata.GetLoadBalancerIP(ctx)
			if err != nil {
				return nil, fmt.Errorf("retrieving load balancer IP failed: %w", err)
			}
		}
	}

	// Step 2: configure kubeadm init config
	initConfig := k.configProvider.InitConfiguration(k.cloudControllerManager.Supported())
	initConfig.SetNodeIP(nodeIP)
	initConfig.SetCertSANs([]string{publicIP, nodeIP})
	initConfig.SetNodeName(nodeName)
	initConfig.SetProviderID(providerID)
	initConfig.SetControlPlaneEndpoint(controlPlaneEndpointIP)
	initConfigYAML, err := initConfig.Marshal()
	if err != nil {
		return nil, fmt.Errorf("encoding kubeadm init configuration as YAML: %w", err)
	}
	if err := k.clusterUtil.InitCluster(ctx, initConfigYAML); err != nil {
		return nil, fmt.Errorf("kubeadm init: %w", err)
	}
	kubeConfig, err := k.GetKubeconfig()
	if err != nil {
		return nil, fmt.Errorf("reading kubeconfig after cluster initialization: %w", err)
	}
	k.client.SetKubeconfig(kubeConfig)

	// Step 3: configure & start kubernetes controllers

	setupPodNetworkInput := k8sapi.SetupPodNetworkInput{
		CloudProvider:     k.cloudProvider,
		NodeName:          nodeName,
		FirstNodePodCIDR:  nodePodCIDR,
		SubnetworkPodCIDR: subnetworkPodCIDR,
		ProviderID:        providerID,
	}
	if err = k.clusterUtil.SetupPodNetwork(ctx, setupPodNetworkInput); err != nil {
		return nil, fmt.Errorf("setting up pod network: %w", err)
	}

	kms := resources.NewKMSDeployment(k.cloudProvider, kmsConfig.MasterSecret)
	if err = k.clusterUtil.SetupKMS(k.client, kms); err != nil {
		return nil, fmt.Errorf("setting up kms: %w", err)
	}

	if err := k.setupActivationService(k.cloudProvider, k.initialMeasurementsJSON, id); err != nil {
		return nil, fmt.Errorf("setting up activation service failed: %w", err)
	}

	if err := k.setupCCM(ctx, subnetworkPodCIDR, cloudServiceAccountURI, instance); err != nil {
		return nil, fmt.Errorf("setting up cloud controller manager: %w", err)
	}
	if err := k.setupCloudNodeManager(); err != nil {
		return nil, fmt.Errorf("setting up cloud node manager: %w", err)
	}

	if err := k.setupClusterAutoscaler(instance, cloudServiceAccountURI, autoscalingNodeGroups); err != nil {
		return nil, fmt.Errorf("setting up cluster autoscaler: %w", err)
	}

	accessManager := resources.NewAccessManagerDeployment(sshUsers)
	if err := k.clusterUtil.SetupAccessManager(k.client, accessManager); err != nil {
		return nil, fmt.Errorf("failed to setup access-manager: %w", err)
	}

	if err := k.clusterUtil.SetupVerificationService(
		k.client, resources.NewVerificationDaemonSet(k.cloudProvider),
	); err != nil {
		return nil, fmt.Errorf("failed to setup verification service: %w", err)
	}

	go k.clusterUtil.FixCilium(nodeName)

	return k.GetKubeconfig()
}

// JoinCluster joins existing Kubernetes cluster.
func (k *KubeWrapper) JoinCluster(ctx context.Context, args *kubeadm.BootstrapTokenDiscovery, certKey string, peerRole role.Role) error {
	// TODO: k8s version should be user input
	if err := k.clusterUtil.InstallComponents(ctx, "1.23.6"); err != nil {
		return err
	}

	// Step 1: retrieve cloud metadata for Kubernetes configuration
	nodeInternalIP, err := k.getIPAddr()
	if err != nil {
		return err
	}
	nodeName := nodeInternalIP
	var providerID string
	if k.providerMetadata.Supported() {
		instance, err := k.providerMetadata.Self(ctx)
		if err != nil {
			return fmt.Errorf("retrieving own instance metadata failed: %w", err)
		}
		providerID = instance.ProviderID
		nodeName = instance.Name
		if len(instance.PrivateIPs) > 0 {
			nodeInternalIP = instance.PrivateIPs[0]
		}
	}
	nodeName = k8sCompliantHostname(nodeName)

	// Step 2: configure kubeadm join config

	joinConfig := k.configProvider.JoinConfiguration(k.cloudControllerManager.Supported())
	joinConfig.SetApiServerEndpoint(args.APIServerEndpoint)
	joinConfig.SetToken(args.Token)
	joinConfig.AppendDiscoveryTokenCaCertHash(args.CACertHashes[0])
	joinConfig.SetNodeIP(nodeInternalIP)
	joinConfig.SetNodeName(nodeName)
	joinConfig.SetProviderID(providerID)
	if peerRole == role.ControlPlane {
		joinConfig.SetControlPlane(nodeInternalIP, certKey)
	}
	joinConfigYAML, err := joinConfig.Marshal()
	if err != nil {
		return fmt.Errorf("encoding kubeadm join configuration as YAML: %w", err)
	}
	if err := k.clusterUtil.JoinCluster(ctx, joinConfigYAML); err != nil {
		return fmt.Errorf("joining cluster: %v; %w ", string(joinConfigYAML), err)
	}

	go k.clusterUtil.FixCilium(nodeName)

	return nil
}

// GetKubeconfig returns the current nodes kubeconfig of stored on disk.
func (k *KubeWrapper) GetKubeconfig() ([]byte, error) {
	kubeconf, err := k.kubeconfigReader.ReadKubeconfig()
	if err != nil {
		return nil, err
	}
	// replace the cluster.Server endpoint (127.0.0.1:16443) in admin.conf with the first bootstrapper endpoint (10.118.0.1:6443)
	// kube-api server listens on 10.118.0.1:6443
	// 127.0.0.1:16443 is the high availability balancer nginx endpoint, runnining localy on all nodes
	// alternatively one could also start a local high availability balancer.
	return []byte(strings.ReplaceAll(string(kubeconf), "127.0.0.1:16443", "10.118.0.1:6443")), nil
}

// GetKubeadmCertificateKey return the key needed to join the Cluster as Control-Plane (has to be executed on a control-plane; errors otherwise).
func (k *KubeWrapper) GetKubeadmCertificateKey(ctx context.Context) (string, error) {
	return k.clusterUtil.GetControlPlaneJoinCertificateKey(ctx)
}

// GetJoinToken returns a bootstrap (join) token.
func (k *KubeWrapper) GetJoinToken(ctx context.Context, ttl time.Duration) (*kubeadm.BootstrapTokenDiscovery, error) {
	return k.clusterUtil.CreateJoinToken(ctx, ttl)
}

func (k *KubeWrapper) setupActivationService(csp string, measurementsJSON []byte, id attestationtypes.ID) error {
	idJSON, err := json.Marshal(id)
	if err != nil {
		return err
	}

	activationConfiguration := resources.NewActivationDaemonset(csp, string(measurementsJSON), string(idJSON))

	return k.clusterUtil.SetupActivationService(k.client, activationConfiguration)
}

func (k *KubeWrapper) setupCCM(ctx context.Context, subnetworkPodCIDR, cloudServiceAccountURI string, instance metadata.InstanceMetadata) error {
	if !k.cloudControllerManager.Supported() {
		return nil
	}
	ccmConfigMaps, err := k.cloudControllerManager.ConfigMaps(instance)
	if err != nil {
		return fmt.Errorf("defining ConfigMaps for CCM failed: %w", err)
	}
	ccmSecrets, err := k.cloudControllerManager.Secrets(ctx, instance.ProviderID, cloudServiceAccountURI)
	if err != nil {
		return fmt.Errorf("defining Secrets for CCM failed: %w", err)
	}

	cloudControllerManagerConfiguration := resources.NewDefaultCloudControllerManagerDeployment(
		k.cloudControllerManager.Name(), k.cloudControllerManager.Image(), k.cloudControllerManager.Path(), subnetworkPodCIDR,
		k.cloudControllerManager.ExtraArgs(), k.cloudControllerManager.Volumes(), k.cloudControllerManager.VolumeMounts(), k.cloudControllerManager.Env(),
	)
	if err := k.clusterUtil.SetupCloudControllerManager(k.client, cloudControllerManagerConfiguration, ccmConfigMaps, ccmSecrets); err != nil {
		return fmt.Errorf("failed to setup cloud-controller-manager: %w", err)
	}

	return nil
}

func (k *KubeWrapper) setupCloudNodeManager() error {
	if !k.cloudNodeManager.Supported() {
		return nil
	}
	cloudNodeManagerConfiguration := resources.NewDefaultCloudNodeManagerDeployment(
		k.cloudNodeManager.Image(), k.cloudNodeManager.Path(), k.cloudNodeManager.ExtraArgs(),
	)
	if err := k.clusterUtil.SetupCloudNodeManager(k.client, cloudNodeManagerConfiguration); err != nil {
		return fmt.Errorf("failed to setup cloud-node-manager: %w", err)
	}

	return nil
}

func (k *KubeWrapper) setupClusterAutoscaler(instance metadata.InstanceMetadata, cloudServiceAccountURI string, autoscalingNodeGroups []string) error {
	if !k.clusterAutoscaler.Supported() {
		return nil
	}
	caSecrets, err := k.clusterAutoscaler.Secrets(instance.ProviderID, cloudServiceAccountURI)
	if err != nil {
		return fmt.Errorf("defining Secrets for cluster-autoscaler failed: %w", err)
	}

	clusterAutoscalerConfiguration := resources.NewDefaultAutoscalerDeployment(k.clusterAutoscaler.Volumes(), k.clusterAutoscaler.VolumeMounts(), k.clusterAutoscaler.Env())
	clusterAutoscalerConfiguration.SetAutoscalerCommand(k.clusterAutoscaler.Name(), autoscalingNodeGroups)
	if err := k.clusterUtil.SetupAutoscaling(k.client, clusterAutoscalerConfiguration, caSecrets); err != nil {
		return fmt.Errorf("failed to setup cluster-autoscaler: %w", err)
	}

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

// StartKubelet starts the kubelet service.
func (k *KubeWrapper) StartKubelet() error {
	return k.clusterUtil.StartKubelet()
}

type fakeK8SClient struct {
	kubeconfig []byte
}

// NewFakeK8SClient creates a new, fake k8s client where apply always works.
func NewFakeK8SClient([]byte) (k8sapi.Client, error) {
	return &fakeK8SClient{}, nil
}

// Apply fakes applying Kubernetes resources.
func (f *fakeK8SClient) Apply(resources resources.Marshaler, forceConflicts bool) error {
	return nil
}

// SetKubeconfig stores the kubeconfig given to it.
func (f *fakeK8SClient) SetKubeconfig(kubeconfig []byte) {
	f.kubeconfig = kubeconfig
}
