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
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi/resources"
	kubewaiter "github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubeWaiter"
	"github.com/edgelesssys/constellation/v2/internal/cloud/azureshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
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

type kubeAPIWaiter interface {
	Wait(ctx context.Context, kubernetesClient kubewaiter.KubernetesClient) error
}

// KubeWrapper implements Cluster interface.
type KubeWrapper struct {
	cloudProvider           string
	clusterUtil             clusterUtil
	helmClient              helmClient
	kubeAPIWaiter           kubeAPIWaiter
	configProvider          configurationProvider
	client                  k8sapi.Client
	kubeconfigReader        configReader
	providerMetadata        ProviderMetadata
	initialMeasurementsJSON []byte
	getIPAddr               func() (string, error)
}

// New creates a new KubeWrapper with real values.
// TODO: Upon changing this function, please refactor it to reduce the number of arguments to <= 5.
//
//revive:disable-next-line
func New(cloudProvider string, clusterUtil clusterUtil, configProvider configurationProvider, client k8sapi.Client,
	providerMetadata ProviderMetadata, initialMeasurementsJSON []byte, helmClient helmClient, kubeAPIWaiter kubeAPIWaiter,
) *KubeWrapper {
	return &KubeWrapper{
		cloudProvider:           cloudProvider,
		clusterUtil:             clusterUtil,
		helmClient:              helmClient,
		kubeAPIWaiter:           kubeAPIWaiter,
		configProvider:          configProvider,
		client:                  client,
		kubeconfigReader:        &KubeconfigReader{fs: afero.Afero{Fs: afero.NewOsFs()}},
		providerMetadata:        providerMetadata,
		initialMeasurementsJSON: initialMeasurementsJSON,
		getIPAddr:               getIPAddr,
	}
}

