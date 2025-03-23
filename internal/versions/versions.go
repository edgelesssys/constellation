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
	case semver.MajorMinor(string(V1_29)):
		return string(V1_29)
	case semver.MajorMinor(string(V1_30)):
		return string(V1_30)
	case semver.MajorMinor(string(V1_31)):
		return string(V1_31)
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
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:v20250225.0.0@sha256:2b947decf2cbd5c0fdd0815010812231098936ccf5768733ff972b3056a15623" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.17.0@sha256:bf1c5758b3d266dd6234422d156c67ffdd47f50f70ce17d5cef1de6065030337" // renovate:container
	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.20.0-pre.0.20241128161539-b99bd53066cf@sha256:726d2e3a15d6fa83d2d64e6f0ecc9b725c5f135ff2be51d1e0e4052d3f5e784e" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.20.0-pre.0.20241128161539-b99bd53066cf@sha256:ab030e79c7bb93aea93f5a787c95c566cab41f4a45e2012577c1f31615f8e687" // renovate:container
	// MetricbeatImage is the container image of filebeat, used for log collection by debugd.
	MetricbeatImage = "ghcr.io/edgelesssys/constellation/metricbeat-debugd:v2.20.0-pre.0.20241128161539-b99bd53066cf@sha256:f406201b3897acbaae41d75d42ebeda02ad08ddad1a3a4527a1b1d86a27ca4cb" // renovate:container

	// currently supported versions.
	//nolint:revive
	V1_29 ValidK8sVersion = "v1.29.15" // renovate:kubernetes-release
	//nolint:revive
	V1_30 ValidK8sVersion = "v1.30.11" // renovate:kubernetes-release
	//nolint:revive
	V1_31 ValidK8sVersion = "v1.31.7" // renovate:kubernetes-release

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_30
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate hash-generator

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_29: {
		ClusterVersion: "v1.29.15", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.6.2/cni-plugins-linux-amd64-v1.6.2.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:b8e811578fb66023f90d2e238d80cec3bdfca4b44049af74c374d4fae0f9c090",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.32.0/crictl-v1.32.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:f050b71d3a73a91a4e0990b90143ed04dcd100cc66f953736fcb6a2730e283c4",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://dl.k8s.io/v1.29.15/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:cd0228a5820f98bbb6371344e5d6645f6486d767c30f927a1d0ec8d17eca4da5",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.29.15/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:d0744fbaa1e67fc65c4a3409f083e01a4ede58181c759b2feeb08b1ef10d6201",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.29.15/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:3473e14c7b024a6e5403c6401b273b3faff8e5b1fed022d633815eb3168e4516",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI5LjE1QHNoYTI1NjpmZDgyYzc0ZTA3NzNhMTAzOTYwNTU5MDQ3NTMxMjY0MTFiM2E5NTg0Y2M0NTNlMWM3MTUyYzgxMDE4YTkzM2I2In1d",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI5LjE1QHNoYTI1Njo0ZjA1YmUyYzA2NjdkOWY0OTc1YmNjNDNkNWUxMzZiMjQzNjk0NmY4NGM4ZjdkYzJkMmRhMTQzOTJlNzYxYTcxIn1d",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI5LjE1QHNoYTI1NjoyNzYxMDhhNDU0MWE1MTg5NGEwMTA2MzMyMzBmN2I2ZDEwZTkyZTczMDI3NGYyNGJkMjFlODI3ZTY0MjQzZDY2In1d",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjE2LTBAc2hhMjU2OmM2YTlkMTFjYzVjMDRiMTE0Y2NkZWYzOWE5MjY1ZWVlZjgxOGUzZDAyZjUzNTliZTAzNWFlNzg0MDk3ZmRlYzUifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		// Check for newer versions at https://github.com/kubernetes/cloud-provider-aws/releases.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.29.8@sha256:3f8e7be967f76b328123d53846c21dcd930b60094f9f4abd8bf5ab0fe108e6e4", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.29.13@sha256:aa3d83f3f1b8dc22a9b2d4e35d417292bffc6eac2c6a2e22a04ebbd6561148b2", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.29.13@sha256:fd8aa37ecaf8e8075555be84a98aff28527ffddbcd084f4e3b3cd31c80748253", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v29.5.1@sha256:ebbc6f5755725b6c2c81ca1d1580e2feba83572c41608b739c50f85b2e5de936", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.29.5@sha256:76865740be7c965b35ee1524931bb4abfe4c27b5bfad280e84068cd6653ee7bb", // renovate:container
	},
	V1_30: {
		ClusterVersion: "v1.30.11", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.6.2/cni-plugins-linux-amd64-v1.6.2.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:b8e811578fb66023f90d2e238d80cec3bdfca4b44049af74c374d4fae0f9c090",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.32.0/crictl-v1.32.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:f050b71d3a73a91a4e0990b90143ed04dcd100cc66f953736fcb6a2730e283c4",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://dl.k8s.io/v1.30.11/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:59177fc92e2b2bb988f7d8d39682ea9e3d9d883273c9c8b51b39502d9b965431",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.30.11/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:06ff7ff15b7fa9af60189fdece5f7c56efa8b637c38b4a498715ca2f04ccfcb2",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.30.11/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:228a8b2679f84de9192a1ac5ad527c9ab73b0f76c452ed74f11da812bbcfaa42",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjMwLjExQHNoYTI1Njo3N2M1NDM0Njk2NTAzNmFjYzdhYzk1YzMyMDA1OTdlZGUzNmRiOTI0NjE3OTI0OGRkZTIxYzFhM2VjYzFjYWYwIn1d",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjMwLjExQHNoYTI1NjpkODg3NGYzZmI0NTU5MWVjZGFjNjdhMzAzNWM3MzA4MDhmMThiM2FiMTMxNDc0OTVjN2Q3N2ViMTk2MGQ0ZjZmIn1d",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjMwLjExQHNoYTI1NjpjNjk5ZjhjOTdhZTdlYzgxOWM4YmQ4NzhkM2RiMTA0YmE3MmZjNDQwZDgxMGQ5MDMwZTA5Mjg2YjY5NjAxN2I1In1d",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjE2LTBAc2hhMjU2OmM2YTlkMTFjYzVjMDRiMTE0Y2NkZWYzOWE5MjY1ZWVlZjgxOGUzZDAyZjUzNTliZTAzNWFlNzg0MDk3ZmRlYzUifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		// Check for newer versions at https://github.com/kubernetes/cloud-provider-aws/releases.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.30.8@sha256:f4e82b924e967656d8df9f467c9c1915509e94e228b53840a7f9f2367e1ba9f5", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.30.10@sha256:98912ffdf1264738fd07a19e41e88c74b77c81ab40e1383a588e7f2a6ec3469d", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.30.10@sha256:4cb8c3e4e79e166f0dbbb76bdd37c6dce931ac4cb63f32aea1327b22fb5317f1", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v30.1.4@sha256:0c3695a18d3825492196facb092e5fe56e466fa8517cde5a206fe21630c1da13", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.30.4@sha256:f508cac11c8300f27529ed73f8d80f9b1949f819e8f8787f28afcb8e47ceb2b4", // renovate:container
	},
	V1_31: {
		ClusterVersion: "v1.31.7", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.6.2/cni-plugins-linux-amd64-v1.6.2.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:b8e811578fb66023f90d2e238d80cec3bdfca4b44049af74c374d4fae0f9c090",
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
				Url:         "https://dl.k8s.io/v1.31.7/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:279e766a1a7c0dce2efae452c9de1e52b169df31c4b75c9d3b7d51f767ae6d42",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.31.7/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:be84c87c7b40977edf67fb8ee231abb273b93bbab5bb770af0f3f37c0d7c4b81",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://dl.k8s.io/v1.31.7/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:80a3c83f00241cd402bc8688464e5e3eedd52a461ee41d882f19cf04ad6d0379",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjMxLjdAc2hhMjU2OjIyYzE5Y2M3MGZlNTgwNmQwYTJjYjI4YTZiNmIzM2ZkMzRlNmY5ZTUwNjE2YmRmNmQ1MzY0OWJjZmFmYmMyNzcifV0=",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjMxLjdAc2hhMjU2OjZhYmU3YTBhY2NlY2YyOWRiNmViYWIxOGExMGY4NDQ2NzhmZmVkNjkzZDc5ZTJlNTFhMThhNmYyYjQ1MzBjYmIifV0=",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjMxLjdAc2hhMjU2OmZiODAyNDliY2I3N2VlNzJiMWM5ZmE1YjcwYmMyOGE4M2VkMTA3YzljYTcxOTU3ODQxYWQ5MWRiMzc5OTYzYmYifV0=",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjE2LTBAc2hhMjU2OmM2YTlkMTFjYzVjMDRiMTE0Y2NkZWYzOWE5MjY1ZWVlZjgxOGUzZDAyZjUzNTliZTAzNWFlNzg0MDk3ZmRlYzUifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		// Check for newer versions at https://github.com/kubernetes/cloud-provider-aws/releases.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.31.5@sha256:6f94a695828de888c8efc57e3243db9e788709767f687a6a8cb07c2e42158862", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.31.4@sha256:839d5b203c3c48eeee1c30002dfae7e481ba650f103c970f2b4b734065bff1f6", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.31.4@sha256:8a1bd6d6eb559ecb814a8745e3fcdbc1ace360398d332caeac3f35fb1bf842dd", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v30.1.4@sha256:0c3695a18d3825492196facb092e5fe56e466fa8517cde5a206fe21630c1da13", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "registry.k8s.io/provider-os/openstack-cloud-controller-manager:v1.32.0@sha256:25e0539888590240483b60dec84c2387fd3cc48bd81dc10a3f6b01fef2585548", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.31.2@sha256:2f2ae9f88573d45d8c72d22abff97fb77fd8d9e55f40e57aa282957e56fd3a1a", // renovate:container
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
