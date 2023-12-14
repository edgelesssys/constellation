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
	"path"
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
// It accepts a full version (e.g. 1.26.7) and validates it.
// Returns an empty string if the given version is invalid.
// strict controls whether the patch version is checked or not.
// If strict is false, the patch version validation is skipped.
func NewValidK8sVersion(k8sVersion string, strict bool) (ValidK8sVersion, error) {
	prefixedVersion := compatibility.EnsurePrefixV(k8sVersion)
	var supported bool
	if strict {
		supported = isSupportedK8sVersionStrict(prefixedVersion)
	} else {
		supported = isSupportedK8sVersion(prefixedVersion)
	}
	if !supported {
		return "", fmt.Errorf("invalid Kubernetes version: %s; supported versions are %v", prefixedVersion, SupportedK8sVersions())
	}
	return ValidK8sVersion(prefixedVersion), nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (v *ValidK8sVersion) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var version string
	if err := unmarshal(&version); err != nil {
		return err
	}
	if !hasPatchVersion(version) {
		return fmt.Errorf("Kubernetes version %s does not specify a patch, supported versions are %s", version, strings.Join(SupportedK8sVersions(), ", "))
	}
	valid, err := NewValidK8sVersion(version, false) // allow any patch version to not force K8s patch upgrades
	if err != nil {
		return fmt.Errorf("unsupported Kubernetes version %s, supported versions are %s", version, strings.Join(SupportedK8sVersions(), ", "))
	}
	*v = valid
	return nil
}

// ResolveK8sPatchVersion transforms a MAJOR.MINOR definition into a supported
// MAJOR.MINOR.PATCH release.
func ResolveK8sPatchVersion(k8sVersion string) (string, error) {
	k8sVersion = compatibility.EnsurePrefixV(k8sVersion)
	if !semver.IsValid(k8sVersion) {
		return "", fmt.Errorf("Kubernetes version does not specify a valid semantic version: %s", k8sVersion)
	}
	if hasPatchVersion(k8sVersion) {
		return k8sVersion, nil
	}
	extendedVersion := k8sVersionFromMajorMinor(k8sVersion)
	if extendedVersion == "" {
		return "", fmt.Errorf("Kubernetes version %s is not valid. Supported versions: %s",
			strings.TrimPrefix(k8sVersion, "v"), supportedVersions())
	}

	return extendedVersion, nil
}

// k8sVersionFromMajorMinor takes a semver in format MAJOR.MINOR
// and returns the version in format MAJOR.MINOR.PATCH with the
// supported patch version as PATCH.
func k8sVersionFromMajorMinor(version string) string {
	switch version {
	case semver.MajorMinor(string(V1_26)):
		return string(V1_26)
	case semver.MajorMinor(string(V1_27)):
		return string(V1_27)
	case semver.MajorMinor(string(V1_28)):
		return string(V1_28)
	default:
		return ""
	}
}

