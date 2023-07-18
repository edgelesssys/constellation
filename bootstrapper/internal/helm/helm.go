/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package helm is used to install Constellation microservices and other services during cluster initialization.
package helm

import (
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
	corev1 "k8s.io/api/core/v1"
)

// Client is used to install microservice during cluster initialization. It is a wrapper for a helm install action.
type Client struct {
	*helm.Installer
	log *logger.Logger
}

// New creates a new client with the given logger.
func New(log *logger.Logger) (*Client, error) {
	installer, err := helm.NewInstaller(log, constants.ControlPlaneAdminConfFilename)
	if err != nil {
		return nil, err
	}
	return &Client{installer, log}, nil
}

// InstallConstellationServices installs the constellation-services chart. In the future this chart should bundle all microservices.
func (h *Client) InstallConstellationServices(ctx context.Context, release helm.Release, extraVals map[string]any) error {
	mergedVals := helm.MergeMaps(release.Values, extraVals)
	return h.InstallChartWithValues(ctx, release, mergedVals, nil)
}

// InstallCertManager installs the cert-manager chart.
func (h *Client) InstallCertManager(ctx context.Context, release helm.Release) error {
	timeout := 10 * time.Minute
	return h.InstallChart(ctx, release, &timeout)
}

// InstallOperators installs the Constellation Operators.
func (h *Client) InstallOperators(ctx context.Context, release helm.Release, extraVals map[string]any) error {
	mergedVals := helm.MergeMaps(release.Values, extraVals)
	return h.InstallChartWithValues(ctx, release, mergedVals, nil)
}

// InstallCilium sets up the cilium pod network.
func (h *Client) InstallCilium(ctx context.Context, kubectl k8sapi.Client, release helm.Release, in k8sapi.SetupPodNetworkInput) error {
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
		return h.installCiliumGeneric(ctx, release, in.LoadBalancerEndpoint)
	case "gcp":
		return h.installCiliumGCP(ctx, release, in.NodeName, in.FirstNodePodCIDR, in.SubnetworkPodCIDR, in.LoadBalancerEndpoint)
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

	return h.InstallChart(ctx, release, nil)
}

func (h *Client) installCiliumGCP(ctx context.Context, release helm.Release, nodeName, nodePodCIDR, subnetworkPodCIDR, kubeAPIEndpoint string) error {
	out, err := exec.CommandContext(ctx, constants.KubectlPath, "--kubeconfig", constants.ControlPlaneAdminConfFilename, "patch", "node", nodeName, "-p", "{\"spec\":{\"podCIDR\": \""+nodePodCIDR+"\"}}").CombinedOutput()
	if err != nil {
		err = errors.New(string(out))
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

	return h.InstallChart(ctx, release, nil)
}
