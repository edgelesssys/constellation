package resources

import (
	"google.golang.org/protobuf/proto"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	cniConfJSON = `{"name":"cbr0","cniVersion":"0.3.1","plugins":[{"type":"flannel","delegate":{"hairpinMode":true,"isDefaultGateway":true}},{"type":"portmap","capabilities":{"portMappings":true}}]}`
	netConfJSON = `{"Network":"10.244.0.0/16","Backend":{"Type":"vxlan"}}`
)

// Reference: https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml
// Changes compared to the reference: added the wireguard interface "wg0" to the args of the "kube-flannel" container of the DaemonSet.

type FlannelDeployment struct {
	PodSecurityPolicy  policy.PodSecurityPolicy
	ClusterRole        rbac.ClusterRole
	ClusterRoleBinding rbac.ClusterRoleBinding
	ServiceAccount     k8s.ServiceAccount
	ConfigMap          k8s.ConfigMap
	DaemonSet          apps.DaemonSet
}

func NewDefaultFlannelDeployment() *FlannelDeployment {
	return &FlannelDeployment{
		PodSecurityPolicy: policy.PodSecurityPolicy{
			TypeMeta: v1.TypeMeta{
				APIVersion: "policy/v1beta1",
				Kind:       "PodSecurityPolicy",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "psp.flannel.unprivileged",
				Annotations: map[string]string{
					"seccomp.security.alpha.kubernetes.io/allowedProfileNames": "docker/default",
					"seccomp.security.alpha.kubernetes.io/defaultProfileName":  "docker/default",
					"apparmor.security.beta.kubernetes.io/allowedProfileNames": "runtime/default",
					"apparmor.security.beta.kubernetes.io/defaultProfileName":  "runtime/default",
				},
			},
			Spec: policy.PodSecurityPolicySpec{
				Privileged: false,
				Volumes: []policy.FSType{
					policy.FSType("configMap"),
					policy.FSType("secret"),
					policy.FSType("emptyDir"),
					policy.FSType("hostPath"),
				},
				AllowedHostPaths: []policy.AllowedHostPath{
					{PathPrefix: "/etc/cni/net.d"},
					{PathPrefix: "/etc/kube-flannel"},
					{PathPrefix: "/run/flannel"},
				},
				ReadOnlyRootFilesystem: false,
				RunAsUser: policy.RunAsUserStrategyOptions{
					Rule: policy.RunAsUserStrategyRunAsAny,
				},
				SupplementalGroups: policy.SupplementalGroupsStrategyOptions{
					Rule: policy.SupplementalGroupsStrategyRunAsAny,
				},
				FSGroup: policy.FSGroupStrategyOptions{
					Rule: policy.FSGroupStrategyRunAsAny,
				},
				AllowPrivilegeEscalation:        proto.Bool(false),
				DefaultAllowPrivilegeEscalation: proto.Bool(false),
				AllowedCapabilities: []k8s.Capability{
					k8s.Capability("NET_ADMIN"),
					k8s.Capability("NET_RAW"),
				},
				HostPID:     false,
				HostIPC:     false,
				HostNetwork: true,
				HostPorts: []policy.HostPortRange{
					{Min: 0, Max: 65535},
				},
				SELinux: policy.SELinuxStrategyOptions{
					Rule: policy.SELinuxStrategyRunAsAny,
				},
			},
		},
		ClusterRole: rbac.ClusterRole{
			TypeMeta: v1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRole",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "flannel",
			},
			Rules: []rbac.PolicyRule{
				{
					APIGroups:     []string{"extensions"},
					Resources:     []string{"podsecuritypolicies"},
					Verbs:         []string{"use"},
					ResourceNames: []string{"psp.flannel.unprivileged"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"pods"},
					Verbs:     []string{"get"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"nodes"},
					Verbs:     []string{"list", "watch"},
				},
				{
					APIGroups: []string{""},
					Resources: []string{"nodes/status"},
					Verbs:     []string{"patch"},
				},
			},
		},
		ClusterRoleBinding: rbac.ClusterRoleBinding{
			TypeMeta: v1.TypeMeta{
				Kind:       "ClusterRoleBinding",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name: "flannel",
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "flannel",
			},
			Subjects: []rbac.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "flannel",
					Namespace: "kube-system",
				},
			},
		},
		ServiceAccount: k8s.ServiceAccount{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "flannel",
				Namespace: "kube-system",
			},
		},
		ConfigMap: k8s.ConfigMap{
			TypeMeta: v1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "kube-flannel-cfg",
				Namespace: "kube-system",
				Labels: map[string]string{
					"tier": "node",
					"app":  "flannel",
				},
			},
			Data: map[string]string{
				"cni-conf.json": cniConfJSON,
				"net-conf.json": netConfJSON,
			},
		},
		DaemonSet: apps.DaemonSet{
			TypeMeta: v1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      "kube-flannel-ds",
				Namespace: "kube-system",
				Labels: map[string]string{
					"tier": "node",
					"app":  "flannel",
				},
			},
			Spec: apps.DaemonSetSpec{
				Selector: &v1.LabelSelector{
					MatchLabels: map[string]string{"app": "flannel"},
				},
				Template: k8s.PodTemplateSpec{
					ObjectMeta: v1.ObjectMeta{
						Labels: map[string]string{
							"tier": "node",
							"app":  "flannel",
						},
					},
					Spec: k8s.PodSpec{
						Affinity: &k8s.Affinity{
							NodeAffinity: &k8s.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &k8s.NodeSelector{
									NodeSelectorTerms: []k8s.NodeSelectorTerm{
										{MatchExpressions: []k8s.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/os",
												Operator: k8s.NodeSelectorOpIn,
												Values:   []string{"linux"},
											},
										}},
									},
								},
							},
						},
						HostNetwork:       true,
						PriorityClassName: "system-node-critical",
						Tolerations: []k8s.Toleration{
							{
								Operator: k8s.TolerationOpExists,
								Effect:   k8s.TaintEffectNoSchedule,
							},
						},
						ServiceAccountName: "flannel",
						InitContainers: []k8s.Container{
							{
								Name:    "install-cni-plugin",
								Image:   "rancher/mirrored-flannelcni-flannel-cni-plugin:v1.0.0",
								Command: []string{"cp"},
								Args:    []string{"-f", "/flannel", "/opt/cni/bin/flannel"},
								VolumeMounts: []k8s.VolumeMount{
									{
										Name:      "cni-plugin",
										MountPath: "/opt/cni/bin",
									},
								},
							},
							{
								Name:    "install-cni",
								Image:   "quay.io/coreos/flannel:v0.15.1",
								Command: []string{"cp"},
								Args:    []string{"-f", "/etc/kube-flannel/cni-conf.json", "/etc/cni/net.d/10-flannel.conflist"},
								VolumeMounts: []k8s.VolumeMount{
									{
										Name:      "cni",
										MountPath: "/etc/cni/net.d",
									},
									{
										Name:      "flannel-cfg",
										MountPath: "/etc/kube-flannel/",
									},
								},
							},
						},
						Containers: []k8s.Container{
							{
								Name:    "kube-flannel",
								Image:   "quay.io/coreos/flannel:v0.15.1",
								Command: []string{"/opt/bin/flanneld"},
								Args:    []string{"--ip-masq", "--kube-subnet-mgr", "--iface", "wg0"},
								Resources: k8s.ResourceRequirements{
									Requests: k8s.ResourceList{
										"cpu":    resource.MustParse("100m"),
										"memory": resource.MustParse("50Mi"),
									},
									Limits: k8s.ResourceList{
										"cpu":    resource.MustParse("100m"),
										"memory": resource.MustParse("50Mi"),
									},
								},
								SecurityContext: &k8s.SecurityContext{
									Privileged: proto.Bool(false),
									Capabilities: &k8s.Capabilities{
										Add: []k8s.Capability{k8s.Capability("NET_ADMIN"), k8s.Capability("NET_RAW")},
									},
								},
								Env: []k8s.EnvVar{
									{
										Name: "POD_NAME",
										ValueFrom: &k8s.EnvVarSource{
											FieldRef: &k8s.ObjectFieldSelector{FieldPath: "metadata.name"},
										},
									},
									{
										Name: "POD_NAMESPACE",
										ValueFrom: &k8s.EnvVarSource{
											FieldRef: &k8s.ObjectFieldSelector{FieldPath: "metadata.namespace"},
										},
									},
								},
								VolumeMounts: []k8s.VolumeMount{
									{
										Name:      "run",
										MountPath: "/run/flannel",
									},
									{
										Name:      "flannel-cfg",
										MountPath: "/etc/kube-flannel/",
									},
								},
							},
						},
						Volumes: []k8s.Volume{
							{
								Name: "run",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/run/flannel",
									},
								},
							},
							{
								Name: "cni-plugin",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/opt/cni/bin",
									},
								},
							},
							{
								Name: "cni",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/etc/cni/net.d",
									},
								},
							},
							{
								Name: "flannel-cfg",
								VolumeSource: k8s.VolumeSource{
									ConfigMap: &k8s.ConfigMapVolumeSource{
										LocalObjectReference: k8s.LocalObjectReference{
											Name: "kube-flannel-cfg",
										},
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

func (f *FlannelDeployment) Marshal() ([]byte, error) {
	return MarshalK8SResources(f)
}
