package k8sapi

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/icholy/replace"
	"golang.org/x/text/transform"
)

const (
	cniPluginsDir           = "/opt/cni/bin"
	binDir                  = "/run/state/bin"
	kubeadmPath             = "/run/state/bin/kubeadm"
	kubeletPath             = "/run/state/bin/kubelet"
	kubectlPath             = "/run/state/bin/kubectl"
	kubeletServiceEtcPath   = "/etc/systemd/system/kubelet.service"
	kubeletServiceStatePath = "/run/state/systemd/system/kubelet.service"
	kubeadmConfEtcPath      = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
	kubeadmConfStatePath    = "/run/state/systemd/system/kubelet.service.d/10-kubeadm.conf"
	executablePerm          = 0o544
	systemdUnitPerm         = 0o644
)

// versionConfigs holds download URLs for all required kubernetes components for every supported version.
var versionConfigs map[string]kubernetesVersion = map[string]kubernetesVersion{
	"1.23.6": {
		CNIPluginsURL:     "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.1/crictl-v1.24.1-linux-amd64.tar.gz",
		KubeletServiceURL: "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service",
		KubeadmConfURL:    "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf",
		KubeletURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.23.6/bin/linux/amd64/kubelet",
		KubeadmURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.23.6/bin/linux/amd64/kubeadm",
		KubectlURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.23.6/bin/linux/amd64/kubectl",
	},
}

type kubernetesVersion struct {
	CNIPluginsURL     string
	CrictlURL         string
	KubeletServiceURL string
	KubeadmConfURL    string
	KubeletURL        string
	KubeadmURL        string
	KubectlURL        string
}

// installK8sComponents installs kubernetes components for this version.
// reference: https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#installing-kubeadm-kubelet-and-kubectl .
func (k *kubernetesVersion) installK8sComponents(ctx context.Context, inst installer) error {
	if err := inst.Install(
		ctx, k.CNIPluginsURL, []string{cniPluginsDir}, executablePerm, true,
	); err != nil {
		return fmt.Errorf("failed to install cni plugins: %w", err)
	}
	if err := inst.Install(
		ctx, k.CrictlURL, []string{binDir}, executablePerm, true,
	); err != nil {
		return fmt.Errorf("failed to install crictl: %w", err)
	}
	if err := inst.Install(
		ctx, k.KubeletServiceURL, []string{kubeletServiceEtcPath, kubeletServiceStatePath}, systemdUnitPerm, false, replace.String("/usr/bin", binDir),
	); err != nil {
		return fmt.Errorf("failed to install kubelet service: %w", err)
	}
	if err := inst.Install(
		ctx, k.KubeadmConfURL, []string{kubeadmConfEtcPath, kubeadmConfStatePath}, systemdUnitPerm, false, replace.String("/usr/bin", binDir),
	); err != nil {
		return fmt.Errorf("failed to install kubeadm conf: %w", err)
	}
	if err := inst.Install(
		ctx, k.KubeletURL, []string{kubeletPath}, executablePerm, false,
	); err != nil {
		return fmt.Errorf("failed to install kubelet: %w", err)
	}
	if err := inst.Install(
		ctx, k.KubeadmURL, []string{kubeadmPath}, executablePerm, false,
	); err != nil {
		return fmt.Errorf("failed to install kubeadm: %w", err)
	}
	if err := inst.Install(
		ctx, k.KubectlURL, []string{kubectlPath}, executablePerm, false,
	); err != nil {
		return fmt.Errorf("failed to install kubectl: %w", err)
	}
	return nil
}

type installer interface {
	Install(
		ctx context.Context, sourceURL string, destinations []string, perm fs.FileMode,
		extract bool, transforms ...transform.Transformer,
	) error
}
