/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package k8sapi

import (
	"context"
	"crypto/rand"
	"crypto/x509"
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

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubelet"
	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/role"
	kubeconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/edgelesssys/constellation/v2/internal/crypto"
	"github.com/edgelesssys/constellation/v2/internal/file"
	"github.com/edgelesssys/constellation/v2/internal/logger"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"golang.org/x/text/transform"
	corev1 "k8s.io/api/core/v1"
)

const (
	// kubeletStartTimeout is the maximum time given to the kubelet service to (re)start.
	kubeletStartTimeout = 10 * time.Minute
	// crdTimeout is the maximum time given to the CRDs to be created.
	crdTimeout = 30 * time.Second
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
		ctx, versionConf.KubectlURL, []string{constants.KubectlPath}, executablePerm, false,
	); err != nil {
		return fmt.Errorf("installing kubectl: %w", err)
	}

	return enableSystemdUnit(ctx, kubeletServicePath)
}

func (k *KubernetesUtil) InitCluster(
	ctx context.Context, initConfig []byte, nodeName string, ips []net.IP, controlPlaneEndpoint string, conformanceMode bool, log *logger.Logger,
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

	log.Infof("Preparing node for Konnectivity")
	if err := k.prepareControlPlaneForKonnectivity(ctx, controlPlaneEndpoint); err != nil {
		return fmt.Errorf("setup konnectivity: %w", err)
	}

	// initialize the cluster
	log.Infof("Initializing the cluster using kubeadm init")
	skipPhases := "--skip-phases=preflight,certs"
	if !conformanceMode {
		skipPhases += ",addon/kube-proxy"
	}

	cmd = exec.CommandContext(ctx, kubeadmPath, "init", "-v=5", skipPhases, "--config", initConfigFile.Name())
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

func (k *KubernetesUtil) prepareControlPlaneForKonnectivity(ctx context.Context, loadBalancerEndpoint string) error {
	if !strings.Contains(loadBalancerEndpoint, ":") {
		loadBalancerEndpoint = net.JoinHostPort(loadBalancerEndpoint, strconv.Itoa(constants.KubernetesPort))
	}

	if err := os.MkdirAll("/etc/kubernetes/manifests", os.ModePerm); err != nil {
		return fmt.Errorf("creating static pods directory: %w", err)
	}

	konnectivityServerYaml, err := resources.NewKonnectivityServerStaticPod().Marshal()
	if err != nil {
		return fmt.Errorf("generating konnectivity server static pod: %w", err)
	}
	if err := os.WriteFile("/etc/kubernetes/manifests/konnectivity-server.yaml", konnectivityServerYaml, 0o644); err != nil {
		return fmt.Errorf("writing konnectivity server pod: %w", err)
	}

	egressConfigYaml, err := resources.NewEgressSelectorConfiguration().Marshal()
	if err != nil {
		return fmt.Errorf("generating egress selector configuration: %w", err)
	}
	if err := os.WriteFile("/etc/kubernetes/egress-selector-configuration.yaml", egressConfigYaml, 0o644); err != nil {
		return fmt.Errorf("writing egress selector config: %w", err)
	}

	if err := k.createSignedKonnectivityCert(); err != nil {
		return fmt.Errorf("generating konnectivity server certificate: %w", err)
	}

	if out, err := exec.CommandContext(ctx, constants.KubectlPath, "config", "set-credentials", "--kubeconfig", "/etc/kubernetes/konnectivity-server.conf", "system:konnectivity-server",
		"--client-certificate", "/etc/kubernetes/konnectivity.crt", "--client-key", "/etc/kubernetes/konnectivity.key", "--embed-certs=true").CombinedOutput(); err != nil {
		return fmt.Errorf("konnectivity kubeconfig set-credentials: %w, %s", err, string(out))
	}
	if out, err := exec.CommandContext(ctx, constants.KubectlPath, "--kubeconfig", "/etc/kubernetes/konnectivity-server.conf", "config", "set-cluster", "kubernetes", "--server", "https://"+loadBalancerEndpoint,
		"--certificate-authority", "/etc/kubernetes/pki/ca.crt", "--embed-certs=true").CombinedOutput(); err != nil {
		return fmt.Errorf("konnectivity kubeconfig set-cluster: %w, %s", err, string(out))
	}
	if out, err := exec.CommandContext(ctx, constants.KubectlPath, "--kubeconfig", "/etc/kubernetes/konnectivity-server.conf", "config", "set-context", "system:konnectivity-server@kubernetes",
		"--cluster", "kubernetes", "--user", "system:konnectivity-server").CombinedOutput(); err != nil {
		return fmt.Errorf("konnectivity kubeconfig set-context: %w, %s", err, string(out))
	}
	if out, err := exec.CommandContext(ctx, constants.KubectlPath, "--kubeconfig", "/etc/kubernetes/konnectivity-server.conf", "config", "use-context", "system:konnectivity-server@kubernetes").CombinedOutput(); err != nil {
		return fmt.Errorf("konnectivity kubeconfig use-context: %w, %s", err, string(out))
	}
	// cleanup
	if err := os.Remove("/etc/kubernetes/konnectivity.crt"); err != nil {
		return fmt.Errorf("removing konnectivity certificate: %w", err)
	}
	if err := os.Remove("/etc/kubernetes/konnectivity.key"); err != nil {
		return fmt.Errorf("removing konnectivity key: %w", err)
	}
	return nil
}

func (k *KubernetesUtil) SetupKonnectivity(kubectl Client, konnectivityAgentsDaemonSet kubernetes.Marshaler) error {
	return kubectl.Apply(konnectivityAgentsDaemonSet, true)
}

type SetupPodNetworkInput struct {
	CloudProvider        string
	NodeName             string
	FirstNodePodCIDR     string
	SubnetworkPodCIDR    string
	ProviderID           string
	LoadBalancerEndpoint string
}

// FixCilium fixes https://github.com/cilium/cilium/issues/19958 but instead of a rollout restart of
// the cilium daemonset, it only restarts the local cilium pod.
func (k *KubernetesUtil) FixCilium(log *logger.Logger) {
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

	// get cilium container id
	out, err := exec.CommandContext(context.Background(), "/run/state/bin/crictl", "ps", "--name", "cilium-agent", "-q").CombinedOutput()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Getting cilium container id failed: %s", out)
		return
	}
	outLines := strings.Split(string(out), "\n")
	if len(outLines) < 2 {
		log.Errorf("Getting cilium container id returned invalid output: %s", out)
		return
	}
	containerID := outLines[len(outLines)-2]

	// get cilium pod id
	out, err = exec.CommandContext(context.Background(), "/run/state/bin/crictl", "inspect", "-o", "go-template", "--template", "{{ .info.sandboxID }}", containerID).CombinedOutput()
	if err != nil {
		log.With(zap.Error(err)).Errorf("Getting cilium pod id failed: %s", out)
		return
	}
	outLines = strings.Split(string(out), "\n")
	if len(outLines) < 2 {
		log.Errorf("Getting cilium pod id returned invalid output: %s", out)
		return
	}
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

// SetupAutoscaling deploys the k8s cluster autoscaler.
func (k *KubernetesUtil) SetupAutoscaling(kubectl Client, clusterAutoscalerConfiguration kubernetes.Marshaler, secrets kubernetes.Marshaler) error {
	if err := kubectl.Apply(secrets, true); err != nil {
		return fmt.Errorf("applying cluster-autoscaler Secrets: %w", err)
	}
	return kubectl.Apply(clusterAutoscalerConfiguration, true)
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
func (k *KubernetesUtil) JoinCluster(ctx context.Context, joinConfig []byte, peerRole role.Role, controlPlaneEndpoint string, log *logger.Logger) error {
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

	if peerRole == role.ControlPlane {
		log.Infof("Prep Init Kubernetes cluster")
		if err := k.prepareControlPlaneForKonnectivity(ctx, controlPlaneEndpoint); err != nil {
			return fmt.Errorf("setup konnectivity: %w", err)
		}
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
	if err := enableSystemdUnit(ctx, kubeletServicePath); err != nil {
		return fmt.Errorf("enabling kubelet systemd unit: %w", err)
	}
	return startSystemdUnit(ctx, "kubelet.service")
}

// createSignedKubeletCert manually creates a Kubernetes CA signed kubelet certificate for the bootstrapper node.
// This is necessary because this node does not request a certificate from the join service.
func (k *KubernetesUtil) createSignedKubeletCert(nodeName string, ips []net.IP) error {
	// Create CSR
	certRequestRaw, kubeletKey, err := kubelet.GetCertificateRequest(nodeName, ips)
	if err != nil {
		return err
	}
	if err := k.file.Write(kubelet.KeyFilename, kubeletKey, file.OptMkdirAll); err != nil {
		return err
	}

	certRequest, err := x509.ParseCertificateRequest(certRequestRaw)
	if err != nil {
		return err
	}

	// Prepare certificate signing
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

	parentCert, parentKey, err := k.getKubernetesCACertAndKey()
	if err != nil {
		return err
	}

	// Sign the certificate
	certRaw, err := x509.CreateCertificate(rand.Reader, certTmpl, parentCert, certRequest.PublicKey, parentKey)
	if err != nil {
		return err
	}

	// Write the certificate
	kubeletCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certRaw,
	})

	return k.file.Write(kubelet.CertificateFilename, kubeletCert, file.OptMkdirAll)
}

// createSignedKonnectivityCert manually creates a Kubernetes CA signed certificate for the Konnectivity server.
func (k *KubernetesUtil) createSignedKonnectivityCert() error {
	// Create CSR
	certRequestRaw, keyPem, err := resources.GetKonnectivityCertificateRequest()
	if err != nil {
		return err
	}
	if err := k.file.Write(resources.KonnectivityKeyFilename, keyPem, file.OptMkdirAll); err != nil {
		return err
	}

	certRequest, err := x509.ParseCertificateRequest(certRequestRaw)
	if err != nil {
		return err
	}

	// Prepare certificate signing
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
	}

	parentCert, parentKey, err := k.getKubernetesCACertAndKey()
	if err != nil {
		return err
	}

	// Sign the certificate
	certRaw, err := x509.CreateCertificate(rand.Reader, certTmpl, parentCert, certRequest.PublicKey, parentKey)
	if err != nil {
		return err
	}

	// Write the certificate
	konnectivityCert := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certRaw,
	})

	return k.file.Write(resources.KonnectivityCertificateFilename, konnectivityCert, file.OptMkdirAll)
}

// getKubernetesCACertAndKey returns the Kubernetes CA certificate and key.
// The key of type `any` can be consumed by `x509.CreateCertificate()`.
func (k *KubernetesUtil) getKubernetesCACertAndKey() (*x509.Certificate, any, error) {
	parentCertRaw, err := k.file.Read(filepath.Join(
		kubeconstants.KubernetesDir,
		kubeconstants.DefaultCertificateDir,
		kubeconstants.CACertName,
	))
	if err != nil {
		return nil, nil, err
	}
	parentCert, err := crypto.PemToX509Cert(parentCertRaw)
	if err != nil {
		return nil, nil, err
	}

	parentKeyRaw, err := k.file.Read(filepath.Join(
		kubeconstants.KubernetesDir,
		kubeconstants.DefaultCertificateDir,
		kubeconstants.CAKeyName,
	))
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}
	return parentCert, parentKey, nil
}
