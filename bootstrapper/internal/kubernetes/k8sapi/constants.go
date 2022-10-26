/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package k8sapi

const (
	// Paths and permissions necessary for Kubernetes installation.
	cniPluginsDir      = "/opt/cni/bin"
	binDir             = "/run/state/bin"
	kubeadmPath        = "/run/state/bin/kubeadm"
	kubeletPath        = "/run/state/bin/kubelet"
	kubeletServicePath = "/usr/lib/systemd/system/kubelet.service"
	executablePerm     = 0o544
)
