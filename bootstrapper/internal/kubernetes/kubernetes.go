package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/bootstrapper/role"
	"github.com/edgelesssys/constellation/bootstrapper/util"
	attestationtypes "github.com/edgelesssys/constellation/internal/attestation/types"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/spf13/afero"
	"go.uber.org/zap"
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
	id attestationtypes.ID, kmsConfig KMSConfig, sshUsers map[string]string, logger *zap.Logger,
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
	var controlPlaneEndpointIP string // this is the IP in "kubeadm init --control-plane-endpoint=<IP/DNS>:<port>" hence the unfortunate name
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
			if k.cloudProvider == "gcp" {
				if err := manuallySetLoadbalancerIP(ctx, controlPlaneEndpointIP); err != nil {
					return nil, fmt.Errorf("setting load balancer IP failed: %w", err)
				}
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
	if err := k.clusterUtil.InitCluster(ctx, initConfigYAML, logger); err != nil {
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

	if err := k.setupJoinService(k.cloudProvider, k.initialMeasurementsJSON, id); err != nil {
		return nil, fmt.Errorf("setting up join service failed: %w", err)
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

	if k.cloudProvider == "gcp" {
		if err := k.clusterUtil.SetupGCPGuestAgent(k.client, resources.NewGCPGuestAgentDaemonset()); err != nil {
			return nil, fmt.Errorf("failed to setup gcp guest agent: %w", err)
		}
	}

	k.clusterUtil.FixCilium(nodeName)

	return k.GetKubeconfig()
}

// JoinCluster joins existing Kubernetes cluster.
func (k *KubeWrapper) JoinCluster(ctx context.Context, args *kubeadm.BootstrapTokenDiscovery, peerRole role.Role, logger *zap.Logger) error {
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
	joinConfig.SetAPIServerEndpoint(args.APIServerEndpoint)
	joinConfig.SetToken(args.Token)
	joinConfig.AppendDiscoveryTokenCaCertHash(args.CACertHashes[0])
	joinConfig.SetNodeIP(nodeInternalIP)
	joinConfig.SetNodeName(nodeName)
	joinConfig.SetProviderID(providerID)
	if peerRole == role.ControlPlane {
		joinConfig.SetControlPlane(nodeInternalIP)
	}
	joinConfigYAML, err := joinConfig.Marshal()
	if err != nil {
		return fmt.Errorf("encoding kubeadm join configuration as YAML: %w", err)
	}
	if err := k.clusterUtil.JoinCluster(ctx, joinConfigYAML, logger); err != nil {
		return fmt.Errorf("joining cluster: %v; %w ", string(joinConfigYAML), err)
	}

	k.clusterUtil.FixCilium(nodeName)

	return nil
}

// GetKubeconfig returns the current nodes kubeconfig of stored on disk.
func (k *KubeWrapper) GetKubeconfig() ([]byte, error) {
	return k.kubeconfigReader.ReadKubeconfig()
}

func (k *KubeWrapper) setupJoinService(csp string, measurementsJSON []byte, id attestationtypes.ID) error {
	idJSON, err := json.Marshal(id)
	if err != nil {
		return err
	}

	joinConfiguration := resources.NewJoinServiceDaemonset(csp, string(measurementsJSON), string(idJSON))

	return k.clusterUtil.SetupJoinService(k.client, joinConfiguration)
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

// manuallySetLoadbalancerIP sets the loadbalancer IP of the first control plane during init.
// The GCP guest agent does this usually, but is deployed in the cluster that doesn't exist
// at this point. This is a workaround to set the loadbalancer IP manually, so kubeadm and kubelet
// can talk to the local Kubernetes API server using the loadbalancer IP.
func manuallySetLoadbalancerIP(ctx context.Context, ip string) error {
	// https://github.com/GoogleCloudPlatform/guest-agent/blob/792fce795218633bcbde505fb3457a0b24f26d37/google_guest_agent/addresses.go#L179
	if !strings.Contains(ip, "/") {
		ip = ip + "/32"
	}
	args := fmt.Sprintf("route add to local %s scope host dev ens3 proto 66", ip)
	_, err := exec.CommandContext(ctx, "ip", strings.Split(args, " ")...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("ip route add (code %v) with: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return fmt.Errorf("ip route add: %w", err)
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
