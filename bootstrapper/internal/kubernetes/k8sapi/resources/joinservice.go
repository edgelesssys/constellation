/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package resources

import (
	"fmt"

	"github.com/edgelesssys/constellation/internal/cloud/cloudprovider"
	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/kubernetes"
	"github.com/edgelesssys/constellation/internal/versions"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type joinServiceDaemonset struct {
	ClusterRole        rbac.ClusterRole
	ClusterRoleBinding rbac.ClusterRoleBinding
	ConfigMap          k8s.ConfigMap
	DaemonSet          apps.DaemonSet
	ServiceAccount     k8s.ServiceAccount
	Service            k8s.Service
}

// NewJoinServiceDaemonset returns a daemonset for the join service.
func NewJoinServiceDaemonset(csp, measurementsJSON, enforcedPCRsJSON, initialIdKeyDigest, enforceIdKeyDigest string, measurementSalt []byte) *joinServiceDaemonset {
	joinConfigData := map[string]string{
		constants.MeasurementsFilename: measurementsJSON,
		constants.EnforcedPCRsFilename: enforcedPCRsJSON,
	}
	if cloudprovider.FromString(csp) == cloudprovider.Azure {
		joinConfigData[constants.EnforceIdKeyDigestFilename] = enforceIdKeyDigest
		joinConfigData[constants.IdKeyDigestFilename] = initialIdKeyDigest
	}

	return &joinServiceDaemonset{
		ClusterRole: rbac.ClusterRole{
			TypeMeta: meta.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: "join-service",
				Labels: map[string]string{
					"k8s-app": "join-service",
				},
			},
			Rules: []rbac.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"secrets"},
					Verbs:     []string{"get", "list", "create", "update"},
				},
				{
					APIGroups: []string{"rbac.authorization.k8s.io"},
					Resources: []string{"roles", "rolebindings"},
					Verbs:     []string{"create", "update"},
				},
			},
		},
		ClusterRoleBinding: rbac.ClusterRoleBinding{
			TypeMeta: meta.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: "join-service",
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "join-service",
			},
			Subjects: []rbac.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "join-service",
					Namespace: constants.ConstellationNamespace,
				},
			},
		},
		DaemonSet: apps.DaemonSet{
			TypeMeta: meta.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "join-service",
				Namespace: constants.ConstellationNamespace,
				Labels: map[string]string{
					"k8s-app":                       "join-service",
					"component":                     "join-service",
					"kubernetes.io/cluster-service": "true",
				},
			},
			Spec: apps.DaemonSetSpec{
				Selector: &meta.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "join-service",
					},
				},
				Template: k8s.PodTemplateSpec{
					ObjectMeta: meta.ObjectMeta{
						Labels: map[string]string{
							"k8s-app": "join-service",
						},
					},
					Spec: k8s.PodSpec{
						PriorityClassName:  "system-cluster-critical",
						ServiceAccountName: "join-service",
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
						Containers: []k8s.Container{
							{
								Name:  "join-service",
								Image: versions.JoinImage,
								Ports: []k8s.ContainerPort{
									{
										ContainerPort: constants.JoinServicePort,
										Name:          "tcp",
									},
								},
								SecurityContext: &k8s.SecurityContext{
									Privileged: func(b bool) *bool { return &b }(true),
								},
								Args: []string{
									fmt.Sprintf("--cloud-provider=%s", csp),
									fmt.Sprintf("--kms-endpoint=kms.kube-system:%d", constants.KMSPort),
								},
								VolumeMounts: []k8s.VolumeMount{
									{
										Name:      "config",
										ReadOnly:  true,
										MountPath: constants.ServiceBasePath,
									},
									{
										Name:      "kubeadm",
										ReadOnly:  true,
										MountPath: "/etc/kubernetes",
									},
								},
							},
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
														Name: constants.JoinConfigMap,
													},
												},
											},
											{
												ConfigMap: &k8s.ConfigMapProjection{
													LocalObjectReference: k8s.LocalObjectReference{
														Name: constants.K8sVersion,
													},
												},
											},
											{
												ConfigMap: &k8s.ConfigMapProjection{
													LocalObjectReference: k8s.LocalObjectReference{
														Name: constants.InternalConfigMap,
													},
												},
											},
										},
									},
								},
							},
							{
								Name: "kubeadm",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/etc/kubernetes",
									},
								},
							},
						},
					},
				},
			},
		},
		ServiceAccount: k8s.ServiceAccount{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "join-service",
				Namespace: constants.ConstellationNamespace,
			},
		},
		Service: k8s.Service{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "join-service",
				Namespace: constants.ConstellationNamespace,
			},
			Spec: k8s.ServiceSpec{
				Type: k8s.ServiceTypeNodePort,
				Ports: []k8s.ServicePort{
					{
						Name:       "grpc",
						Protocol:   k8s.ProtocolTCP,
						Port:       constants.JoinServicePort,
						TargetPort: intstr.IntOrString{IntVal: constants.JoinServicePort},
						NodePort:   constants.JoinServiceNodePort,
					},
				},
				Selector: map[string]string{
					"k8s-app": "join-service",
				},
			},
		},
		ConfigMap: k8s.ConfigMap{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      constants.JoinConfigMap,
				Namespace: constants.ConstellationNamespace,
			},
			Data: joinConfigData,
			BinaryData: map[string][]byte{
				constants.MeasurementSaltFilename: measurementSalt,
			},
		},
	}
}

// Marshal the daemonset using the Kubernetes resource marshaller.
func (a *joinServiceDaemonset) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(a)
}
