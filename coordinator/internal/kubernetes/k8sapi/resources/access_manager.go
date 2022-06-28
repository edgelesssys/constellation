package resources

import (
	"github.com/edgelesssys/constellation/internal/secrets"
	"google.golang.org/protobuf/proto"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// accessManagerDeployment holds the configuration for the SSH user creation pods. User/Key definitions are stored in the ConfigMap, and the manager is deployed on each node by the DaemonSet.
type accessManagerDeployment struct {
	ConfigMap       k8s.ConfigMap
	ServiceAccount  k8s.ServiceAccount
	Role            rbac.Role
	RoleBinding     rbac.RoleBinding
	DaemonSet       apps.DaemonSet
	ImagePullSecret k8s.Secret
}

// NewAccessManagerDeployment creates a new *accessManagerDeployment which manages the SSH users for the cluster.
func NewAccessManagerDeployment(sshUsers map[string]string) *accessManagerDeployment {
	return &accessManagerDeployment{
		ServiceAccount: k8s.ServiceAccount{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: v1.ObjectMeta{
				Labels: map[string]string{
					"app.kubernetes.io/instance":   "constellation",
					"app.kubernetes.io/name":       "constellation-access-manager",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-access-manager",
				Namespace: "kube-system",
			},
			AutomountServiceAccountToken: proto.Bool(true),
		},
		ConfigMap: k8s.ConfigMap{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "ssh-users",
				Namespace: "kube-system",
			},
			Data: sshUsers,
		},
		DaemonSet: apps.DaemonSet{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "constellation-access-manager",
				Namespace: "kube-system",
				Labels: map[string]string{
					"app.kubernetes.io/instance": "constellation",
					"app.kubernetes.io/name":     "constellation-access-manager",
				},
			},
			Spec: apps.DaemonSetSpec{
				Selector: &v1.LabelSelector{
					MatchLabels: map[string]string{
						"app.kubernetes.io/instance": "constellation",
						"app.kubernetes.io/name":     "constellation-access-manager",
					},
				},
				Template: k8s.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{
							"app.kubernetes.io/instance": "constellation",
							"app.kubernetes.io/name":     "constellation-access-manager",
						},
					},
					Spec: k8s.PodSpec{
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
						},
						ImagePullSecrets: []k8s.LocalObjectReference{
							{
								Name: secrets.PullSecretName,
							},
						},
						Containers: []k8s.Container{
							{
								Name:            "pause",
								Image:           "gcr.io/google_containers/pause",
								ImagePullPolicy: k8s.PullIfNotPresent,
							},
						},
						InitContainers: []k8s.Container{
							{
								Name:  "constellation-access-manager",
								Image: accessManagerImage,
								VolumeMounts: []k8s.VolumeMount{
									{
										Name:      "host",
										MountPath: "/host",
									},
								},
								SecurityContext: &k8s.SecurityContext{
									Capabilities: &k8s.Capabilities{
										Add: []k8s.Capability{
											"SYS_CHROOT",
										},
									},
								},
							},
						},
						ServiceAccountName: "constellation-access-manager",
						Volumes: []k8s.Volume{
							{
								Name: "host",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/",
									},
								},
							},
						},
					},
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
					"app.kubernetes.io/name":       "constellation-access-manager",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-access-manager",
				Namespace: "kube-system",
			},
			Rules: []rbac.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{
						"configmaps",
					},
					ResourceNames: []string{
						"ssh-users",
					},
					Verbs: []string{
						"get",
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
					"app.kubernetes.io/name":       "constellation-access-manager",
					"app.kubernetes.io/managed-by": "Constellation",
				},
				Name:      "constellation-access-manager",
				Namespace: "kube-system",
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     "constellation-access-manager",
			},
			Subjects: []rbac.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "constellation-access-manager",
					Namespace: "kube-system",
				},
			},
		},
		ImagePullSecret: NewImagePullSecret(),
	}
}

// Marshal marshals the access-manager deployment as YAML documents.
func (c *accessManagerDeployment) Marshal() ([]byte, error) {
	return MarshalK8SResources(c)
}