// InitCluster initializes a new Kubernetes cluster and applies pod network provider.
// TODO: Upon changing this function, please refactor it to reduce the number of arguments to <= 5.
//
//revive:disable-next-line
func (k *KubeWrapper) InitCluster(
	ctx context.Context, cloudServiceAccountURI, versionString string, measurementSalt []byte, enforcedPCRs []uint32,
	enforceIDKeyDigest bool, idKeyDigest []byte, azureCVM bool,
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

	var nodePodCIDR string
	var validIPs []net.IP

	// Step 1: retrieve cloud metadata for Kubernetes configuration
	log.Infof("Retrieving node metadata")
	instance, err := k.providerMetadata.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving own instance metadata: %w", err)
	}
	if instance.VPCIP != "" {
		validIPs = append(validIPs, net.ParseIP(instance.VPCIP))
	}
	nodeName := k8sCompliantHostname(instance.Name)
	nodeIP := instance.VPCIP
	subnetworkPodCIDR := instance.SecondaryIPRange
	if len(instance.AliasIPRanges) > 0 {
		nodePodCIDR = instance.AliasIPRanges[0]
	}

	// this is the endpoint in "kubeadm init --control-plane-endpoint=<IP/DNS>:<port>"
	controlPlaneEndpoint, err := k.providerMetadata.GetLoadBalancerEndpoint(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving load balancer endpoint: %w", err)
	}

	log.With(
		zap.String("nodeName", nodeName),
		zap.String("providerID", instance.ProviderID),
		zap.String("nodeIP", nodeIP),
		zap.String("controlPlaneEndpoint", controlPlaneEndpoint),
		zap.String("podCIDR", subnetworkPodCIDR),
	).Infof("Setting information for node")

	// Step 2: configure kubeadm init config
	ccmSupported := cloudprovider.FromString(k.cloudProvider) == cloudprovider.Azure ||
		cloudprovider.FromString(k.cloudProvider) == cloudprovider.GCP
	initConfig := k.configProvider.InitConfiguration(ccmSupported, k8sVersion)
	initConfig.SetNodeIP(nodeIP)
	initConfig.SetCertSANs([]string{nodeIP})
	initConfig.SetNodeName(nodeName)
	initConfig.SetProviderID(instance.ProviderID)
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

	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := k.kubeAPIWaiter.Wait(waitCtx, k.client); err != nil {
		return nil, fmt.Errorf("waiting for Kubernetes API to be available: %w", err)
	}

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

	loadBalancerIP := controlPlaneEndpoint
	if strings.Contains(controlPlaneEndpoint, ":") {
		loadBalancerIP, _, err = net.SplitHostPort(controlPlaneEndpoint)
		if err != nil {
			return nil, fmt.Errorf("splitting host port: %w", err)
		}
	}
	serviceConfig := constellationServicesConfig{k.initialMeasurementsJSON, idKeyDigest, measurementSalt, subnetworkPodCIDR, cloudServiceAccountURI, loadBalancerIP}
	extraVals, err := k.setupExtraVals(ctx, serviceConfig)
	if err != nil {
		return nil, fmt.Errorf("setting up extraVals: %w", err)
	}

	if err = k.helmClient.InstallConstellationServices(ctx, helmReleases.ConstellationServices, extraVals); err != nil {
		return nil, fmt.Errorf("installing constellation-services: %w", err)
	}

	if err := k.setupInternalConfigMap(ctx, strconv.FormatBool(azureCVM)); err != nil {
		return nil, fmt.Errorf("failed to setup internal ConfigMap: %w", err)
	}

	// cert-manager is necessary for our operator deployments.
	// They are currently only deployed on GCP & Azure. This is why we deploy cert-manager only on GCP & Azure.
	if k.cloudProvider == "gcp" || k.cloudProvider == "azure" {
		if err = k.helmClient.InstallCertManager(ctx, helmReleases.CertManager); err != nil {
			return nil, fmt.Errorf("installing cert-manager: %w", err)
		}
	}

	operatorVals, err := k.setupOperatorVals(ctx)
	if err != nil {
		return nil, fmt.Errorf("setting up operator vals: %w", err)
	}

	if err = k.helmClient.InstallOperators(ctx, helmReleases.Operators, operatorVals); err != nil {
		return nil, fmt.Errorf("installing operators: %w", err)
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
	log.Infof("Retrieving node metadata")
	instance, err := k.providerMetadata.Self(ctx)
	if err != nil {
		return fmt.Errorf("retrieving own instance metadata: %w", err)
	}
	providerID := instance.ProviderID
	nodeInternalIP := instance.VPCIP
	nodeName := k8sCompliantHostname(instance.Name)

	loadbalancerEndpoint, err := k.providerMetadata.GetLoadBalancerEndpoint(ctx)
	if err != nil {
		return fmt.Errorf("retrieving own instance metadata: %w", err)
	}

	log.With(
		zap.String("nodeName", nodeName),
		zap.String("providerID", providerID),
		zap.String("nodeIP", nodeInternalIP),
	).Infof("Setting information for node")

	// Step 2: configure kubeadm join config
	ccmSupported := cloudprovider.FromString(k.cloudProvider) == cloudprovider.Azure ||
		cloudprovider.FromString(k.cloudProvider) == cloudprovider.GCP
	joinConfig := k.configProvider.JoinConfiguration(ccmSupported)
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
func (k *KubeWrapper) setupExtraVals(ctx context.Context, serviceConfig constellationServicesConfig) (map[string]any, error) {
	extraVals := map[string]any{
		"join-service": map[string]any{
			"measurements":    string(serviceConfig.initialMeasurementsJSON),
			"measurementSalt": base64.StdEncoding.EncodeToString(serviceConfig.measurementSalt),
		},
		"ccm": map[string]any{},
		"verification-service": map[string]any{
			"loadBalancerIP": serviceConfig.loadBalancerIP,
		},
	}

	instance, err := k.providerMetadata.Self(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving current instance: %w", err)
	}

	switch cloudprovider.FromString(k.cloudProvider) {
	case cloudprovider.GCP:
		uid, err := k.providerMetadata.UID(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting uid: %w", err)
		}

		projectID, _, _, err := gcpshared.SplitProviderID(instance.ProviderID)
		if err != nil {
			return nil, fmt.Errorf("splitting providerID: %w", err)
		}

		serviceAccountKey, err := gcpshared.ServiceAccountKeyFromURI(serviceConfig.cloudServiceAccountURI)
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
			"subnetworkPodCIDR": serviceConfig.subnetworkPodCIDR,
		}

	case cloudprovider.Azure:
		ccmAzure, ok := k.providerMetadata.(ccmConfigGetter)
		if !ok {
			return nil, errors.New("invalid cloud provider metadata for Azure")
		}

		ccmConfig, err := ccmAzure.GetCCMConfig(ctx, instance.ProviderID, serviceConfig.cloudServiceAccountURI)
		if err != nil {
			return nil, fmt.Errorf("creating ccm secret: %w", err)
		}

		ccmVals, ok := extraVals["ccm"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid ccm values")
		}
		ccmVals["Azure"] = map[string]any{
			"azureConfig":       string(ccmConfig),
			"subnetworkPodCIDR": serviceConfig.subnetworkPodCIDR,
		}

		joinVals, ok := extraVals["join-service"].(map[string]any)
		if !ok {
			return nil, errors.New("invalid join-service values")
		}
		joinVals["idkeydigest"] = hex.EncodeToString(serviceConfig.idkeydigest)

		subscriptionID, resourceGroup, err := azureshared.BasicsFromProviderID(instance.ProviderID)
		if err != nil {
			return nil, err
		}
		creds, err := azureshared.ApplicationCredentialsFromURI(serviceConfig.cloudServiceAccountURI)
		if err != nil {
			return nil, err
		}

		extraVals["autoscaler"] = map[string]any{
			"Azure": map[string]any{
				"clientID":       creds.AppClientID,
				"clientSecret":   creds.ClientSecretValue,
				"resourceGroup":  resourceGroup,
				"subscriptionID": subscriptionID,
				"tenantID":       creds.TenantID,
			},
		}

	}
	return extraVals, nil
}

func (k *KubeWrapper) setupOperatorVals(ctx context.Context) (map[string]any, error) {
	uid, err := k.providerMetadata.UID(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving constellation UID: %w", err)
	}

	return map[string]any{
		"constellation-operator": map[string]any{
			"constellationUID": uid,
		},
	}, nil
}

type ccmConfigGetter interface {
	GetCCMConfig(ctx context.Context, providerID, cloudServiceAccountURI string) ([]byte, error)
}

type constellationServicesConfig struct {
	initialMeasurementsJSON []byte
	idkeydigest             []byte
	measurementSalt         []byte
	subnetworkPodCIDR       string
	cloudServiceAccountURI  string
	loadBalancerIP          string
}
