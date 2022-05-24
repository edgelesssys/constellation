package aws

import (
	"context"

	"github.com/edgelesssys/constellation/coordinator/cloudprovider"
	"github.com/edgelesssys/constellation/coordinator/cloudprovider/cloudtypes"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	k8s "k8s.io/api/core/v1"
)

// CloudControllerManager holds the AWS cloud-controller-manager configuration.
type CloudControllerManager struct{}

// Image returns the container image used to provide cloud-controller-manager for the cloud-provider.
func (c CloudControllerManager) Image() string {
	return cloudprovider.CloudControllerManagerImageAWS
}

// Path returns the path used by cloud-controller-manager executable within the container image.
func (c CloudControllerManager) Path() string {
	return "/aws-cloud-controller-manager"
}

// Name returns the cloud-provider name as used by k8s cloud-controller-manager (k8s.gcr.io/cloud-controller-manager).
func (c CloudControllerManager) Name() string {
	return "aws"
}

// ExtraArgs returns a list of arguments to append to the cloud-controller-manager command.
func (c CloudControllerManager) ExtraArgs() []string {
	return []string{}
}

// ConfigMaps returns a list of ConfigMaps to deploy together with the k8s cloud-controller-manager
// Reference: https://kubernetes.io/docs/concepts/configuration/configmap/ .
func (c CloudControllerManager) ConfigMaps(instance cloudtypes.Instance) (resources.ConfigMaps, error) {
	return resources.ConfigMaps{}, nil
}

// Secrets returns a list of secrets to deploy together with the k8s cloud-controller-manager.
// Reference: https://kubernetes.io/docs/concepts/configuration/secret/ .
func (c CloudControllerManager) Secrets(ctx context.Context, instance cloudtypes.Instance, cloudServiceAccountURI string) (resources.Secrets, error) {
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

// Supported is used to determine if cloud controller manager is implemented for this cloud provider.
func (c CloudControllerManager) Supported() bool {
	return false
}
