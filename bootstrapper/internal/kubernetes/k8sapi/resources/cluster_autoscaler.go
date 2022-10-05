/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package resources

import (
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"google.golang.org/protobuf/proto"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	rbac "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type AutoscalerDeployment struct {
	PodDisruptionBudget policy.PodDisruptionBudget
	ServiceAccount      k8s.ServiceAccount
	ClusterRole         rbac.ClusterRole
	ClusterRoleBinding  rbac.ClusterRoleBinding
	Role                rbac.Role
	RoleBinding         rbac.RoleBinding
	Service             k8s.Service
	Deployment          apps.Deployment
}

// NewDefaultAutoscalerDeployment creates a new *autoscalerDeployment, customized for the CSP.
func NewDefaultAutoscalerDeployment(extraVolumes []k8s.Volume, extraVolumeMounts []k8s.VolumeMount, env []k8s.EnvVar, k8sVersion versions.ValidK8sVersion) *AutoscalerDeployment {
	return &AutoscalerDeployment{
		PodDisruptionBudget: policy.PodDisruptionBudget{
			TypeMeta: v1.TypeMeta{
				APIVersion: "policy/v1",
				Kind:       "PodDisruptionBudget",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "constellation",
					"app.kubernetes.io/name":       "cluster-autoscaler",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-cluster-autoscaler",
				Namespace: "default",
			},
			Spec: policy.PodDisruptionBudgetSpec{
				Selector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"app.kubernetes.io/instance": "constellation",
						"app.kubernetes.io/name":     "cluster-autoscaler",
					},
				},
				MaxUnavailable: &intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: 1,
				},
			},
		},
		ServiceAccount: k8s.ServiceAccount{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "constellation",
					"app.kubernetes.io/name":       "cluster-autoscaler",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-cluster-autoscaler",
				Namespace: "kube-system",
			},
			AutomountServiceAccountToken: proto.Bool(true),
		},
		ClusterRole: rbac.ClusterRole{
			TypeMeta: v1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "constellation",
					"app.kubernetes.io/name":       "cluster-autoscaler",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name: "constellation-cluster-autoscaler",
			},
			Rules: []rbac.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{
						"events",
						"endpoints",
					},
					Verbs: []string{
						"create",
						"patch",
					},
				},
				{
					APIGroups: []string{""},
					Resources: []string{
						"pods/eviction",
					},
					Verbs: []string{
						"create",
					},
				},
				{
					APIGroups: []string{""},
					Resources: []string{
						"pods/status",
					},
					Verbs: []string{
						"update",
					},
				},
				{
					APIGroups: []string{""},
					Resources: []string{
						"endpoints",
					},
					ResourceNames: []string{
						"cluster-autoscaler",
					},
					Verbs: []string{
						"get",
						"update",
					},
				},
				{
					APIGroups: []string{""},
					Resources: []string{
						"nodes",
					},
					Verbs: []string{
						"watch",
						"list",
						"get",
						"update",
					},
				},
				{
					APIGroups: []string{""},
					Resources: []string{
						"namespaces",
						"pods",
						"services",
						"replicationcontrollers",
						"persistentvolumeclaims",
						"persistentvolumes",
					},
					Verbs: []string{
						"watch",
						"list",
						"get",
					},
				},
				{
					APIGroups: []string{
						"batch",
					},
					Resources: []string{
						"jobs",
						"cronjobs",
					},
					Verbs: []string{
						"watch",
						"list",
						"get",
					},
				},
				{
					APIGroups: []string{
						"batch",
						"extensions",
					},
					Resources: []string{
						"jobs",
					},
					Verbs: []string{
						"get",
						"list",
						"patch",
						"watch",
					},
				},
				{
					APIGroups: []string{
						"extensions",
					},
					Resources: []string{
						"replicasets",
						"daemonsets",
					},
					Verbs: []string{
						"watch",
						"list",
						"get",
					},
				},
				{
					APIGroups: []string{
						"policy",
					},
					Resources: []string{
						"poddisruptionbudgets",
					},
					Verbs: []string{
						"watch",
						"list",
					},
				},
				{
					APIGroups: []string{
						"apps",
					},
					Resources: []string{
						"daemonsets",
						"replicasets",
						"statefulsets",
					},
					Verbs: []string{
						"watch",
						"list",
						"get",
					},
				},
				{
					APIGroups: []string{
						"storage.k8s.io",
					},
					Resources: []string{
						"storageclasses",
						"csinodes",
						"csidrivers",
						"csistoragecapacities",
					},
					Verbs: []string{
						"watch",
						"list",
						"get",
					},
				},
				{
					APIGroups: []string{""},
					Resources: []string{
						"configmaps",
					},
					Verbs: []string{
						"list",
						"watch",
					},
				},
				{
					APIGroups: []string{
						"coordination.k8s.io",
					},
					Resources: []string{
						"leases",
					},
					Verbs: []string{
						"create",
					},
				},
				{
					APIGroups: []string{
						"coordination.k8s.io",
					},
					ResourceNames: []string{
						"cluster-autoscaler",
					},
					Resources: []string{
						"leases",
					},
					Verbs: []string{
						"get",
						"update",
					},
				},
			},
		},
		ClusterRoleBinding: rbac.ClusterRoleBinding{
			TypeMeta: v1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "constellation",
					"app.kubernetes.io/name":       "cluster-autoscaler",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-cluster-autoscaler",
				Namespace: "kube-system",
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "constellation-cluster-autoscaler",
			},
			Subjects: []rbac.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "constellation-cluster-autoscaler",
					Namespace: "kube-system",
				},
			},
		},
		Role: rbac.Role{
			TypeMeta: v1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "Role",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "constellation",
					"app.kubernetes.io/name":       "cluster-autoscaler",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-cluster-autoscaler",
				Namespace: "kube-system",
			},
			Rules: []rbac.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{
						"configmaps",
					},
					Verbs: []string{
						"create",
					},
				},
				{
					APIGroups: []string{""},
					Resources: []string{
						"configmaps",
					},
					ResourceNames: []string{
						"cluster-autoscaler-status",
					},
					Verbs: []string{
						"delete",
						"get",
						"update",
					},
				},
			},
		},
		RoleBinding: rbac.RoleBinding{
			TypeMeta: v1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "RoleBinding",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "constellation",
					"app.kubernetes.io/name":       "cluster-autoscaler",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-cluster-autoscaler",
				Namespace: "kube-system",
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     "constellation-cluster-autoscaler",
			},
			Subjects: []rbac.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "constellation-cluster-autoscaler",
					Namespace: "kube-system",
				},
			},
		},
		Service: k8s.Service{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "constellation",
					"app.kubernetes.io/name":       "cluster-autoscaler",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-cluster-autoscaler",
				Namespace: "default",
			},
			Spec: k8s.ServiceSpec{
				Ports: []k8s.ServicePort{
					{
						Port:       8085,
						Protocol:   k8s.ProtocolTCP,
						TargetPort: intstr.FromInt(8085),
						Name:       "http",
					},
				},
				Selector: map[string]string{
					"app.kubernetes.io/instance": "constellation",
					"app.kubernetes.io/name":     "cluster-autoscaler",
				},
				Type: k8s.ServiceTypeClusterIP,
			},
		},
		Deployment: apps.Deployment{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "constellation",
					"app.kubernetes.io/name":       "cluster-autoscaler",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-cluster-autoscaler",
				Namespace: "kube-system",
			},
			Spec: apps.DeploymentSpec{
				Replicas: proto.Int32(0),
				Selector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"app.kubernetes.io/instance": "constellation",
						"app.kubernetes.io/name":     "cluster-autoscaler",
					},
				},
				Template: k8s.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{
							"app.kubernetes.io/instance": "constellation",
							"app.kubernetes.io/name":     "cluster-autoscaler",
						},
					},
					Spec: k8s.PodSpec{
						PriorityClassName: "system-cluster-critical",
						DNSPolicy:         k8s.DNSClusterFirst,
						Containers: []k8s.Container{
							{
								Name:            "cluster-autoscaler",
								Image:           versions.VersionConfigs[k8sVersion].ClusterAutoscalerImage,
								ImagePullPolicy: k8s.PullIfNotPresent,
								LivenessProbe: &k8s.Probe{
									ProbeHandler: k8s.ProbeHandler{
										HTTPGet: &k8s.HTTPGetAction{
											Path: "/health-check",
											Port: intstr.FromInt(8085),
										},
									},
								},
								Ports: []k8s.ContainerPort{
									{
										ContainerPort: 8085,
									},
								},
								VolumeMounts: extraVolumeMounts,
								Env:          env,
							},
						},
						Volumes:            extraVolumes,
						ServiceAccountName: "constellation-cluster-autoscaler",
						Tolerations: []k8s.Toleration{
							{
								Key:      "node-role.kubernetes.io/master",
								Operator: k8s.TolerationOpExists,
								Effect:   k8s.TaintEffectNoSchedule,
							},
							{
								Key:      "node-role.kubernetes.io/control-plane",
								Operator: k8s.TolerationOpExists,
								Effect:   k8s.TaintEffectNoSchedule,
							},
							{
								Key:      "node.cloudprovider.kubernetes.io/uninitialized",
								Operator: k8s.TolerationOpEqual,
								Value:    "true",
								Effect:   k8s.TaintEffectNoSchedule,
							},
						},
						NodeSelector: map[string]string{
							"node-role.kubernetes.io/control-plane": "",
						},
					},
				},
			},
		},
	}
}

func (a *AutoscalerDeployment) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(a)
}
