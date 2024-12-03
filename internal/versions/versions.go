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
	case semver.MajorMinor(string(V1_28)):
		return string(V1_28)
	case semver.MajorMinor(string(V1_29)):
		return string(V1_29)
	case semver.MajorMinor(string(V1_30)):
		return string(V1_30)
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

	// GcpGuestImage image for GCP guest agent.
	// Check for new versions at https://github.com/GoogleCloudPlatform/guest-agent/releases and update in /.github/workflows/build-gcp-guest-agent.yml.
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:v20240816.0.0@sha256:a6f871346da12d95a1961cb247343ccaa708039f49999ce56d00e35f3f701b97" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.17.0@sha256:bf1c5758b3d266dd6234422d156c67ffdd47f50f70ce17d5cef1de6065030337" // renovate:container
	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.17.0-pre.0.20240627193502-8aed4bb0fe45@sha256:d6c5a06049e5c1b9d7ba4b83367fa0c06ba2d1b65e1d299f3e00f465f310642b" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.17.0-pre.0.20240627193502-8aed4bb0fe45@sha256:606adccf544a15e6b9ae9e11eec707668660bc1af346ff72559404e36da5baa2" // renovate:container
	// MetricbeatImage is the container image of filebeat, used for log collection by debugd.
	MetricbeatImage = "ghcr.io/edgelesssys/constellation/metricbeat-debugd:v2.17.0-pre.0.20240627193502-8aed4bb0fe45@sha256:690b9d36cc334a7f83b58ca905169bb9f1c955b7a436c0044a07f4ce15a90594" // renovate:container

	// currently supported versions.
	//nolint:revive
	V1_28 ValidK8sVersion = "v1.28.15" // renovate:kubernetes-release
	//nolint:revive
	V1_29 ValidK8sVersion = "v1.29.11" // renovate:kubernetes-release
	//nolint:revive
	V1_30 ValidK8sVersion = "v1.30.7" // renovate:kubernetes-release

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_29
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate hash-generator

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_28: {
		ClusterVersion: "v1.28.15", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.6.0/cni-plugins-linux-amd64-v1.6.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:682b49ff8933a997a52107161f1745f8312364b4c7f605ccdf7a77499130d89d",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.31.1/crictl-v1.31.1-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:0a03ba6b1e4c253d63627f8d210b2ea07675a8712587e697657b236d06d7d231",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://dl.k8s.io/v1.28.15/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:b07a27fd5bd2419c9c623de15c1dd339af84eb27e9276c81070071065db00036",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.28.15/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:0555b2c2fd30efcdb44b7fba5460c3dc3d3e39f2301e1eef7894a9f8976e1b4c",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.28.15/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:1f7651ad0b50ef4561aa82e77f3ad06599b5e6b0b2a5fb6c4f474d95a77e41c5",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI4LjE1QHNoYTI1Njo2ZGZhODRmNWQ2YmU3MTFhZTBkMTk3NTgyMDFkMzM3ZTgzNmFiN2RlNzMzMDZmZjE0NzI1Y2VhYTk3OGZlYThmIn1d",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI4LjE1QHNoYTI1NjpkYWRkMmEzNzg0NzgzMDE4YTdlZTg1ODhkMTFmNzg3ZmVlNGQ1NDI0ZjJjZGQ2Y2U4OWQzYmExODQ0YTZjMTc1In1d",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI4LjE1QHNoYTI1Njo4MmY0YTQyMzE3NDUwODU4ZDNkNzBmZGU1YjNjMGYyMjE1M2VhMTU1ZmQwNTNmMDk4NjU5OTlhNDY2MWYyZGNhIn1d",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjE1LTBAc2hhMjU2OmE2ZGM2M2U2ZThjZmEwMzA3ZDc4NTE3NjJmYTZiNjI5YWZiMThmMjhkOGFhM2ZhYjVhNmU5MWI0YWY2MDAyNmEifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		// Check for newer versions at https://github.com/kubernetes/cloud-provider-aws/releases.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.28.9@sha256:168905b591796fbd07cb35cd0e3f206fdb7efb30e325c9bf7fa70d1b48989f73", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.28.13@sha256:8b853f4f54a09c363806714189435933a8575ac6dca27e991976bd685603113e", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.28.13@sha256:525ca9c8a44bbdfa9acc0a417776bb822a1bbdaaf27d9776b8dcf5b3519c346a", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v28.10.0@sha256:f3b6fa7faea27b4a303c91b3bc7ee192b050e21e27579e9f3da90ae4ba38e626", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.28.7@sha256:77906954da9171425c8c8d3286091818143b6dcf9039abd49b8f33f1502978a1", // renovate:container
	},
	V1_29: {
		ClusterVersion: "v1.29.11", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.6.0/cni-plugins-linux-amd64-v1.6.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:682b49ff8933a997a52107161f1745f8312364b4c7f605ccdf7a77499130d89d",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.31.1/crictl-v1.31.1-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:0a03ba6b1e4c253d63627f8d210b2ea07675a8712587e697657b236d06d7d231",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://dl.k8s.io/v1.29.11/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:1aaa9025cceac0c9a4df295a58aa79d8932a5b13a43c8910412c9ef970c42d21",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.29.11/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:6cf3567bd69a14859fb80fb39a09196dc2de1729ae72566e7e4819c5600e49c6",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.29.11/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:14d7ea4ada60ff15ef3b7734a83c4d05cff164d4843b6f4c081a50b86547c17d",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI5LjExQHNoYTI1Njo2NTQ1ZjZmNjJjMWYwZjBhZmZmNTM0YzIyNzZiZDI0MmZmZGFjNTRlNzNiMjJjYzEwMzQzNTIyYzcyYjlmNzcxIn1d",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI5LjExQHNoYTI1Njo1OTU1NDgyZDg5NWFhNGE0MTNhNzBhZjUwNjcwNjA1OTIzOGIzNjM3ZmQ3MzA1OTRlNTljOTkwOTU0MjIxZjEzIn1d",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI5LjExQHNoYTI1NjpmZTA3OGYyNmJmMTQzMDFmNzk0NjhkNzlhNWQyZDQzMWIzNDZhNGFiYjFjYWI0ZTliMWY5YmE4OTE0OTY0MTYxIn1d",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjE1LTBAc2hhMjU2OmE2ZGM2M2U2ZThjZmEwMzA3ZDc4NTE3NjJmYTZiNjI5YWZiMThmMjhkOGFhM2ZhYjVhNmU5MWI0YWY2MDAyNmEifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		// Check for newer versions at https://github.com/kubernetes/cloud-provider-aws/releases.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.29.6@sha256:8074b8828a33fb273833e8fd374dda6a0ab10335ae8e19684fbd61eeff7d3594", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.29.11@sha256:2ecdca660c03b17110a4ee732230424ce0377c5b1756a4408666e40938ee976a", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.29.11@sha256:17888b0ebaec6735214b85d20bdcc8062f051bc27e835454e9ef89734d34aa4b", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v29.5.1@sha256:ebbc6f5755725b6c2c81ca1d1580e2feba83572c41608b739c50f85b2e5de936", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.29.5@sha256:76865740be7c965b35ee1524931bb4abfe4c27b5bfad280e84068cd6653ee7bb", // renovate:container
	},
	V1_30: {
		ClusterVersion: "v1.30.7", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.6.0/cni-plugins-linux-amd64-v1.6.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:682b49ff8933a997a52107161f1745f8312364b4c7f605ccdf7a77499130d89d",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.31.1/crictl-v1.31.1-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:0a03ba6b1e4c253d63627f8d210b2ea07675a8712587e697657b236d06d7d231",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://dl.k8s.io/v1.30.7/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:4b2fb90661e2b7be8b34b5bd405652590f6351a9f55194430fa16d547a7c9d16",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.30.7/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:3635901a3a6e5b0124547151de700a1dbd4571531327a32edbec43bb155b877b",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.30.7/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:3950ec7b81b9aa5a3856dd6155e42e1ad12a2fa8b050502f5a9652050f94a8ee",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjMwLjdAc2hhMjU2OjEzZjRmMGY1ODUwYjM5NzQyMTAxYzY1NmIxYmJkNTA5MGVhY2YyNzA4NGFkODliNDg5ZmM4MjRlZjQ4MmVkOWMifV0=",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjMwLjdAc2hhMjU2OjViOGRjMjZjMDViMjczY2UxOThmYmFmNGViMTc5ZjNiNmQ2OTE5YzZlMjExNmQzNmRjN2Y4MjU1NTM3NGM2ODcifV0=",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjMwLjdAc2hhMjU2OjAwYjE2ZGI5OTExMDFiNzM2MWYyZjE4MDM1ZTFjNjUyNmMwY2U2YzlhOTU2ODUyNDgyNGI0YmNjZGYxYWZiZDYifV0=",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjE1LTBAc2hhMjU2OmE2ZGM2M2U2ZThjZmEwMzA3ZDc4NTE3NjJmYTZiNjI5YWZiMThmMjhkOGFhM2ZhYjVhNmU5MWI0YWY2MDAyNmEifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		// Check for newer versions at https://github.com/kubernetes/cloud-provider-aws/releases.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.30.3@sha256:30a1758dec30814178c787e2d50f46bb141e9f0bb2e16190ddd19df15f957874", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.30.7@sha256:03b2876f481507781a27b56a6e66c1928b7b93774e787e52a5239aefa41191e4", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.30.7@sha256:f18feb78e36eef88f0e23d98d798476d2bf6837de11892fe118ab043afdcd497", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v30.1.0@sha256:64d2d5d4d2b5fb426c307c64ada9a61b64e797b56d9768363f145f2bd957998f", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.30.3@sha256:08fd86ee093760849ac4fd579eb90185b669fc20aa56c156aa34ea7b73dd5e34", // renovate:container
	},
}

// KubernetesVersion bundles download Urls to all version-releated binaries necessary for installing/deploying a particular Kubernetes version.
type KubernetesVersion struct {
	ClusterVersion                       string
	KubernetesComponents                 components.Components
	CloudControllerManagerImageAWS       string // k8s version dependency.
	CloudControllerManagerImageAzure     string // k8s version dependency.
	CloudControllerManagerImageGCP       string // Published by .github/workflows/build-ccm-gcp.yml because of https://github.com/kubernetes/cloud-provider-gcp/issues/289.
	CloudControllerManagerImageOpenStack string // k8s version dependency.
	CloudNodeManagerImageAzure           string // k8s version dependency. Same version as Azure's CCM image above.
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
