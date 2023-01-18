/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
)

// ValidK8sVersion represents any of the three currently supported k8s versions.
type ValidK8sVersion string

// NewValidK8sVersion validates the given string and produces a new ValidK8sVersion object.
func NewValidK8sVersion(k8sVersion string) (ValidK8sVersion, error) {
	if IsSupportedK8sVersion(k8sVersion) {
		return ValidK8sVersion(k8sVersion), nil
	}
	return "", fmt.Errorf("invalid k8sVersion supplied: %s", k8sVersion)
}

// IsSupportedK8sVersion checks if a given Kubernetes version is supported by Constellation.
func IsSupportedK8sVersion(version string) bool {
	switch version {
	case string(V1_23):
		return true
	case string(V1_24):
		return true
	case string(V1_25):
		return true
	case string(V1_26):
		return true
	default:
		return false
	}
}

// IsPreviewK8sVersion checks if a given Kubernetes version is still in preview and not fully supported.
func IsPreviewK8sVersion(version ValidK8sVersion) bool {
	return false
}

const (
	//
	// Constellation images.
	// These images are built in a way that they support all versions currently listed in VersionConfigs.
	//

	// KonnectivityAgentImage agent image for konnectivity service.
	KonnectivityAgentImage = "registry.k8s.io/kas-network-proxy/proxy-agent:v0.0.35@sha256:8970dca5c4c9df1d566c3c3c91ef2e743e410a8623d42062eb48e7245f1eef93" // renovate:container
	// KonnectivityServerImage server image for konnectivity service.
	KonnectivityServerImage = "registry.k8s.io/kas-network-proxy/proxy-server:v0.0.35@sha256:d863f7fd0da4392b9753dc6c9195a658e80d70e0be8c9adb410d77cf20b75c76" // renovate:container
	// JoinImage image of Constellation join service.
	JoinImage = "ghcr.io/edgelesssys/constellation/join-service:v2.5.0-pre.0.20230116153610-19cea2ede4da@sha256:b1a28e76bc0f3a40b38b59de56ff1143d79e8b5d192777e9eccf3b6f7c546c65" // renovate:container
	// KeyServiceImage image of Constellation KMS server.
	KeyServiceImage = "ghcr.io/edgelesssys/constellation/kmsserver:v2.5.0-pre.0.20230112123617-d0e9f427d1ba@sha256:d4319308eb62e2ee079cc86858acdd1faccc404edec7bfabecf35861284a55f3" // renovate:container
	// VerificationImage image of Constellation verification service.
	VerificationImage = "ghcr.io/edgelesssys/constellation/verification-service:v2.5.0-pre.0.20230116125211-d37bd077d8c6@sha256:c7371bf1c0d6839b87f6fbebbc3864105c051883ea2b969062ca6f96e954d382" // renovate:container
	// GcpGuestImage image for GCP guest agent.
	// Check for new versions at https://github.com/GoogleCloudPlatform/guest-agent/releases and update in /.github/workflows/build-gcp-guest-agent.yml.
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:20220927.00@sha256:3dea1ae3f162d2353e6584b325f0e325a39cda5f380f41e5a0ee43c6641d3905" // renovate:container
	// ConstellationOperatorImage is the image for the constellation node operator.
	ConstellationOperatorImage = "ghcr.io/edgelesssys/constellation/node-operator:v2.5.0-pre.0.20230117150550-f534f1f351f3@sha256:91c119dfa8f3adbf076f288cb29b6ec66e5e788e726fa3e0d239714b21ea518f" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.14.0@sha256:2dffb6ffdbbe997d317799fc709baf030d678bde0be0264931ff6b3e94fd89ab" // renovate:container

	// QEMUMetadataImage image of QEMU metadata api service.
	QEMUMetadataImage = "ghcr.io/edgelesssys/constellation/qemu-metadata-api:v2.5.0-pre.0.20230116125211-d37bd077d8c6@sha256:ceacc71836183daf7e4877ccfa6ea1fb1dc947d16db37a097d95efc4b21a3605" // renovate:container
	// LibvirtImage image that provides libvirt.
	LibvirtImage = "ghcr.io/edgelesssys/constellation/libvirt:v2.2.0@sha256:81ddc30cd679a95379e94e2f154861d9112bcabfffa96330c09a4917693f7cce" // renovate:container

	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.5.0-pre.0.20230113154941-67f8336b9d2c@sha256:f101d92c153537da3eb17e3ddae8a7144da8d5d9adf60fa9be8fd8a744237815" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.5.0-pre.0.20230113154941-67f8336b9d2c@sha256:76c45894627bc2bbb5be98be0ce252881e979141d32762e0b4a39dbc2b07cf23" // renovate:container

	// currently supported versions.
	//nolint:revive
	V1_23 ValidK8sVersion = "1.23"
	//nolint:revive
	V1_24 ValidK8sVersion = "1.24"
	//nolint:revive
	V1_25 ValidK8sVersion = "1.25"
	//nolint:revive
	V1_26 ValidK8sVersion = "1.26"

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_25
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate go run hash-generator/generate.go

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_23: {
		ClusterVersion: "v1.23.15", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:b275772da4026d2161bf8a8b41ed4786754c8a93ebfb6564006d5da7f23831e5",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				URL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.26.0/crictl-v1.26.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:cda5e2143bf19f6b548110ffba0fe3565e03e8743fadd625fee3d62fc4134eed",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.23.15/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:5cf382d911c13c9cc8f770251b3a2fd9399c70ac50337874f670b9078f88231d",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.23.15/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:63329e21be8367628f71978cfc140c74ce9cb0336abd9c4802ca7d20d5dec3c3",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.23.15/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:adab29cf67e04e48f566ce185e3904b5deb389ae1e4d57548fcf8947a49a26f5",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.23.2@sha256:5caf74bfe1c6e1b7b7d40345db52b54eeea7229a8fd73c7db9488ef87dc7a496", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v23.0.0@sha256:bf54ecb58fef5b1358d1dd25b1068598a74adbc7e7622b42a2708d1ed4bdc4bc", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.23.25@sha256:bc24df9f9b46bf28e69778892835d730b90b94a5315b9de51eacd2292c7dd499", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.23.25@sha256:ecf4cbd18a98255151e5692469c658a3067cc84b3b7749a8af5bfa10cb4a060f", // renovate:container
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.23.1@sha256:cd2101ba67f3d6ec719f7792d4bdaa3a50e1b716f3a9ccee8931086496c655b7", // renovate:container
	},
	V1_24: {
		ClusterVersion: "v1.24.9", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:b275772da4026d2161bf8a8b41ed4786754c8a93ebfb6564006d5da7f23831e5",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				URL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.26.0/crictl-v1.26.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:cda5e2143bf19f6b548110ffba0fe3565e03e8743fadd625fee3d62fc4134eed",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.24.9/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:8753b9ae0c3e22f09dafdb4178492582c28874f70844de38dc43eb3fad5ca8bb",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.24.9/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:20406971ae71886f7f8ee7b9a33c885391ae64da561fb679d5819f2ccc19ac9f",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.24.9/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:7e13f33b7379b6c25c3ae055e4389eb3eef168e563f37b5c5f1be672e46b686e",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.24.4@sha256:56f1e111977989a403ae2bb53a2b4d1565d1ce132016efe47cfbe45b635ec9cd", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v24.0.0@sha256:80e2910509ccb4d99b2e08182c2101fbed64f0663194adae08fc1cf878ecc58b", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.12@sha256:4b04685ceb59ed6b64b32e3ee29f37245dcf8f8c53ed95ec4ef455d6b9488ff7", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.12@sha256:fefa39c3e19c6c7500e1500f56c7be3a1a541b375ebd683d4d9d589c5870b0db", // renovate:container
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.24.0@sha256:5bd22353ae7f30c9abfaa08189281367ef47ea1b3d09eb13eb26bd13de241e72", // renovate:container
	},
	V1_25: {
		ClusterVersion: "v1.25.5", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:b275772da4026d2161bf8a8b41ed4786754c8a93ebfb6564006d5da7f23831e5",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				URL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.26.0/crictl-v1.26.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:cda5e2143bf19f6b548110ffba0fe3565e03e8743fadd625fee3d62fc4134eed",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.5/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:16b23e1254830805b892cfccf2687eb3edb4ea54ffbadb8cc2eee6d3b1fab8e6",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.5/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:af0b25c7a995c2d208ef0b9d24b70fe6f390ebb1e3987f4e0f548854ba9a3b87",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.5/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:6a660cd44db3d4bfe1563f6689cbe2ffb28ee4baf3532e04fff2d7b909081c29",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.2@sha256:dcccdfba225e93ba2060a4c0b9072b50b0a564354c37bba6ed3ce89c326db58c", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v25.2.0@sha256:86fa9d31ed0b3d0d8806f13d6e7debd3471028b2cb7cca3a876d8a31612a7ba5", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.6@sha256:b0792c5173725e8b22351d18ebcbfb416b0fe592f583014c1076eb42200e1a55", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.6@sha256:81938646521dcc795f681b8febf40fbec92a971293eb35369936924421bfd348", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.0@sha256:f509ffab618dbd07d129b69ec56963aac7f61aaa792851206b54a2f0bbe046df", // renovate:container
	},
	V1_26: {
		ClusterVersion: "v1.26.0", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:b275772da4026d2161bf8a8b41ed4786754c8a93ebfb6564006d5da7f23831e5",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				URL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.26.0/crictl-v1.26.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:cda5e2143bf19f6b548110ffba0fe3565e03e8743fadd625fee3d62fc4134eed",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.0/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:b64949fe696c77565edbe4100a315b6bf8f0e2325daeb762f7e865f16a6e54b5",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.0/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:72631449f26b7203701a1b99f6914f31859583a0e247c3ac0f6aaf59ca80af19",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.0/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:b6769d8ac6a0ed0f13b307d289dc092ad86180b08f5b5044af152808c04950ae",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.26.0@sha256:fdeb61e3e42ecd9cca868d550ebdb88dd6341d9e91fcfa9a37e227dab2ad22cb", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v25.2.0@sha256:86fa9d31ed0b3d0d8806f13d6e7debd3471028b2cb7cca3a876d8a31612a7ba5", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.26.2@sha256:bd5cff08104068c0ca5b6bdefc4a5fa7d96b60119b698d175fd14309f4f90dcd", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.26.2@sha256:8f9a21851278d5ba634ca39ccb0e75756dd4dbb63aeeb492274e0f05301fccd1", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.26.1@sha256:c0b4ef409e23a79b28e2e9710d7317dbddeab141f4021895ebe90422eba1055c", // renovate:container
	},
}

// KubernetesVersion bundles download URLs to all version-releated binaries necessary for installing/deploying a particular Kubernetes version.
type KubernetesVersion struct {
	ClusterVersion                   string
	KubernetesComponents             components.Components
	CloudControllerManagerImageAWS   string // k8s version dependency.
	CloudControllerManagerImageGCP   string // Using self-built image until resolved: https://github.com/kubernetes/cloud-provider-gcp/issues/289
	CloudControllerManagerImageAzure string // k8s version dependency.
	CloudNodeManagerImageAzure       string // k8s version dependency. Same version as above.
	ClusterAutoscalerImage           string // Matches k8s versioning scheme.
}

// versionFromDockerImage returns the version tag from the image name, e.g. "v1.22.2" from "foocr.io/org/repo:v1.22.2@sha256:3009fj0...".
func versionFromDockerImage(imageName string) string {
	beforeAt, _, _ := strings.Cut(imageName, "@")
	_, version, ok := strings.Cut(beforeAt, ":")
	if !ok {
		panic(fmt.Errorf("failed to extract version from image name, no ':' found in %s", imageName))
	}
	return version
}
