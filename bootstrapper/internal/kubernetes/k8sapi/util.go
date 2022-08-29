package k8sapi

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubelet"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/kubernetes"
	kubeconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/edgelesssys/constellation/internal/crypto"
	"github.com/edgelesssys/constellation/internal/deploy/helm"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/versions"
	"github.com/icholy/replace"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"golang.org/x/text/transform"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
)

const (
	// kubeConfig is the path to the Kubernetes admin config (used for authentication).
	kubeConfig = "/etc/kubernetes/admin.conf"
	// kubeletStartTimeout is the maximum time given to the kubelet service to (re)start.
	kubeletStartTimeout = 10 * time.Minute
	// crdTimeout is the maximum time given to the CRDs to be created.
	crdTimeout = 15 * time.Second
	// helmTimeout is the maximum time given to the helm client.
	helmTimeout = 5 * time.Minute
)

// Client provides the functions to talk to the k8s API.
type Client interface {
	Apply(resources kubernetes.Marshaler, forceConflicts bool) error
	SetKubeconfig(kubeconfig []byte)
	CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error
	AddTolerationsToDeployment(ctx context.Context, tolerations []corev1.Toleration, name string, namespace string) error
	AddNodeSelectorsToDeployment(ctx context.Context, selectors map[string]string, name string, namespace string) error
	WaitForCRDs(ctx context.Context, crds []string) error
}

type installer interface {
	Install(
		ctx context.Context, sourceURL string, destinations []string, perm fs.FileMode,
		extract bool, transforms ...transform.Transformer,
	) error
}

// KubernetesUtil provides low level management of the kubernetes cluster.
type KubernetesUtil struct {
	inst installer
	file file.Handler
}

// NewKubernetesUtil creates a new KubernetesUtil.
func NewKubernetesUtil() *KubernetesUtil {
	return &KubernetesUtil{
		inst: newOSInstaller(),
		file: file.NewHandler(afero.NewOsFs()),
	}
}

// InstallComponents installs kubernetes components in the version specified.
func (k *KubernetesUtil) InstallComponents(ctx context.Context, version versions.ValidK8sVersion) error {
	versionConf := versions.VersionConfigs[version]

	if err := k.inst.Install(
		ctx, versionConf.CNIPluginsURL, []string{cniPluginsDir}, executablePerm, true,
	); err != nil {
		return fmt.Errorf("installing cni plugins: %w", err)
	}
	if err := k.inst.Install(
		ctx, versionConf.CrictlURL, []string{binDir}, executablePerm, true,
	); err != nil {
		return fmt.Errorf("installing crictl: %w", err)
	}
	if err := k.inst.Install(
		ctx, versionConf.KubeletServiceURL, []string{kubeletServiceEtcPath, kubeletServiceStatePath}, systemdUnitPerm, false, replace.String("/usr/bin", binDir),
	); err != nil {
		return fmt.Errorf("installing kubelet service: %w", err)
	}
	if err := k.inst.Install(
		ctx, versionConf.KubeadmConfURL, []string{kubeadmConfEtcPath, kubeadmConfStatePath}, systemdUnitPerm, false, replace.String("/usr/bin", binDir),
	); err != nil {
		return fmt.Errorf("installing kubeadm conf: %w", err)
	}
	if err := k.inst.Install(
		ctx, versionConf.KubeletURL, []string{kubeletPath}, executablePerm, false,
	); err != nil {
		return fmt.Errorf("installing kubelet: %w", err)
	}
	if err := k.inst.Install(
		ctx, versionConf.KubeadmURL, []string{kubeadmPath}, executablePerm, false,
	); err != nil {
		return fmt.Errorf("installing kubeadm: %w", err)
	}
	if err := k.inst.Install(
		ctx, versionConf.KubectlURL, []string{kubectlPath}, executablePerm, false,
	); err != nil {
		return fmt.Errorf("installing kubectl: %w", err)
	}

	return enableSystemdUnit(ctx, kubeletServiceEtcPath)
}

