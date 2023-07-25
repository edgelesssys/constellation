/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	corev1 "k8s.io/api/core/v1"
)

// var validHostnameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

// k8sCompliantHostname transforms a hostname to an RFC 1123 compliant, lowercase subdomain as required by Kubernetes node names.
// The following regex is used by k8s for validation: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/ .
// Only a simple heuristic is used for now (to lowercase, replace underscores).
//func K8sCompliantHostname(in string) (string, error) {
//	hostname := strings.ToLower(in)
//	hostname = strings.ReplaceAll(hostname, "_", "-")
//	if !validHostnameRegex.MatchString(hostname) {
//		return "", fmt.Errorf("failed to generate a Kubernetes compliant hostname for %s", in)
//	}
//	return hostname, nil
//}

// helmClient bundles functions related to microservice deployment.
// Only microservices that can be deployed purely via Helm are deployed with this interface.
type helmClient interface {
	InstallChart(context.Context, Release) error
	InstallChartWithValues(ctx context.Context, release Release, extraValues map[string]any) error
}

// Client provides the functions to talk to the k8s API.
type k8sClient interface {
	Initialize(kubeconfig []byte) error
	CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error
	AddTolerationsToDeployment(ctx context.Context, tolerations []corev1.Toleration, name string, namespace string) error
	AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string, namespace string) error
	ListAllNamespaces(ctx context.Context) (*corev1.NamespaceList, error)
	AnnotateNode(ctx context.Context, nodeName, annotationKey, annotationValue string) error
	EnforceCoreDNSSpread(ctx context.Context) error
}

// func GetSetupPodNetwork(ctx context.Context, metadata *aws.Cloud, cloudProvider cloudprovider.Provider) (*SetupPodNetworkInput, error) {
//	var validIPs []net.IP
//	var nodePodCIDR string

//	instance, err := metadata.Self(ctx)
//	if err != nil {
//		return nil, err // log.With(zap.Error(err)).Fatalf("Failed to get instance metadata")
//	}
//	if instance.VPCIP != "" {
//		validIPs = append(validIPs, net.ParseIP(instance.VPCIP))
//	}
//	nodeName, err := K8sCompliantHostname(instance.Name)
//	if err != nil {
//		return nil, fmt.Errorf("generating node name: %w", err)
//	}

//	subnetworkPodCIDR := instance.SecondaryIPRange
//	if len(instance.AliasIPRanges) > 0 {
//		nodePodCIDR = instance.AliasIPRanges[0]
//	}

//	// this is the endpoint in "kubeadm init --control-plane-endpoint=<IP/DNS>:<port>"
//	// TODO(malt3): switch over to DNS name on AWS and Azure
//	// soon as every apiserver certificate of every control-plane node
//	// has the dns endpoint in its SAN list.
//	controlPlaneHost, controlPlanePort, err := k.providerMetadata.GetLoadBalancerEndpoint(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("retrieving load balancer endpoint: %w", err)
//	}

//	log.Infof("Starting Kubernetes controllers and deployments")
//	setupPodNetworkInput := &SetupPodNetworkInput{
//		CloudProvider:     cloudProvider.String(),
//		NodeName:          nodeName,
//		FirstNodePodCIDR:  nodePodCIDR,
//		SubnetworkPodCIDR: subnetworkPodCIDR,
//		LoadBalancerHost:  controlPlaneHost,
//		LoadBalancerPort:  controlPlanePort,
//	}
//	return setupPodNetworkInput, nil
//}

// SetupPodNetworkInput holds all configuration options to setup the pod network.
type SetupPodNetworkInput struct {
	CloudProvider     string
	NodeName          string
	FirstNodePodCIDR  string
	SubnetworkPodCIDR string
	LoadBalancerHost  string
	LoadBalancerPort  string
}

// InstallCilium sets up the cilium pod network.
func InstallCilium(ctx context.Context, helmInstaller helmClient, kubectl k8sClient, release Release, in SetupPodNetworkInput) error {
	timeoutS := int64(10)
	// allow coredns to run on uninitialized nodes (required by cloud-controller-manager)
	tolerations := []corev1.Toleration{
		{
			Key:    "node.cloudprovider.kubernetes.io/uninitialized",
			Value:  "true",
			Effect: corev1.TaintEffectNoSchedule,
		},
		{
			Key:               "node.kubernetes.io/unreachable",
			Operator:          corev1.TolerationOpExists,
			Effect:            corev1.TaintEffectNoExecute,
			TolerationSeconds: &timeoutS,
		},
	}
	if err := kubectl.AddTolerationsToDeployment(ctx, tolerations, "coredns", "kube-system"); err != nil {
		return fmt.Errorf("failed to add tolerations to coredns deployment: %w", err)
	}
	if err := kubectl.EnforceCoreDNSSpread(ctx); err != nil {
		return fmt.Errorf("failed to enforce CoreDNS spread: %w", err)
	}

	switch in.CloudProvider {
	case "aws", "azure", "openstack", "qemu":
		return installCiliumGeneric(ctx, helmInstaller, release, in.LoadBalancerHost, in.LoadBalancerPort)
	case "gcp":
		return installCiliumGCP(ctx, helmInstaller, release, in.NodeName, in.FirstNodePodCIDR, in.SubnetworkPodCIDR, in.LoadBalancerHost, in.LoadBalancerPort)
	default:
		return fmt.Errorf("unsupported cloud provider %q", in.CloudProvider)
	}
}

// installCiliumGeneric installs cilium with the given load balancer endpoint.
// This is used for cloud providers that do not require special server-side configuration.
// Currently this is AWS, Azure, and QEMU.
func installCiliumGeneric(ctx context.Context, helmInstaller helmClient, release Release, kubeAPIHost, kubeAPIPort string) error {
	if release.Values != nil {
		release.Values["k8sServiceHost"] = kubeAPIHost
		release.Values["k8sServicePort"] = kubeAPIPort
	}
	return helmInstaller.InstallChart(ctx, release)
}

func installCiliumGCP(ctx context.Context, helmInstaller helmClient, release Release, nodeName, nodePodCIDR, subnetworkPodCIDR, kubeAPIHost, kubeAPIPort string) error {
	out, err := exec.CommandContext(ctx, constants.KubectlPath, "--kubeconfig", constants.ControlPlaneAdminConfFilename, "patch", "node", nodeName, "-p", "{\"spec\":{\"podCIDR\": \""+nodePodCIDR+"\"}}").CombinedOutput()
	if err != nil {
		err = errors.New(string(out))
		return err
	}

	// configure pod network CIDR
	release.Values["ipv4NativeRoutingCIDR"] = subnetworkPodCIDR
	release.Values["strictModeCIDR"] = subnetworkPodCIDR
	release.Values["k8sServiceHost"] = kubeAPIHost
	release.Values["k8sServicePort"] = kubeAPIPort

	return helmInstaller.InstallChart(ctx, release)
}
