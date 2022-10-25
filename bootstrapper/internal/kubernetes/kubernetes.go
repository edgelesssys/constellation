/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// configReader provides kubeconfig as []byte.
type configReader interface {
	ReadKubeconfig() ([]byte, error)
}

// configurationProvider provides kubeadm init and join configuration.
type configurationProvider interface {
	InitConfiguration(externalCloudProvider bool, k8sVersion versions.ValidK8sVersion) k8sapi.KubeadmInitYAML
	JoinConfiguration(externalCloudProvider bool) k8sapi.KubeadmJoinYAML
}

// KubeWrapper implements Cluster interface.
type KubeWrapper struct {
	cloudProvider           string
	clusterUtil             clusterUtil
	helmClient              helmClient
	configProvider          configurationProvider
	client                  k8sapi.Client
	kubeconfigReader        configReader
	cloudControllerManager  CloudControllerManager
	clusterAutoscaler       ClusterAutoscaler
	providerMetadata        ProviderMetadata
	initialMeasurementsJSON []byte
	getIPAddr               func() (string, error)
}

// New creates a new KubeWrapper with real values.
func New(cloudProvider string, clusterUtil clusterUtil, configProvider configurationProvider, client k8sapi.Client, cloudControllerManager CloudControllerManager,
	clusterAutoscaler ClusterAutoscaler, providerMetadata ProviderMetadata, initialMeasurementsJSON []byte, helmClient helmClient,
) *KubeWrapper {
	return &KubeWrapper{
		cloudProvider:           cloudProvider,
		clusterUtil:             clusterUtil,
		helmClient:              helmClient,
		configProvider:          configProvider,
		client:                  client,
		kubeconfigReader:        &KubeconfigReader{fs: afero.Afero{Fs: afero.NewOsFs()}},
		cloudControllerManager:  cloudControllerManager,
		clusterAutoscaler:       clusterAutoscaler,
		providerMetadata:        providerMetadata,
		initialMeasurementsJSON: initialMeasurementsJSON,
		getIPAddr:               getIPAddr,
	}
}

