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
// Returns an empty string if the given version is invalid.
// strict controls whether the patch version is checked or not.
// If strict is false, the patch version is ignored and the returned
// ValidK8sVersion is a supported patch version for the given major.minor version.
func NewValidK8sVersion(k8sVersion string, strict bool) (ValidK8sVersion, error) {
	var supported bool
	if strict {
		supported = isSupportedK8sVersionStrict(k8sVersion)
	} else {
		supported = isSupportedK8sVersion(k8sVersion)
	}
	if !supported {
		return "", fmt.Errorf("invalid Kubernetes version: %s; supported versions are %v", k8sVersion, SupportedK8sVersions())
	}
	if !strict {
		k8sVersion, _ = supportedVersionForMajorMinor(k8sVersion)
	}

	return ValidK8sVersion(k8sVersion), nil
}

// IsSupportedK8sVersion checks if a given Kubernetes minor version is supported by Constellation.
// Note: the patch version is not checked!
func isSupportedK8sVersion(version string) bool {
	for _, valid := range SupportedK8sVersions() {
		if semver.MajorMinor(valid) == semver.MajorMinor(version) {
			return true
		}
	}
	return false
}

// IsSupportedK8sVersion checks if a given Kubernetes version is supported by Constellation.
func isSupportedK8sVersionStrict(version string) bool {
	for _, valid := range SupportedK8sVersions() {
		if valid == version {
			return true
		}
	}
	return false
}

// IsPreviewK8sVersion checks if a given Kubernetes version is still in preview and not fully supported.
func IsPreviewK8sVersion(_ ValidK8sVersion) bool {
	return false
}

func supportedVersionForMajorMinor(majorMinor string) (string, bool) {
	majorMinor = semver.MajorMinor(majorMinor)
	for _, valid := range SupportedK8sVersions() {
		if semver.MajorMinor(valid) == majorMinor {
			return valid, true
		}
	}
	return "", false
}