func (k *KubernetesUtil) InitCluster(
	ctx context.Context, initConfig []byte, nodeName string, ips []net.IP, log *logger.Logger,
) error {
	// TODO: audit policy should be user input
	auditPolicy, err := resources.NewDefaultAuditPolicy().Marshal()
	if err != nil {
		return fmt.Errorf("generating default audit policy: %w", err)
	}
	if err := os.WriteFile(auditPolicyPath, auditPolicy, 0o644); err != nil {
		return fmt.Errorf("writing default audit policy: %w", err)
	}

	initConfigFile, err := os.CreateTemp("", "kubeadm-init.*.yaml")
	if err != nil {
		return fmt.Errorf("creating init config file %v: %w", initConfigFile.Name(), err)
	}

	if _, err := initConfigFile.Write(initConfig); err != nil {
		return fmt.Errorf("writing kubeadm init yaml config %v: %w", initConfigFile.Name(), err)
	}

	// preflight
	log.Infof("Running kubeadm preflight checks")
	cmd := exec.CommandContext(ctx, kubeadmPath, "init", "phase", "preflight", "-v=5", "--config", initConfigFile.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("kubeadm init phase preflight failed (code %v) with: %s", exitErr.ExitCode(), out)
		}
		return fmt.Errorf("kubeadm init: %w", err)
	}

	// create CA certs
	log.Infof("Creating Kubernetes control-plane certificates and keys")
	cmd = exec.CommandContext(ctx, kubeadmPath, "init", "phase", "certs", "all", "-v=5", "--config", initConfigFile.Name())
	out, err = cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("kubeadm init phase certs all failed (code %v) with: %s", exitErr.ExitCode(), out)
		}
		return fmt.Errorf("kubeadm init: %w", err)
	}

	// create kubelet key and CA signed certificate for the node
	log.Infof("Creating signed kubelet certificate")
	if err := k.createSignedKubeletCert(nodeName, ips); err != nil {
		return err
	}

	// initialize the cluster
	log.Infof("Initializing the cluster using kubeadm init")
	cmd = exec.CommandContext(ctx, kubeadmPath, "init", "-v=5", "--skip-phases=preflight,certs,addon/kube-proxy", "--config", initConfigFile.Name())
	out, err = cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("kubeadm init failed (code %v) with: %s", exitErr.ExitCode(), out)
		}
		return fmt.Errorf("kubeadm init: %w", err)
	}
	log.With(zap.String("output", string(out))).Infof("kubeadm init succeeded")
	return nil
}

func (k *KubernetesUtil) SetupHelmDeployments(ctx context.Context, kubectl Client, helmDeployments []byte, in SetupPodNetworkInput, log *logger.Logger) error {
	var helmDeploy helm.Deployments
	if err := json.Unmarshal(helmDeployments, &helmDeploy); err != nil {
		return fmt.Errorf("unmarshalling helm deployments: %w", err)
	}
	settings := cli.New()
	settings.KubeConfig = kubeConfig

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), constants.HelmNamespace,
		"secret", log.Infof); err != nil {
		return err
	}

	helmClient := action.NewInstall(actionConfig)
	helmClient.Namespace = constants.HelmNamespace
	helmClient.ReleaseName = "cilium"
	helmClient.Wait = true
	helmClient.Timeout = helmTimeout

	if err := k.deployCilium(ctx, in, helmClient, helmDeploy.Cilium, kubectl); err != nil {
		return fmt.Errorf("deploying cilium: %w", err)
	}

	return nil
}

type SetupPodNetworkInput struct {
	CloudProvider        string
	NodeName             string
	FirstNodePodCIDR     string
	SubnetworkPodCIDR    string
	ProviderID           string
	LoadBalancerEndpoint string
}

// deployCilium sets up the cilium pod network.
func (k *KubernetesUtil) deployCilium(ctx context.Context, in SetupPodNetworkInput, helmClient *action.Install, ciliumDeployment helm.Deployment, kubectl Client) error {
	switch in.CloudProvider {
	case "gcp":
		return k.deployCiliumGCP(ctx, helmClient, kubectl, ciliumDeployment, in.NodeName, in.FirstNodePodCIDR, in.SubnetworkPodCIDR, in.LoadBalancerEndpoint)
	case "azure":
		return k.deployCiliumAzure(ctx, helmClient, ciliumDeployment, in.LoadBalancerEndpoint)
	case "qemu":
		return k.deployCiliumQEMU(ctx, helmClient, ciliumDeployment, in.SubnetworkPodCIDR)
	default:
		return fmt.Errorf("unsupported cloud provider %q", in.CloudProvider)
	}
}

