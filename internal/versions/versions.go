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
	KonnectivityAgentImage  = "us.gcr.io/k8s-artifacts-prod/kas-network-proxy/proxy-agent:v0.0.33@sha256:48f2a4ec3e10553a81b8dd1c6fa5fe4bcc9617f78e71c1ca89c6921335e2d7da"
	KonnectivityServerImage = "registry.k8s.io/kas-network-proxy/proxy-server:v0.0.33@sha256:2c111f004bec24888d8cfa2a812a38fb8341350abac67dcd0ac64e709dfe389c"
	JoinImage               = "ghcr.io/edgelesssys/constellation/join-service:v2.1.0@sha256:b0a78ae13d8ff4ed6bc920f5a27ba575e689986ef2d1ee9dd9ba4410a5e30e56"
	AccessManagerImage      = "ghcr.io/edgelesssys/constellation/access-manager:v2.1.0@sha256:e1c083702a5a8e34eb14f0514b8d2b1bbeeb100ab837012d05482da041fc5c40"
	KmsImage                = "ghcr.io/edgelesssys/constellation/kmsserver:v2.1.0@sha256:f56f901bd805550ac8232b4f7fdc8091a03ab16de5deddb1edd22b607413f406"
	VerificationImage       = "ghcr.io/edgelesssys/constellation/verification-service:v2.1.0@sha256:7a1e6bec4cda270924c3495466fa536a2b6cd2d2f9c0be319fc6368710c255e8"
	// Check for new versions at https://github.com/GoogleCloudPlatform/guest-agent/releases and update in /.github/workflows/build-gcp-guest-agent.yml.
	GcpGuestImage            = "ghcr.io/edgelesssys/gcp-guest-agent:20220927.00@sha256:3dea1ae3f162d2353e6584b325f0e325a39cda5f380f41e5a0ee43c6641d3905"
	NodeOperatorCatalogImage = "ghcr.io/edgelesssys/constellation/node-operator-catalog:v2.2.0-pre.0.20221024145821-b35b74b77278@sha256:f1ee4e2642fd758083344df10a98195213dda299fbbc720bf57873e392e001f1"
	// TODO: switch node maintenance operator catalog back to upstream quay.io/medik8s/node-maintenance-operator-catalog
	// once https://github.com/medik8s/node-maintenance-operator/issues/49 is resolved.
	NodeMaintenanceOperatorCatalogImage = "ghcr.io/edgelesssys/constellation/node-maintenance-operator-catalog:v0.13.1-alpha1@sha256:d382c3aaf9bc470cde6f6c05c2c6ff5c9dcfd90540d5b11f9cf69c4e1dd1ca9d"

	QEMUMetadataImage = "ghcr.io/edgelesssys/constellation/qemu-metadata-api:v2.1.0@sha256:abfc36fcd02a145412074cdbb54597878594aa1cfb0ffd66e36d3b3e95ee9e7f"
	LibvirtImage      = "ghcr.io/edgelesssys/constellation/libvirt:v2.2.0-pre.0.20221021080602-f3d78a573fb2@sha256:f42fa5f009415f2c6631b83e8831790d324c27d5f3ae883c59ea7bfeba50facd"

	// ConstellationQEMUImageURL is the artifact URL for QEMU qcow2 images.
	ConstellationQEMUImageURL = "https://d1gl9j3ejrmbpr.cloudfront.net/mini-constellation-v2.1.0.qcow2"

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
		PatchVersion:  "1.23.12",
		CNIPluginsURL: "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:     "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.1/crictl-v1.24.1-linux-amd64.tar.gz",
		KubeletURL:    "https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubelet",
		KubeadmURL:    "https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubeadm",
		KubectlURL:    "https://storage.googleapis.com/kubernetes-release/release/v1.23.12/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.23.2@sha256:5caf74bfe1c6e1b7b7d40345db52b54eeea7229a8fd73c7db9488ef87dc7a496",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v23.0.0@sha256:476616939b85345d7188815045847fcbea8d502464083407cdbb6c934e35820d",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.23.21@sha256:8dea59c658fb7f4ada03bb15bd8a6a5ff8a03b27ff1da2f2b9dc30ee7ee78831",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.23.21@sha256:196fc0f5cd2e73114fdb4c819c7a5e6503f3049549158b25a2e5f1395728953f",
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.23.1@sha256:cd2101ba67f3d6ec719f7792d4bdaa3a50e1b716f3a9ccee8931086496c655b7",
	},
	V1_24: {
		PatchVersion:  "1.24.6",
		CNIPluginsURL: "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:     "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.24.2/crictl-v1.24.2-linux-amd64.tar.gz",
		KubeletURL:    "https://storage.googleapis.com/kubernetes-release/release/v1.24.6/bin/linux/amd64/kubelet",
		KubeadmURL:    "https://storage.googleapis.com/kubernetes-release/release/v1.24.6/bin/linux/amd64/kubeadm",
		KubectlURL:    "https://storage.googleapis.com/kubernetes-release/release/v1.24.6/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.24.1@sha256:4b75b09cc5b3959d06a8c2fb84f165e8163ec0153eaa6a48ece6c8113e78e720",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v24.0.0@sha256:8ee4261980019d3ee8517e12f36fc313fe3ea3e44dd40ee2e004b57f6e5ef171",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.24.8@sha256:109de6e558fa3f68c5d786adddf2eb864fde36c1887a3a04e0d11c85a39954c0",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.24.8@sha256:b98d0a9a683d86c58ea27956f3c7e1263f95541e41b5053c6209c4c5fc647cba",
		// External service image. Depends on k8s version.
		ClusterAutoscalerImage: "k8s.gcr.io/autoscaling/cluster-autoscaler:v1.24.0@sha256:5bd22353ae7f30c9abfaa08189281367ef47ea1b3d09eb13eb26bd13de241e72",
	},
	V1_25: {
		PatchVersion:  "1.25.2",
		CNIPluginsURL: "https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz",
		CrictlURL:     "https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.25.0/crictl-v1.25.0-linux-amd64.tar.gz",
		KubeletURL:    "https://storage.googleapis.com/kubernetes-release/release/v1.25.2/bin/linux/amd64/kubelet",
		KubeadmURL:    "https://storage.googleapis.com/kubernetes-release/release/v1.25.2/bin/linux/amd64/kubeadm",
		KubectlURL:    "https://storage.googleapis.com/kubernetes-release/release/v1.25.2/bin/linux/amd64/kubectl",
		// CloudControllerManagerImageAWS is the CCM image used on AWS.
		CloudControllerManagerImageAWS: "registry.k8s.io/provider-aws/cloud-controller-manager:v1.25.1@sha256:85d3f1e9dacc72531445989bb10999e1e70ebc409d11be57e5baa5f031a893b0",
		// CloudControllerManagerImageGCP is the CCM image used on GCP.
		// TODO: use newer "cloud-provider-gcp" from https://github.com/kubernetes/cloud-provider-gcp when newer releases are available.
		CloudControllerManagerImageGCP: "ghcr.io/edgelesssys/cloud-provider-gcp:v25.2.0@sha256:86fa9d31ed0b3d0d8806f13d6e7debd3471028b2cb7cca3a876d8a31612a7ba5",
		// CloudControllerManagerImageAzure is the CCM image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudControllerManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager:v1.25.2@sha256:d6f128bfdfc2496b67afff35ae22269cd6ac87af4e367a7802d67a65938b2f4e",
		// CloudNodeManagerImageAzure is the cloud-node-manager image used on Azure.
		// Check for newer versions at https://github.com/kubernetes-sigs/cloud-provider-azure/blob/master/README.md.
		CloudNodeManagerImageAzure: "mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager:v1.25.2@sha256:a372695fe6ec9297e20943728990ac4f2213741a9e3cb26c7b93f722298f6aaf",
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
	KubeletURL                       string // k8s version dependency.
	KubeadmURL                       string // k8s version dependency.
	KubectlURL                       string // k8s version dependency.
	CloudControllerManagerImageAWS   string // k8s version dependency.
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
