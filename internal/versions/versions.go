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
	KonnectivityAgentImage = "registry.k8s.io/kas-network-proxy/proxy-agent:v0.29.0@sha256:0349e45ae85f9aec3fafc289498b0058b2c0c4725fe4e7f69fef361182976833" // renovate:container
	// KonnectivityServerImage server image for konnectivity service.
	KonnectivityServerImage = "registry.k8s.io/kas-network-proxy/proxy-server:v0.29.0@sha256:f0c74d7420d1320bc8ece0bb57125f5a1a021d3310bc990ac7749ec3dab4eab7" // renovate:container
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
	V1_26 ValidK8sVersion = "v1.26.12" // renovate:kubernetes-release
	//nolint:revive
	V1_27 ValidK8sVersion = "v1.27.9" // renovate:kubernetes-release
	//nolint:revive
	V1_28 ValidK8sVersion = "v1.28.5" // renovate:kubernetes-release

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_27
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate hash-generator

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_26: {
		ClusterVersion: "v1.26.12", // renovate:kubernetes-release
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
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.12/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:aed0a351b01f1e6a84a0992ef1265bb0c9994b900162c075df58d0d02517d3df",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.12/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:5a5d65acefb50010859be8ffba8e6e059d552ae357e3101c12c62e747a9416a2",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.26.12/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:8e6af8d68e7b9d2a1eb43255c0da793276e549a34a2b9c3c87a9c26438e7fd71",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI2LjEyQHNoYTI1NjozYzMxMWRjMjY1NzY1YzBiOTNlMzJmMjVkMzI4Y2NkODk5NGI1ZTAwOWQ5YzAyODcyM2I5OTYyNTRlNTIwYjdlIn1d",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI2LjEyQHNoYTI1NjoyNGEzNDNiMDViYmRlY2VkMzRkOWNmNzMwNDMzYzNhZTgxMDBiMTRmNjAyYTAyOGJhNWM0M2JjNDc5NzUzY2M1In1d",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI2LjEyQHNoYTI1NjoxZDMzMjU1YTE2MjE5YzJhMzFiNWZmZDk1MjczMDA2MjUzY2MzNDlmZDIxMzIzOTQ0ZDMxZjk2MDBmODdjNDA2In1d",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjEwLTBAc2hhMjU2OjIyZjg5MmQ3NjcyYWRjMGI5Yzg2ZGY2Nzc5MmFmZGI4YjJkYzA4ODgwZjQ5ZjY2OWVhYWE1OWM0N2Q3OTA4YzIifV0=",
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
		ClusterVersion: "v1.27.9", // renovate:kubernetes-release
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
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.9/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:ede60eea3acbac3f35dbb23d7b148f45cf169ebbb20af102d3ce141fc0bac60c",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.9/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:78dddac376fa2f04116022cb44ed39ccb9cb0104e05c5b21b220d5151e5c0f86",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.9/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:d0caae91072297b2915dd65f6ef3055d27646dce821ec67d18da35ba9a8dc85b",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI3LjlAc2hhMjU2OjkyMjc0ZTgyZTI0NTJkYzA0ZmVkOTMzY2U2ZTQyMWM2NGUzMGQxNGQ4NjhkMDdiZmIwNzY5N2E5NjE0YTFkYjgifV0=",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI3LjlAc2hhMjU2OjczNWFlZmY1YTJlNjI4MmUwZWI3MjQ1YmUyNGIzZGEwNzYyOTdmOWU0ZmJhMmIzMjA5NGNjZjYxYTA4Y2NjYzIifV0=",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI3LjlAc2hhMjU2OmYzOTc1OTU2YWQyMzY2N2NhOGY1NTdkNzY0MDQyNTNjYjdlODE1Y2E3Zjc3YWVkOTBlMWFlN2Q2NWU4OGYyYjEifV0=",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjEwLTBAc2hhMjU2OjIyZjg5MmQ3NjcyYWRjMGI5Yzg2ZGY2Nzc5MmFmZGI4YjJkYzA4ODgwZjQ5ZjY2OWVhYWE1OWM0N2Q3OTA4YzIifV0=",
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
		ClusterVersion: "v1.28.5", // renovate:kubernetes-release
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
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.28.5/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:bf37335da58182783a8c63866ec1f895b4c436e3ed96bdd87fe3f8ae8004ba1d",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.28.5/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:2b54078c5ea9e85b27f162f508e0bf834a2753e52a57e896812ec3dca92fe9cd",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.28.5/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:2a44c0841b794d85b7819b505da2ff3acd5950bd1bcd956863714acc80653574",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI4LjVAc2hhMjU2OjRiYjZmNDZiYWE5ODA1MjM5OWVlMjI3MGQ1OTEyZWRiOTdkNGY4NjAyZWEyZTI3MDBmMDUyN2E4ODcyMjgxMTIifV0=",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI4LjVAc2hhMjU2OjZlOGM5MTcxZjc0YTRlM2ZhZGVkY2U4Zjg2NWY1ODA5MmE1OTc2NTBhNzA5NzAyZjJiMTIyZjZjYTNiNmNkMzIifV0=",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI4LjVAc2hhMjU2OjlhNDhlMzNlNDU0YzkwNGNmNTNjMTQ0YThkMGFlM2Y1YTRjMmM1YmQwODZiODk1M2FkN2Q1YTYzN2I5YWEwMDcifV0=",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjEwLTBAc2hhMjU2OjIyZjg5MmQ3NjcyYWRjMGI5Yzg2ZGY2Nzc5MmFmZGI4YjJkYzA4ODgwZjQ5ZjY2OWVhYWE1OWM0N2Q3OTA4YzIifV0=",
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
