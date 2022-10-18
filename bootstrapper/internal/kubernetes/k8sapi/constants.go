/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package k8sapi

const (
	// Paths and permissions necessary for Kubernetes installation.
	cniPluginsDir           = "/opt/cni/bin"
	binDir                  = "/run/state/bin"
	kubeadmPath             = "/run/state/bin/kubeadm"
	kubeletPath             = "/run/state/bin/kubelet"
	kubeletServiceEtcPath   = "/etc/systemd/system/kubelet.service"
	kubeletServiceStatePath = "/run/state/systemd/system/kubelet.service"
	kubeadmConfEtcPath      = "/etc/systemd/system/kubelet.service.d/10-kubeadm.conf"
	kubeadmConfStatePath    = "/run/state/systemd/system/kubelet.service.d/10-kubeadm.conf"
	executablePerm          = 0o544
	systemdUnitPerm         = 0o644
)
