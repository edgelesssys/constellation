package k8sapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

const (
	// kubeConfig is the path to the Kubernetes admin config (used for authentication).
	kubeConfig = "/etc/kubernetes/admin.conf"
	// kubeletStartTimeout is the maximum time given to the kubelet service to (re)start.
	kubeletStartTimeout = 10 * time.Minute
)

var (
	kubernetesKeyRegexp = regexp.MustCompile("[a-f0-9]{64}")
	providerIDRegex     = regexp.MustCompile(`^azure:///subscriptions/([^/]+)/resourceGroups/([^/]+)/providers/Microsoft.Compute/virtualMachineScaleSets/([^/]+)/virtualMachines/([^/]+)$`)
)

// Client provides the functionality of `kubectl apply`.
type Client interface {
	Apply(resources resources.Marshaler, forceConflicts bool) error
	SetKubeconfig(kubeconfig []byte)
	// TODO: add tolerations
}

type ClusterUtil interface {
	InstallComponents(ctx context.Context, version string) error
	InitCluster(initConfig []byte) error
	JoinCluster(joinConfig []byte) error
	SetupPodNetwork(kubectl Client, podNetworkConfiguration resources.Marshaler) error
	SetupAccessManager(kubectl Client, accessManagerConfiguration resources.Marshaler) error
	SetupAutoscaling(kubectl Client, clusterAutoscalerConfiguration resources.Marshaler, secrets resources.Marshaler) error
	SetupCloudControllerManager(kubectl Client, cloudControllerManagerConfiguration resources.Marshaler, configMaps resources.Marshaler, secrets resources.Marshaler) error
	SetupCloudNodeManager(kubectl Client, cloudNodeManagerConfiguration resources.Marshaler) error
	SetupKMS(kubectl Client, kmsConfiguration resources.Marshaler) error
	StartKubelet() error
	RestartKubelet() error
	GetControlPlaneJoinCertificateKey() (string, error)
	CreateJoinToken(ttl time.Duration) (*kubeadm.BootstrapTokenDiscovery, error)
}

// KubernetesUtil provides low level management of the kubernetes cluster.
type KubernetesUtil struct {
	inst installer
}

// NewKubernetesUtils creates a new KubernetesUtil.
func NewKubernetesUtil() *KubernetesUtil {
	return &KubernetesUtil{
		inst: newOSInstaller(),
	}
}

// InstallComponents installs kubernetes components in the version specified.
func (k *KubernetesUtil) InstallComponents(ctx context.Context, version string) error {
	var versionConf kubernetesVersion
	var ok bool
	if versionConf, ok = versionConfigs[version]; !ok {
		return fmt.Errorf("unsupported kubernetes version %q", version)
	}
	if err := versionConf.installK8sComponents(ctx, k.inst); err != nil {
		return err
	}

	return enableSystemdUnit(ctx, kubeletServiceEtcPath)
}

func (k *KubernetesUtil) InitCluster(ctx context.Context, initConfig []byte) error {
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
	defer os.Remove(initConfigFile.Name())

	if _, err := initConfigFile.Write(initConfig); err != nil {
		return fmt.Errorf("writing kubeadm init yaml config %v: %w", initConfigFile.Name(), err)
	}

	cmd := exec.CommandContext(ctx, kubeadmPath, "init", "--config", initConfigFile.Name())
	_, err = cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("kubeadm init failed (code %v) with: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return fmt.Errorf("kubeadm init: %w", err)
	}
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
		return k.setupQemuPodNetwork(ctx)
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

func (k *KubernetesUtil) setupQemuPodNetwork(ctx context.Context) error {
	ciliumInstall := exec.CommandContext(ctx, "cilium", "install", "--encryption", "wireguard", "--helm-set", "ipam.operator.clusterPoolIPv4PodCIDRList=10.244.0.0/16,endpointRoutes.enabled=true")
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

// SetupActivationService deploys the Constellation node activation service.
func (k *KubernetesUtil) SetupActivationService(kubectl Client, activationServiceConfiguration resources.Marshaler) error {
	return kubectl.Apply(activationServiceConfiguration, true)
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

// JoinCluster joins existing Kubernetes cluster using kubeadm join.
func (k *KubernetesUtil) JoinCluster(ctx context.Context, joinConfig []byte) error {
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
	defer os.Remove(joinConfigFile.Name())

	if _, err := joinConfigFile.Write(joinConfig); err != nil {
		return fmt.Errorf("writing kubeadm init yaml config %v: %w", joinConfigFile.Name(), err)
	}

	// run `kubeadm join` to join a worker node to an existing Kubernetes cluster
	cmd := exec.CommandContext(ctx, kubeadmPath, "join", "--config", joinConfigFile.Name())
	if _, err := cmd.Output(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("kubeadm join failed (code %v) with: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return fmt.Errorf("kubeadm join: %w", err)
	}

	return nil
}

// SetupKMS deploys the KMS deployment.
func (k *KubernetesUtil) SetupKMS(kubectl Client, kmsConfiguration resources.Marshaler) error {
	if err := kubectl.Apply(kmsConfiguration, true); err != nil {
		return fmt.Errorf("applying KMS configuration: %w", err)
	}
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

// GetControlPlaneJoinCertificateKey return the key which can be used in combination with the joinArgs
// to join the Cluster as control-plane.
func (k *KubernetesUtil) GetControlPlaneJoinCertificateKey(ctx context.Context) (string, error) {
	// Key will be valid for 1h (no option to reduce the duration).
	// https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-init-phase/#cmd-phase-upload-certs
	output, err := exec.CommandContext(ctx, kubeadmPath, "init", "phase", "upload-certs", "--upload-certs").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("kubeadm upload-certs failed (code %v) with: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return "", fmt.Errorf("kubeadm upload-certs: %w", err)
	}
	// Example output:
	/*
		[upload-certs] Storing the certificates in ConfigMap "kubeadm-certs" in the "kube-system" Namespace
		[upload-certs] Using certificate key:
		9555b74008f24687eb964bd90a164ecb5760a89481d9c55a77c129b7db438168
	*/
	key := kubernetesKeyRegexp.FindString(string(output))
	if key == "" {
		return "", fmt.Errorf("failed to parse kubeadm output: %s", string(output))
	}
	return key, nil
}

// CreateJoinToken creates a new bootstrap (join) token.
func (k *KubernetesUtil) CreateJoinToken(ctx context.Context, ttl time.Duration) (*kubeadm.BootstrapTokenDiscovery, error) {
	output, err := exec.CommandContext(ctx, kubeadmPath, "token", "create", "--ttl", ttl.String(), "--print-join-command").Output()
	if err != nil {
		return nil, fmt.Errorf("kubeadm token create: %w", err)
	}
	// `kubeadm token create [...] --print-join-command` outputs the following format:
	// kubeadm join [API_SERVER_ENDPOINT] --token [TOKEN] --discovery-token-ca-cert-hash [DISCOVERY_TOKEN_CA_CERT_HASH]
	return ParseJoinCommand(string(output))
}