// InitCluster initializes a new Kubernetes cluster and applies pod network provider.
func (k *KubeWrapper) InitCluster(
	ctx context.Context, cloudServiceAccountURI, versionString string, measurementSalt []byte, enforcedPCRs []uint32,
	enforceIDKeyDigest bool, idKeyDigest []byte, azureCVM bool, sshUsers map[string]string,
	helmReleasesRaw []byte, conformanceMode bool, log *logger.Logger,
) ([]byte, error) {
	k8sVersion, err := versions.NewValidK8sVersion(versionString)
	if err != nil {
		return nil, err
	}
	log.With(zap.String("version", string(k8sVersion))).Infof("Installing Kubernetes components")
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
	var nodePodCIDR string
	var subnetworkPodCIDR string
	var controlPlaneEndpoint string // this is the endpoint in "kubeadm init --control-plane-endpoint=<IP/DNS>:<port>"
	var nodeIP string
	var validIPs []net.IP

	// Step 1: retrieve cloud metadata for Kubernetes configuration
	if k.providerMetadata.Supported() {
		log.Infof("Retrieving node metadata")
		instance, err = k.providerMetadata.Self(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving own instance metadata: %w", err)
		}
		if instance.VPCIP != "" {
			validIPs = append(validIPs, net.ParseIP(instance.VPCIP))
		}
		nodeName = k8sCompliantHostname(instance.Name)
		providerID = instance.ProviderID
		nodeIP = instance.VPCIP
		subnetworkPodCIDR = instance.SecondaryIPRange

		if len(instance.AliasIPRanges) > 0 {
			nodePodCIDR = instance.AliasIPRanges[0]
		}
		controlPlaneEndpoint, err = k.providerMetadata.GetLoadBalancerEndpoint(ctx)
		if err != nil {
			return nil, fmt.Errorf("retrieving load balancer endpoint: %w", err)
		}
	}
	log.With(
		zap.String("nodeName", nodeName),
		zap.String("providerID", providerID),
		zap.String("nodeIP", nodeIP),
		zap.String("controlPlaneEndpoint", controlPlaneEndpoint),
		zap.String("podCIDR", subnetworkPodCIDR),
	).Infof("Setting information for node")

	// Step 2: configure kubeadm init config
	initConfig := k.configProvider.InitConfiguration(k.cloudControllerManager.Supported(), k8sVersion)
	initConfig.SetNodeIP(nodeIP)
	initConfig.SetCertSANs([]string{nodeIP})
	initConfig.SetNodeName(nodeName)
	initConfig.SetProviderID(providerID)
	initConfig.SetControlPlaneEndpoint(controlPlaneEndpoint)
	initConfigYAML, err := initConfig.Marshal()
	if err != nil {
		return nil, fmt.Errorf("encoding kubeadm init configuration as YAML: %w", err)
	}
	log.Infof("Initializing Kubernetes cluster")
	if err := k.clusterUtil.InitCluster(ctx, initConfigYAML, nodeName, validIPs, controlPlaneEndpoint, conformanceMode, log); err != nil {
		return nil, fmt.Errorf("kubeadm init: %w", err)
	}
	kubeConfig, err := k.GetKubeconfig()
	if err != nil {
		return nil, fmt.Errorf("reading kubeconfig after cluster initialization: %w", err)
	}
	k.client.SetKubeconfig(kubeConfig)

	// Step 3: configure & start kubernetes controllers
	log.Infof("Starting Kubernetes controllers and deployments")
	setupPodNetworkInput := k8sapi.SetupPodNetworkInput{
		CloudProvider:        k.cloudProvider,
		NodeName:             nodeName,
		FirstNodePodCIDR:     nodePodCIDR,
		SubnetworkPodCIDR:    subnetworkPodCIDR,
		LoadBalancerEndpoint: controlPlaneEndpoint,
	}

	var helmReleases helm.Releases
	if err := json.Unmarshal(helmReleasesRaw, &helmReleases); err != nil {
		return nil, fmt.Errorf("unmarshalling helm releases: %w", err)
	}

	if err = k.helmClient.InstallCilium(ctx, k.client, helmReleases.Cilium, setupPodNetworkInput); err != nil {
		return nil, fmt.Errorf("installing pod network: %w", err)
	}

	var controlPlaneIP string
	if strings.Contains(controlPlaneEndpoint, ":") {
		controlPlaneIP, _, err = net.SplitHostPort(controlPlaneEndpoint)
		if err != nil {
			return nil, fmt.Errorf("parsing control plane endpoint: %w", err)
		}
	} else {
		controlPlaneIP = controlPlaneEndpoint
	}
	if err = k.clusterUtil.SetupKonnectivity(k.client, resources.NewKonnectivityAgents(controlPlaneIP)); err != nil {
		return nil, fmt.Errorf("setting up konnectivity: %w", err)
	}

	extraVals, err := k.setupExtraVals(ctx, k.initialMeasurementsJSON, idKeyDigest, measurementSalt, subnetworkPodCIDR, cloudServiceAccountURI)
	if err != nil {
		return nil, fmt.Errorf("setting up extraVals: %w", err)
	}

	if err = k.helmClient.InstallConstellationServices(ctx, helmReleases.ConstellationServices, extraVals); err != nil {
		return nil, fmt.Errorf("installing constellation-services: %w", err)
	}

	if err := k.setupInternalConfigMap(ctx, strconv.FormatBool(azureCVM)); err != nil {
		return nil, fmt.Errorf("failed to setup internal ConfigMap: %w", err)
	}

	if err := k.setupClusterAutoscaler(instance, cloudServiceAccountURI, k8sVersion); err != nil {
		return nil, fmt.Errorf("setting up cluster autoscaler: %w", err)
	}

	// TODO: remove access manager or re-enable with support for readonly /etc
	// accessManager := resources.NewAccessManagerDeployment(sshUsers)
	// if err := k.clusterUtil.SetupAccessManager(k.client, accessManager); err != nil {
	// 	return nil, fmt.Errorf("failed to setup access-manager: %w", err)
	// }

	if err := k.clusterUtil.SetupVerificationService(
		k.client, resources.NewVerificationDaemonSet(k.cloudProvider, controlPlaneEndpoint),
	); err != nil {
		return nil, fmt.Errorf("failed to setup verification service: %w", err)
	}

	if err := k.setupOperators(ctx); err != nil {
		return nil, fmt.Errorf("setting up operators: %w", err)
	}

	if k.cloudProvider == "gcp" {
		if err := k.clusterUtil.SetupGCPGuestAgent(k.client, resources.NewGCPGuestAgentDaemonset()); err != nil {
			return nil, fmt.Errorf("failed to setup gcp guest agent: %w", err)
		}
	}

	// Store the received k8sVersion in a ConfigMap, overwriting existing values (there shouldn't be any).
	// Joining nodes determine the kubernetes version they will install based on this ConfigMap.
	if err := k.setupK8sVersionConfigMap(ctx, k8sVersion); err != nil {
		return nil, fmt.Errorf("failed to setup k8s version ConfigMap: %w", err)
	}

	k.clusterUtil.FixCilium(log)

	return k.GetKubeconfig()
}

