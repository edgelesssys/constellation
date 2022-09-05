/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package qemu

import (
	"github.com/edgelesssys/constellation/internal/kubernetes"
	k8s "k8s.io/api/core/v1"
)

// Autoscaler holds the QEMU cluster-autoscaler configuration.
type Autoscaler struct{}

// Name returns the cloud-provider name as used by k8s cluster-autoscaler.
func (a Autoscaler) Name() string {
	return "qemu"
}

// Secrets returns a list of secrets to deploy together with the k8s cluster-autoscaler.
func (a Autoscaler) Secrets(providerID, cloudServiceAccountURI string) (kubernetes.Secrets, error) {
	return kubernetes.Secrets{}, nil
}

// Volumes returns a list of volumes to deploy together with the k8s cluster-autoscaler.
func (a Autoscaler) Volumes() []k8s.Volume {
	return []k8s.Volume{}
}

// VolumeMounts returns a list of volume mounts to deploy together with the k8s cluster-autoscaler.
func (a Autoscaler) VolumeMounts() []k8s.VolumeMount {
	return []k8s.VolumeMount{}
}

// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cluster-autoscaler.
func (a Autoscaler) Env() []k8s.EnvVar {
	return []k8s.EnvVar{}
}

// Supported is used to determine if we support autoscaling for the cloud provider.
func (a Autoscaler) Supported() bool {
	return false
}
