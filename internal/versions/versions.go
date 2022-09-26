/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import "fmt"

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
	case string(V1_22):
		return true
	case string(V1_23):
		return true
	case string(V1_24):
		return true
	case string(V1_25):
		return true
	default:
		return false
	}
}

// IsPreviewK8sVersion checks if a given Kubernetes version is still in preview and not fully supported.
func IsPreviewK8sVersion(version ValidK8sVersion) bool {
	return version == V1_25
}

const (
	// Constellation images.
	// These images are built in a way that they support all versions currently listed in VersionConfigs.
	KonnectivityAgentImage = "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:v0.0.32"
	// TODO: switch back to official image once cilium node2node encryption is enabled.
	// KonnectivityServerImage = "registry.k8s.io/kas-network-proxy/proxy-server:v0.0.32"
	KonnectivityServerImage  = "ghcr.io/3u13r/constellation-konnectivity-server:latest@sha256:a737d2c50f5a2612ffccf2c47349bcccf60fb6e79dfa9af5a50352ae9bc24753"
	JoinImage                = "ghcr.io/edgelesssys/constellation/join-service:v2.0.0"
	AccessManagerImage       = "ghcr.io/edgelesssys/constellation/access-manager:v2.0.0"
	KmsImage                 = "ghcr.io/edgelesssys/constellation/kmsserver:v2.0.0"
	VerificationImage        = "ghcr.io/edgelesssys/constellation/verification-service:v2.0.0"
	GcpGuestImage            = "ghcr.io/edgelesssys/gcp-guest-agent:20220713.00"
	NodeOperatorCatalogImage = "ghcr.io/edgelesssys/constellation/node-operator-catalog"
	NodeOperatorVersion      = "v2.0.0"
	// TODO: switch node maintenance operator catalog back to upstream quay.io/medik8s/node-maintenance-operator-catalog
	// once https://github.com/medik8s/node-maintenance-operator/issues/49 is resolved.
	NodeMaintenanceOperatorCatalogImage = "ghcr.io/edgelesssys/constellation/node-maintenance-operator-catalog"
	NodeMaintenanceOperatorVersion      = "v0.13.1-alpha1"

	// currently supported versions.
	V1_22   ValidK8sVersion = "1.22"
	V1_23   ValidK8sVersion = "1.23"
	V1_24   ValidK8sVersion = "1.24"
	V1_25   ValidK8sVersion = "1.25"
	Default ValidK8sVersion = V1_23
)

// versionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs map[ValidK8sVersion]KubernetesVersion = map[ValidK8sVersion]KubernetesVersion{
	V1_22: {
		PatchVersion:      "1.22.12",
		CNIPluginsURL:     "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.1/crictl-v1.24.1-linux-amd64.tar.gz",
		KubeletServiceURL: "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service",
		KubeadmConfURL:    "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf",
		KubeletURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.22.12/bin/linux/amd64/kubelet",
		KubeadmURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.22.12/bin/linux/amd64/kubeadm",
		KubectlURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.22.12/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v22",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.1.18",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.1.18",
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.22.3",
	},
	V1_23: {
		PatchVersion:      "1.23.9",
		CNIPluginsURL:     "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.1/crictl-v1.24.1-linux-amd64.tar.gz",
		KubeletServiceURL: "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service",
		KubeadmConfURL:    "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf",
		KubeletURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.23.9/bin/linux/amd64/kubelet",
		KubeadmURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.23.9/bin/linux/amd64/kubeadm",
		KubectlURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.23.9/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v23",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.23.11",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.23.11",
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.1",
	},
	V1_24: {
		PatchVersion:      "1.24.3",
		CNIPluginsURL:     "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.1/crictl-v1.24.1-linux-amd64.tar.gz",
		KubeletServiceURL: "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service",
		KubeadmConfURL:    "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf",
		KubeletURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.24.3/bin/linux/amd64/kubelet",
		KubeadmURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.24.3/bin/linux/amd64/kubeadm",
		KubectlURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.24.3/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v24",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.3",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.3",
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.1",
	},
	V1_25: {
		PatchVersion:      "1.25.0",
		CNIPluginsURL:     "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.25.0/crictl-v1.25.0-linux-amd64.tar.gz",
		KubeletServiceURL: "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service",
		KubeadmConfURL:    "https://raw.githubusercontent.com/kubernetes/release/v0.13.0/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf",
		KubeletURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.25.0/bin/linux/amd64/kubelet",
		KubeadmURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.25.0/bin/linux/amd64/kubeadm",
		KubectlURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.25.0/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v24",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.4",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.4",
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.1",
	},
}

// KubernetesVersion bundles download URLs to all version-releated binaries necessary for installing/deploying a particular Kubernetes version.
type KubernetesVersion struct {
	PatchVersion                     string
	CNIPluginsURL                    string // No k8s version dependency.
	CrictlURL                        string // k8s version dependency.
	KubeletServiceURL                string // No k8s version dependency.
	KubeadmConfURL                   string // kubeadm/kubelet v1.11+.
	KubeletURL                       string // k8s version dependency.
	KubeadmURL                       string // k8s version dependency.
	KubectlURL                       string // k8s version dependency.
	CloudControllerManagerImageGCP   string // Using self-built image until resolved: https://github.com/kubernetes/cloud-provider-gcp/issues/289
	CloudControllerManagerImageAzure string // k8s version dependency.
	CloudNodeManagerImageAzure       string // k8s version dependency. Same version as above.
	ClusterAutoscalerImage           string // Matches k8s versioning scheme.
}
