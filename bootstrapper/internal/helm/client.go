/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package helm

import (
	"bytes"
	"context"
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
	"helm.sh/helm/v3/pkg/chart/loader"
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
	settings.KubeConfig = constants.ControlPlaneAdminConfFilename

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), constants.HelmNamespace,
		"secret", log.Infof); err != nil {
		return nil, err
	}

	action := action.NewInstall(actionConfig)
	action.Namespace = constants.HelmNamespace
	action.Timeout = timeout

	return &Client{
		action,
	}, nil
}

// InstallConstellationServices installs the constellation-services chart. In the future this chart should bundle all microservices.
func (h *Client) InstallConstellationServices(ctx context.Context, release helm.Release, extraVals map[string]any) error {
	h.ReleaseName = release.ReleaseName
	h.Wait = release.Wait

	mergedVals := helm.MergeMaps(release.Values, extraVals)

	if err := h.install(ctx, release.Chart, mergedVals); err != nil {
		return err
	}

	return nil
}

// InstallCertManager installs the cert-manager chart.
func (h *Client) InstallCertManager(ctx context.Context, release helm.Release) error {
	h.ReleaseName = release.ReleaseName
	h.Wait = release.Wait

	if err := h.install(ctx, release.Chart, release.Values); err != nil {
		return err
	}

	return nil
}

// InstallOperators installs the Constellation Operators.
func (h *Client) InstallOperators(ctx context.Context, release helm.Release, extraVals map[string]any) error {
	h.ReleaseName = release.ReleaseName
	h.Wait = release.Wait

	mergedVals := helm.MergeMaps(release.Values, extraVals)

	if err := h.install(ctx, release.Chart, mergedVals); err != nil {
		return err
	}

	return nil
}

// InstallCilium sets up the cilium pod network.
func (h *Client) InstallCilium(ctx context.Context, kubectl k8sapi.Client, release helm.Release, in k8sapi.SetupPodNetworkInput) error {
	h.ReleaseName = release.ReleaseName
	h.Wait = release.Wait

	switch in.CloudProvider {
	case "aws", "azure", "qemu":
		return h.installCiliumGeneric(ctx, release, in.LoadBalancerEndpoint)
	case "gcp":
		return h.installCiliumGCP(ctx, kubectl, release, in.NodeName, in.FirstNodePodCIDR, in.SubnetworkPodCIDR, in.LoadBalancerEndpoint)
	default:
		return fmt.Errorf("unsupported cloud provider %q", in.CloudProvider)
	}
}

// installCiliumGeneric installs cilium with the given load balancer endpoint.
// This is used for cloud providers that do not require special server-side configuration.
// Currently this is AWS, Azure, and QEMU.
func (h *Client) installCiliumGeneric(ctx context.Context, release helm.Release, kubeAPIEndpoint string) error {
	host := kubeAPIEndpoint
	release.Values["k8sServiceHost"] = host
	release.Values["k8sServicePort"] = strconv.Itoa(constants.KubernetesPort)

	if err := h.install(ctx, release.Chart, release.Values); err != nil {
		return err
	}
	return nil
}

func (h *Client) installCiliumGCP(ctx context.Context, kubectl k8sapi.Client, release helm.Release, nodeName, nodePodCIDR, subnetworkPodCIDR, kubeAPIEndpoint string) error {
	out, err := exec.CommandContext(ctx, constants.KubectlPath, "--kubeconfig", constants.ControlPlaneAdminConfFilename, "patch", "node", nodeName, "-p", "{\"spec\":{\"podCIDR\": \""+nodePodCIDR+"\"}}").CombinedOutput()
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

	if err := h.install(ctx, release.Chart, release.Values); err != nil {
		return err
	}

	return nil
}

func (h *Client) install(ctx context.Context, chartRaw []byte, values map[string]any) error {
	reader := bytes.NewReader(chartRaw)
	chart, err := loader.LoadArchive(reader)
	if err != nil {
		return fmt.Errorf("helm load archive: %w", err)
	}

	_, err = h.RunWithContext(ctx, chart, values)
	if err != nil {
		return fmt.Errorf("helm install: %w", err)
	}
	return nil
}
