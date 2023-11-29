/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package kubernetes provides functionality to bootstrap a Kubernetes cluster, or join an exiting one.
package kubernetes

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubewaiter"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/role"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

var validHostnameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

// configurationProvider provides kubeadm init and join configuration.
type configurationProvider interface {
	InitConfiguration(externalCloudProvider bool, k8sVersion string) k8sapi.KubeadmInitYAML
	JoinConfiguration(externalCloudProvider bool) k8sapi.KubeadmJoinYAML
}

type kubeAPIWaiter interface {
	Wait(ctx context.Context, kubernetesClient kubewaiter.KubernetesClient) error
}

// KubeWrapper implements Cluster interface.
type KubeWrapper struct {
	cloudProvider    string
	clusterUtil      clusterUtil
	kubeAPIWaiter    kubeAPIWaiter
	configProvider   configurationProvider
	client           k8sapi.Client
	providerMetadata ProviderMetadata
	getIPAddr        func() (string, error)
}

// New creates a new KubeWrapper with real values.
func New(cloudProvider string, clusterUtil clusterUtil, configProvider configurationProvider, client k8sapi.Client,
	providerMetadata ProviderMetadata, kubeAPIWaiter kubeAPIWaiter,
) *KubeWrapper {
	return &KubeWrapper{
		cloudProvider:    cloudProvider,
		clusterUtil:      clusterUtil,
		kubeAPIWaiter:    kubeAPIWaiter,
		configProvider:   configProvider,
		client:           client,
		providerMetadata: providerMetadata,
		getIPAddr:        getIPAddr,
	}
}

