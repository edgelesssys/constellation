package k8sapi

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
)

// kubeConfig is the path to the Kubernetes admin config (used for authentication).
const kubeConfig = "/etc/kubernetes/admin.conf"

// Client provides the functionality of `kubectl apply`.
type Client interface {
	Apply(resources resources.Marshaler, forceConflicts bool) error
	SetKubeconfig(kubeconfig []byte)
	// TODO: add tolerations
}

type ClusterUtil interface {
	InitCluster(initConfig []byte) (*kubeadm.BootstrapTokenDiscovery, error)
	JoinCluster(joinConfig []byte) error
	SetupPodNetwork(kubectl Client, podNetworkConfiguration resources.Marshaler) error
	SetupAutoscaling(kubectl Client, clusterAutoscalerConfiguration resources.Marshaler, secrets resources.Marshaler) error
	SetupCloudControllerManager(kubectl Client, cloudControllerManagerConfiguration resources.Marshaler, configMaps resources.Marshaler, secrets resources.Marshaler) error
	SetupCloudNodeManager(kubectl Client, cloudNodeManagerConfiguration resources.Marshaler) error
	RestartKubelet() error
	GetControlPlaneJoinCertificateKey() (string, error)
}

type KubernetesUtil struct{}

func (k *KubernetesUtil) InitCluster(initConfig []byte) (*kubeadm.BootstrapTokenDiscovery, error) {
	initConfigFile, err := os.CreateTemp("", "kubeadm-init.*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to create init config file %v: %w", initConfigFile.Name(), err)
	}
	defer os.Remove(initConfigFile.Name())

	if _, err := initConfigFile.Write(initConfig); err != nil {
		return nil, fmt.Errorf("writing kubeadm init yaml config %v failed: %w", initConfigFile.Name(), err)
	}

	cmd := exec.Command("kubeadm", "init", "--config", initConfigFile.Name())
	stdout, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("kubeadm init failed (code %v) with: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return nil, fmt.Errorf("kubeadm init failed: %w", err)
	}

	stdoutStr := string(stdout)
	indexKubeadmJoin := strings.Index(stdoutStr, "kubeadm join")
	if indexKubeadmJoin < 0 {
		return nil, errors.New("kubeadm init did not return join command")
	}

	joinCommand := strings.ReplaceAll(stdoutStr[indexKubeadmJoin:], "\\\n", " ")
	// `kubeadm init` returns the two join commands, each broken up into two lines with backslash + newline in between.
	// The following functions assume that stdoutStr[indexKubeadmJoin:] look like the following string.

	// -----------------------------------------------------------------------------------------------
	// --- When modifying the kubeadm.InitConfiguration make sure that this assumption still holds ---
	// -----------------------------------------------------------------------------------------------

	// "kubeadm join 127.0.0.1:16443 --token vlhjr4.9l6lhek0b9v65m67 \
	//	--discovery-token-ca-cert-hash sha256:2b5343a162e31b70602e3cab3d87189dc10431e869633c4db63c3bfcd038dee6 \
	//	--control-plane
	//
	// Then you can join any number of worker nodes by running the following on each as root:
	//
	// kubeadm join 127.0.0.1:16443 --token vlhjr4.9l6lhek0b9v65m67 \
	//  --discovery-token-ca-cert-hash sha256:2b5343a162e31b70602e3cab3d87189dc10431e869633c4db63c3bfcd038dee6"

	// Splits the string into a slice, where earch slice-element contains one line from the previous string
	splittedJoinCommand := strings.SplitN(joinCommand, "\n", 2)
	joinConfig, err := ParseJoinCommand(splittedJoinCommand[0])
	if err != nil {
		return nil, err
	}

	// create extra join token without expiration
	cmd = exec.Command("kubeadm", "token", "create", "--ttl", "0")
	joinToken, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("kubeadm token create failed (code %v) with: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return nil, fmt.Errorf("kubeadm token create failed: %w", err)
	}
	joinConfig.Token = strings.TrimSpace(string(joinToken))
	return joinConfig, nil
}

