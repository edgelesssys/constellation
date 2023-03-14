/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package helm is used to install Constellation microservices and other services during cluster initialization.
package helm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/retry"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// timeout is the maximum time given to the helm client.
	timeout = 5 * time.Minute
)

// Client is used to install microservice during cluster initialization. It is a wrapper for a helm install action.
type Client struct {
	*action.Install
	log *logger.Logger
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
		log,
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
		return h.installCiliumGCP(ctx, kubectl, release, in.NodeName, in.NodeIP, in.FirstNodePodCIDR, in.SubnetworkPodCIDR, in.LoadBalancerEndpoint)
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

func (h *Client) installCiliumGCP(ctx context.Context, kubectl k8sapi.Client, release helm.Release, nodeName, nodeIP, nodePodCIDR, subnetworkPodCIDR, kubeAPIEndpoint string) error {
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

	// On GCP we might want to have a higher MTU since we can use jumbo frames.
	// Cilium's auto-detection usually handles this.
	// However, it can get a wrong value when multiple network devices exist (e.g. from Podman).
	// Thus, we set it manually here from the expected external interface.
	// Tracking issue: https://github.com/cilium/cilium/issues/14339
	// This part could be way easier using some tricks (e.g. just hardcode ens3).
	// But not sure if this will blow up in the future, so let's better ask the operating system instead of doing tricks.
	hostMTU, err := getHostMTU(nodeIP, h.log)
	if err != nil {
		h.log.With(zap.Error(err)).Warnf("Failed to determine MTU from host network interface, falling back to Cilium's auto-detection")
	} else {
		h.log.With(zap.Int("mtu", hostMTU)).Infof("Detected host MTU for Cilium")
	}

	// configure pod network CIDR
	release.Values["ipv4NativeRoutingCIDR"] = subnetworkPodCIDR
	release.Values["strictModeCIDR"] = subnetworkPodCIDR
	release.Values["k8sServiceHost"] = host
	if port != "" {
		release.Values["k8sServicePort"] = port
	}
	if hostMTU != 0 {
		release.Values["MTU"] = hostMTU
	}

	if err := h.install(ctx, release.Chart, release.Values); err != nil {
		return err
	}

	return nil
}

// install tries to install the given chart and aborts after ~5 tries.
// The function will wait 30 seconds before retrying a failed installation attempt.
// After 5 minutes the retrier will be canceld and the function returns with an error.
func (h *Client) install(ctx context.Context, chartRaw []byte, values map[string]any) error {
	retriable := func(err error) bool {
		return errors.Is(err, wait.ErrWaitTimeout) ||
			strings.Contains(err.Error(), "connection refused")
	}

	reader := bytes.NewReader(chartRaw)
	chart, err := loader.LoadArchive(reader)
	if err != nil {
		return fmt.Errorf("helm load archive: %w", err)
	}

	doer := installDoer{
		h,
		chart,
		values,
		h.log,
	}
	retrier := retry.NewIntervalRetrier(doer, 30*time.Second, retriable)

	// Since we have no precise retry condition we want to stop retrying after 5 minutes.
	// The helm library only reports a timeout error in the error cases we currently know.
	// Other errors will not be retried.
	newCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if err := retrier.Do(newCtx); err != nil {
		return fmt.Errorf("helm install: %w", err)
	}
	return nil
}

// installDoer is a help struct to enable retrying helm's install action.
type installDoer struct {
	client *Client
	chart  *chart.Chart
	values map[string]any
	log    *logger.Logger
}

// Do logs which chart is installed and tries to install it.
func (i installDoer) Do(ctx context.Context) error {
	i.log.With(zap.String("chart", i.chart.Name())).Infof("Trying to install helm chart")

	_, err := i.client.RunWithContext(ctx, i.chart, i.values)

	return err
}

// getHostMTU gets the host network interface and its MTU from a passed node IP.
func getHostMTU(nodeIP string, log *logger.Logger) (int, error) {
	parsedNodeIP := net.ParseIP(nodeIP)
	if parsedNodeIP == nil {
		return 0, fmt.Errorf("failed to parse node IP from string to IP: %s", nodeIP)
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return 0, fmt.Errorf("retrieving network interfaces: %w", err)
	}

	var nodeNetworkInterfaceMTU int
	var foundNodeNetworkInterface bool
	for _, i := range ifaces {
		// Abort if network interface has already been found.
		if foundNodeNetworkInterface {
			break
		}

		addrs, err := i.Addrs()
		if err != nil {
			log.With(zap.String("interface", i.Name)).Warnf("Failed to retrieve interface addresses")
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.Contains(parsedNodeIP) {
					nodeNetworkInterfaceMTU = i.MTU
					foundNodeNetworkInterface = true
					break
				}
			case *net.IPAddr:
				if nodeIP == v.IP.String() {
					nodeNetworkInterfaceMTU = i.MTU
					foundNodeNetworkInterface = true
					break
				}
			}
		}
	}

	if nodeNetworkInterfaceMTU == 0 {
		return 0, fmt.Errorf("did not find network interface with node IP: %s", nodeIP)
	}

	return nodeNetworkInterfaceMTU, nil
}
