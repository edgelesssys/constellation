/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package resources

import (
	"fmt"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const kmsNamespace = "kube-system"

type KMSDeployment struct {
	ServiceAccount     k8s.ServiceAccount
	Service            k8s.Service
	ClusterRole        rbac.ClusterRole
	ClusterRoleBinding rbac.ClusterRoleBinding
	Deployment         apps.DaemonSet
	MasterSecret       k8s.Secret
}

// KMSConfig is the configuration needed to set up Constellation's key management service.
type KMSConfig struct {
	MasterSecret       []byte
	Salt               []byte
	KMSURI             string
	StorageURI         string
	KeyEncryptionKeyID string
	UseExistingKEK     bool
}

// NewKMSDeployment creates a new *kmsDeployment to use as the key management system inside Constellation.
func NewKMSDeployment(csp string, config KMSConfig) *KMSDeployment {
	return &KMSDeployment{
		ServiceAccount: k8s.ServiceAccount{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "kms",
				Namespace: kmsNamespace,
			},
		},
		Service: k8s.Service{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "kms",
				Namespace: kmsNamespace,
			},
			Spec: k8s.ServiceSpec{
				Type: k8s.ServiceTypeClusterIP,
				Ports: []k8s.ServicePort{
					{
						Name:       "grpc",
						Protocol:   k8s.ProtocolTCP,
						Port:       constants.KMSPort,
						TargetPort: intstr.FromInt(constants.KMSPort),
					},
				},
				Selector: map[string]string{
					"k8s-app": "kms",
				},
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
					Namespace: kmsNamespace,
				},
			},
		},
		Deployment: apps.DaemonSet{
			TypeMeta: meta.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			ObjectMeta: meta.ObjectMeta{
				Labels: map[string]string{
					"k8s-app":                       "kms",
					"component":                     "kms",
					"kubernetes.io/cluster-service": "true",
				},
				Name:      "kms",
				Namespace: kmsNamespace,
			},
			Spec: apps.DaemonSetSpec{
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
						PriorityClassName: "system-cluster-critical",
						Tolerations: []k8s.Toleration{
							{
								Key:      "CriticalAddonsOnly",
								Operator: k8s.TolerationOpExists,
							},
							{
								Key:      "node-role.kubernetes.io/master",
								Operator: k8s.TolerationOpEqual,
								Value:    "true",
								Effect:   k8s.TaintEffectNoSchedule,
							},
							{
								Key:      "node-role.kubernetes.io/control-plane",
								Operator: k8s.TolerationOpExists,
								Effect:   k8s.TaintEffectNoSchedule,
							},
							{
								Operator: k8s.TolerationOpExists,
								Effect:   k8s.TaintEffectNoExecute,
							},
							{
								Operator: k8s.TolerationOpExists,
								Effect:   k8s.TaintEffectNoSchedule,
							},
						},
						// Only run on control plane nodes
						NodeSelector: map[string]string{
							"node-role.kubernetes.io/control-plane": "",
						},
						Volumes: []k8s.Volume{
							{
								Name: "config",
								VolumeSource: k8s.VolumeSource{
									Projected: &k8s.ProjectedVolumeSource{
										Sources: []k8s.VolumeProjection{
											{
												ConfigMap: &k8s.ConfigMapProjection{
													LocalObjectReference: k8s.LocalObjectReference{
														Name: "join-config",
													},
													Items: []k8s.KeyToPath{
														{
															Key:  constants.MeasurementsFilename,
															Path: constants.MeasurementsFilename,
														},
													},
												},
											},
											{
												Secret: &k8s.SecretProjection{
													LocalObjectReference: k8s.LocalObjectReference{
														Name: constants.ConstellationMasterSecretStoreName,
													},
													Items: []k8s.KeyToPath{
														{
															Key:  constants.ConstellationMasterSecretKey,
															Path: constants.ConstellationMasterSecretKey,
														},
														{
															Key:  constants.ConstellationMasterSecretSalt,
															Path: constants.ConstellationMasterSecretSalt,
														},
													},
												},
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
								Image: versions.KmsImage,
								Args: []string{
									fmt.Sprintf("--port=%d", constants.KMSPort),
								},
								VolumeMounts: []k8s.VolumeMount{
									{
										Name:      "config",
										ReadOnly:  true,
										MountPath: constants.ServiceBasePath,
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
				Namespace: kmsNamespace,
			},
			Data: map[string][]byte{
				constants.ConstellationMasterSecretKey:  config.MasterSecret,
				constants.ConstellationMasterSecretSalt: config.Salt,
			},
			Type: "Opaque",
		},
	}
}

func (c *KMSDeployment) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(c)
}
