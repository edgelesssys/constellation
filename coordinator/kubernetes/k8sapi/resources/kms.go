package resources

import (
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/secrets"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type kmsDeployment struct {
	ServiceAccount     k8s.ServiceAccount
	ClusterRole        rbac.ClusterRole
	ClusterRoleBinding rbac.ClusterRoleBinding
	Deployment         apps.Deployment
	MasterSecret       k8s.Secret
	ImagePullSecret    k8s.Secret
}

const (
	kmsImage = "ghcr.io/edgelesssys/constellation/kmsserver:latest"
)

// NewKMSDeployment creates a new *kmsDeployment to use as the key management system inside Constellation.
func NewKMSDeployment(masterSecret []byte) *kmsDeployment {
	return &kmsDeployment{
		ServiceAccount: k8s.ServiceAccount{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "kms",
				Namespace: "kube-system",
			},
		},
		ClusterRole: rbac.ClusterRole{
			TypeMeta: meta.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: "kms",
				Labels: map[string]string{
					"k8s-app": "kms",
				},
			},
			Rules: []rbac.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"secrets"},
					Verbs:     []string{"get"},
				},
			},
		},
		ClusterRoleBinding: rbac.ClusterRoleBinding{
			TypeMeta: meta.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: "kms",
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "kms",
			},
			Subjects: []rbac.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "kms",
					Namespace: "kube-system",
				},
			},
		},
		Deployment: apps.Deployment{
			TypeMeta: meta.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			ObjectMeta: meta.ObjectMeta{
				Labels: map[string]string{
					"k8s-app": "kms",
				},
				Name:      "kms",
				Namespace: "kube-system",
			},
			Spec: apps.DeploymentSpec{
				Selector: &meta.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "kms",
					},
				},
				Template: k8s.PodTemplateSpec{
					ObjectMeta: meta.ObjectMeta{
						Labels: map[string]string{
							"k8s-app": "kms",
						},
					},
					Spec: k8s.PodSpec{
						ImagePullSecrets: []k8s.LocalObjectReference{
							{
								Name: secrets.PullSecretName,
							},
						},
						Volumes: []k8s.Volume{
							{
								Name: "mastersecret",
								VolumeSource: k8s.VolumeSource{
									Secret: &k8s.SecretVolumeSource{
										SecretName: constants.ConstellationMasterSecretStoreName,
										Items: []k8s.KeyToPath{
											{
												Key:  constants.ConstellationMasterSecretKey,
												Path: "constellation-mastersecret.base64",
											},
										},
									},
								},
							},
						},
						ServiceAccountName: "kms",
						Containers: []k8s.Container{
							{
								Name:  "kms",
								Image: kmsImage,
								VolumeMounts: []k8s.VolumeMount{
									{
										Name:      "mastersecret",
										ReadOnly:  true,
										MountPath: "/constellation/",
									},
								},
							},
						},
					},
				},
			},
		},
		MasterSecret: k8s.Secret{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      constants.ConstellationMasterSecretStoreName,
				Namespace: "kube-system",
			},
			Data: map[string][]byte{
				constants.ConstellationMasterSecretKey: masterSecret,
			},
			Type: "Opaque",
		},
		ImagePullSecret: NewImagePullSecret(),
	}
}

func (c *kmsDeployment) Marshal() ([]byte, error) {
	return MarshalK8SResources(c)
}
