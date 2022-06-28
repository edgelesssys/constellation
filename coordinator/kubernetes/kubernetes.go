package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/coordinator/role"
	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
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
	}
}

// InitCluster initializes a new Kubernetes cluster and applies pod network provider.
func (k *KubeWrapper) InitCluster(
	ctx context.Context, autoscalingNodeGroups []string, cloudServiceAccountURI, vpnIP string, id attestationtypes.ID, masterSecret []byte, sshUsers map[string]string,
) error {
	// TODO: k8s version should be user input
	if err := k.clusterUtil.InstallComponents(context.TODO(), "1.23.6"); err != nil {
		return err
	}

	nodeName := vpnIP
	var providerID string
	var instance cloudtypes.Instance
	var publicIP string
	var nodePodCIDR string
	var subnetworkPodCIDR string
	// this is the IP in "kubeadm init --control-plane-endpoint=<IP/DNS>:<port>" hence the unfortunate name
	var controlPlaneEndpointIP string
	var nodeIP string
	var err error

	// Step 1: retrieve cloud metadata for Kubernetes configuration
	if k.providerMetadata.Supported() {
		instance, err = k.providerMetadata.Self(context.TODO())
		if err != nil {
			return fmt.Errorf("retrieving own instance metadata failed: %w", err)
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
		subnetworkPodCIDR, err = k.providerMetadata.GetSubnetworkCIDR(context.TODO())
		if err != nil {
			return fmt.Errorf("retrieving subnetwork CIDR failed: %w", err)
		}
		controlPlaneEndpointIP = publicIP
		if k.providerMetadata.SupportsLoadBalancer() {
			controlPlaneEndpointIP, err = k.providerMetadata.GetLoadBalancerIP(context.TODO())
			if err != nil {
				return fmt.Errorf("retrieving load balancer IP failed: %w", err)
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
		return fmt.Errorf("encoding kubeadm init configuration as YAML: %w", err)
	}
	if err := k.clusterUtil.InitCluster(ctx, initConfigYAML); err != nil {
		return fmt.Errorf("kubeadm init: %w", err)
	}
	kubeConfig, err := k.GetKubeconfig()
	if err != nil {
		return fmt.Errorf("reading kubeconfig after cluster initialization: %w", err)
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
		return fmt.Errorf("setting up pod network: %w", err)
	}

	kms := resources.NewKMSDeployment(masterSecret)
	if err = k.clusterUtil.SetupKMS(k.client, kms); err != nil {
		return fmt.Errorf("setting up kms: %w", err)
	}

	if err := k.setupActivationService(k.cloudProvider, k.initialMeasurementsJSON, id); err != nil {
		return fmt.Errorf("setting up activation service failed: %w", err)
	}

	if err := k.setupCCM(context.TODO(), vpnIP, subnetworkPodCIDR, cloudServiceAccountURI, instance); err != nil {
		return fmt.Errorf("setting up cloud controller manager: %w", err)
	}
	if err := k.setupCloudNodeManager(); err != nil {
		return fmt.Errorf("setting up cloud node manager: %w", err)
	}

	if err := k.setupClusterAutoscaler(instance, cloudServiceAccountURI, autoscalingNodeGroups); err != nil {
		return fmt.Errorf("setting up cluster autoscaler: %w", err)
	}

	accessManager := resources.NewAccessManagerDeployment(sshUsers)
	if err := k.clusterUtil.SetupAccessManager(k.client, accessManager); err != nil {
		return fmt.Errorf("failed to setup access-manager: %w", err)
	}

	if err := k.clusterUtil.SetupVerificationService(
		k.client, resources.NewVerificationDaemonSet(k.cloudProvider),
	); err != nil {
		return fmt.Errorf("failed to setup verification service: %w", err)
	}

	go k.clusterUtil.FixCilium(nodeName)

	return nil
}

// JoinCluster joins existing Kubernetes cluster.
func (k *KubeWrapper) JoinCluster(ctx context.Context, args *kubeadm.BootstrapTokenDiscovery, nodeVPNIP, certKey string, peerRole role.Role) error {
	// TODO: k8s version should be user input
	if err := k.clusterUtil.InstallComponents(context.TODO(), "1.23.6"); err != nil {
		return err
	}

	// Step 1: retrieve cloud metadata for Kubernetes configuration
	var providerID string
	nodeName := nodeVPNIP
	nodeInternalIP := nodeVPNIP
	if k.providerMetadata.Supported() {
		instance, err := k.providerMetadata.Self(context.TODO())
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

	if k.cloudControllerManager.Supported() && k.providerMetadata.Supported() {
		if err := k.prepareInstanceForCCM(context.TODO(), nodeVPNIP); err != nil {
			return fmt.Errorf("preparing node for CCM failed: %w", err)
		}
	}

	// Step 2: configure kubeadm join config

	joinConfig := k.configProvider.JoinConfiguration(k.cloudControllerManager.Supported())
	joinConfig.SetApiServerEndpoint(args.APIServerEndpoint)
	joinConfig.SetToken(args.Token)
	joinConfig.AppendDiscoveryTokenCaCertHash(args.CACertHashes[0])
	joinConfig.SetNodeIP(nodeInternalIP)
	joinConfig.SetNodeName(nodeName)
	joinConfig.SetProviderID(providerID)
	if peerRole == role.Coordinator {
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
	// replace the cluster.Server endpoint (127.0.0.1:16443) in admin.conf with the first coordinator endpoint (10.118.0.1:6443)
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

func (k *KubeWrapper) setupCCM(ctx context.Context, vpnIP, subnetworkPodCIDR, cloudServiceAccountURI string, instance cloudtypes.Instance) error {
	if !k.cloudControllerManager.Supported() {
		return nil
	}
	if err := k.prepareInstanceForCCM(context.TODO(), vpnIP); err != nil {
		return fmt.Errorf("preparing node for CCM failed: %w", err)
	}
	ccmConfigMaps, err := k.cloudControllerManager.ConfigMaps(instance)
	if err != nil {
		return fmt.Errorf("defining ConfigMaps for CCM failed: %w", err)
	}
	ccmSecrets, err := k.cloudControllerManager.Secrets(ctx, instance, cloudServiceAccountURI)
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

func (k *KubeWrapper) setupClusterAutoscaler(instance cloudtypes.Instance, cloudServiceAccountURI string, autoscalingNodeGroups []string) error {
	if !k.clusterAutoscaler.Supported() {
		return nil
	}
	caSecrets, err := k.clusterAutoscaler.Secrets(instance, cloudServiceAccountURI)
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

// prepareInstanceForCCM sets the vpn IP in cloud provider metadata.
func (k *KubeWrapper) prepareInstanceForCCM(ctx context.Context, vpnIP string) error {
	if err := k.providerMetadata.SetVPNIP(ctx, vpnIP); err != nil {
		return fmt.Errorf("setting VPN IP for cloud-controller-manager failed: %w", err)
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
