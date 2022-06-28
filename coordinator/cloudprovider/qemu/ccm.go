package qemu

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	k8s "k8s.io/api/core/v1"
)

// CloudControllerManager holds the QEMU cloud-controller-manager configuration.
type CloudControllerManager struct{}

// Image returns the container image used to provide cloud-controller-manager for the cloud-provider.
func (c CloudControllerManager) Image() string {
	return ""
}

// Path returns the path used by cloud-controller-manager executable within the container image.
func (c CloudControllerManager) Path() string {
	return "/qemu-cloud-controller-manager"
}

// Name returns the cloud-provider name as used by k8s cloud-controller-manager (k8s.gcr.io/cloud-controller-manager).
func (c CloudControllerManager) Name() string {
	return "qemu"
}

// ExtraArgs returns a list of arguments to append to the cloud-controller-manager command.
func (c CloudControllerManager) ExtraArgs() []string {
	return []string{}
}

// ConfigMaps returns a list of ConfigMaps to deploy together with the k8s cloud-controller-manager
// Reference: https://kubernetes.io/docs/concepts/configuration/configmap/ .
func (c CloudControllerManager) ConfigMaps(instance metadata.InstanceMetadata) (resources.ConfigMaps, error) {
	return resources.ConfigMaps{}, nil
}

// Secrets returns a list of secrets to deploy together with the k8s cloud-controller-manager.
// Reference: https://kubernetes.io/docs/concepts/configuration/secret/ .
func (c CloudControllerManager) Secrets(ctx context.Context, instance metadata.InstanceMetadata, cloudServiceAccountURI string) (resources.Secrets, error) {
	return resources.Secrets{}, nil
}

// Volumes returns a list of volumes to deploy together with the k8s cloud-controller-manager.
// Reference: https://kubernetes.io/docs/concepts/storage/volumes/ .
func (c CloudControllerManager) Volumes() []k8s.Volume {
	return []k8s.Volume{}
}

// VolumeMounts a list of of volume mounts to deploy together with the k8s cloud-controller-manager.
func (c CloudControllerManager) VolumeMounts() []k8s.VolumeMount {
	return []k8s.VolumeMount{}
}

// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cloud-controller-manager.
func (c CloudControllerManager) Env() []k8s.EnvVar {
	return []k8s.EnvVar{}
}

// PrepareInstance is called on every instance before deploying the cloud-controller-manager.
// Allows for cloud-provider specific hooks.
func (c CloudControllerManager) PrepareInstance(instance metadata.InstanceMetadata, vpnIP string) error {
	// no specific hook required.
	return nil
}

// Supported is used to determine if cloud controller manager is implemented for this cloud provider.
func (c CloudControllerManager) Supported() bool {
	return false
}
