package gcp

import (
	"github.com/edgelesssys/constellation/internal/kubernetes"
	k8s "k8s.io/api/core/v1"
)

// Autoscaler holds the GCP cluster-autoscaler configuration.
type Autoscaler struct{}

// Name returns the cloud-provider name as used by k8s cluster-autoscaler.
func (a *Autoscaler) Name() string {
	return "gce"
}

// Secrets returns a list of secrets to deploy together with the k8s cluster-autoscaler.
func (a *Autoscaler) Secrets(instance, cloudServiceAccountURI string) (kubernetes.Secrets, error) {
	return kubernetes.Secrets{}, nil
}

// Volumes returns a list of volumes to deploy together with the k8s cluster-autoscaler.
func (a *Autoscaler) Volumes() []k8s.Volume {
	return []k8s.Volume{
		{
			Name: "gcekey",
			VolumeSource: k8s.VolumeSource{
				Secret: &k8s.SecretVolumeSource{
					SecretName: "gcekey",
				},
			},
		},
	}
}

// VolumeMounts returns a list of volume mounts to deploy together with the k8s cluster-autoscaler.
func (a *Autoscaler) VolumeMounts() []k8s.VolumeMount {
	return []k8s.VolumeMount{
		{
			Name:      "gcekey",
			ReadOnly:  true,
			MountPath: "/var/secrets/google",
		},
	}
}

// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cluster-autoscaler.
func (a *Autoscaler) Env() []k8s.EnvVar {
	return []k8s.EnvVar{
		{
			Name:  "GOOGLE_APPLICATION_CREDENTIALS",
			Value: "/var/secrets/google/key.json",
		},
	}
}

// Supported is used to determine if we support autoscaling for the cloud provider.
func (a *Autoscaler) Supported() bool {
	return true
}