// JoinCluster joins existing Kubernetes cluster.
func (k *KubeWrapper) JoinCluster(ctx context.Context, args *kubeadm.BootstrapTokenDiscovery, peerRole role.Role, versionString string, log *logger.Logger) error {
	k8sVersion, err := versions.NewValidK8sVersion(versionString)
	if err != nil {
		return err
	}
	log.With(zap.String("version", string(k8sVersion))).Infof("Installing Kubernetes components")
	if err := k.clusterUtil.InstallComponents(ctx, k8sVersion); err != nil {
		return err
	}

	// Step 1: retrieve cloud metadata for Kubernetes configuration
	nodeInternalIP, err := k.getIPAddr()
	if err != nil {
		return err
	}
	nodeName := nodeInternalIP
	var providerID string
	var loadbalancerEndpoint string
	if k.providerMetadata.Supported() {
		log.Infof("Retrieving node metadata")
		instance, err := k.providerMetadata.Self(ctx)
		if err != nil {
			return fmt.Errorf("retrieving own instance metadata: %w", err)
		}
		providerID = instance.ProviderID
		nodeName = instance.Name
		nodeInternalIP = instance.VPCIP
		loadbalancerEndpoint, err = k.providerMetadata.GetLoadBalancerEndpoint(ctx)
		if err != nil {
			return fmt.Errorf("retrieving loadbalancer endpoint: %w", err)
		}
	}
	nodeName = k8sCompliantHostname(nodeName)

	log.With(
		zap.String("nodeName", nodeName),
		zap.String("providerID", providerID),
		zap.String("nodeIP", nodeInternalIP),
	).Infof("Setting information for node")

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
	log.With(zap.String("apiServerEndpoint", args.APIServerEndpoint)).Infof("Joining Kubernetes cluster")
	if err := k.clusterUtil.JoinCluster(ctx, joinConfigYAML, peerRole, loadbalancerEndpoint, log); err != nil {
		return fmt.Errorf("joining cluster: %v; %w ", string(joinConfigYAML), err)
	}

	k.clusterUtil.FixCilium(log)

	return nil
}

// GetKubeconfig returns the current nodes kubeconfig of stored on disk.
func (k *KubeWrapper) GetKubeconfig() ([]byte, error) {
	return k.kubeconfigReader.ReadKubeconfig()
}

func (k *KubeWrapper) setupClusterAutoscaler(instance metadata.InstanceMetadata, cloudServiceAccountURI string, k8sVersion versions.ValidK8sVersion) error {
	if !k.clusterAutoscaler.Supported() {
		return nil
	}
	caSecrets, err := k.clusterAutoscaler.Secrets(instance.ProviderID, cloudServiceAccountURI)
	if err != nil {
		return fmt.Errorf("defining Secrets for cluster-autoscaler: %w", err)
	}

	clusterAutoscalerConfiguration := resources.NewDefaultAutoscalerDeployment(k.clusterAutoscaler.Volumes(), k.clusterAutoscaler.VolumeMounts(), k.clusterAutoscaler.Env(), k8sVersion)
	if err := k.clusterUtil.SetupAutoscaling(k.client, clusterAutoscalerConfiguration, caSecrets); err != nil {
		return fmt.Errorf("setting up cluster-autoscaler: %w", err)
	}

	return nil
}

// setupK8sVersionConfigMap applies a ConfigMap (cf. server-side apply) to consistently store the installed k8s version.
func (k *KubeWrapper) setupK8sVersionConfigMap(ctx context.Context, k8sVersion versions.ValidK8sVersion) error {
	config := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k8s-version",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			constants.K8sVersion: string(k8sVersion),
		},
	}

	// We do not use the client's Apply method here since we are handling a kubernetes-native type.
	// These types don't implement our custom Marshaler interface.
	if err := k.client.CreateConfigMap(ctx, config); err != nil {
		return fmt.Errorf("apply in KubeWrapper.setupK8sVersionConfigMap(..) failed with: %w", err)
	}

	return nil
}

// setupInternalConfigMap applies a ConfigMap (cf. server-side apply) to store information that is not supposed to be user-editable.
func (k *KubeWrapper) setupInternalConfigMap(ctx context.Context, azureCVM string) error {
	config := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.InternalConfigMap,
			Namespace: "kube-system",
		},
		Data: map[string]string{
			constants.AzureCVM: azureCVM,
		},
	}

	// We do not use the client's Apply method here since we are handling a kubernetes-native type.
	// These types don't implement our custom Marshaler interface.
	if err := k.client.CreateConfigMap(ctx, config); err != nil {
		return fmt.Errorf("apply in KubeWrapper.setupInternalConfigMap failed with: %w", err)
	}

	return nil
}

