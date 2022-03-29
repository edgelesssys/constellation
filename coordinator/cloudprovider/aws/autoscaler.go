package aws

import (
	"github.com/edgelesssys/constellation/coordinator/core"
	"github.com/edgelesssys/constellation/coordinator/kubernetes/k8sapi/resources"
	k8s "k8s.io/api/core/v1"
)

// Autoscaler holds the AWS cluster-autoscaler configuration.
type Autoscaler struct{}

// Name returns the cloud-provider name as used by k8s cluster-autoscaler.
func (a Autoscaler) Name() string {
	return "aws"
}

// Secrets returns a list of secrets to deploy together with the k8s cluster-autoscaler.
func (a Autoscaler) Secrets(instance core.Instance, cloudServiceAccountURI string) (resources.Secrets, error) {
	return resources.Secrets{}, nil
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