// SetupPodNetwork sets up the flannel pod network.
func (k *KubernetesUtil) SetupPodNetwork(kubectl Client, podNetworkConfiguration resources.Marshaler) error {
	if err := kubectl.Apply(podNetworkConfiguration, true); err != nil {
		return err
	}

	// allow coredns to run on uninitialized nodes (required by cloud-controller-manager)
	err := exec.Command("kubectl", "--kubeconfig", kubeConfig, "-n", "kube-system", "patch", "deployment", "coredns", "--type", "json", "-p", "[{\"op\":\"add\",\"path\":\"/spec/template/spec/tolerations/-\",\"value\":{\"key\":\"node.cloudprovider.kubernetes.io/uninitialized\",\"value\":\"true\",\"effect\":\"NoSchedule\"}}]").Run()
	if err != nil {
		return err
	}
	return exec.Command("kubectl", "--kubeconfig", kubeConfig, "-n", "kube-system", "patch", "deployment", "coredns", "--type", "json", "-p", "[{\"op\":\"add\",\"path\":\"/spec/template/spec/tolerations/-\",\"value\":{\"key\":\"node.kubernetes.io/network-unavailable\",\"value\":\"\",\"effect\":\"NoSchedule\"}}]").Run()
}

// SetupAutoscaling deploys the k8s cluster autoscaler.
func (k *KubernetesUtil) SetupAutoscaling(kubectl Client, clusterAutoscalerConfiguration resources.Marshaler, secrets resources.Marshaler) error {
	if err := kubectl.Apply(secrets, true); err != nil {
		return fmt.Errorf("applying cluster-autoscaler Secrets failed: %w", err)
	}
	return kubectl.Apply(clusterAutoscalerConfiguration, true)
}

// SetupCloudControllerManager deploys the k8s cloud-controller-manager.
func (k *KubernetesUtil) SetupCloudControllerManager(kubectl Client, cloudControllerManagerConfiguration resources.Marshaler, configMaps resources.Marshaler, secrets resources.Marshaler) error {
	if err := kubectl.Apply(configMaps, true); err != nil {
		return fmt.Errorf("applying ccm ConfigMaps failed: %w", err)
	}
	if err := kubectl.Apply(secrets, true); err != nil {
		return fmt.Errorf("applying ccm Secrets failed: %w", err)
	}
	if err := kubectl.Apply(cloudControllerManagerConfiguration, true); err != nil {
		return fmt.Errorf("applying ccm failed: %w", err)
	}
	return nil
}

// SetupCloudNodeManager deploys the k8s cloud-node-manager.
func (k *KubernetesUtil) SetupCloudNodeManager(kubectl Client, cloudNodeManagerConfiguration resources.Marshaler) error {
	return kubectl.Apply(cloudNodeManagerConfiguration, true)
}

// JoinCluster joins existing Kubernetes cluster using kubeadm join.
func (k *KubernetesUtil) JoinCluster(joinConfig []byte) error {
	joinConfigFile, err := os.CreateTemp("", "kubeadm-join.*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create join config file %v: %w", joinConfigFile.Name(), err)
	}
	defer os.Remove(joinConfigFile.Name())

	if _, err := joinConfigFile.Write(joinConfig); err != nil {
		return fmt.Errorf("writing kubeadm init yaml config %v failed: %w", joinConfigFile.Name(), err)
	}

	// run `kubeadm join` to join a worker node to an existing Kubernetes cluster
	cmd := exec.Command("kubeadm", "join", "--config", joinConfigFile.Name())
	if _, err := cmd.Output(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("kubeadm join failed (code %v) with: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return fmt.Errorf("kubeadm join failed: %w", err)
	}
	return nil
}

// RestartKubelet restarts a kubelet.
func (k *KubernetesUtil) RestartKubelet() error {
	return RestartSystemdUnit("kubelet.service")
}

// GetControlPlaneJoinCertificateKey return the key which can be used in combination with the joinArgs
// to join the Cluster as control-plane.
func (k *KubernetesUtil) GetControlPlaneJoinCertificateKey() (string, error) {
	// Key will be valid for 1h (no option to reduce the duration).
	// https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-init-phase/#cmd-phase-upload-certs
	output, err := exec.Command("kubeadm", "init", "phase", "upload-certs", "--upload-certs").Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("kubeadm upload-certs failed (code %v) with: %s", exitErr.ExitCode(), exitErr.Stderr)
		}
		return "", fmt.Errorf("kubeadm upload-certs failed: %w", err)
	}
	// Example output:
	/*
		[upload-certs] Storing the certificates in ConfigMap "kubeadm-certs" in the "kube-system" Namespace
		[upload-certs] Using certificate key:
		9555b74008f24687eb964bd90a164ecb5760a89481d9c55a77c129b7db438168
	*/
	key := regexp.MustCompile("[a-f0-9]{64}").FindString(string(output))
	if key == "" {
		return "", fmt.Errorf("failed to parse kubeadm output: %s", string(output))
	}
	return key, nil
}
