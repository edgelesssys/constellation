/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
)

const (
	// timeout is the maximum time given to the helm client.
	timeout = 5 * time.Minute
)

// Client is used to install microservice during cluster initialization. It is a wrapper for a helm install action.
type Client struct {
	*action.Install
}

// New creates a new client with the given logger.
func New(log *logger.Logger) (*Client, error) {
	settings := cli.New()
	settings.KubeConfig = constants.CoreOSAdminConfFilename

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), constants.HelmNamespace,
		"secret", log.Infof); err != nil {
		return nil, err
	}

	return &Client{
		action.NewInstall(actionConfig),
	}, nil
}

// InstallCilium sets up the cilium pod network.
func (h *Client) InstallCilium(ctx context.Context, kubectl k8sapi.Client, release helm.Release, in k8sapi.SetupPodNetworkInput) error {
	h.Namespace = constants.HelmNamespace
	h.ReleaseName = release.ReleaseName
	h.Wait = release.Wait
	h.Timeout = timeout

	switch in.CloudProvider {
	case "gcp":
		return h.installlCiliumGCP(ctx, kubectl, release, in.NodeName, in.FirstNodePodCIDR, in.SubnetworkPodCIDR, in.LoadBalancerEndpoint)
	case "azure":
		return h.installCiliumAzure(ctx, release, in.LoadBalancerEndpoint)
	case "qemu":
		return h.installCiliumQEMU(ctx, release, in.SubnetworkPodCIDR, in.LoadBalancerEndpoint)
	default:
		return fmt.Errorf("unsupported cloud provider %q", in.CloudProvider)
	}
}

func (h *Client) installCiliumAzure(ctx context.Context, release helm.Release, kubeAPIEndpoint string) error {
	host := kubeAPIEndpoint
	release.Values["k8sServiceHost"] = host
	release.Values["k8sServicePort"] = strconv.Itoa(constants.KubernetesPort)

	_, err := h.RunWithContext(ctx, release.Chart, release.Values)
	if err != nil {
		return fmt.Errorf("installing cilium: %w", err)
	}
	return nil
}

func (h *Client) installlCiliumGCP(ctx context.Context, kubectl k8sapi.Client, release helm.Release, nodeName, nodePodCIDR, subnetworkPodCIDR, kubeAPIEndpoint string) error {
	out, err := exec.CommandContext(ctx, constants.KubectlPath, "--kubeconfig", constants.CoreOSAdminConfFilename, "patch", "node", nodeName, "-p", "{\"spec\":{\"podCIDR\": \""+nodePodCIDR+"\"}}").CombinedOutput()
	if err != nil {
		err = errors.New(string(out))
		return err
	}

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
	if err = kubectl.AddTolerationsToDeployment(ctx, tolerations, "coredns", "kube-system"); err != nil {
		return err
	}
	selectors := map[string]string{
		"node-role.kubernetes.io/control-plane": "",
	}
	if err = kubectl.AddNodeSelectorsToDeployment(ctx, selectors, "coredns", "kube-system"); err != nil {
		return err
	}

	host, port, err := net.SplitHostPort(kubeAPIEndpoint)
	if err != nil {
		return err
	}

	// configure pod network CIDR
	release.Values["ipv4NativeRoutingCIDR"] = subnetworkPodCIDR
	release.Values["strictModeCIDR"] = subnetworkPodCIDR
	release.Values["k8sServiceHost"] = host
	if port != "" {
		release.Values["k8sServicePort"] = port
	}

	_, err = h.RunWithContext(ctx, release.Chart, release.Values)
	if err != nil {
		return fmt.Errorf("helm install cilium: %w", err)
	}

	return nil
}

func (h *Client) installCiliumQEMU(ctx context.Context, release helm.Release, subnetworkPodCIDR, kubeAPIEndpoint string) error {
	// configure pod network CIDR
	release.Values["ipam"] = map[string]interface{}{
		"operator": map[string]interface{}{
			"clusterPoolIPv4PodCIDRList": []interface{}{
				subnetworkPodCIDR,
			},
		},
	}

	release.Values["k8sServiceHost"] = kubeAPIEndpoint
	release.Values["k8sServicePort"] = strconv.Itoa(constants.KubernetesPort)

	_, err := h.RunWithContext(ctx, release.Chart, release.Values)
	if err != nil {
		return fmt.Errorf("helm install cilium: %w", err)
	}
	return nil
}

// InstallKMS deploys the KMS deployment.
func (h *Client) InstallKMS(ctx context.Context, release helm.Release, kmsConfig KMSConfig) error {
	h.Namespace = constants.HelmNamespace
	h.ReleaseName = release.ReleaseName
	h.Wait = release.Wait
	h.Timeout = timeout

	release.Values["masterSecret"] = base64.StdEncoding.EncodeToString(kmsConfig.MasterSecret[:])
	release.Values["salt"] = base64.StdEncoding.EncodeToString(kmsConfig.Salt[:])

	_, err := h.RunWithContext(ctx, release.Chart, release.Values)
	if err != nil {
		return fmt.Errorf("helm install kms: %w", err)
	}

	return nil
}
