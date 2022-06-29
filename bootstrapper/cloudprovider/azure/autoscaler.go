package azure

import (
	"github.com/edgelesssys/constellation/bootstrapper/internal/kubernetes/k8sapi/resources"
	"github.com/edgelesssys/constellation/internal/azureshared"
	k8s "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Autoscaler holds the Azure cluster-autoscaler configuration.
type Autoscaler struct{}

// Name returns the cloud-provider name as used by k8s cluster-autoscaler.
func (a *Autoscaler) Name() string {
	return "azure"
}

// Secrets returns a list of secrets to deploy together with the k8s cluster-autoscaler.
func (a *Autoscaler) Secrets(providerID string, cloudServiceAccountURI string) (resources.Secrets, error) {
	subscriptionID, resourceGroup, err := azureshared.BasicsFromProviderID(providerID)
	if err != nil {
		return resources.Secrets{}, err
	}
	creds, err := azureshared.ApplicationCredentialsFromURI(cloudServiceAccountURI)
	if err != nil {
		return resources.Secrets{}, err
	}
	return resources.Secrets{
		&k8s.Secret{
			TypeMeta: meta.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "cluster-autoscaler-azure",
				Namespace: "kube-system",
			},
			Data: map[string][]byte{
				"ClientID":       []byte(creds.ClientID),
				"ClientSecret":   []byte(creds.ClientSecret),
				"ResourceGroup":  []byte(resourceGroup),
				"SubscriptionID": []byte(subscriptionID),
				"TenantID":       []byte(creds.TenantID),
				"VMType":         []byte("vmss"),
			},
		},
	}, nil
}

// Volumes returns a list of volumes to deploy together with the k8s cluster-autoscaler.
func (a *Autoscaler) Volumes() []k8s.Volume {
	return []k8s.Volume{}
}

// VolumeMounts returns a list of volume mounts to deploy together with the k8s cluster-autoscaler.
func (a *Autoscaler) VolumeMounts() []k8s.VolumeMount {
	return []k8s.VolumeMount{}
}

// Env returns a list of k8s environment key-value pairs to deploy together with the k8s cluster-autoscaler.
func (a *Autoscaler) Env() []k8s.EnvVar {
	return []k8s.EnvVar{
		{
			Name: "ARM_SUBSCRIPTION_ID",
			ValueFrom: &k8s.EnvVarSource{
				SecretKeyRef: &k8s.SecretKeySelector{
					Key:                  "SubscriptionID",
					LocalObjectReference: k8s.LocalObjectReference{Name: "cluster-autoscaler-azure"},
				},
			},
		},
		{
			Name: "ARM_RESOURCE_GROUP",
			ValueFrom: &k8s.EnvVarSource{
				SecretKeyRef: &k8s.SecretKeySelector{
					Key:                  "ResourceGroup",
					LocalObjectReference: k8s.LocalObjectReference{Name: "cluster-autoscaler-azure"},
				},
			},
		},
		{
			Name: "ARM_TENANT_ID",
			ValueFrom: &k8s.EnvVarSource{
				SecretKeyRef: &k8s.SecretKeySelector{
					Key:                  "TenantID",
					LocalObjectReference: k8s.LocalObjectReference{Name: "cluster-autoscaler-azure"},
				},
			},
		},
		{
			Name: "ARM_CLIENT_ID",
			ValueFrom: &k8s.EnvVarSource{
				SecretKeyRef: &k8s.SecretKeySelector{
					Key:                  "ClientID",
					LocalObjectReference: k8s.LocalObjectReference{Name: "cluster-autoscaler-azure"},
				},
			},
		},
		{
			Name: "ARM_CLIENT_SECRET",
			ValueFrom: &k8s.EnvVarSource{
				SecretKeyRef: &k8s.SecretKeySelector{
					Key:                  "ClientSecret",
					LocalObjectReference: k8s.LocalObjectReference{Name: "cluster-autoscaler-azure"},
				},
			},
		},
		{
			Name: "ARM_VM_TYPE",
			ValueFrom: &k8s.EnvVarSource{
				SecretKeyRef: &k8s.SecretKeySelector{
					Key:                  "VMType",
					LocalObjectReference: k8s.LocalObjectReference{Name: "cluster-autoscaler-azure"},
				},
			},
		},
	}
}

// Supported is used to determine if we support autoscaling for the cloud provider.
func (a *Autoscaler) Supported() bool {
	return true
}
