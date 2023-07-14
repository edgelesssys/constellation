package helm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/deploy/helm"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes/kubectl"
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

func Install(kubeconfig string) {
	loader := NewLoader(cloudprovider.AWS, "v1.26.6", "constell-aws")
	builder := ChartBuilder{
		i: loader,
	}
	builder.AddChart(awsInfo)
	release, err := builder.Load(helm.WaitModeAtomic)
	if err != nil {
		panic(err)
	}
	installer, err := New(logger.New(logger.PlainLog, -1), kubeconfig)
	if err != nil {
		panic(err)
	}

	kubectl := kubectl.New()
	// Build the rest.Config object from the KUBECONFIG file
	cfgB, err := os.ReadFile(kubeconfig)
	if err != nil {
		panic(fmt.Errorf("failed to read kubeconfig file: %w", err))
	}
	err = kubectl.Initialize(cfgB)
	if err != nil {
		panic(err)
	}
	err = installer.InstallAWSLoadBalancerController(context.Background(), kubectl, release.AWSLoadBalancerController)
	if err != nil {
		panic(err)
	}
}

/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

const (
	// timeout is the maximum time given to the helm Installer.
	timeout = 5 * time.Minute
	// maximumRetryAttempts is the maximum number of attempts to retry a helm install.
	maximumRetryAttempts = 3
)

// Installer is used to install microservice during cluster initialization. It is a wrapper for a helm install action.
type Installer struct {
	*action.Install
	log *logger.Logger
}

