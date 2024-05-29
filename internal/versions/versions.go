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
	case semver.MajorMinor(string(V1_27)):
		return string(V1_27)
	case semver.MajorMinor(string(V1_28)):
		return string(V1_28)
	case semver.MajorMinor(string(V1_29)):
		return string(V1_29)
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
	GcpGuestImage = "ghcr.io/edgelesssys/gcp-guest-agent:v20240314.0.0@sha256:56f5f5250056174c82cffe6b3190838f4e001cf6375eeea0b7847d679e0a600f" // renovate:container
	// NodeMaintenanceOperatorImage is the image for the node maintenance operator.
	NodeMaintenanceOperatorImage = "quay.io/medik8s/node-maintenance-operator:v0.15.0@sha256:8cb8dad93283268282c30e75c68f4bd76b28def4b68b563d2f9db9c74225d634" // renovate:container
	// LogstashImage is the container image of logstash, used for log collection by debugd.
	LogstashImage = "ghcr.io/edgelesssys/constellation/logstash-debugd:v2.17.0-pre.0.20240513104207-d76c9ac82de7@sha256:e7d77bea381354c58fd9783eb9e2b846b14ea846d531eb3c54d5dd89917908c4" // renovate:container
	// FilebeatImage is the container image of filebeat, used for log collection by debugd.
	FilebeatImage = "ghcr.io/edgelesssys/constellation/filebeat-debugd:v2.17.0-pre.0.20240513104207-d76c9ac82de7@sha256:6567d682385c06b49f6d56fdf3f20d5c24809bbfced15b816f4717bf837fc776" // renovate:container
	// MetricbeatImage is the container image of filebeat, used for log collection by debugd.
	MetricbeatImage = "ghcr.io/edgelesssys/constellation/metricbeat-debugd:v2.17.0-pre.0.20240513104207-d76c9ac82de7@sha256:d16fe0371fbb383ca3cfa70d4d10ea5124fd87054bfac189ce5458ba350bb25d" // renovate:container

	// currently supported versions.
	//nolint:revive
	V1_27 ValidK8sVersion = "v1.27.13" // renovate:kubernetes-release
	//nolint:revive
	V1_28 ValidK8sVersion = "v1.28.9" // renovate:kubernetes-release
	//nolint:revive
	V1_29 ValidK8sVersion = "v1.29.4" // renovate:kubernetes-release

	// Default k8s version deployed by Constellation.
	Default ValidK8sVersion = V1_28
)

// Regenerate the hashes by running go generate.
// To add another Kubernetes version, add a new entry to the VersionConfigs map below and fill the Hash field with an empty string.
//go:generate hash-generator

// VersionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_27: {
		ClusterVersion: "v1.27.13", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.4.0/cni-plugins-linux-amd64-v1.4.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:c2485ddb3ffc176578ae30ae58137f0b88e50f7c7f2af7d53a569276b2949a33",
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
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.13/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:ed68df2a77f3057ab47f57eacb6e9310e91731e4f43c58a3c3b5c857d78d0080",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.13/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:b88c30b7067f095b7fa02c5560cc50d6e69a5a9fecc606ef477dc7efc86453b9",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.27.13/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:e991f163197cbd85bbff22f656a74d48b69db5addfa43cc04cca0cf5328f57f1",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI3LjEzQHNoYTI1Njo3YTM0YmZlYjcxOWMyNmEwOWJmMjZiZjUxYjhiZjM0N2Q0MWY0ZWQ1ZTMwZTkzNGY2Y2E3M2FjMTk3MjA3NGE3In1d",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI3LjEzQHNoYTI1Njo5NGVmNzdkZDQxNzYzZjMwZDNkOGViNWNhNWZhNzIwZjI4OWMyNGZlNTc5NGE3YTYyZTJkYTEwODJiOTVkMjhlIn1d",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI3LjEzQHNoYTI1NjpkMTA4NDI0NTk5ZmI3NjI1NDQxNmIwYTU2MDlkYjUzMGE4YmQzNzJhNzg4N2ZiNTYzNjc1MzI4ODU5YmRmZWIyIn1d",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjEyLTBAc2hhMjU2OjQ0YThlMjRkY2JiYTM0NzBlZTFmZWUyMWQ1ZTg4ZDEyOGM5MzZlOWI1NWQ0YmM1MWZiZWY4MDg2ZjhlZDEyM2IifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.27.6@sha256:1891fb7f0cef58a657854ce553fcbda1a98b7a7418a17688055ca014f6451f2b", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.27.17@sha256:fc2eb36fbf5ba8e854da02574ce434a7a02669180fdb6dba4c068a072985ac47", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.27.17@sha256:f7ba1998387e669f148b70e5ca7269d75fd9e10cd638ae107f8da0540f6d0ac1", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v27.1.6@sha256:b097b4e5382ea1987db5996a9eaffb94fa224639b3464876f0b1b17f64509ac4", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.27.8@sha256:521727ab44a163de45f8a607f20bce2cb79b3d87da09fa263ebdfab880a09eb7", // renovate:container
	},
	V1_28: {
		ClusterVersion: "v1.28.9", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.4.0/cni-plugins-linux-amd64-v1.4.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:c2485ddb3ffc176578ae30ae58137f0b88e50f7c7f2af7d53a569276b2949a33",
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
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.28.9/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:f3af46cff11c675a80d91ebb38ebc4e85a9f813ce93e56ee131e7fea1491b786",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.28.9/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:a4d8acf0a74cb1d07d96a1a34148f54c6420874221af16d8ec902d9bffc7ef89",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.28.9/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:b4693d0b22f509250694b10c7727c42b427d570af04f2065fe23a55d6c0051f1",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI4LjlAc2hhMjU2OjdkMTFjNjJiMjRjZjVkM2ViMWZmZDZhYzMxNWU2MzI4ZmExYWUyMzk2ZDQ1NzJkNDY3NDA0M2U5YTdkZDRlYzIifV0=",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI4LjlAc2hhMjU2OmZmNmRhMTYxNGIyMTJjNGNiMDI3ZGI2NjQwYzJmMDAyYTA3MjNiMzg1N2JlOWE2OWUzZWM2YWQ0ZWE4YTVlMWMifV0=",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI4LjlAc2hhMjU2OmMyZDJmOTcyM2M0MjU0MDEwYjE4Y2YzODcwYjdmMTZkNzE3MTlkNTlhOWMxNGY5NWNjNWZhMmU1MjRkODU4NjkifV0=",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjEyLTBAc2hhMjU2OjQ0YThlMjRkY2JiYTM0NzBlZTFmZWUyMWQ1ZTg4ZDEyOGM5MzZlOWI1NWQ0YmM1MWZiZWY4MDg2ZjhlZDEyM2IifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.28.5@sha256:ac0579afacbf2db0e9b2dda6c777c3f27ce877c17dc29db623998ab09858ac0e", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.27.17@sha256:fc2eb36fbf5ba8e854da02574ce434a7a02669180fdb6dba4c068a072985ac47", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.27.17@sha256:f7ba1998387e669f148b70e5ca7269d75fd9e10cd638ae107f8da0540f6d0ac1", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v27.1.6@sha256:b097b4e5382ea1987db5996a9eaffb94fa224639b3464876f0b1b17f64509ac4", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.27.8@sha256:521727ab44a163de45f8a607f20bce2cb79b3d87da09fa263ebdfab880a09eb7", // renovate:container
	},
	V1_29: {
		ClusterVersion: "v1.29.4", // renovate:kubernetes-release
		KubernetesComponents: components.Components{
			{
				Url:         "https://github.com/containernetworking/plugins/releases/download/v1.4.0/cni-plugins-linux-amd64-v1.4.0.tgz", // renovate:cni-plugins-release
				Hash:        "sha256:c2485ddb3ffc176578ae30ae58137f0b88e50f7c7f2af7d53a569276b2949a33",
				InstallPath: constants.CniPluginsDir,
				Extract:     true,
			},
			{
				Url:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.29.0/crictl-v1.29.0-linux-amd64.tar.gz", // renovate:crictl-release
				Hash:        "sha256:d16a1ffb3938f5a19d5c8f45d363bd091ef89c0bc4d44ad16b933eede32fdcbb",
				InstallPath: constants.BinDir,
				Extract:     true,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.29.4/bin/linux/amd64/kubelet", // renovate:kubernetes-release
				Hash:        "sha256:58571f0ed62543a9bbac541e52c15d8385083113a463e23aec1341d0b5043939",
				InstallPath: constants.KubeletPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.29.4/bin/linux/amd64/kubeadm", // renovate:kubernetes-release
				Hash:        "sha256:ea20ab064f716ab7f69a36d72df340257b31c9721ea86e1cf9d70b35999ddeea",
				InstallPath: constants.KubeadmPath,
				Extract:     false,
			},
			{
				Url:         "https://storage.googleapis.com/kubernetes-release/release/v1.29.4/bin/linux/amd64/kubectl", // renovate:kubernetes-release
				Hash:        "sha256:10e343861c3cb0010161e703307ba907add2aeeeaffc6444779ad915f9889c88",
				InstallPath: constants.KubectlPath,
				Extract:     false,
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtYXBpc2VydmVyOnYxLjI5LjRAc2hhMjU2OjgwYmVmODIwZTk1NzdmMTQ5ZTUyMTkxOTZmNGRlZjQ5MTJhZWQ1ZmFiYmYwN2Q3YzQxNWQyNDEwNWY5ZWMwNTkifV0=",
				InstallPath: patchFilePath("kube-apiserver"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtY29udHJvbGxlci1tYW5hZ2VyOnYxLjI5LjRAc2hhMjU2OjM4YTgzMjZkZjk3ZDI4NDY4MjQ5MTcwYWE2MzI1MGI4Y2U3Yjg3OTMxNTVmZjczZTI0ZDcwOTQyNGQ1YzlmMDkifV0=",
				InstallPath: patchFilePath("kube-controller-manager"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2t1YmUtc2NoZWR1bGVyOnYxLjI5LjRAc2hhMjU2OjhkYzZhODY1MzE2NDA0YWM3ODVmYmU4OWQ3MTY3ZmVkOThlNjljYTk4MDgyMDZiNTU2YWRmYTZkZDZkNDk3ZDAifV0=",
				InstallPath: patchFilePath("kube-scheduler"),
			},
			{
				Url:         "data:application/json;base64,W3sib3AiOiJyZXBsYWNlIiwicGF0aCI6Ii9zcGVjL2NvbnRhaW5lcnMvMC9pbWFnZSIsInZhbHVlIjoicmVnaXN0cnkuazhzLmlvL2V0Y2Q6My41LjEyLTBAc2hhMjU2OjQ0YThlMjRkY2JiYTM0NzBlZTFmZWUyMWQ1ZTg4ZDEyOGM5MzZlOWI1NWQ0YmM1MWZiZWY4MDg2ZjhlZDEyM2IifV0=",
				InstallPath: patchFilePath("etcd"),
			},
		},
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.29.2@sha256:db37f9ad7cb64d2248fee5354e5d2d9311bea5f188bdf24ffc3380a4c00066c9", // renovate:container
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.28.9@sha256:7abf813ad41b8f1ed91f50bfefb6f285b367664c57758af0b5a623b65f55b34b", // renovate:container
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.28.9@sha256:65285c13cc3eced1005a1c6c5f727570d781ac25f421a9e5cf169de8d7e1d6a9", // renovate:container
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO(3u13r): use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v29.0.1@sha256:66651afaf19f7da2f60f80fd733ee8cc7d161767f34f0da4f111ff35387c5217", // renovate:container
		// CloudControllerManagerImageOpenStack is the CCM image used on OpenStack.
		CloudControllerManagerImageOpenStack: "docker.io/k8scloudprovider/openstack-cloud-controller-manager:v1.26.4@sha256:05e846fb13481b6dbe4a1e50491feb219e8f5101af6cf662a086115735624db0", // renovate:container
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "registry.k8s.io/autoscaling/cluster-autoscaler:v1.29.3@sha256:ea973cf727ff98a0c8c8f770a0787c26a86604182565ce180a5665936f3b38bc", // renovate:container
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