const (
	//
	// Constellation images.
	// These images are built in a way that they support all versions currently listed in VersionConfigs.
	//

	// KonnectivityAgentImage agent image for konnectivity service.
	KonnectivityAgentImage = "registry.k8s.io/kas-network-proxy/proxy-agent:v0.1.2@sha256:cd3046d253d26ffb5907c625e0d0c2be05c5693c90e12116980851739fc0ead8" // renovate:container
	// KonnectivityServerImage server image for konnectivity service.
	KonnectivityServerImage = "registry.k8s.io/kas-network-proxy/proxy-server:v0.1.2@sha256:79933c3779bc30e33bb7509dff913e70f6ba78ad441f4827f0f3e840ce5f3ddb" // renovate:container
	// GcpGuestImage image for GCP guest agent.
	// Check for new versions at https://github.com/GoogleCloudPlatform/guest-agent/releases and update in /.github/workflows/build-gcp-guest-agent.yml.
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:v20230628.0.0@sha256:e76f66b20be7e30f7e3bfd2b37e068a39874706cd9c0a43c1500e74fe39df797" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.15.0@sha256:8cb8dad93283268282c30e75c68f4bd76b28def4b68b563d2f9db9c74225d634" // renovate:container
	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.7.0-pre.0.20230405123345-6bf3c63115a5@sha256:1e2c396538be7571138272f8a54e3412d4ff91ee370880f89894501a2555706a" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.7.0-pre.0.20230405123345-6bf3c63115a5@sha256:abd739853af4981c3a4b338bb3a27433284525d7ebdb84adfc77f1873c41de93" // renovate:container

	// currently supported versions.
	//nolint:revive
	V1_25 ValidK8sVersion = "v1.25.12" // renovate:kubernetes-release
	//nolint:revive
	V1_26 ValidK8sVersion = "v1.26.7" // renovate:kubernetes-release
	//nolint:revive
	V1_27 ValidK8sVersion = "v1.27.4" // renovate:kubernetes-release

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_26
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate hash-generator

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_25: {
		ClusterVersion: "v1.25.12", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.2.0/cni-plugins-linux-amd64-v1.2.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:f3a841324845ca6bf0d4091b4fc7f97e18a623172158b72fc3fdcdb9d42d2d37",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				URL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:d335d6e16c309fbc3ff1a29a7e49bb253b5c9b4b030990bf7c6b48687f985cee",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.12/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:7aa7d0b4512e6d79ada2017c054b07aaf30d4dc0d740449364a5e2c26e2c1842",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.12/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:293252f0a1727bfad4ef4fe99d704a56ecea45e39b0ea77f629c55da39e377da",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.12/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:75842752ea07cb8ee2210df40faa7c61e1317e76d5c7968e380cae83447d4a0f",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.3@sha256:47eb1c1e6a3bd6d0fb44ac4992885b6218f1448ea339de778d8b703df463c06f", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.17@sha256:db168815dd1801bbde18e585875c155bece8dd2077514abb91de6ec82bc5e97b", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.17@sha256:89f7bcf23a7674714103800306b2efd18f8b9f42477c6ac7687a99692adb236d", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v25.2.0@sha256:86fa9d31ed0b3d0d8806f13d6e7debd3471028b2cb7cca3a876d8a31612a7ba5", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.25.5", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.3@sha256:a8c8f3c5e5fc532c37ab9cf5388d0061ecbf78090b91c3d059b09fa3ffd4ac53", // renovate:container
	},
	V1_26: {
		ClusterVersion: "v1.26.7", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.2.0/cni-plugins-linux-amd64-v1.2.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:f3a841324845ca6bf0d4091b4fc7f97e18a623172158b72fc3fdcdb9d42d2d37",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				URL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:d335d6e16c309fbc3ff1a29a7e49bb253b5c9b4b030990bf7c6b48687f985cee",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.7/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:2926ea2cd7fcd644d24a258bdf21e1a8cfd95412b1079914ca46466dae1d74f2",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.7/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:812e6d0e94a3fc77d3e9d09dbe709190b77408936cc4e960d916e8401be11090",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.7/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:d9dc7741e5f279c28ef32fbbe1daa8ebc36622391c33470efed5eb8426959971",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.26.6@sha256:33445ab57f48938fe989ffe311dacee0044b82f2bd23cb7f7b563275926f0ce9", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.26.13@sha256:d44cd3490d3ab7a4bf11faa4c8b918864be041f8b183dcedc75caf6fb9d1fdf1", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.26.13@sha256:ba8c73fc49495ed69d4242eee3068609ff2752d4c3f51de740385b05a4c303f1", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v26.0.1@sha256:db2b15a20ad690784a6015bfad55c4dff15826be8cf9f6ac77d70abd11b1f70c", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.2", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.26.4@sha256:f771284ff54ecfedf40c7af70c5450600786c98989aeb69cdcf7e7bb7ac5a20d", // renovate:container
	},
	V1_27: {
		ClusterVersion: "v1.27.4", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.2.0/cni-plugins-linux-amd64-v1.2.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:f3a841324845ca6bf0d4091b4fc7f97e18a623172158b72fc3fdcdb9d42d2d37",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				URL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:d335d6e16c309fbc3ff1a29a7e49bb253b5c9b4b030990bf7c6b48687f985cee",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.4/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:385f65878dc8b48df0f2bd369535ff273390518b5ac2cc1a1684d65619324704",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.4/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:7be21d6fb3707fbbe8f0db0403db6234c8af773b941f931bf8248759ee988bcd",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.4/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:4685bfcf732260f72fce58379e812e091557ef1dfc1bc8084226c7891dd6028f",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.27.2@sha256:42be09a2b13b4e69b42905639d6b005ebe1ca490aabefad427256abf2cc892c7", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.27.7@sha256:e27c4ddc8b9efdac8509a2f681eaa98152309f6b2f08d14230b11c60a9b8b87f", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.27.7@sha256:998453989b03ac6c24e53aabbf271fa181e054b3a061fe6caa565186ae79bd0c", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v26.0.1@sha256:db2b15a20ad690784a6015bfad55c4dff15826be8cf9f6ac77d70abd11b1f70c", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.2", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.27.3@sha256:0e1ab1bfeb1beaa82f59356ef36364503df22aeb8f8d0d7383bac449b4e808fb", // renovate:container
	},
}

// KubernetesVersion bundles download URLs to all version-releated binaries necessary for installing/deploying a particular Kubernetes version.
type KubernetesVersion struct {
	ClusterVersion                       string
	KubernetesComponents                 components.Components
	CloudControllerManagerImageAWS       string // k8s version dependency.
	CloudControllerManagerImageAzure     string // k8s version dependency.
	CloudControllerManagerImageGCP       string // Using self-built image until resolved: https://github.com/kubernetes/cloud-provider-gcp/issues/289
	CloudControllerManagerImageOpenStack string // k8s version dependency.
	CloudNodeManagerImageAzure           string // k8s version dependency. Same version as above.
	ClusterAutoscalerImage               string // Matches k8s versioning scheme.
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
