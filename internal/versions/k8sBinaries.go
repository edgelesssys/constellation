package versions

// versionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs map[string]KubernetesVersion = map[string]KubernetesVersion{
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

type KubernetesVersion struct {
	CNIPluginsURL     string
	CrictlURL         string
	KubeletServiceURL string
	KubeadmConfURL    string
	KubeletURL        string
	KubeadmURL        string
	KubectlURL        string
}

func IsSupportedK8sVersion(version string) bool {
	if _, ok := VersionConfigs[version]; !ok {
		return false
	}
	return true
}
