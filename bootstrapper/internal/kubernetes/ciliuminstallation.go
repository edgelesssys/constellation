/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package kubernetes provides functionality to bootstrap a Kubernetes cluster, or join an exiting one.
package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
)

// installCilium sets up the cilium pod network.
func installCilium(ctx context.Context, helmInstaller helmClient, release helm.Release, in k8sapi.SetupPodNetworkInput) error {
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
func installCiliumGeneric(ctx context.Context, helmInstaller helmClient, release helm.Release, kubeAPIHost, kubeAPIPort string) error {
	if release.Values != nil {
		release.Values["k8sServiceHost"] = kubeAPIHost
		release.Values["k8sServicePort"] = kubeAPIPort
	}
	return helmInstaller.InstallChart(ctx, release)
}

func installCiliumGCP(ctx context.Context, helmInstaller helmClient, release helm.Release, nodeName, nodePodCIDR, subnetworkPodCIDR, kubeAPIHost, kubeAPIPort string) error {
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
