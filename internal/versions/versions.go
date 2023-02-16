/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package versions defines the supported versions of Constellation components.

Binaries and container image versions are pinned by their hashes, the generate tool can be found in the hash-generator subpackage.
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
	JoinImage = "ghcr.io/edgelesssys/constellation/join-service:v2.5.2@sha256:223c845045659fcee648de020aea2c24332f15d05a5694519c15bfe99929cf09" // renovate:container
	// KeyServiceImage image of Constellation KMS server.
	KeyServiceImage = "ghcr.io/edgelesssys/constellation/key-service:v2.5.2@sha256:869e22f0976fb3f2e2d6e1734d72528476b36c45c2899a8c5877390cfc607e76" // renovate:container
	// VerificationImage image of Constellation verification service.
	VerificationImage = "ghcr.io/edgelesssys/constellation/verification-service:v2.5.2@sha256:c2e67c06ba6ea159204b974c9085123cd19d6550202607e3202b65b138eb655b" // renovate:container
	// GcpGuestImage image for GCP guest agent.
	// Check for new versions at https://github.com/GoogleCloudPlatform/guest-agent/releases and update in /.github/workflows/build-gcp-guest-agent.yml.
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:20220927.00@sha256:3dea1ae3f162d2353e6584b325f0e325a39cda5f380f41e5a0ee43c6641d3905" // renovate:container
	// ConstellationOperatorImage is the image for the constellation node operator.
	ConstellationOperatorImage = "ghcr.io/edgelesssys/constellation/node-operator:v2.5.2@sha256:73cdd7a018e0084346f337836c870914b2b9b61fe71a464e975d973e90a97ff5" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.14.0@sha256:2dffb6ffdbbe997d317799fc709baf030d678bde0be0264931ff6b3e94fd89ab" // renovate:container

	// QEMUMetadataImage image of QEMU metadata api service.
	QEMUMetadataImage = "ghcr.io/edgelesssys/constellation/qemu-metadata-api:v2.5.2@sha256:3efb686de1172128f4878765aaf8ae000fcb7670c25326a3e813d793e4950dc3" // renovate:container
	// LibvirtImage image that provides libvirt.
	LibvirtImage = "ghcr.io/edgelesssys/constellation/libvirt:v2.2.0@sha256:81ddc30cd679a95379e94e2f154861d9112bcabfffa96330c09a4917693f7cce" // renovate:container

	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.5.0-pre.0.20230120132332-a31d79e9cb71@sha256:17f8555581d8916d8121c6ce00f85974e62df55898a890c9855e830856c8cdf7" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.5.0-pre.0.20230120132332-a31d79e9cb71@sha256:9cdfa372c836325979aeeab74f23c1b31e9d757ef8ea95a362133c649a464b02" // renovate:container

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
		ClusterVersion: "v1.23.16", // renovate:kubernetes-release
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
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.23.16/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:ab4c64a54c3c6de57108b7b83bdcf03413285f06adccfa2bbc5535d20a684682",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.23.16/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:511b8d915b8e7ab3c0e6aef905c7ac7901c39d2c67fbc099dda28b498a9871d8",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.23.16/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:5f914edc9dbfbe1b8b8dc0f5dbbac28720a8dffeb940e3339c371e3612c37e48",
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
		ClusterVersion: "v1.24.10", // renovate:kubernetes-release
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
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.24.10/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:34b1731df37d1762662bd91f1cba57a9d2ee86296813c48c4e52a9d7955a1b9e",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.24.10/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:5e29917dc277a8bc4b90bf9dbed8d3dca903fd7cbf7f12c2e256fe22e9f2a1f9",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.24.10/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:d8e9cd9bb073ff09e2f2a74cf48e94a9b9d4f2fa2e2dd91b68b01f64e7061a3b",
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
		ClusterVersion: "v1.25.6", // renovate:kubernetes-release
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
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.6/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:8485ac4a60455b77a9b518c13b3aeb0d32338ab7e9894a0b5d217fea585cd2be",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.6/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:d8bf16d1a808dce10d4eb9b391ddd6ee8a81e94c669441f20b1227083dbc4417",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.6/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:ba876aef0e9d7e2e8fedac036ec194de5ec9b6d2953e30ff82a2758c6ba32174",
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
		ClusterVersion: "v1.26.1", // renovate:kubernetes-release
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
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.1/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:8b99dd73f309ca1ac4005db638e82f949ffcfb877a060089ec0e729503db8198",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.1/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:1531abfe96e2e9d8af9219192c65d04df8507a46a081ae1e101478e95d2b63da",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.1/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:d57be22cfa25f7427cfb538cfc8853d763878f8b36c76ce93830f6f2d67c6e5d",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.26.0@sha256:fdeb61e3e42ecd9cca868d550ebdb88dd6341d9e91fcfa9a37e227dab2ad22cb", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v26.0.1@sha256:3d4d0187d7d9f9af1b98811cfd8b6c3e37a4800f8d22f11320f07cf072382358", // renovate:container
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
