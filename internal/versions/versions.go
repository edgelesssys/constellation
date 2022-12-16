/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/bootstrapper/initproto"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/joinservice/joinproto"
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
	KonnectivityAgentImage = "registry.k8s.io/kas-network-proxy/proxy-agent:v0.0.33@sha256:48f2a4ec3e10553a81b8dd1c6fa5fe4bcc9617f78e71c1ca89c6921335e2d7da" // renovate:container
	// KonnectivityServerImage server image for konnectivity service.
	KonnectivityServerImage = "registry.k8s.io/kas-network-proxy/proxy-server:v0.0.33@sha256:2c111f004bec24888d8cfa2a812a38fb8341350abac67dcd0ac64e709dfe389c" // renovate:container
	// JoinImage image of Constellation join service.
	JoinImage = "ghcr.io/edgelesssys/constellation/join-service:v2.3.0@sha256:0ffbbd200939480b1d88a5c7cde41ec8f39a3343a2208b459ddd103b8a7133a7" // renovate:container
	// KmsImage image of Constellation KMS server.
	KmsImage = "ghcr.io/edgelesssys/constellation/kmsserver:v2.3.0@sha256:386ba219acffa06d8c202b3746cf7551d782a9656190ed5515cfb96f010bdae7" // renovate:container
	// VerificationImage image of Constellation verification service.
	VerificationImage = "ghcr.io/edgelesssys/constellation/verification-service:v2.3.0@sha256:1281bb548b43b9a056101db684d3700e5eef5e4e72da3e7c8f913baf3fec28fc" // renovate:container
	// GcpGuestImage image for GCP guest agent.
	// Check for new versions at https://github.com/GoogleCloudPlatform/guest-agent/releases and update in /.github/workflows/build-gcp-guest-agent.yml.
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:20220927.00@sha256:3dea1ae3f162d2353e6584b325f0e325a39cda5f380f41e5a0ee43c6641d3905" // renovate:container
	// ConstellationOperatorImage is the image for the constellation node operator.
	ConstellationOperatorImage = "ghcr.io/edgelesssys/constellation/node-operator:v2.3.0@sha256:37ebbc2d32c1884f7a54a5250512ca1cb7e6a434e9eabc6842b1a2b58d8f747d" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "ghcr.io/edgelesssys/constellation/node-maintenance-operator:v0.13.1-alpha1@sha256:e011d428dba3ef66a2a4656a2bf58bcfe89836c62b0a75676f5c12350502a3cf" // renovate:container

	// QEMUMetadataImage image of QEMU metadata api service.
	QEMUMetadataImage = "ghcr.io/edgelesssys/constellation/qemu-metadata-api:v2.3.0@sha256:18a1cfe8bf274d3b2d2c6f3f9c495c5a18b99421c1bfd73ef7fd2a9db7027a56" // renovate:container
	// LibvirtImage image that provides libvirt.
	LibvirtImage = "ghcr.io/edgelesssys/constellation/libvirt:v2.2.0@sha256:81ddc30cd679a95379e94e2f154861d9112bcabfffa96330c09a4917693f7cce" // renovate:container

	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.3.0-pre.0.20221212170906-a77f38efbb31@sha256:cef4d17ed639765f127c16fb5fd896cd3cd2db3c71591929e4df3b4959218395" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.3.0-pre.0.20221212170906-a77f38efbb31@sha256:305a27bcb8f76caa23902c753e0342d5bcfc9c89074a1848c157d8e74c5b5049" // renovate:container

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
//go:generate go run generateHashes.go

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_23: {
		PatchVersion: "v1.23.15", // renovate:kubernetes-release
		KubernetesComponents: ComponentVersions{
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
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.23.23@sha256:a5ff0f4c2ee3438ff5372442f657552dec549afb4fa04aeab90a15f37a466125", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.23.23@sha256:a82d73fb1ee10e3041b4f03cfe4ab5bb8edc8329c45bf1d42ff9e06340137de3", // renovate:container
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.23.1@sha256:cd2101ba67f3d6ec719f7792d4bdaa3a50e1b716f3a9ccee8931086496c655b7", // renovate:container
	},
	V1_24: {
		PatchVersion: "v1.24.9", // renovate:kubernetes-release
		KubernetesComponents: ComponentVersions{
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
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.24.1@sha256:4b75b09cc5b3959d06a8c2fb84f165e8163ec0153eaa6a48ece6c8113e78e720", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v24.0.0@sha256:80e2910509ccb4d99b2e08182c2101fbed64f0663194adae08fc1cf878ecc58b", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.10@sha256:846d631cf2a1abc5450d62e72a5e055377bbb9f7bf3d0aed9dd52acfe26c0e8a", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.10@sha256:fed0573c5200e2ba6874a08b4fa875523958d6e6cebc4831f5798ae8caf4ac8e", // renovate:container
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.24.0@sha256:5bd22353ae7f30c9abfaa08189281367ef47ea1b3d09eb13eb26bd13de241e72", // renovate:container
	},
	V1_25: {
		PatchVersion: "v1.25.5", // renovate:kubernetes-release
		KubernetesComponents: ComponentVersions{
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
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.1@sha256:85d3f1e9dacc72531445989bb10999e1e70ebc409d11be57e5baa5f031a893b0", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v25.2.0@sha256:86fa9d31ed0b3d0d8806f13d6e7debd3471028b2cb7cca3a876d8a31612a7ba5", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.4@sha256:f851de1325e7fffb61ab817db310743574e7d96576984d3351ddde2c840b3ebd", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.4@sha256:5227c3820a60df390107fa0a0865bf19745f21fc3c323c779ac71e3b70e46846", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.0@sha256:f509ffab618dbd07d129b69ec56963aac7f61aaa792851206b54a2f0bbe046df", // renovate:container
	},
	V1_26: {
		PatchVersion: "v1.26.0", // renovate:kubernetes-release
		KubernetesComponents: ComponentVersions{
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
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.1@sha256:85d3f1e9dacc72531445989bb10999e1e70ebc409d11be57e5baa5f031a893b0", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v25.2.0@sha256:86fa9d31ed0b3d0d8806f13d6e7debd3471028b2cb7cca3a876d8a31612a7ba5", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.4@sha256:f851de1325e7fffb61ab817db310743574e7d96576984d3351ddde2c840b3ebd", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.4@sha256:5227c3820a60df390107fa0a0865bf19745f21fc3c323c779ac71e3b70e46846", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.0@sha256:f509ffab618dbd07d129b69ec56963aac7f61aaa792851206b54a2f0bbe046df", // renovate:container
	},
}

// KubernetesVersion bundles download URLs to all version-releated binaries necessary for installing/deploying a particular Kubernetes version.
type KubernetesVersion struct {
	PatchVersion                     string
	KubernetesComponents             ComponentVersions
	CloudControllerManagerImageAWS   string // k8s version dependency.
	CloudControllerManagerImageGCP   string // Using self-built image until resolved: https://github.com/kubernetes/cloud-provider-gcp/issues/289
	CloudControllerManagerImageAzure string // k8s version dependency.
	CloudNodeManagerImageAzure       string // k8s version dependency. Same version as above.
	ClusterAutoscalerImage           string // Matches k8s versioning scheme.
}

// ComponentVersion is a version of a particular artifact.
type ComponentVersion struct {
	URL         string
	Hash        string
	InstallPath string
	Extract     bool
}

// ComponentVersions is a list of ComponentVersion.
type ComponentVersions []ComponentVersion

// NewComponentVersionsFromInitProto converts a protobuf KubernetesVersion to ComponentVersions.
func NewComponentVersionsFromInitProto(protoComponents []*initproto.KubernetesComponent) ComponentVersions {
	components := ComponentVersions{}
	for _, protoComponent := range protoComponents {
		if protoComponent == nil {
			continue
		}
		components = append(components, ComponentVersion{URL: protoComponent.Url, Hash: protoComponent.Hash, InstallPath: protoComponent.InstallPath, Extract: protoComponent.Extract})
	}
	return components
}

// NewComponentVersionsFromJoinProto converts a protobuf KubernetesVersion to ComponentVersions.
func NewComponentVersionsFromJoinProto(protoComponents []*joinproto.KubernetesComponent) ComponentVersions {
	components := ComponentVersions{}
	for _, protoComponent := range protoComponents {
		if protoComponent == nil {
			continue
		}
		components = append(components, ComponentVersion{URL: protoComponent.Url, Hash: protoComponent.Hash, InstallPath: protoComponent.InstallPath, Extract: protoComponent.Extract})
	}
	return components
}

// ToInitProto converts a ComponentVersions to a protobuf KubernetesVersion.
func (c ComponentVersions) ToInitProto() []*initproto.KubernetesComponent {
	protoComponents := []*initproto.KubernetesComponent{}
	for _, component := range c {
		protoComponents = append(protoComponents, &initproto.KubernetesComponent{Url: component.URL, Hash: component.Hash, InstallPath: component.InstallPath, Extract: component.Extract})
	}
	return protoComponents
}

// ToJoinProto converts a ComponentVersions to a protobuf KubernetesVersion.
func (c ComponentVersions) ToJoinProto() []*joinproto.KubernetesComponent {
	protoComponents := []*joinproto.KubernetesComponent{}
	for _, component := range c {
		protoComponents = append(protoComponents, &joinproto.KubernetesComponent{Url: component.URL, Hash: component.Hash, InstallPath: component.InstallPath, Extract: component.Extract})
	}
	return protoComponents
}

// GetHash returns the hash over all component hashes.
func (c ComponentVersions) GetHash() string {
	sha := sha256.New()
	for _, component := range c {
		sha.Write([]byte(component.Hash))
	}

	return fmt.Sprintf("sha256:%x", sha.Sum(nil))
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
