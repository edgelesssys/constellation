/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package kubernetes provides functionality to bootstrap a Kubernetes cluster, or join an exiting one.
package kubernetes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/kubewaiter"
	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/cloud/gcpshared"
	"github.com/edgelesssys/constellation/v2/internal/cloud/openstack"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
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
	helmClient       helmClient
	kubeAPIWaiter    kubeAPIWaiter
	configProvider   configurationProvider
	client           k8sapi.Client
	providerMetadata ProviderMetadata
	getIPAddr        func() (string, error)
}

// New creates a new KubeWrapper with real values.
func New(cloudProvider string, clusterUtil clusterUtil, configProvider configurationProvider, client k8sapi.Client,
	providerMetadata ProviderMetadata, helmClient helmClient, kubeAPIWaiter kubeAPIWaiter,
) *KubeWrapper {
	return &KubeWrapper{
		cloudProvider:    cloudProvider,
		clusterUtil:      clusterUtil,
		helmClient:       helmClient,
		kubeAPIWaiter:    kubeAPIWaiter,
		configProvider:   configProvider,
		client:           client,
		providerMetadata: providerMetadata,
		getIPAddr:        getIPAddr,
	}
}

// InitCluster initializes a new Kubernetes cluster and applies pod network provider.
func (k *KubeWrapper) InitCluster(
	ctx context.Context, cloudServiceAccountURI, versionString, clusterName string, measurementSalt []byte,
	helmReleasesRaw []byte, conformanceMode bool, kubernetesComponents components.Components, apiServerCertSANs []string, log *logger.Logger,
) ([]byte, error) {
	log.With(zap.String("version", versionString)).Infof("Installing Kubernetes components")
	if err := k.clusterUtil.InstallComponents(ctx, kubernetesComponents); err != nil {
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
	nodeName, err := k8sCompliantHostname(instance.Name)
	if err != nil {
		return nil, fmt.Errorf("generating node name: %w", err)
	}

	nodeIP := instance.VPCIP
	subnetworkPodCIDR := instance.SecondaryIPRange
	if len(instance.AliasIPRanges) > 0 {
		nodePodCIDR = instance.AliasIPRanges[0]
	}

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
	initConfigYAML, err := initConfig.Marshal()
	if err != nil {
		return nil, fmt.Errorf("encoding kubeadm init configuration as YAML: %w", err)
	}
	log.Infof("Initializing Kubernetes cluster")
	kubeConfig, err := k.clusterUtil.InitCluster(ctx, initConfigYAML, nodeName, clusterName, validIPs, controlPlaneHost, controlPlanePort, conformanceMode, log)
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

	// Annotate Node with the hash of the installed components
	if err := k.client.AnnotateNode(ctx, nodeName,
		constants.NodeKubernetesComponentsAnnotationKey, k8sComponentsConfigMap,
	); err != nil {
		return nil, fmt.Errorf("annotating node with Kubernetes components hash: %w", err)
	}

	// Step 3: configure & start kubernetes controllers
	log.Infof("Starting Kubernetes controllers and deployments")
	setupPodNetworkInput := k8sapi.SetupPodNetworkInput{
		CloudProvider:     k.cloudProvider,
		NodeName:          nodeName,
		FirstNodePodCIDR:  nodePodCIDR,
		SubnetworkPodCIDR: subnetworkPodCIDR,
		LoadBalancerHost:  controlPlaneHost,
		LoadBalancerPort:  controlPlanePort,
	}

	var helmReleases helm.Releases
	if err := json.Unmarshal(helmReleasesRaw, &helmReleases); err != nil {
		return nil, fmt.Errorf("unmarshalling helm releases: %w", err)
	}

	log.Infof("Installing Cilium")
	ciliumVals, err := k.setupCiliumVals(ctx, setupPodNetworkInput)
	if err != nil {
		return nil, fmt.Errorf("setting up cilium vals: %w", err)
	}
	if err := k.helmClient.InstallChartWithValues(ctx, helmReleases.Cilium, ciliumVals); err != nil {
		return nil, fmt.Errorf("installing cilium pod network: %w", err)
	}

	log.Infof("Waiting for Cilium to become healthy")
	timeToStartWaiting := time.Now()
	// TODO(3u13r): Reduce the timeout when we switched the package repository - this is only this high because we once
	// saw polling times of ~16 minutes when hitting a slow PoP from Fastly (GitHub's / ghcr.io CDN).
	waitCtx, cancel = context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()
	if err := k.clusterUtil.WaitForCilium(waitCtx, log); err != nil {
		return nil, fmt.Errorf("waiting for Cilium to become healthy: %w", err)
	}
	timeUntilFinishedWaiting := time.Since(timeToStartWaiting)
	log.With(zap.Duration("duration", timeUntilFinishedWaiting)).Infof("Cilium became healthy")

	log.Infof("Restarting Cilium")
	if err := k.clusterUtil.FixCilium(ctx); err != nil {
		log.With(zap.Error(err)).Errorf("FixCilium failed")
		// Continue and don't throw an error here - things might be okay.
	}

	serviceConfig := constellationServicesConfig{
		measurementSalt:        measurementSalt,
		subnetworkPodCIDR:      subnetworkPodCIDR,
		cloudServiceAccountURI: cloudServiceAccountURI,
		loadBalancerIP:         controlPlaneHost,
	}
	constellationVals, err := k.setupExtraVals(ctx, serviceConfig)
	if err != nil {
		return nil, fmt.Errorf("setting up extraVals: %w", err)
	}

	log.Infof("Setting up internal-config ConfigMap")
	if err := k.setupInternalConfigMap(ctx); err != nil {
		return nil, fmt.Errorf("failed to setup internal ConfigMap: %w", err)
	}

	log.Infof("Installing Constellation microservices")
	if err = k.helmClient.InstallChartWithValues(ctx, helmReleases.ConstellationServices, constellationVals); err != nil {
		return nil, fmt.Errorf("installing constellation-services: %w", err)
	}

	// cert-manager provides CRDs used by other deployments,
	// so it should be installed as early as possible, but after the services cert-manager depends on.
	log.Infof("Installing cert-manager")
	if err = k.helmClient.InstallChart(ctx, helmReleases.CertManager); err != nil {
		return nil, fmt.Errorf("installing cert-manager: %w", err)
	}

	// Install CSI drivers if enabled by the user.
	if helmReleases.CSI != nil {
		var csiVals map[string]any
		if cloudprovider.FromString(k.cloudProvider) == cloudprovider.OpenStack {
			creds, err := openstack.AccountKeyFromURI(serviceConfig.cloudServiceAccountURI)
			if err != nil {
				return nil, err
			}
			cinderIni := creds.CloudINI().CinderCSIConfiguration()
			csiVals = map[string]any{
				"cinder-config": map[string]any{
					"secretData": cinderIni,
				},
			}
		}

		log.Infof("Installing CSI deployments")
		if err := k.helmClient.InstallChartWithValues(ctx, *helmReleases.CSI, csiVals); err != nil {
			return nil, fmt.Errorf("installing CSI snapshot CRDs: %w", err)
		}
	}

	if helmReleases.AWSLoadBalancerController != nil {
		log.Infof("Installing AWS Load Balancer Controller")
		if err = k.helmClient.InstallChart(ctx, *helmReleases.AWSLoadBalancerController); err != nil {
			return nil, fmt.Errorf("installing AWS Load Balancer Controller: %w", err)
		}
	}

	operatorVals, err := k.setupOperatorVals(ctx)
	if err != nil {
		return nil, fmt.Errorf("setting up operator vals: %w", err)
	}

	// Constellation operators require CRDs from cert-manager.
	// They must be installed after it.
	log.Infof("Installing operators")
	if err = k.helmClient.InstallChartWithValues(ctx, helmReleases.ConstellationOperators, operatorVals); err != nil {
		return nil, fmt.Errorf("installing operators: %w", err)
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
	if err := k.clusterUtil.JoinCluster(ctx, joinConfigYAML, peerRole, loadBalancerHost, loadBalancerPort, log); err != nil {
		return fmt.Errorf("joining cluster: %v; %w ", string(joinConfigYAML), err)
	}

	log.Infof("Waiting for Cilium to become healthy")
	if err := k.clusterUtil.WaitForCilium(context.Background(), log); err != nil {
		return fmt.Errorf("waiting for Cilium to become healthy: %w", err)
	}

	log.Infof("Restarting Cilium")
	if err := k.clusterUtil.FixCilium(context.Background()); err != nil {
		log.With(zap.Error(err)).Errorf("FixCilium failed")
		// Continue and don't throw an error here - things might be okay.
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

	if err := k.client.CreateConfigMap(ctx, componentsConfig); err != nil {
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
	if err := k.client.CreateConfigMap(ctx, config); err != nil {
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
func (k *KubeWrapper) StartKubelet(log *logger.Logger) error {
	if err := k.clusterUtil.StartKubelet(); err != nil {
		return fmt.Errorf("starting kubelet: %w", err)
	}

	log.Infof("Waiting for Cilium to become healthy")
	if err := k.clusterUtil.WaitForCilium(context.Background(), log); err != nil {
		return fmt.Errorf("waiting for Cilium to become healthy: %w", err)
	}

	log.Infof("Restarting Cilium")
	if err := k.clusterUtil.FixCilium(context.Background()); err != nil {
		log.With(zap.Error(err)).Errorf("FixCilium failed")
		// Continue and don't throw an error here - things might be okay.
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

// setupExtraVals create a helm values map for consumption by helm-install.
// Will move to a more dedicated place once that place becomes apparent.
func (k *KubeWrapper) setupExtraVals(ctx context.Context, serviceConfig constellationServicesConfig) (map[string]any, error) {
	extraVals := map[string]any{
		"join-service": map[string]any{
			"measurementSalt": base64.StdEncoding.EncodeToString(serviceConfig.measurementSalt),
		},
		"verification-service": map[string]any{
			"loadBalancerIP": serviceConfig.loadBalancerIP,
		},
		"konnectivity": map[string]any{
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

		extraVals["ccm"] = map[string]any{
			"GCP": map[string]any{
				"projectID":         projectID,
				"uid":               uid,
				"secretData":        string(rawKey),
				"subnetworkPodCIDR": serviceConfig.subnetworkPodCIDR,
			},
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

		extraVals["ccm"] = map[string]any{
			"Azure": map[string]any{
				"azureConfig": string(ccmConfig),
			},
		}

	case cloudprovider.OpenStack:
		creds, err := openstack.AccountKeyFromURI(serviceConfig.cloudServiceAccountURI)
		if err != nil {
			return nil, err
		}
		credsIni := creds.CloudINI().FullConfiguration()
		networkIDsGetter, ok := k.providerMetadata.(openstackMetadata)
		if !ok {
			return nil, errors.New("generating yawol configuration: cloud provider metadata does not implement OpenStack specific methods")
		}
		networkIDs, err := networkIDsGetter.GetNetworkIDs(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting network IDs: %w", err)
		}
		if len(networkIDs) == 0 {
			return nil, errors.New("getting network IDs: no network IDs found")
		}
		extraVals["ccm"] = map[string]any{
			"OpenStack": map[string]any{
				"secretData": credsIni,
			},
		}
		yawolIni := creds.CloudINI().YawolConfiguration()
		extraVals["yawol-config"] = map[string]any{
			"secretData": yawolIni,
		}
		extraVals["yawol-controller"] = map[string]any{
			"yawolNetworkID": networkIDs[0],
			"yawolAPIHost":   fmt.Sprintf("https://%s:%d", serviceConfig.loadBalancerIP, constants.KubernetesPort),
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

func (k *KubeWrapper) setupCiliumVals(ctx context.Context, in k8sapi.SetupPodNetworkInput) (map[string]any, error) {
	vals := map[string]any{
		"k8sServiceHost": in.LoadBalancerHost,
		"k8sServicePort": in.LoadBalancerPort,
	}

	// GCP requires extra configuration for Cilium
	if cloudprovider.FromString(k.cloudProvider) == cloudprovider.GCP {
		if out, err := exec.CommandContext(
			ctx, constants.KubectlPath,
			"--kubeconfig", constants.ControlPlaneAdminConfFilename,
			"patch", "node", in.NodeName, "-p", "{\"spec\":{\"podCIDR\": \""+in.FirstNodePodCIDR+"\"}}",
		).CombinedOutput(); err != nil {
			err = errors.New(string(out))
			return nil, fmt.Errorf("%s: %w", out, err)
		}

		vals["ipv4NativeRoutingCIDR"] = in.SubnetworkPodCIDR
		vals["strictModeCIDR"] = in.SubnetworkPodCIDR

	}

	return vals, nil
}

type ccmConfigGetter interface {
	GetCCMConfig(ctx context.Context, providerID, cloudServiceAccountURI string) ([]byte, error)
}

type constellationServicesConfig struct {
	measurementSalt        []byte
	subnetworkPodCIDR      string
	cloudServiceAccountURI string
	loadBalancerIP         string
}

type openstackMetadata interface {
	GetNetworkIDs(ctx context.Context) ([]string, error)
}
