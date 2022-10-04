/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"fmt"
	"strings"
)

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
	return false
}

const (
	// Constellation images.
	// These images are built in a way that they support all versions currently listed in VersionConfigs.
	KonnectivityAgentImage   = "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:v0.0.33@sha256:48f2a4ec3e10553a81b8dd1c6fa5fe4bcc9617f78e71c1ca89c6921335e2d7da"
	KonnectivityServerImage  = "registry.k8s.io/kas-network-proxy/proxy-server:v0.0.33@sha256:2c111f004bec24888d8cfa2a812a38fb8341350abac67dcd0ac64e709dfe389c"
	JoinImage                = "ghcr.io/edgelesssys/constellation/join-service:v2.1.0-pre.0.20220928143744-c46d6e390f2d@sha256:723aead5b049da4e4c64b602f2ab4dbd9f464c63ac274ac4ae2f4c12b38db2fb"
	AccessManagerImage       = "ghcr.io/edgelesssys/constellation/access-manager:v2.0.0@sha256:194b8b5f20658867fd291882ea25bf472bd706bcd5575f37bc9ec92dcdbc2f20"
	KmsImage                 = "ghcr.io/edgelesssys/constellation/kmsserver:v2.0.0@sha256:8c603d9c76c0a1823dab651a0102c6e337bd9ae72b6da31343763e5cc2baaeeb"
	VerificationImage        = "ghcr.io/edgelesssys/constellation/verification-service:v2.0.0@sha256:61929d4b0092462d417b7fe36450688619c0a0c74f44f3c4e215023782b489ae"
	GcpGuestImage            = "ghcr.io/edgelesssys/gcp-guest-agent:20220713.00@sha256:6dca3b81b54ce204f3c7898190a28397d2c79b8f07001f6cf90cb492ef05e9f0"
	NodeOperatorCatalogImage = "ghcr.io/edgelesssys/constellation/node-operator-catalog:v2.1.0-pre.0.20220927150426-f69db6f26ecb@sha256:4a9b0d36309fa788bd70a0ee4d691e4d4b5473d0aa430cc7fd70e4921f0dedd8"
	// TODO: switch node maintenance operator catalog back to upstream quay.io/medik8s/node-maintenance-operator-catalog
	// once https://github.com/medik8s/node-maintenance-operator/issues/49 is resolved.
	NodeMaintenanceOperatorCatalogImage = "ghcr.io/edgelesssys/constellation/node-maintenance-operator-catalog:v0.13.1-alpha1@sha256:d382c3aaf9bc470cde6f6c05c2c6ff5c9dcfd90540d5b11f9cf69c4e1dd1ca9d"

	// ConstellationQEMUImageURL is the artifact URL for QEMU qcow2 images.
	// TODO: Replace with actual artifact URL once we have one (S3, Google Bucket or similar).
	ConstellationQEMUImageURL = "http://localhost/files/constellation.qcow2"

	// currently supported versions.
	//nolint:revive
	V1_23 ValidK8sVersion = "1.23"
	//nolint:revive
	V1_24 ValidK8sVersion = "1.24"
	//nolint:revive
	V1_25 ValidK8sVersion = "1.25"

	Default ValidK8sVersion = V1_24
)

var (
	NodeOperatorVersion            = versionFromDockerImage(NodeOperatorCatalogImage)
	NodeMaintenanceOperatorVersion = versionFromDockerImage(NodeMaintenanceOperatorCatalogImage)
)

