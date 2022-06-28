package resources

import (
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type cloudNodeManagerDeployment struct {
	ServiceAccount     k8s.ServiceAccount
	ClusterRole        rbac.ClusterRole
	ClusterRoleBinding rbac.ClusterRoleBinding
	DaemonSet          apps.DaemonSet
}

// NewDefaultCloudNodeManagerDeployment creates a new *cloudNodeManagerDeployment, customized for the CSP.
func NewDefaultCloudNodeManagerDeployment(image, path string, extraArgs []string) *cloudNodeManagerDeployment {
	command := []string{
		path,
		"--node-name=$(NODE_NAME)",
	}
	command = append(command, extraArgs...)
	return &cloudNodeManagerDeployment{
		ServiceAccount: k8s.ServiceAccount{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "cloud-node-manager",
				Namespace: "kube-system",
				Labels: map[string]string{
					"k8s-app":                         "cloud-node-manager",
					"kubernetes.io/cluster-service":   "true",
					"addonmanager.kubernetes.io/mode": "Reconcile",
				},
			},
		},
		ClusterRole: rbac.ClusterRole{
			TypeMeta: meta.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: "cloud-node-manager",
				Labels: map[string]string{
					"k8s-app":                         "cloud-node-manager",
					"kubernetes.io/cluster-service":   "true",
					"addonmanager.kubernetes.io/mode": "Reconcile",
				},
			},
			Rules: []rbac.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"nodes"},
					Verbs:     []string{"watch", "list", "get", "update", "patch"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"nodes/status"},
					Verbs:     []string{"patch"},
				},
			},
		},
		ClusterRoleBinding: rbac.ClusterRoleBinding{
			TypeMeta: meta.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: "cloud-node-manager",
				Labels: map[string]string{
					"k8s-app":                         "cloud-node-manager",
					"kubernetes.io/cluster-service":   "true",
					"addonmanager.kubernetes.io/mode": "Reconcile",
				},
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "cloud-node-manager",
			},
			Subjects: []rbac.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "cloud-node-manager",
					Namespace: "kube-system",
				},
			},
		},
		DaemonSet: apps.DaemonSet{
			TypeMeta: meta.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "cloud-node-manager",
				Namespace: "kube-system",
				Labels: map[string]string{
					"component":                       "cloud-node-manager",
					"kubernetes.io/cluster-service":   "true",
					"addonmanager.kubernetes.io/mode": "Reconcile",
				},
			},
			Spec: apps.DaemonSetSpec{
				Selector: &meta.LabelSelector{
					MatchLabels: map[string]string{"k8s-app": "cloud-node-manager"},
				},
				Template: k8s.PodTemplateSpec{
					ObjectMeta: meta.ObjectMeta{
						Labels:      map[string]string{"k8s-app": "cloud-node-manager"},
						Annotations: map[string]string{"cluster-autoscaler.kubernetes.io/daemonset-pod": "true"},
					},
					Spec: k8s.PodSpec{
						PriorityClassName:  "system-node-critical",
						ServiceAccountName: "cloud-node-manager",
						HostNetwork:        true,
						NodeSelector:       map[string]string{"kubernetes.io/os": "linux"},
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
						Containers: []k8s.Container{
							{
								Name:            "cloud-node-manager",
								Image:           image,
								ImagePullPolicy: k8s.PullIfNotPresent,
								Command:         command,
								Env: []k8s.EnvVar{
									{
										Name: "NODE_NAME",
										ValueFrom: &k8s.EnvVarSource{
											FieldRef: &k8s.ObjectFieldSelector{
												FieldPath: "spec.nodeName",
											},
										},
									},
								},
								Resources: k8s.ResourceRequirements{
									Requests: k8s.ResourceList{
										k8s.ResourceCPU:    resource.MustParse("50m"),
										k8s.ResourceMemory: resource.MustParse("50Mi"),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Marshal marshals the cloud-node-manager deployment as YAML documents.
func (c *cloudNodeManagerDeployment) Marshal() ([]byte, error) {
	return MarshalK8SResources(c)
}
