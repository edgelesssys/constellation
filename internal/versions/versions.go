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
	"sort"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/versions/components"
	"golang.org/x/mod/semver"
)

// SupportedK8sVersions returns a list (sorted) of supported Kubernetes versions.
func SupportedK8sVersions() []string {
	validVersions := make([]string, len(VersionConfigs))
	i := 0
	for _, conf := range VersionConfigs {
		validVersions[i] = compatibility.EnsurePrefixV(conf.ClusterVersion)
		i++
	}
	validVersionsSorted := semver.ByVersion(validVersions)
	sort.Sort(validVersionsSorted)

	return validVersionsSorted
}

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
	for _, valid := range SupportedK8sVersions() {
		if valid == version {
			return true
		}
	}
	return false
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
	KonnectivityAgentImage = "registry.k8s.io/kas-network-proxy/proxy-agent:v0.1.1@sha256:939c42e815e6b6af3181f074652c0d18fe429fcee9b49c1392aee7e92887cfef" // renovate:container
	// KonnectivityServerImage server image for konnectivity service.
	KonnectivityServerImage = "registry.k8s.io/kas-network-proxy/proxy-server:v0.1.1@sha256:b1389e7014425a1752aac55f5043ef4c52edaef0e223bf4d48ed1324e298087c" // renovate:container
	// JoinImage image of Constellation join service.
	JoinImage = "ghcr.io/edgelesssys/constellation/join-service:v2.7.0-pre.0.20230310103644-64e1f553d10f@sha256:2f40f9d1825174876a816e9f2956dd22a35b86a467dbbad13a2893a9f7fc887e" // renovate:container
	// KeyServiceImage image of Constellation KMS server.
	KeyServiceImage = "ghcr.io/edgelesssys/constellation/key-service:v2.7.0-pre.0.20230310103644-64e1f553d10f@sha256:b55ae3e8b588bc2b6cb04f8b9f4b55a6689abf83801805816e8821dbff876ffe" // renovate:container
	// VerificationImage image of Constellation verification service.
	VerificationImage = "ghcr.io/edgelesssys/constellation/verification-service:v2.7.0-pre.0.20230310103644-64e1f553d10f@sha256:609d98a5066538b180bae3713aa8d245438b5af1fcf7958fcc87be6b50934ce4" // renovate:container
	// GcpGuestImage image for GCP guest agent.
	// Check for new versions at https://github.com/GoogleCloudPlatform/guest-agent/releases and update in /.github/workflows/build-gcp-guest-agent.yml.
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:20230221.0@sha256:8be328a5d8d601170b82481d413cf326b20c5219c016633f1651e35d95f1d6f1" // renovate:container
	// ConstellationOperatorImage is the image for the constellation node operator.
	ConstellationOperatorImage = "ghcr.io/edgelesssys/constellation/node-operator:v2.7.0-pre.0.20230310174017-cb2d2b0b89ea@sha256:fc8ca604fe192fd1168bb250b1312ec3b5957c23252bfeaccf89e432d8b49f5f" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.14.0@sha256:2dffb6ffdbbe997d317799fc709baf030d678bde0be0264931ff6b3e94fd89ab" // renovate:container

	// QEMUMetadataImage image of QEMU metadata api service.
	QEMUMetadataImage = "ghcr.io/edgelesssys/constellation/qemu-metadata-api:v2.7.0-pre.0.20230310103644-64e1f553d10f@sha256:82218cff0aa080ae79948928a58769d230991fb6b59e5b24ddbef8a91c71d5c7" // renovate:container
	// LibvirtImage image that provides libvirt.
	LibvirtImage = "ghcr.io/edgelesssys/constellation/libvirt:v2.7.0-pre.0.20230309142342-bdba9d8ba6c3@sha256:98e04e2732d53490f5e00d117c7e22863dc501660f939a244b7d94a09290c572" // renovate:container

	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.7.0-pre.0.20230313093557-a658368d4082@sha256:fd343d0d5d64ddbefafbc7256ef67b7c0b824188160d25f80b0443a0fb8c5cf4" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.7.0-pre.0.20230313093557-a658368d4082@sha256:947ece0e58a398e7a37a5c36b3c2f49c95b206e785d96f396cc15dddb262c1f8" // renovate:container

	// currently supported versions.
	//nolint:revive
	V1_24 ValidK8sVersion = "v1.24.11" // renovate:kubernetes-release
	//nolint:revive
	V1_25 ValidK8sVersion = "v1.25.7" // renovate:kubernetes-release
	//nolint:revive
	V1_26 ValidK8sVersion = "v1.26.2" // renovate:kubernetes-release

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_25
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate go run hash-generator/generate.go

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_24: {
		ClusterVersion: "v1.24.11", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.2.0/cni-plugins-linux-amd64-v1.2.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:f3a841324845ca6bf0d4091b4fc7f97e18a623172158b72fc3fdcdb9d42d2d37",
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
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.24.11/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:1d4849fa3aee4847b8ccecc7e14b16caeb0d507a20d0250cf57e71155248f74f",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.24.11/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:9123cdda3118772794198f784e4f39f34212ebaa919bdddaf611190d7c4abf38",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.24.11/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:c8bdf1b12d5ac91d163c07e61b9527ef718bec6a00f4fd4cf071591218f59be5",
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
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.15@sha256:7a1c3838813992965ddffce53894814d8ad7fd2cb57ab63ca99cd6d260d55dc8", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.15@sha256:f603eb39a850fe297115ef3a31ef5181c4e82eb67cb938e47ab126191e5dd609", // renovate:container
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.24.0@sha256:5bd22353ae7f30c9abfaa08189281367ef47ea1b3d09eb13eb26bd13de241e72", // renovate:container
	},
	V1_25: {
		ClusterVersion: "v1.25.7", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.2.0/cni-plugins-linux-amd64-v1.2.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:f3a841324845ca6bf0d4091b4fc7f97e18a623172158b72fc3fdcdb9d42d2d37",
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
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.7/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:2e3216ac291c78d82fb8988c15d9fd4cf14e2ddd9b17ff91e3abf2e5f3e14fd9",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.7/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:54e369043d5c7ac320ccbd51757019274dbfefce36c9abee746e387ac8203704",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.7/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:6cdbaf3fdd1032fc8e560ccc0a75b5bd6fa5b6cb45491e9677872f511131ad3d",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.3@sha256:47eb1c1e6a3bd6d0fb44ac4992885b6218f1448ea339de778d8b703df463c06f", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v25.2.0@sha256:86fa9d31ed0b3d0d8806f13d6e7debd3471028b2cb7cca3a876d8a31612a7ba5", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.9@sha256:f84018518a4e1a66a53836541fce00e3446ab2b174e66e4f99e7a34d2eef9288", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.9@sha256:695eaf2af202ef4490ac08fe857bae7fb1e93eb597a4f9050cf0a600ebc58028", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.0@sha256:f509ffab618dbd07d129b69ec56963aac7f61aaa792851206b54a2f0bbe046df", // renovate:container
	},
	V1_26: {
		ClusterVersion: "v1.26.2", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.2.0/cni-plugins-linux-amd64-v1.2.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:f3a841324845ca6bf0d4091b4fc7f97e18a623172158b72fc3fdcdb9d42d2d37",
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
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.2/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:e6dd2ee432a093492936ff8505f084b5ed41662f50231f1c11ae08ee8582a3f5",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.2/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:277d880dc6d79994fd333e49d42943b7c9183b1c4ffdbf9da59f806acec7fd82",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.2/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:fcf86d21fb1a49b012bce7845cf00081d2dd7a59f424b28621799deceb5227b3",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.26.0@sha256:fdeb61e3e42ecd9cca868d550ebdb88dd6341d9e91fcfa9a37e227dab2ad22cb", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v26.0.1@sha256:db2b15a20ad690784a6015bfad55c4dff15826be8cf9f6ac77d70abd11b1f70c", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.26.5@sha256:c03adf0aed398701a7d58433d9692b522c594c7e30c01db70304da3cf59389a7", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.26.5@sha256:7322001fc7dec8a85a9c5b962f97166d972fd140247ce85af34191bf6665219e", // renovate:container
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