// supportedVersions prints the supported version without v prefix and without patch version.
// Should only be used when accepting Kubernetes versions from --kubernetes.
func supportedVersions() string {
	builder := strings.Builder{}
	for i, version := range SupportedK8sVersions() {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(strings.TrimPrefix(semver.MajorMinor(version), "v"))
	}
	return builder.String()
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

// hasPatchVersion returns if the given version has specified a patch version.
func hasPatchVersion(version string) bool {
	return semver.MajorMinor(version) != version
}

// patchFilePath returns the canonical path for kubeadm patch files for the given component.
// See https://pkg.go.dev/k8s.io/kubernetes@v1.27.7/cmd/kubeadm/app/apis/kubeadm/v1beta3#InitConfiguration.
func patchFilePath(component string) string {
	return path.Join(constants.KubeadmPatchDir, fmt.Sprintf("%s+json.json", component))
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
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:v20231016.0.0@sha256:c51ebfc2b67f5a39daba88039e7f8f171d7084656c49c092cc53b0a2318209b2" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.15.0@sha256:8cb8dad93283268282c30e75c68f4bd76b28def4b68b563d2f9db9c74225d634" // renovate:container
	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.13.0-pre.0.20231031120927-9a282df84686@sha256:bbbf8d0507359f4adb2eaab7607edb56f03d2d16265c0277cc98c1951e699135" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.13.0-pre.0.20231031120927-9a282df84686@sha256:6416d9f0a71720de1f4503a6be4099a2759e8c33b6d69144179932c027a48124" // renovate:container
	// MetricbeatImage is the container image of filebeat, used for log collection by debugd.
	MetricbeatImage = "ghcr.io/edgelesssys/constellation/metricbeat-debugd:v2.13.0-pre.0.20231031120927-9a282df84686@sha256:7c4a36f30702c6d228813f07030f33321ba1cd31a475c5a14751445515597318" // renovate:container

	// currently supported versions.
	//nolint:revive
	V1_26 ValidK8sVersion = "v1.26.11" // renovate:kubernetes-release
	//nolint:revive
	V1_27 ValidK8sVersion = "v1.27.8" // renovate:kubernetes-release
	//nolint:revive
	V1_28 ValidK8sVersion = "v1.28.4" // renovate:kubernetes-release

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_27
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate hash-generator

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_26: {
		ClusterVersion: "v1.26.11", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.3.0/cni-plugins-linux-amd64-v1.3.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:754a71ed60a4bd08726c3af705a7d55ee3df03122b12e389fdba4bea35d7dd7e",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.28.0/crictl-v1.28.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:8dc78774f7cbeaf787994d386eec663f0a3cf24de1ea4893598096cb39ef2508",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.11/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:a62953f20fa9fedff50c6c5423e68981e3382d92cf04174d5bca5f4d084de0c5",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.11/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:58f886e39e517ba1a92493f136e80f1b6ea9362966ad9d2accdf2133004161f2",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.11/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:27c34a0870230d9dd723e1e01114634e396cd2a3d25ced263b769a4bd53e4edd",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI2LjExQHNoYTI1NjozOTUzNWQwZWZlODk1YWU5MWI1NTExZmRhZGI1MmVjOTMyOWYzODk4NzYxMTYzYThjMGRlMjAzZTIzZTMzODUzIn1d",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI2LjExQHNoYTI1NjpjZGJlZmZmMTU0ZDRjY2I1ZDhlOGIxNmI4MDRjYmM2Y2M5MzI2YTc2MGI5ZjkxNDIyMjcwOGY5OTExOThkNTdjIn1d",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI2LjExQHNoYTI1NjowNjg0ZTIzMTcyZDkyMDMxNDk3MTU4MGFiMTE1YTViNjc5YWMxZmFlMmNiOTRkODNlOTEwNWMwYjFlOTNhMWJjIn1d",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjctMEBzaGEyNTY6NTFlYWU4MzgxZGNiMTA3ODI4OWZhN2I0ZjNkZjI2MzBjZGMxOGQwOWZiNTZmOGU1NmI0MWM0MGUxOTFkNmM4MyJ9XQ==",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.26.6@sha256:33445ab57f48938fe989ffe311dacee0044b82f2bd23cb7f7b563275926f0ce9", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.26.16@sha256:92abc79a8a339cc7ab47abae35075b4f9771e5a25a9ada7c5040b1b3c7c7046e", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.26.16@sha256:82ae9ba5483c4dd900f65c008cbeb390f62d93983374ec601f269d3597d4da8b", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v26.4.0@sha256:dbe983cceabb3df98112b083d844229c85a1bbdfef2060c79f4cd49afe2a07f3", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.26.4@sha256:f771284ff54ecfedf40c7af70c5450600786c98989aeb69cdcf7e7bb7ac5a20d", // renovate:container
	},
	V1_27: {
		ClusterVersion: "v1.27.8", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.3.0/cni-plugins-linux-amd64-v1.3.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:754a71ed60a4bd08726c3af705a7d55ee3df03122b12e389fdba4bea35d7dd7e",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.28.0/crictl-v1.28.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:8dc78774f7cbeaf787994d386eec663f0a3cf24de1ea4893598096cb39ef2508",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.8/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:2e0557b38c5b9a1263eed25a0b84d741453ed9c0c7bd916f80eadaf7edfb7784",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.8/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:f8864769b8b2d7a14f53eb983f23317ff14d68ab76aba71e9de17ce84c38d4eb",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.8/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:027b3161e99fa0a7fa529e8f17f73ee2c0807c81c721ca7cf307f6b41c17bc57",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI3LjhAc2hhMjU2OjcwYjA1YjYxZDg0NmViYjY5YTkwN2ZlMjU1ZDM5YTZmNmMxMGQ1Y2E5NTA0ZjNkMmMwZGZmM2Y4NjQ2OTBkMzMifV0=",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI3LjhAc2hhMjU2OmU1OWM2MzczZDI2YjY4NGE5MWNmYTU5NDJjMGY3MzcxYmRhOWI0YmI3Njg5ZTNmOTBmN2VlNGY5NjUxZWUyMmIifV0=",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI3LjhAc2hhMjU2OjYyMzdlNzEwMGNjZGJiZDVlMGU3Y2ZmNzc5NjgzMWMxODVhMzk0NzE5OTgyM2YzOTEyODNjNzlkMDBhZmYwNzAifV0=",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjktMEBzaGEyNTY6ZTAxM2QwZDVlNGUyNWQwMGM2MWE3ZmY4Mzk5MjdhMWYzNjQ3OTY3OGYxMWU0OTUwMmI1M2E1ZTBiMTRmMTBjMyJ9XQ==",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.27.2@sha256:42be09a2b13b4e69b42905639d6b005ebe1ca490aabefad427256abf2cc892c7", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.27.10@sha256:3366e0e51c56643968c7e607cb27c2545948cfab5bff3bed85e314d93a689d8e", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.27.10@sha256:754d4eb709d0c5955af8bc46f5beccf0fa8c09551855a3810145b09af27d6656", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v27.1.6@sha256:b097b4e5382ea1987db5996a9eaffb94fa224639b3464876f0b1b17f64509ac4", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.27.3@sha256:0e1ab1bfeb1beaa82f59356ef36364503df22aeb8f8d0d7383bac449b4e808fb", // renovate:container
	},
	V1_28: {
		ClusterVersion: "v1.28.4", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.3.0/cni-plugins-linux-amd64-v1.3.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:754a71ed60a4bd08726c3af705a7d55ee3df03122b12e389fdba4bea35d7dd7e",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.28.0/crictl-v1.28.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:8dc78774f7cbeaf787994d386eec663f0a3cf24de1ea4893598096cb39ef2508",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.28.4/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:db2a473b73c3754d4011590f2f0aa877657608499590c6b0f8b40bec96a3e9ba",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.28.4/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:b4d2531b7cddf782f59555436bc098485b5fa6c05afccdeecf0d62d21d84f5bd",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.28.4/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:893c92053adea6edbbd4e959c871f5c21edce416988f968bec565d115383f7b8",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI4LjRAc2hhMjU2OjViMjhhMzY0NDY3Y2Y3ZTEzNDM0M2JiM2VlMmM2ZDQwNjgyYjQ3M2E3NDNhNzIxNDJjN2JiZTI1NzY3ZDM2ZWIifV0=",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI4LjRAc2hhMjU2OjY1NDg2YzhjMzM4Zjk2ZGMwMjJkZDFhMGFiZTg3NjNlMzhmMzUwOTViODRiMjA4Yzc4ZjQ0ZDllOTk0NDdkMWMifV0=",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI4LjRAc2hhMjU2OjMzNWJiYTllODYxYjg4ZmE4YjdiYjkyNTBiY2Q2OWI3YTMzZjgzZGE0ZmVlOTNmOWZjMGVlZGM2ZjM0ZTI4YmEifV0=",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjktMEBzaGEyNTY6ZTAxM2QwZDVlNGUyNWQwMGM2MWE3ZmY4Mzk5MjdhMWYzNjQ3OTY3OGYxMWU0OTUwMmI1M2E1ZTBiMTRmMTBjMyJ9XQ==",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.28.1@sha256:79b423ac8bc52d00f932b40de11fc3047a5ed1cbec47cda23bcf8f45ef583ed1", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.27.10@sha256:3366e0e51c56643968c7e607cb27c2545948cfab5bff3bed85e314d93a689d8e", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.27.10@sha256:754d4eb709d0c5955af8bc46f5beccf0fa8c09551855a3810145b09af27d6656", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v27.1.6@sha256:b097b4e5382ea1987db5996a9eaffb94fa224639b3464876f0b1b17f64509ac4", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.27.3@sha256:0e1ab1bfeb1beaa82f59356ef36364503df22aeb8f8d0d7383bac449b4e808fb", // renovate:container
	},
}

// KubernetesVersion bundles download Urls to all version-releated binaries necessary for installing/deploying a particular Kubernetes version.
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