func (k *KubernetesUtil) deployCiliumAzure(ctx context.Context, helmClient *action.Install, ciliumDeployment helm.Deployment, kubeAPIEndpoint string) error {
	host := kubeAPIEndpoint
	ciliumDeployment.Values["k8sServiceHost"] = host
	ciliumDeployment.Values["k8sServicePort"] = strconv.Itoa(constants.KubernetesPort)

	_, err := helmClient.RunWithContext(ctx, ciliumDeployment.Chart, ciliumDeployment.Values)
	if err != nil {
		return fmt.Errorf("installing cilium: %w", err)
	}
	return nil
}

func (k *KubernetesUtil) deployCiliumGCP(ctx context.Context, helmClient *action.Install, kubectl Client, ciliumDeployment helm.Deployment, nodeName, nodePodCIDR, subnetworkPodCIDR, kubeAPIEndpoint string) error {
	out, err := exec.CommandContext(ctx, kubectlPath, "--kubeconfig", kubeConfig, "patch", "node", nodeName, "-p", "{\"spec\":{\"podCIDR\": \""+nodePodCIDR+"\"}}").CombinedOutput()
	if err != nil {
		err = errors.New(string(out))
		return err
	}

	// allow coredns to run on uninitialized nodes (required by cloud-controller-manager)
	tolerations := []corev1.Toleration{
		{
			Key:    "node.cloudprovider.kubernetes.io/uninitialized",
			Value:  "true",
			Effect: "NoSchedule",
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
	ciliumDeployment.Values["ipv4NativeRoutingCIDR"] = subnetworkPodCIDR
	ciliumDeployment.Values["strictModeCIDR"] = subnetworkPodCIDR
	ciliumDeployment.Values["k8sServiceHost"] = host
	if port != "" {
		ciliumDeployment.Values["k8sServicePort"] = port
	}

	_, err = helmClient.RunWithContext(ctx, ciliumDeployment.Chart, ciliumDeployment.Values)
	if err != nil {
		return fmt.Errorf("installing cilium: %w", err)
	}

	return nil
}

// FixCilium fixes https://github.com/cilium/cilium/issues/19958 but instead of a rollout restart of
// the cilium daemonset, it only restarts the local cilium pod.
func (k *KubernetesUtil) FixCilium(nodeNameK8s string, log *logger.Logger) {
	// wait for cilium pod to be healthy
	client := http.Client{}
	for {
		time.Sleep(5 * time.Second)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1:9879/healthz", http.NoBody)
		if err != nil {
			log.With(zap.Error(err)).Errorf("Unable to create request")
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			log.With(zap.Error(err)).Warnf("Waiting for local cilium daemonset pod not healthy")
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == 200 {
			break
		}
	}

	// get cilium pod name
	out, err := exec.CommandContext(context.Background(), "/bin/bash", "-c", "/run/state/bin/crictl ps -o json | jq -r '.containers[] | select(.metadata.name == \"cilium-agent\") | .podSandboxId'").CombinedOutput()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Getting pod id failed: %s", out)
		return
	}
	outLines := strings.Split(string(out), "\n")
	podID := outLines[len(outLines)-2]

	// stop and delete pod
	out, err = exec.CommandContext(context.Background(), "/run/state/bin/crictl", "stopp", podID).CombinedOutput()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Stopping cilium agent pod failed: %s", out)
		return
	}
	out, err = exec.CommandContext(context.Background(), "/run/state/bin/crictl", "rmp", podID).CombinedOutput()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Removing cilium agent pod failed: %s", out)
	}
}

func (k *KubernetesUtil) deployCiliumQEMU(ctx context.Context, helmClient *action.Install, ciliumDeployment helm.Deployment, subnetworkPodCIDR string) error {
	// configure pod network CIDR
	ciliumDeployment.Values["ipam"] = map[string]interface{}{
		"operator": map[string]interface{}{
			"clusterPoolIPv4PodCIDRList": []interface{}{
				subnetworkPodCIDR,
			},
		},
	}

	_, err := helmClient.RunWithContext(ctx, ciliumDeployment.Chart, ciliumDeployment.Values)
	if err != nil {
		return fmt.Errorf("installing cilium: %w", err)
	}
	return nil
}

// SetupAutoscaling deploys the k8s cluster autoscaler.
func (k *KubernetesUtil) SetupAutoscaling(kubectl Client, clusterAutoscalerConfiguration kubernetes.Marshaler, secrets kubernetes.Marshaler) error {
	if err := kubectl.Apply(secrets, true); err != nil {
		return fmt.Errorf("applying cluster-autoscaler Secrets: %w", err)
	}
	return kubectl.Apply(clusterAutoscalerConfiguration, true)
}