// InitCluster initializes a new Kubernetes cluster and applies pod network provider.
func (k *KubeWrapper) InitCluster(
	ctx context.Context, versionString, clusterName string, conformanceMode bool, kubernetesComponents components.Components, apiServerCertSANs []string, serviceCIDR string, log *logger.Logger,
) ([]byte, error) {
	log.With(zap.String("version", versionString)).Infof("Installing Kubernetes components")
	if err := k.clusterUtil.InstallComponents(ctx, kubernetesComponents); err != nil {
		return nil, err
	}

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
	nodeName, err := k8sCompliantHostname(instance.Name)
	if err != nil {
		return nil, fmt.Errorf("generating node name: %w", err)
	}

	nodeIP := instance.VPCIP
	subnetworkPodCIDR := instance.SecondaryIPRange

	// this is the endpoint in "kubeadm init --control-plane-endpoint=<IP/DNS>:<port>"
	// TODO(malt3): switch over to DNS name on AWS and Azure
	// soon as every apiserver certificate of every control-plane node
	// has the dns endpoint in its SAN list.
	controlPlaneHost, controlPlanePort, err := k.providerMetadata.GetLoadBalancerEndpoint(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving load balancer endpoint: %w", err)
	}

	certSANs := []string{nodeIP}
	certSANs = append(certSANs, apiServerCertSANs...)

	log.With(
		zap.String("nodeName", nodeName),
		zap.String("providerID", instance.ProviderID),
		zap.String("nodeIP", nodeIP),
		zap.String("controlPlaneHost", controlPlaneHost),
		zap.String("controlPlanePort", controlPlanePort),
		zap.String("certSANs", strings.Join(certSANs, ",")),
		zap.String("podCIDR", subnetworkPodCIDR),
	).Infof("Setting information for node")

	// Step 2: configure kubeadm init config
	ccmSupported := cloudprovider.FromString(k.cloudProvider) == cloudprovider.Azure ||
		cloudprovider.FromString(k.cloudProvider) == cloudprovider.GCP ||
		cloudprovider.FromString(k.cloudProvider) == cloudprovider.AWS
	initConfig := k.configProvider.InitConfiguration(ccmSupported, versionString)
	initConfig.SetNodeIP(nodeIP)
	initConfig.SetClusterName(clusterName)
	initConfig.SetCertSANs(certSANs)
	initConfig.SetNodeName(nodeName)
	initConfig.SetProviderID(instance.ProviderID)
	initConfig.SetControlPlaneEndpoint(controlPlaneHost)
	initConfig.SetServiceSubnet(serviceCIDR)
	initConfigYAML, err := initConfig.Marshal()
	if err != nil {
		return nil, fmt.Errorf("encoding kubeadm init configuration as YAML: %w", err)
	}
	log.Infof("Initializing Kubernetes cluster")
	kubeConfig, err := k.clusterUtil.InitCluster(ctx, initConfigYAML, nodeName, clusterName, validIPs, conformanceMode, log)
	if err != nil {
		return nil, fmt.Errorf("kubeadm init: %w", err)
	}

	err = k.client.Initialize(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("initializing kubectl client: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := k.kubeAPIWaiter.Wait(waitCtx, k.client); err != nil {
		return nil, fmt.Errorf("waiting for Kubernetes API to be available: %w", err)
	}

	if err := k.client.EnforceCoreDNSSpread(ctx); err != nil {
		return nil, fmt.Errorf("configuring CoreDNS deployment: %w", err)
	}

	// Setup the K8s components ConfigMap.
	k8sComponentsConfigMap, err := k.setupK8sComponentsConfigMap(ctx, kubernetesComponents, versionString)
	if err != nil {
		return nil, fmt.Errorf("failed to setup k8s version ConfigMap: %w", err)
	}

	if cloudprovider.FromString(k.cloudProvider) == cloudprovider.GCP {
		// GCP uses direct routing, so we need to set the pod CIDR of the first control-plane node for Cilium.
		var nodePodCIDR string
		if len(instance.AliasIPRanges) > 0 {
			nodePodCIDR = instance.AliasIPRanges[0]
		}
		if err := k.client.PatchFirstNodePodCIDR(ctx, nodePodCIDR); err != nil {
			return nil, fmt.Errorf("patching first node pod CIDR: %w", err)
		}
	}

	// Annotate Node with the hash of the installed components
	if err := k.client.AnnotateNode(ctx, nodeName,
		constants.NodeKubernetesComponentsAnnotationKey, k8sComponentsConfigMap,
	); err != nil {
		return nil, fmt.Errorf("annotating node with Kubernetes components hash: %w", err)
	}

	log.Infof("Setting up internal-config ConfigMap")
	if err := k.setupInternalConfigMap(ctx); err != nil {
		return nil, fmt.Errorf("failed to setup internal ConfigMap: %w", err)
	}
	return kubeConfig, nil
}

// JoinCluster joins existing Kubernetes cluster.
func (k *KubeWrapper) JoinCluster(ctx context.Context, args *kubeadm.BootstrapTokenDiscovery, peerRole role.Role, k8sComponents components.Components, log *logger.Logger) error {
	log.With("k8sComponents", k8sComponents).Infof("Installing provided kubernetes components")
	if err := k.clusterUtil.InstallComponents(ctx, k8sComponents); err != nil {
		return fmt.Errorf("installing kubernetes components: %w", err)
	}

	// Step 1: retrieve cloud metadata for Kubernetes configuration
	log.Infof("Retrieving node metadata")
	instance, err := k.providerMetadata.Self(ctx)
	if err != nil {
		return fmt.Errorf("retrieving own instance metadata: %w", err)
	}
	providerID := instance.ProviderID
	nodeInternalIP := instance.VPCIP
	nodeName, err := k8sCompliantHostname(instance.Name)
	if err != nil {
		return fmt.Errorf("generating node name: %w", err)
	}

	loadBalancerHost, loadBalancerPort, err := k.providerMetadata.GetLoadBalancerEndpoint(ctx)
	if err != nil {
		return fmt.Errorf("retrieving own instance metadata: %w", err)
	}

	// override join endpoint to go over lb
	args.APIServerEndpoint = net.JoinHostPort(loadBalancerHost, loadBalancerPort)

	log.With(
		zap.String("nodeName", nodeName),
		zap.String("providerID", providerID),
		zap.String("nodeIP", nodeInternalIP),
		zap.String("loadBalancerHost", loadBalancerHost),
		zap.String("loadBalancerPort", loadBalancerPort),
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
	if err := k.clusterUtil.JoinCluster(ctx, joinConfigYAML, log); err != nil {
		return fmt.Errorf("joining cluster: %v; %w ", string(joinConfigYAML), err)
	}

	return nil
}

// setupK8sComponentsConfigMap applies a ConfigMap (cf. server-side apply) to store the installed k8s components.
// It returns the name of the ConfigMap.
func (k *KubeWrapper) setupK8sComponentsConfigMap(ctx context.Context, components components.Components, clusterVersion string) (string, error) {
	componentsConfig, err := kubernetes.ConstructK8sComponentsCM(components, clusterVersion)
	if err != nil {
		return "", fmt.Errorf("constructing k8s-components ConfigMap: %w", err)
	}

	if err := k.client.CreateConfigMap(ctx, &componentsConfig); err != nil {
		return "", fmt.Errorf("apply in KubeWrapper.setupK8sVersionConfigMap(..) for components config map failed with: %w", err)
	}

	return componentsConfig.ObjectMeta.Name, nil
}

// setupInternalConfigMap applies a ConfigMap (cf. server-side apply) to store information that is not supposed to be user-editable.
func (k *KubeWrapper) setupInternalConfigMap(ctx context.Context) error {
	config := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.InternalConfigMap,
			Namespace: "kube-system",
		},
		Data: map[string]string{},
	}

	// We do not use the client's Apply method here since we are handling a kubernetes-native type.
	// These types don't implement our custom Marshaler interface.
	if err := k.client.CreateConfigMap(ctx, &config); err != nil {
		return fmt.Errorf("apply in KubeWrapper.setupInternalConfigMap failed with: %w", err)
	}

	return nil
}

// k8sCompliantHostname transforms a hostname to an RFC 1123 compliant, lowercase subdomain as required by Kubernetes node names.
// The following regex is used by k8s for validation: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/ .
// Only a simple heuristic is used for now (to lowercase, replace underscores).
func k8sCompliantHostname(in string) (string, error) {
	hostname := strings.ToLower(in)
	hostname = strings.ReplaceAll(hostname, "_", "-")
	if !validHostnameRegex.MatchString(hostname) {
		return "", fmt.Errorf("failed to generate a Kubernetes compliant hostname for %s", in)
	}
	return hostname, nil
}

// StartKubelet starts the kubelet service.
func (k *KubeWrapper) StartKubelet() error {
	if err := k.clusterUtil.StartKubelet(); err != nil {
		return fmt.Errorf("starting kubelet: %w", err)
	}

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