// versionConfigs holds download URLs for all required kubernetes components for every supported version.
var VersionConfigs = map[ValidK8sVersion]KubernetesVersion{
	V1_23: {
		PatchVersion:      "1.23.12",
		CNIPluginsURL:     "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.1/crictl-v1.24.1-linux-amd64.tar.gz",
		KubeletServiceURL: "https://raw.githubusercontent.com/kubernetes/release/v0.14.0/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service",
		KubeadmConfURL:    "https://raw.githubusercontent.com/kubernetes/release/v0.14.0/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf",
		KubeletURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubelet",
		KubeadmURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubeadm",
		KubectlURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v23@sha256:476616939b85345d7188815045847fcbea8d502464083407cdbb6c934e35820d",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.23.21@sha256:8dea59c658fb7f4ada03bb15bd8a6a5ff8a03b27ff1da2f2b9dc30ee7ee78831",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.23.21@sha256:196fc0f5cd2e73114fdb4c819c7a5e6503f3049549158b25a2e5f1395728953f",
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.1@sha256:cd2101ba67f3d6ec719f7792d4bdaa3a50e1b716f3a9ccee8931086496c655b7",
	},
	V1_24: {
		PatchVersion:      "1.24.6",
		CNIPluginsURL:     "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.2/crictl-v1.24.2-linux-amd64.tar.gz",
		KubeletServiceURL: "https://raw.githubusercontent.com/kubernetes/release/v0.14.0/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service",
		KubeadmConfURL:    "https://raw.githubusercontent.com/kubernetes/release/v0.14.0/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf",
		KubeletURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.24.6/bin/linux/amd64/kubelet",
		KubeadmURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.24.6/bin/linux/amd64/kubeadm",
		KubectlURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.24.6/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v24@sha256:8ee4261980019d3ee8517e12f36fc313fe3ea3e44dd40ee2e004b57f6e5ef171",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.8@sha256:109de6e558fa3f68c5d786adddf2eb864fde36c1887a3a04e0d11c85a39954c0",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.8@sha256:b98d0a9a683d86c58ea27956f3c7e1263f95541e41b5053c6209c4c5fc647cba",
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.1@sha256:cd2101ba67f3d6ec719f7792d4bdaa3a50e1b716f3a9ccee8931086496c655b7",
	},
	V1_25: {
		PatchVersion:      "1.25.2",
		CNIPluginsURL:     "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:         "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.25.0/crictl-v1.25.0-linux-amd64.tar.gz",
		KubeletServiceURL: "https://raw.githubusercontent.com/kubernetes/release/v0.14.0/cmd/kubepkg/templates/latest/deb/kubelet/lib/systemd/system/kubelet.service",
		KubeadmConfURL:    "https://raw.githubusercontent.com/kubernetes/release/v0.14.0/cmd/kubepkg/templates/latest/deb/kubeadm/10-kubeadm.conf",
		KubeletURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.25.2/bin/linux/amd64/kubelet",
		KubeadmURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.25.2/bin/linux/amd64/kubeadm",
		KubectlURL:        "https://storage.googleapis.com/kubernetes-release/release/v1.25.2/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v24@sha256:8ee4261980019d3ee8517e12f36fc313fe3ea3e44dd40ee2e004b57f6e5ef171",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.0@sha256:df926a931f9c8cb62bf135efeb5dd9540bc155ec51fc2e79f8d0cefbb7d6188b",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.0@sha256:137a5e0aaf6c4e7a59a00157614c700b1f538c04d5412924d98cc5865c7dbf78",
		// External service image. Depends on k8s version.
		// Check for new versions at https://github.com/kubernetes/autoscaler/releases.
		ClusterAutoscalerImage: "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.25.0@sha256:f509ffab618dbd07d129b69ec56963aac7f61aaa792851206b54a2f0bbe046df",
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

// versionFromDockerImage returns the version tag from the image name, e.g. "v1.22.2" from "foocr.io/org/repo:v1.22.2@sha256:3009fj0...".
func versionFromDockerImage(imageName string) string {
	beforeAt, _, _ := strings.Cut(imageName, "@")
	_, version, ok := strings.Cut(beforeAt, ":")
	if !ok {
		panic(fmt.Errorf("failed to extract version from image name, no ':' found in %s", imageName))
	}
	return version
}