// SetupJoinService deploys the Constellation node join service.
func (k *KubernetesUtil) SetupJoinService(kubectl Client, joinServiceConfiguration kubernetes.Marshaler) error {
	return kubectl.Apply(joinServiceConfiguration, true)
}

// SetupGCPGuestAgent deploys the GCP guest agent daemon set.
func (k *KubernetesUtil) SetupGCPGuestAgent(kubectl Client, guestAgentDaemonset kubernetes.Marshaler) error {
	return kubectl.Apply(guestAgentDaemonset, true)
}

// SetupCloudControllerManager deploys the k8s cloud-controller-manager.
func (k *KubernetesUtil) SetupCloudControllerManager(kubectl Client, cloudControllerManagerConfiguration kubernetes.Marshaler, configMaps kubernetes.Marshaler, secrets kubernetes.Marshaler) error {
	if err := kubectl.Apply(configMaps, true); err != nil {
		return fmt.Errorf("applying ccm ConfigMaps: %w", err)
	}
	if err := kubectl.Apply(secrets, true); err != nil {
		return fmt.Errorf("applying ccm Secrets: %w", err)
	}
	if err := kubectl.Apply(cloudControllerManagerConfiguration, true); err != nil {
		return fmt.Errorf("applying ccm: %w", err)
	}
	return nil
}

// SetupCloudNodeManager deploys the k8s cloud-node-manager.
func (k *KubernetesUtil) SetupCloudNodeManager(kubectl Client, cloudNodeManagerConfiguration kubernetes.Marshaler) error {
	return kubectl.Apply(cloudNodeManagerConfiguration, true)
}

// SetupAccessManager deploys the constellation-access-manager for deploying SSH keys on control-plane & worker nodes.
func (k *KubernetesUtil) SetupAccessManager(kubectl Client, accessManagerConfiguration kubernetes.Marshaler) error {
	return kubectl.Apply(accessManagerConfiguration, true)
}

// SetupKMS deploys the KMS deployment.
func (k *KubernetesUtil) SetupKMS(kubectl Client, kmsConfiguration kubernetes.Marshaler) error {
	if err := kubectl.Apply(kmsConfiguration, true); err != nil {
		return fmt.Errorf("applying KMS configuration: %w", err)
	}
	return nil
}

// SetupVerificationService deploys the verification service.
func (k *KubernetesUtil) SetupVerificationService(kubectl Client, verificationServiceConfiguration kubernetes.Marshaler) error {
	return kubectl.Apply(verificationServiceConfiguration, true)
}

func (k *KubernetesUtil) SetupOperatorLifecycleManager(ctx context.Context, kubectl Client, olmCRDs, olmConfiguration kubernetes.Marshaler, crdNames []string) error {
	if err := kubectl.Apply(olmCRDs, true); err != nil {
		return fmt.Errorf("applying OLM CRDs: %w", err)
	}
	crdReadyTimeout, cancel := context.WithTimeout(ctx, crdTimeout)
	defer cancel()
	if err := kubectl.WaitForCRDs(crdReadyTimeout, crdNames); err != nil {
		return fmt.Errorf("waiting for OLM CRDs: %w", err)
	}
	return kubectl.Apply(olmConfiguration, true)
}

func (k *KubernetesUtil) SetupNodeMaintenanceOperator(kubectl Client, nodeMaintenanceOperatorConfiguration kubernetes.Marshaler) error {
	return kubectl.Apply(nodeMaintenanceOperatorConfiguration, true)
}

func (k *KubernetesUtil) SetupNodeOperator(ctx context.Context, kubectl Client, nodeOperatorConfiguration kubernetes.Marshaler) error {
	return kubectl.Apply(nodeOperatorConfiguration, true)
}

