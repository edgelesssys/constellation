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
	"regexp"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/bootstrapper/internal/kubelet"
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/bootstrapper/util"
	"github.com/edgelesssys/constellation/internal/file"
	"github.com/edgelesssys/constellation/internal/logger"
	"github.com/edgelesssys/constellation/internal/versions"
	"github.com/icholy/replace"
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"golang.org/x/text/transform"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

const (
	// kubeConfig is the path to the Kubernetes admin config (used for authentication).
	kubeConfig = "/etc/kubernetes/admin.conf"
	// kubeletStartTimeout is the maximum time given to the kubelet service to (re)start.
	kubeletStartTimeout = 10 * time.Minute
)

var providerIDRegex = regexp.MustCompile(`^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachineScaleSets/([^/]+)/virtualMachines/([^/]+)$`)

// Client provides the functionality of `kubectl apply`.
type Client interface {
	Apply(resources resources.Marshaler, forceConflicts bool) error
	SetKubeconfig(kubeconfig []byte)
	CreateConfigMap(ctx context.Context, configMap corev1.ConfigMap) error
	// TODO: add tolerations
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
	cmd = exec.CommandContext(ctx, kubeadmPath, "init", "-v=5", "--skip-phases=preflight,certs", "--config", initConfigFile.Name())
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

type SetupPodNetworkInput struct {
	CloudProvider     string
	NodeName          string
	FirstNodePodCIDR  string
	SubnetworkPodCIDR string
	ProviderID        string
}

// SetupPodNetwork sets up the cilium pod network.
func (k *KubernetesUtil) SetupPodNetwork(ctx context.Context, in SetupPodNetworkInput) error {
	switch in.CloudProvider {
	case "gcp":
		return k.setupGCPPodNetwork(ctx, in.NodeName, in.FirstNodePodCIDR, in.SubnetworkPodCIDR)
	case "azure":
		return k.setupAzurePodNetwork(ctx, in.ProviderID, in.SubnetworkPodCIDR)
	case "qemu":
		return k.setupQemuPodNetwork(ctx, in.SubnetworkPodCIDR)
	default:
		return fmt.Errorf("unsupported cloud provider %q", in.CloudProvider)
	}
}

func (k *KubernetesUtil) setupAzurePodNetwork(ctx context.Context, providerID, subnetworkPodCIDR string) error {
	matches := providerIDRegex.FindStringSubmatch(providerID)
	if len(matches) != 5 {
		return fmt.Errorf("error splitting providerID %q", providerID)
	}

	ciliumInstall := exec.CommandContext(ctx, "cilium", "install", "--azure-resource-group", matches[2], "--ipam", "azure",
		"--helm-set",
		"tunnel=disabled,enableIPv4Masquerade=true,azure.enabled=true,debug.enabled=true,ipv4NativeRoutingCIDR="+subnetworkPodCIDR+
			",endpointRoutes.enabled=true,encryption.enabled=true,encryption.type=wireguard,l7Proxy=false,egressMasqueradeInterfaces=eth0")
	ciliumInstall.Env = append(os.Environ(), "KUBECONFIG="+kubeConfig)
	out, err := ciliumInstall.CombinedOutput()
	if err != nil {
		err = errors.New(string(out))
		return err
	}
	return nil
}

func (k *KubernetesUtil) setupGCPPodNetwork(ctx context.Context, nodeName, nodePodCIDR, subnetworkPodCIDR string) error {
	out, err := exec.CommandContext(ctx, kubectlPath, "--kubeconfig", kubeConfig, "patch", "node", nodeName, "-p", "{\"spec\":{\"podCIDR\": \""+nodePodCIDR+"\"}}").CombinedOutput()
	if err != nil {
		err = errors.New(string(out))
		return err
	}

	// allow coredns to run on uninitialized nodes (required by cloud-controller-manager)
	err = exec.CommandContext(ctx, kubectlPath, "--kubeconfig", kubeConfig, "-n", "kube-system", "patch", "deployment", "coredns", "--type", "json", "-p", "[{\"op\":\"add\",\"path\":\"/spec/template/spec/tolerations/-\",\"value\":{\"key\":\"node.cloudprovider.kubernetes.io/uninitialized\",\"value\":\"true\",\"effect\":\"NoSchedule\"}},{\"op\":\"add\",\"path\":\"/spec/template/spec/nodeSelector\",\"value\":{\"node-role.kubernetes.io/control-plane\":\"\"}}]").Run()
	if err != nil {
		return err
	}

	ciliumInstall := exec.CommandContext(ctx, "cilium", "install", "--ipam", "kubernetes", "--ipv4-native-routing-cidr", subnetworkPodCIDR,
		"--helm-set", "endpointRoutes.enabled=true,tunnel=disabled,encryption.enabled=true,encryption.type=wireguard,l7Proxy=false")
	ciliumInstall.Env = append(os.Environ(), "KUBECONFIG="+kubeConfig)
	out, err = ciliumInstall.CombinedOutput()
	if err != nil {
		err = errors.New(string(out))
		return err
	}

	return nil
}

// FixCilium fixes https://github.com/cilium/cilium/issues/19958 but instead of a rollout restart of
// the cilium daemonset, it only restarts the local cilium pod.
func (k *KubernetesUtil) FixCilium(nodeNameK8s string) {
	// wait for cilium pod to be healthy
	for {
		time.Sleep(5 * time.Second)
		resp, err := http.Get("http://127.0.0.1:9876/healthz")
		if err != nil {
			fmt.Printf("waiting for local cilium daemonset pod not healthy: %v\n", err)
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
		fmt.Printf("getting pod id failed: %v: %v\n", err, string(out))
		return
	}
	outLines := strings.Split(string(out), "\n")
	fmt.Println(outLines)
	podID := outLines[len(outLines)-2]

	// stop and delete pod
	out, err = exec.CommandContext(context.Background(), "/run/state/bin/crictl", "stopp", podID).CombinedOutput()
	if err != nil {
		fmt.Printf("stopping cilium agent pod failed: %v: %v\n", err, string(out))
		return
	}
	out, err = exec.CommandContext(context.Background(), "/run/state/bin/crictl", "rmp", podID).CombinedOutput()
	if err != nil {
		fmt.Printf("removing cilium agent pod failed: %v: %v\n", err, string(out))
	}
}

func (k *KubernetesUtil) setupQemuPodNetwork(ctx context.Context, subnetworkPodCIDR string) error {
	ciliumInstall := exec.CommandContext(ctx, "cilium", "install", "--encryption", "wireguard", "--helm-set", "ipam.operator.clusterPoolIPv4PodCIDRList="+subnetworkPodCIDR+",endpointRoutes.enabled=true")
	ciliumInstall.Env = append(os.Environ(), "KUBECONFIG="+kubeConfig)
	out, err := ciliumInstall.CombinedOutput()
	if err != nil {
		err = errors.New(string(out))
		return err
	}
	return nil
}

// SetupAutoscaling deploys the k8s cluster autoscaler.
func (k *KubernetesUtil) SetupAutoscaling(kubectl Client, clusterAutoscalerConfiguration resources.Marshaler, secrets resources.Marshaler) error {
	if err := kubectl.Apply(secrets, true); err != nil {
		return fmt.Errorf("applying cluster-autoscaler Secrets: %w", err)
	}
	return kubectl.Apply(clusterAutoscalerConfiguration, true)
}

// SetupJoinService deploys the Constellation node join service.
func (k *KubernetesUtil) SetupJoinService(kubectl Client, joinServiceConfiguration resources.Marshaler) error {
	return kubectl.Apply(joinServiceConfiguration, true)
}

// SetupGCPGuestAgent deploys the GCP guest agent daemon set.
func (k *KubernetesUtil) SetupGCPGuestAgent(kubectl Client, guestAgentDaemonset resources.Marshaler) error {
	return kubectl.Apply(guestAgentDaemonset, true)
}

// SetupCloudControllerManager deploys the k8s cloud-controller-manager.
func (k *KubernetesUtil) SetupCloudControllerManager(kubectl Client, cloudControllerManagerConfiguration resources.Marshaler, configMaps resources.Marshaler, secrets resources.Marshaler) error {
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
func (k *KubernetesUtil) SetupCloudNodeManager(kubectl Client, cloudNodeManagerConfiguration resources.Marshaler) error {
	return kubectl.Apply(cloudNodeManagerConfiguration, true)
}

// SetupAccessManager deploys the constellation-access-manager for deploying SSH keys on control-plane & worker nodes.
func (k *KubernetesUtil) SetupAccessManager(kubectl Client, accessManagerConfiguration resources.Marshaler) error {
	return kubectl.Apply(accessManagerConfiguration, true)
}

// SetupKMS deploys the KMS deployment.
func (k *KubernetesUtil) SetupKMS(kubectl Client, kmsConfiguration resources.Marshaler) error {
	if err := kubectl.Apply(kmsConfiguration, true); err != nil {
		return fmt.Errorf("applying KMS configuration: %w", err)
	}
	return nil
}

// SetupVerificationService deploys the verification service.
func (k *KubernetesUtil) SetupVerificationService(kubectl Client, verificationServiceConfiguration resources.Marshaler) error {
	return kubectl.Apply(verificationServiceConfiguration, true)
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
		constants.KubernetesDir,
		constants.DefaultCertificateDir,
		constants.CACertName,
	))
	if err != nil {
		return err
	}
	parentCertPEM, _ := pem.Decode(parentCertRaw)
	parentCert, err := x509.ParseCertificate(parentCertPEM.Bytes)
	if err != nil {
		return err
	}

	parentKeyRaw, err := k.file.Read(filepath.Join(
		constants.KubernetesDir,
		constants.DefaultCertificateDir,
		constants.CAKeyName,
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
		err = fmt.Errorf("unsupported key type %q", parentCertPEM.Type)
	}
	if err != nil {
		return err
	}

	certRequest, err := x509.ParseCertificateRequest(certRequestRaw)
	if err != nil {
		return err
	}

	serialNumber, err := util.GenerateCertificateSerialNumber()
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