// setupOperators deploys the operator lifecycle manager and subscriptions to operators.
func (k *KubeWrapper) setupOperators(ctx context.Context) error {
	if err := k.clusterUtil.SetupOperatorLifecycleManager(ctx, k.client, &resources.OperatorLifecycleManagerCRDs{}, &resources.OperatorLifecycleManager{}, resources.OLMCRDNames); err != nil {
		return fmt.Errorf("setting up OLM: %w", err)
	}

	if err := k.clusterUtil.SetupNodeMaintenanceOperator(k.client, resources.NewNodeMaintenanceOperatorDeployment()); err != nil {
		return fmt.Errorf("setting up node maintenance operator: %w", err)
	}

	uid, err := k.providerMetadata.UID(ctx)
	if err != nil {
		return fmt.Errorf("retrieving constellation UID: %w", err)
	}

	if err := k.clusterUtil.SetupNodeOperator(ctx, k.client, resources.NewNodeOperatorDeployment(k.cloudProvider, uid)); err != nil {
		return fmt.Errorf("setting up constellation node operator: %w", err)
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
func (k *KubeWrapper) StartKubelet(log *logger.Logger) error {
	if err := k.clusterUtil.StartKubelet(); err != nil {
		return fmt.Errorf("starting kubelet: %w", err)
	}

	k.clusterUtil.FixCilium(log)
	return nil
}

// getIPAddr retrieves to default sender IP used for outgoing connection.
func getIPAddr() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

// setupExtraVals create a helm values map for consumption by helm-install.
// Will move to a more dedicated place once that place becomes apparent.
func (k *KubeWrapper) setupExtraVals(ctx context.Context, initialMeasurementsJSON []byte, idkeydigest []byte, measurementSalt []byte, subnetworkPodCIDR string, cloudServiceAccountURI string) (map[string]any, error) {
	extraVals := map[string]any{
		"join-service": map[string]any{
			"measurements":    string(initialMeasurementsJSON),
			"measurementSalt": base64.StdEncoding.EncodeToString(measurementSalt),
		},
		"ccm": map[string]any{},
	}

	instance, err := k.providerMetadata.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving current instance: %w", err)
	}

	switch cloudprovider.FromString(k.cloudProvider) {
	case cloudprovider.GCP:
		{
			uid, err := k.providerMetadata.UID(ctx)
			if err != nil {
				return nil, fmt.Errorf("getting uid: %w", err)
			}

			projectID, _, _, err := gcpshared.SplitProviderID(instance.ProviderID)
			if err != nil {
				return nil, fmt.Errorf("splitting providerID: %w", err)
			}

			serviceAccountKey, err := gcpshared.ServiceAccountKeyFromURI(cloudServiceAccountURI)
			if err != nil {
				return nil, fmt.Errorf("getting service account key: %w", err)
			}
			rawKey, err := json.Marshal(serviceAccountKey)
			if err != nil {
				return nil, fmt.Errorf("marshaling service account key: %w", err)
			}

			ccmVals, ok := extraVals["ccm"].(map[string]any)
			if !ok {
				return nil, errors.New("invalid ccm values")
			}
			ccmVals["GCP"] = map[string]any{
				"projectID":         projectID,
				"uid":               uid,
				"secretData":        string(rawKey),
				"subnetworkPodCIDR": subnetworkPodCIDR,
			}
		}
	case cloudprovider.Azure:
		{
			// TODO: After refactoring the ProviderMetadata interface this section should be rewritten.
			// Currently, we have to rely on the Secrets(..) method, as GetNetworkSecurityGroupName & GetLoadBalancerName
			// rely on Azure specific API endpoints.
			ccmSecrets, err := k.cloudControllerManager.Secrets(ctx, instance.ProviderID, cloudServiceAccountURI)
			if err != nil {
				return nil, fmt.Errorf("creating ccm secret: %w", err)
			}
			if len(ccmSecrets) < 1 {
				return nil, errors.New("missing secret")
			}
			rawConfig := ccmSecrets[0].Data["azure.json"]

			ccmVals, ok := extraVals["ccm"].(map[string]any)
			if !ok {
				return nil, errors.New("invalid ccm values")
			}
			ccmVals["Azure"] = map[string]any{
				"azureConfig":       string(rawConfig),
				"subnetworkPodCIDR": subnetworkPodCIDR,
			}

			joinVals, ok := extraVals["join-service"].(map[string]any)
			if !ok {
				return nil, errors.New("invalid join-service values")
			}
			joinVals["idkeydigest"] = hex.EncodeToString(idkeydigest)
		}
	}
	return extraVals, nil
}
