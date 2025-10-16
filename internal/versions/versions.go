/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
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
	case semver.MajorMinor(string(V1_30)):
		return string(V1_30)
	case semver.MajorMinor(string(V1_31)):
		return string(V1_31)
	case semver.MajorMinor(string(V1_32)):
		return string(V1_32)
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
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:v20250818.0.0@sha256:80237f67052301ab842bee20820dd25b233cc1c2626e5a2fbbb9243343922de9" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.18.0@sha256:79f2f7e3b116618fc68a16d3e4524ba6e8c8dea47d7db445a4b4f9c1d4f0e0e8" // renovate:container
	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.24.0-pre.0.20250715233448-3b9f7530fbba@sha256:207e24406a5fddd62189b8cd89ea3751b78113109a2da828d9b5e9fbb1b44cc7" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.24.0-pre.0.20250715233448-3b9f7530fbba@sha256:05bc4fb1679430de5a664f0c8cb43407a00f0217eeb3cbb54f7c1088ff968da6" // renovate:container
	// MetricbeatImage is the container image of filebeat, used for log collection by debugd.
	MetricbeatImage = "ghcr.io/edgelesssys/constellation/metricbeat-debugd:v2.24.0-pre.0.20250715233448-3b9f7530fbba@sha256:86532d8a3f37236dd8d4a30e0446c18a296643046d4127f66a643c263d556957" // renovate:container

	// currently supported versions.
	//nolint:revive
	V1_30 ValidK8sVersion = "v1.30.14" // renovate:kubernetes-release
	//nolint:revive
	V1_31 ValidK8sVersion = "v1.31.13" // renovate:kubernetes-release
	//nolint:revive
	V1_32 ValidK8sVersion = "v1.32.9" // renovate:kubernetes-release

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_31
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate hash-generator

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_30: {
		ClusterVersion: "v1.30.14", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.8.0/cni-plugins-linux-amd64-v1.8.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:ab3bda535f9d90766cccc90d3dddb5482003dd744d7f22bcf98186bf8eea8be6",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.33.0/crictl-v1.33.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:8307399e714626e69d1213a4cd18c8dec3d0201ecdac009b1802115df8973f0f",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://dl.k8s.io/v1.30.14/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:46baa60748b179164e80f5565d99dad642d554fb431925d211ffa921b917d5c7",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.30.14/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:bf1f8af81af8ecf003cbc03a8700c6e94a74c183ee092bbc77b92270ada2be70",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.30.14/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:7ccac981ece0098284d8961973295f5124d78eab7b89ba5023f35591baa16271",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjMwLjE0QHNoYTI1NjpiZTA3OWZlODVkNmI2ODA0Yjg5YWI0ZmRkNmEzNWNkNTYzNDFlOTllYTgwOTg4MWNmZTM3OTYyZjQ0MGRjMWJlIn1d",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjMwLjE0QHNoYTI1NjplYmE0MWQ3NmI2YWYxMGFmOTQxMTQ3ZTJiNDQ5YzZkNjhlYWE0MmMxZmQwZmM4ZjBhYjhlYzJmZjBhYjg0OTY0In1d",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjMwLjE0QHNoYTI1Njo3NGE1Y2Y5Y2ZhOWZjYzIyNDZmNjhjNjUwZjFmNWM3YWRkMjBkYTIxNDVmMTM4MDBmZDk3YmExZDY5ZmMwNmM4In1d",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjIxLTBAc2hhMjU2OmQ1OGMwMzVkZjU1NzA4MGEyNzM4N2Q2ODcwOTJlM2ZjMmI2NGM2ZDBlMzE2MmRjNTE0NTNhMTE1Zjg0N2QxMjEifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		// Check for newer versions at https://github.com/kubernetes/cloud-provider-aws/releases.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.30.9@sha256:066bacbf4485dc90f9003680b5500bc67305a9eff81a49a4373dd5d6d4d376e6", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.30.15@sha256:eebb5232e88e73b806329d3b3a73375b1b491e66d88d9c4854979072ef159e12", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.30.15@sha256:174749b03a966056c3368d81fb568b99cbfcf45da248eb8b046bd876ffff2587", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v30.1.4@sha256:0c3695a18d3825492196facb092e5fe56e466fa8517cde5a206fe21630c1da13", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.30.6@sha256:bd403d2cce02b8b5c42acc35d21843467d30c864d3986f9cd51c34b28e68e587", // renovate:container
	},
	V1_31: {
		ClusterVersion: "v1.31.13", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.8.0/cni-plugins-linux-amd64-v1.8.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:ab3bda535f9d90766cccc90d3dddb5482003dd744d7f22bcf98186bf8eea8be6",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.33.0/crictl-v1.33.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:8307399e714626e69d1213a4cd18c8dec3d0201ecdac009b1802115df8973f0f",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://dl.k8s.io/v1.31.13/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:b7733cb0e36031ab8f69cf3f73cc364730141b107d24ecb75bee5b3122dc6b6e",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.31.13/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:e0c91c606e1402e943666f6fda8452140851a1b2bb4f69918dd7176e8d86845c",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.31.13/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:ce427a4df97b044bf6beaf38be6b3aaafb1b5c782c406958b3c71ebb25a48fe3",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjMxLjEzQHNoYTI1Njo5YWJlYjhhMmQzZTUzZTM1NmQxZjJlNWQ1ZGMyMDgxY2YyOGYyMzI0MjY1MWIwNTUyYzllMzhmNGE3YWU5NjBlIn1d",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjMxLjEzQHNoYTI1NjpmYWNjOTEyODg2OTdhMjg4YTY5MTUyMDk0OWZlNGVlYzQwMDU5ZWYwNjVjODlkYThlMTA0ODFkMTRlMTMxYjA5In1d",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjMxLjEzQHNoYTI1NjpjNWNlMTUwZGNjZTI0MTlmZGVmOWY5ODc1ZmVmNDMwMTQzNTVjY2ViZjkzNzg0NmVkM2EyOTcxOTUzZjliMjQxIn1d",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjIxLTBAc2hhMjU2OmQ1OGMwMzVkZjU1NzA4MGEyNzM4N2Q2ODcwOTJlM2ZjMmI2NGM2ZDBlMzE2MmRjNTE0NTNhMTE1Zjg0N2QxMjEifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		// Check for newer versions at https://github.com/kubernetes/cloud-provider-aws/releases.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.31.8@sha256:f846203ea9e0f510176b08c1c0406ffec03498c610d66442953b7823ca99622f", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.31.9@sha256:5842bc4bb651e5931717953ee3e4851083c4db1185d1cbf61eb538da21bfbf15", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.31.9@sha256:77641062e704a709f73d484ad50dde2cbf408491e1d3ec0f5b750eb7dd1e5df7", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v30.1.4@sha256:0c3695a18d3825492196facb092e5fe56e466fa8517cde5a206fe21630c1da13", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.33.1@sha256:de8a6da8c31c7b967625451a7169309d6f77aee1ff64b3f8e6ba8d8810ce2a22", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.31.5@sha256:b90105c3470ad5059e8a7c11e2c63117dd8d499ae893d4cc76968e748a886b74", // renovate:container
	},
	V1_32: {
		ClusterVersion: "v1.32.9", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.8.0/cni-plugins-linux-amd64-v1.8.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:ab3bda535f9d90766cccc90d3dddb5482003dd744d7f22bcf98186bf8eea8be6",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.33.0/crictl-v1.33.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:8307399e714626e69d1213a4cd18c8dec3d0201ecdac009b1802115df8973f0f",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://dl.k8s.io/v1.32.9/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:fd7711d1f0c1e263e9332004858fc4a6c39462e3e2ee485706eea5297966ed9c",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.32.9/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:183b3b12e39b3ed2dc2db25cbc17769610cdd5f02e9d1325ba747d54978d8f5f",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.32.9/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:509ae171bac7ad3b98cc49f5594d6bc84900cf6860f155968d1059fde3be5286",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjMyLjlAc2hhMjU2OjZkZjExY2MyYWQ5Njc5YjExMTdiZTM0ZDNhMDIzMGFkZDg4YmMwYTA4ZmQ3YTNlYmMyNmI2ODA1NzVlOGRlOTcifV0=",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjMyLjlAc2hhMjU2OjI0M2M0YjhlM2JjZTI3MWZjYjFiNzgwMDhhYjk5NmFiNjk3NmIxYTIwMDk2ZGVhYzA4MzM4ZmNkMTc5Nzk5MjIifV0=",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjMyLjlAc2hhMjU2OjUwYzQ5NTIwZGJkMGU4YjQwNzZiNmE1Yzc3ZDgwMTRkZjA5ZWEzZDU5YTczZThiYWZkMjY3OGQ1MWViYjkyZDUifV0=",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjIxLTBAc2hhMjU2OmQ1OGMwMzVkZjU1NzA4MGEyNzM4N2Q2ODcwOTJlM2ZjMmI2NGM2ZDBlMzE2MmRjNTE0NTNhMTE1Zjg0N2QxMjEifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		// Check for newer versions at https://github.com/kubernetes/cloud-provider-aws/releases.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.32.5@sha256:542aa2b272f931fc5b0bf260866eea1135aee33ba6b84bfa4564adda35bf1732", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.32.8@sha256:d44a8e0e7e347343d2bd8d04acbf874f22f47483ac988646129d70bc1e82a107", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.32.8@sha256:87cbaa3239ce627b3e3adea1d4faaf41eadaef770831eff1ea49fc8efdf3e5c8", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v30.1.4@sha256:0c3695a18d3825492196facb092e5fe56e466fa8517cde5a206fe21630c1da13", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.33.1@sha256:de8a6da8c31c7b967625451a7169309d6f77aee1ff64b3f8e6ba8d8810ce2a22", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.32.4@sha256:77cfe4c344b8650a9335d5d42df4397838764977d6f112be73e29da74dab8124", // renovate:container
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
