package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/internal/cloud/metadata"
	"github.com/edgelesssys/constellation/internal/gcpshared"
	"github.com/edgelesssys/constellation/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/internal/versions"
	k8s "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CloudControllerManager holds the gcp cloud-controller-manager configuration.
type CloudControllerManager struct{}

// Image returns the container image used to provide cloud-controller-manager for the cloud-provider.
func (c *CloudControllerManager) Image(k8sVersion versions.ValidK8sVersion) (string, error) {
	return versions.VersionConfigs[k8sVersion].CloudControllerManagerImageGCP, nil
}

// Path returns the path used by cloud-controller-manager executable within the container image.
func (c *CloudControllerManager) Path() string {
	return "/cloud-controller-manager"
}

// Name returns the cloud-provider name as used by k8s cloud-controller-manager (k8s.gcr.io/cloud-controller-manager).
func (c *CloudControllerManager) Name() string {
	return "gce"
}

// ExtraArgs returns a list of arguments to append to the cloud-controller-manager command.
func (c *CloudControllerManager) ExtraArgs() []string {
	return []string{
		"--use-service-account-credentials",
		"--controllers=cloud-node,cloud-node-lifecycle,nodeipam,service,route",
		"--cloud-config=/etc/gce/gce.conf",
		"--cidr-allocator-type=CloudAllocator",
		"--allocate-node-cidrs=true",
		"--configure-cloud-routes=false",
	}
}

// ConfigMaps returns a list of ConfigMaps to deploy together with the k8s cloud-controller-manager
// Reference: https://kubernetes.io/docs/concepts/configuration/configmap/ .
func (c *CloudControllerManager) ConfigMaps(instance metadata.InstanceMetadata) (resources.ConfigMaps, error) {
	// GCP CCM expects cloud config to contain the GCP project-id and other configuration.
	// reference: https://github.com/kubernetes/cloud-provider-gcp/blob/master/cluster/gce/gci/configure-helper.sh#L791-L892
	var config strings.Builder
	config.WriteString("[global]\n")
	projectID, _, _, err := gcpshared.SplitProviderID(instance.ProviderID)
	if err != nil {
		return resources.ConfigMaps{}, err
	}
	config.WriteString(fmt.Sprintf("project-id = %s\n", projectID))
	config.WriteString("use-metadata-server = true\n")

	nameParts := strings.Split(instance.Name, "-")
	config.WriteString("node-tags = constellation-" + nameParts[len(nameParts)-2] + "\n")

	return resources.ConfigMaps{
		&k8s.ConfigMap{
			TypeMeta: v1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "gceconf",
				Namespace: "kube-system",
			},
			Data: map[string]string{
				"gce.conf": config.String(),
			},
		},
	}, nil
}

// Secrets returns a list of secrets to deploy together with the k8s cloud-controller-manager.
// Reference: https://kubernetes.io/docs/concepts/configuration/secret/ .
func (c *CloudControllerManager) Secrets(ctx context.Context, _ string, cloudServiceAccountURI string) (resources.Secrets, error) {
	serviceAccountKey, err := gcpshared.ServiceAccountKeyFromURI(cloudServiceAccountURI)
	if err != nil {
		return resources.Secrets{}, err
	}
	rawKey, err := json.Marshal(serviceAccountKey)
	if err != nil {
		return resources.Secrets{}, err
	}

	return resources.Secrets{
		&k8s.Secret{
			TypeMeta: v1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "gcekey",
				Namespace: "kube-system",
			},
			Data: map[string][]byte{
				"key.json": rawKey,
			},
		},
	}, nil
}

// Volumes returns a list of volumes to deploy together with the k8s cloud-controller-manager.
// Reference: https://kubernetes.io/docs/concepts/storage/volumes/ .
func (c *CloudControllerManager) Volumes() []k8s.Volume {
	return []k8s.Volume{
		{
			Name: "gceconf",
			VolumeSource: k8s.VolumeSource{
				ConfigMap: &k8s.ConfigMapVolumeSource{
					LocalObjectReference: k8s.LocalObjectReference{
						Name: "gceconf",
					},
				},
			},
		},
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

// VolumeMounts returns a list of volume mounts to deploy together with the k8s cloud-controller-manager.
func (c *CloudControllerManager) VolumeMounts() []k8s.VolumeMount {
	return []k8s.VolumeMount{
		{
			Name:      "gceconf",
			ReadOnly:  true,
			MountPath: "/etc/gce",
		},
		{
			Name:      "gcekey",
			ReadOnly:  true,
			MountPath: "/var/secrets/google",
		},
	}
}

// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cloud-controller-manager.
func (c *CloudControllerManager) Env() []k8s.EnvVar {
	return []k8s.EnvVar{
		{
			Name:  "GOOGLE_APPLICATION_CREDENTIALS",
			Value: "/var/secrets/google/key.json",
		},
	}
}

// Supported is used to determine if cloud controller manager is implemented for this cloud provider.
func (c *CloudControllerManager) Supported() bool {
	return true
}