// New creates a new Installer with the given logger.
func New(log *logger.Logger, kubeconfig string) (*Installer, error) {
	settings := cli.New()
	settings.KubeConfig = kubeconfig

	actionConfig := &action.Configuration{}
	if err := actionConfig.Init(settings.RESTClientGetter(), constants.HelmNamespace,
		"secret", log.Infof); err != nil {
		return nil, err
	}

	action := action.NewInstall(actionConfig)
	action.Namespace = constants.HelmNamespace
	action.Timeout = timeout

	return &Installer{
		action,
		log,
	}, nil
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

// InstallAWSLoadBalancerController installs the AWS Load Balancer Controller.
// fails when --skip-helm-wait due to needing cert-manager to be ready
func (h *Installer) InstallAWSLoadBalancerController(ctx context.Context, kubectl k8sClient, release helm.Release) error {
	h.ReleaseName = release.ReleaseName
	if err := h.setWaitMode(release.WaitMode); err != nil {
		return err
	}
	err := h.install(ctx, release.Chart, release.Values)
	if err != nil {
		return err
	}
	return nil
}

//// InstallConstellationServices installs the constellation-services chart. In the future this chart should bundle all microservices.
//func (h *Installer) InstallConstellationServices(ctx context.Context, release helm.Release, extraVals map[string]any) error {
//	h.ReleaseName = release.ReleaseName
//	if err := h.setWaitMode(release.WaitMode); err != nil {
//		return err
//	}

//	mergedVals := helm.MergeMaps(release.Values, extraVals)

//	return h.install(ctx, release.Chart, mergedVals)
//}

//// InstallCertManager installs the cert-manager chart.
//func (h *Installer) InstallCertManager(ctx context.Context, release helm.Release) error {
//	h.ReleaseName = release.ReleaseName
//	h.Timeout = 10 * time.Minute
//	if err := h.setWaitMode(release.WaitMode); err != nil {
//		return err
//	}

//	return h.install(ctx, release.Chart, release.Values)
//}

//// InstallOperators installs the Constellation Operators.
//func (h *Installer) InstallOperators(ctx context.Context, release helm.Release, extraVals map[string]any) error {
//	h.ReleaseName = release.ReleaseName
//	if err := h.setWaitMode(release.WaitMode); err != nil {
//		return err
//	}

//	mergedVals := helm.MergeMaps(release.Values, extraVals)

//	return h.install(ctx, release.Chart, mergedVals)
//}

//// InstallCilium sets up the cilium pod network.
//func (h *Installer) InstallCilium(ctx context.Context, kubectl k8sapi.Installer, release helm.Release, in k8sapi.SetupPodNetworkInput) error {
//	h.ReleaseName = release.ReleaseName
//	if err := h.setWaitMode(release.WaitMode); err != nil {
//		return err
//	}

//	timeoutS := int64(10)
//	// allow coredns to run on uninitialized nodes (required by cloud-controller-manager)
//	tolerations := []corev1.Toleration{
//		{
//			Key:    "node.cloudprovider.kubernetes.io/uninitialized",
//			Value:  "true",
//			Effect: corev1.TaintEffectNoSchedule,
//		},
//		{
//			Key:               "node.kubernetes.io/unreachable",
//			Operator:          corev1.TolerationOpExists,
//			Effect:            corev1.TaintEffectNoExecute,
//			TolerationSeconds: &timeoutS,
//		},
//	}
//	if err := kubectl.AddTolerationsToDeployment(ctx, tolerations, "coredns", "kube-system"); err != nil {
//		return fmt.Errorf("failed to add tolerations to coredns deployment: %w", err)
//	}
//	if err := kubectl.EnforceCoreDNSSpread(ctx); err != nil {
//		return fmt.Errorf("failed to enforce CoreDNS spread: %w", err)
//	}

//	switch in.CloudProvider {
//	case "aws", "azure", "openstack", "qemu":
//		return h.installCiliumGeneric(ctx, release, in.LoadBalancerEndpoint)
//	case "gcp":
//		return h.installCiliumGCP(ctx, release, in.NodeName, in.FirstNodePodCIDR, in.SubnetworkPodCIDR, in.LoadBalancerEndpoint)
//	default:
//		return fmt.Errorf("unsupported cloud provider %q", in.CloudProvider)
//	}
//}

//// installCiliumGeneric installs cilium with the given load balancer endpoint.
//// This is used for cloud providers that do not require special server-side configuration.
//// Currently this is AWS, Azure, and QEMU.
//func (h *Installer) installCiliumGeneric(ctx context.Context, release helm.Release, kubeAPIEndpoint string) error {
//	host := kubeAPIEndpoint
//	release.Values["k8sServiceHost"] = host
//	release.Values["k8sServicePort"] = strconv.Itoa(constants.KubernetesPort)

//	return h.install(ctx, release.Chart, release.Values)
//}

// func (h *Installer) installCiliumGCP(ctx context.Context, release helm.Release, nodeName, nodePodCIDR, subnetworkPodCIDR, kubeAPIEndpoint string) error {
//	out, err := exec.CommandContext(ctx, constants.KubectlPath, "--kubeconfig", constants.ControlPlaneAdminConfFilename, "patch", "node", nodeName, "-p", "{\"spec\":{\"podCIDR\": \""+nodePodCIDR+"\"}}").CombinedOutput()
//	if err != nil {
//		err = errors.New(string(out))
//		return err
//	}

//	host, port, err := net.SplitHostPort(kubeAPIEndpoint)
//	if err != nil {
//		return err
//	}

//	// configure pod network CIDR
//	release.Values["ipv4NativeRoutingCIDR"] = subnetworkPodCIDR
//	release.Values["strictModeCIDR"] = subnetworkPodCIDR
//	release.Values["k8sServiceHost"] = host
//	if port != "" {
//		release.Values["k8sServicePort"] = port
//	}

//	return h.install(ctx, release.Chart, release.Values)
//}

// install tries to install the given chart and aborts after ~5 tries.
// The function will wait 30 seconds before retrying a failed installation attempt.
// After 3 tries, the retrier will be canceled and the function returns with an error.
func (h *Installer) install(ctx context.Context, chartRaw []byte, values map[string]any) error {
	var retries int
	retriable := func(err error) bool {
		// abort after maximumRetryAttempts tries.
		if retries >= maximumRetryAttempts {
			return false
		}
		retries++
		// only retry if atomic is set
		// otherwise helm doesn't uninstall
		// the release on failure
		if !h.Atomic {
			return false
		}
		// check if error is retriable
		return wait.Interrupted(err) ||
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

	retryLoopStartTime := time.Now()
	if err := retrier.Do(ctx); err != nil {
		return fmt.Errorf("helm install: %w", err)
	}
	retryLoopFinishDuration := time.Since(retryLoopStartTime)
	h.log.With(zap.String("chart", chart.Name()), zap.Duration("duration", retryLoopFinishDuration)).Infof("Helm chart installation finished")

	return nil
}

func (h *Installer) setWaitMode(waitMode helm.WaitMode) error {
	switch waitMode {
	case helm.WaitModeNone:
		h.Wait = false
		h.Atomic = false
	case helm.WaitModeWait:
		h.Wait = true
		h.Atomic = false
	case helm.WaitModeAtomic:
		h.Wait = true
		h.Atomic = true
	default:
		return fmt.Errorf("unknown wait mode %q", waitMode)
	}
	return nil
}

// installDoer is a help struct to enable retrying helm's install action.
type installDoer struct {
	Installer *Installer
	chart     *chart.Chart
	values    map[string]any
	log       *logger.Logger
}

// Do logs which chart is installed and tries to install it.
func (i installDoer) Do(ctx context.Context) error {
	i.log.With(zap.String("chart", i.chart.Name())).Infof("Trying to install Helm chart")

	if _, err := i.Installer.RunWithContext(ctx, i.chart, i.values); err != nil {
		i.log.With(zap.Error(err), zap.String("chart", i.chart.Name())).Errorf("Helm chart installation failed")
		return err
	}

	return nil
}
