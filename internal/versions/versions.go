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
func NewValidK8sVersion(k8sVersion string, strict bool) (ValidK8sVersion, error) {
	var supported bool
	if strict {
		supported = isSupportedK8sVersionStrict(k8sVersion)
	} else {
		supported = isSupportedK8sVersion(k8sVersion)
	}

	if supported {
		return ValidK8sVersion(k8sVersion), nil
	}
	return "", fmt.Errorf("invalid Kubernetes version: %s", k8sVersion)
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
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.14.0@sha256:2dffb6ffdbbe997d317799fc709baf030d678bde0be0264931ff6b3e94fd89ab" // renovate:container
	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.7.0-pre.0.20230405123345-6bf3c63115a5@sha256:1e2c396538be7571138272f8a54e3412d4ff91ee370880f89894501a2555706a" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.7.0-pre.0.20230405123345-6bf3c63115a5@sha256:abd739853af4981c3a4b338bb3a27433284525d7ebdb84adfc77f1873c41de93" // renovate:container

	// currently supported versions.
	//nolint:revive
	V1_25 ValidK8sVersion = "v1.25.11" // renovate:kubernetes-release
	//nolint:revive
	V1_26 ValidK8sVersion = "v1.26.6" // renovate:kubernetes-release
	//nolint:revive
	V1_27 ValidK8sVersion = "v1.27.3" // renovate:kubernetes-release

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_26
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate hash-generator

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_25: {
		ClusterVersion: "v1.25.11", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.2.0/cni-plugins-linux-amd64-v1.2.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:f3a841324845ca6bf0d4091b4fc7f97e18a623172158b72fc3fdcdb9d42d2d37",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				URL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:0c1a0f9900c15ee7a55e757bcdc220faca5dd2e1cfc120459ad1f04f08598127",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.11/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:4801700e29405e49a7e51cccb806decd65ca3a5068d459a40be3b4c5846b9a46",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.11/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:6ff43cc8266a21c7b62878a0a9507b085bbb079a37b095fab5bcd31f2dbd80e0",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.25.11/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:d12bc7d26313546827683ff7b79d0cb2e7ac17cdad4dce138ed518e478b148a7",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.3@sha256:47eb1c1e6a3bd6d0fb44ac4992885b6218f1448ea339de778d8b703df463c06f", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.15@sha256:504be9fe564fa3e36a5b44d2e7990b2a0c096ceecfde81df34e46334495aef67", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.15@sha256:496c91b640adfb27654a112ed4b86afbc72753854142432c44ac6e1db8b03add", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v25.2.0@sha256:86fa9d31ed0b3d0d8806f13d6e7debd3471028b2cb7cca3a876d8a31612a7ba5", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.25.5", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.25.2@sha256:e1507a57738ccba5fbe7b313aad80f0c9822680eadca1a742b84c988f17287e5", // renovate:container
	},
	V1_26: {
		ClusterVersion: "v1.26.6", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				URL:         "https://github.com/containernetworking/plugins/releases/download/v1.2.0/cni-plugins-linux-amd64-v1.2.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:f3a841324845ca6bf0d4091b4fc7f97e18a623172158b72fc3fdcdb9d42d2d37",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				URL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:0c1a0f9900c15ee7a55e757bcdc220faca5dd2e1cfc120459ad1f04f08598127",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.6/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:da82477404414eb342d6b93533f372aa1c41956a57517453ef3d39ebbfdf8cc2",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.6/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:ba699c3c26aaf64ef46d34621de9f3b62e37656943e09f23dc3bf5aa7b3f5094",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.6/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:ee23a539b5600bba9d6a404c6d4ea02af3abee92ad572f1b003d6f5a30c6f8ab",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.26.1@sha256:2a43d2d5611ba920c49e23127cfd474fb7932fcade1671dddbef757921fcdb40", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.26.11@sha256:793943d1931f7af57b342446bf965a074ea770f6935f92b129ff0fd21ad15e2f", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.26.11@sha256:fa9d8296ceede786842670c8b027b0759651bd3f0100c31fa586321dfc98aca6", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v26.0.1@sha256:db2b15a20ad690784a6015bfad55c4dff15826be8cf9f6ac77d70abd11b1f70c", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.2", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.26.3@sha256:7572c43b32f5e6912cd7d087dc20a908b6f34186f000cacc698883f44be0db23", // renovate:container
	},
	V1_27: {
		ClusterVersion: "v1.27.3", // renovate:kubernetes-release
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
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.3/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:c0e18da6a55830cf4910ecd7261597c66ea3f8f58cf44d4adb6bdcb6e2e6f0bf",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.3/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:2cd663f25c2490bd614a6c0ad9089a47ef315caf0dbdf78efd787d5653b1c6e3",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				URL:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.3/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:fba6c062e754a120bc8105cde1344de200452fe014a8759e06e4eec7ed258a09",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.27.1@sha256:c02832d2e4bb96ac4ea14a466982d261069f9bb366f2ad68889f9a5b10b8d1b0", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.27.5@sha256:672a8b02cb75acbdbd12855395377527d16e7d3fab3a19e130d64cbd25a03692", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.27.5@sha256:30e859785838c9de1854436d9d4c1a3a54a0144fb1f232b20d17e7f1102d0928", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v26.0.1@sha256:db2b15a20ad690784a6015bfad55c4dff15826be8cf9f6ac77d70abd11b1f70c", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.2", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.27.2@sha256:dfbb63a82d437253febc68540ced60ed796494edf242d1d2cb5b9665570e99c0", // renovate:container
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