// JoinCluster joins existing Kubernetes cluster using kubeadm join.
func (k *KubernetesUtil) JoinCluster(ctx context.Context, joinConfig []byte, log *logger.Logger) error {
	// TODO: audit policy should be user input
	auditPolicy, err := resources.NewDefaultAuditPolicy().Marshal()
	if err != nil {
		return fmt.Errorf("generating default audit policy: %w", err)
	}
	if err := os.WriteFile(auditPolicyPath, auditPolicy, 0o644); err != nil {
		return fmt.Errorf("writing default audit policy: %w", err)
	}

	joinConfigFile, err := os.CreateTemp("", "kubeadm-join.*.yaml")
	if err != nil {
		return fmt.Errorf("creating join config file %v: %w", joinConfigFile.Name(), err)
	}

	if _, err := joinConfigFile.Write(joinConfig); err != nil {
		return fmt.Errorf("writing kubeadm init yaml config %v: %w", joinConfigFile.Name(), err)
	}

	// run `kubeadm join` to join a worker node to an existing Kubernetes cluster
	cmd := exec.CommandContext(ctx, kubeadmPath, "join", "-v=5", "--config", joinConfigFile.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("kubeadm join failed (code %v) with: %s (full err: %s)", exitErr.ExitCode(), out, err)
		}
		return fmt.Errorf("kubeadm join: %w", err)
	}
	log.With(zap.String("output", string(out))).Infof("kubeadm join succeeded")

	return nil
}

// StartKubelet enables and starts the kubelet systemd unit.
func (k *KubernetesUtil) StartKubelet() error {
	ctx, cancel := context.WithTimeout(context.TODO(), kubeletStartTimeout)
	defer cancel()
	if err := enableSystemdUnit(ctx, kubeletServiceEtcPath); err != nil {
		return fmt.Errorf("enabling kubelet systemd unit: %w", err)
	}
	return startSystemdUnit(ctx, "kubelet.service")
}

// RestartKubelet restarts a kubelet.
func (k *KubernetesUtil) RestartKubelet() error {
	ctx, cancel := context.WithTimeout(context.TODO(), kubeletStartTimeout)
	defer cancel()
	return restartSystemdUnit(ctx, "kubelet.service")
}

// createSignedKubeletCert manually creates a Kubernetes CA signed kubelet certificate for the bootstrapper node.
// This is necessary because this node does not request a certificate from the join service.
func (k *KubernetesUtil) createSignedKubeletCert(nodeName string, ips []net.IP) error {
	certRequestRaw, kubeletKey, err := kubelet.GetCertificateRequest(nodeName, ips)
	if err != nil {
		return err
	}
	if err := k.file.Write(kubelet.KeyFilename, kubeletKey, file.OptMkdirAll); err != nil {
		return err
	}

	parentCertRaw, err := k.file.Read(filepath.Join(
		kubeconstants.KubernetesDir,
		kubeconstants.DefaultCertificateDir,
		kubeconstants.CACertName,
	))
	if err != nil {
		return err
	}
	parentCert, err := crypto.PemToX509Cert(parentCertRaw)
	if err != nil {
		return err
	}

	parentKeyRaw, err := k.file.Read(filepath.Join(
		kubeconstants.KubernetesDir,
		kubeconstants.DefaultCertificateDir,
		kubeconstants.CAKeyName,
	))
	if err != nil {
		return err
	}
	parentKeyPEM, _ := pem.Decode(parentKeyRaw)
	var parentKey any
	switch parentKeyPEM.Type {
	case "EC PRIVATE KEY":
		parentKey, err = x509.ParseECPrivateKey(parentKeyPEM.Bytes)
	case "RSA PRIVATE KEY":
		parentKey, err = x509.ParsePKCS1PrivateKey(parentKeyPEM.Bytes)
	case "PRIVATE KEY":
		parentKey, err = x509.ParsePKCS8PrivateKey(parentKeyPEM.Bytes)
	default:
		err = fmt.Errorf("unsupported key type %q", parentKeyPEM.Type)
	}
	if err != nil {
		return err
	}

	certRequest, err := x509.ParseCertificateRequest(certRequestRaw)
	if err != nil {
		return err
	}

	serialNumber, err := crypto.GenerateCertificateSerialNumber()
	if err != nil {
		return err
	}

	now := time.Now()
	// Create the kubelet certificate
	// For a reference on the certificate fields, see: https://kubernetes.io/docs/setup/best-practices/certificates/
	certTmpl := &x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    now.Add(-2 * time.Hour),
		NotAfter:     now.Add(24 * 365 * time.Hour),
		Subject:      certRequest.Subject,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		IsCA:                  false,
		BasicConstraintsValid: true,
		IPAddresses:           certRequest.IPAddresses,
	}

	certRaw, err := x509.CreateCertificate(rand.Reader, certTmpl, parentCert, certRequest.PublicKey, parentKey)
	if err != nil {
		return err
	}
	kubeletCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certRaw,
	})

	return k.file.Write(kubelet.CertificateFilename, kubeletCert, file.OptMkdirAll)
}
