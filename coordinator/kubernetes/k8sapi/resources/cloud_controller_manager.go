package resources

import (
	"fmt"

	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type cloudControllerManagerDeployment struct {
	ServiceAccount     k8s.ServiceAccount
	ClusterRoleBinding rbac.ClusterRoleBinding
	DaemonSet          apps.DaemonSet
}

// references:
// https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/admin/cloud/ccm-example.yaml
// https://kubernetes.io/docs/tasks/administer-cluster/running-cloud-controller/#cloud-controller-manager

func NewDefaultCloudControllerManagerDeployment(cloudProvider string, image string, path string) *cloudControllerManagerDeployment {
	return &cloudControllerManagerDeployment{
		ServiceAccount: k8s.ServiceAccount{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "cloud-controller-manager",
				Namespace: "kube-system",
			},
		},
		ClusterRoleBinding: rbac.ClusterRoleBinding{
			TypeMeta: meta.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: meta.ObjectMeta{
				Name: "system:cloud-controller-manager",
			},
			RoleRef: rbac.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "cluster-admin",
			},
			Subjects: []rbac.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "cloud-controller-manager",
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
				Labels: map[string]string{
					"k8s-app": "cloud-controller-manager",
				},
				Name:      "cloud-controller-manager",
				Namespace: "kube-system",
			},
			Spec: apps.DaemonSetSpec{
				Selector: &meta.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "cloud-controller-manager",
					},
				},
				Template: k8s.PodTemplateSpec{
					ObjectMeta: meta.ObjectMeta{
						Labels: map[string]string{
							"k8s-app": "cloud-controller-manager",
						},
					},
					Spec: k8s.PodSpec{
						ServiceAccountName: "cloud-controller-manager",
						Containers: []k8s.Container{
							{
								Name:  "cloud-controller-manager",
								Image: image,
								Command: []string{
									path,
									fmt.Sprintf("--cloud-provider=%s", cloudProvider),
									"--leader-elect=true",
									"--allocate-node-cidrs=false",
									"--configure-cloud-routes=false",
									"--controllers=cloud-node,cloud-node-lifecycle",
									"--use-service-account-credentials",
									"--cluster-cidr=10.244.0.0/16",
									"-v=2",
								},
								VolumeMounts: []k8s.VolumeMount{
									{
										MountPath: "/etc/kubernetes",
										Name:      "etckubernetes",
										ReadOnly:  true,
									},
									{
										MountPath: "/etc/ssl",
										Name:      "etcssl",
										ReadOnly:  true,
									},
									{
										MountPath: "/etc/pki",
										Name:      "etcpki",
										ReadOnly:  true,
									},
									{
										MountPath: "/etc/gce.conf",
										Name:      "gceconf",
										ReadOnly:  true,
									},
								},
							},
						},
						Volumes: []k8s.Volume{
							{
								Name: "etckubernetes",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{Path: "/etc/kubernetes"},
								},
							},
							{
								Name: "etcssl",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{Path: "/etc/ssl"},
								},
							},
							{
								Name: "etcpki",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{Path: "/etc/pki"},
								},
							},
							{
								Name: "gceconf",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{Path: "/etc/gce.conf"},
								},
							},
						},
						Tolerations: []k8s.Toleration{
							{
								Key:    "node.cloudprovider.kubernetes.io/uninitialized",
								Value:  "true",
								Effect: k8s.TaintEffectNoSchedule,
							},
							{
								Key:    "node-role.kubernetes.io/master",
								Effect: k8s.TaintEffectNoSchedule,
							},
						},
						NodeSelector: map[string]string{
							"node-role.kubernetes.io/master": "",
						},
					},
				},
			},
		},
	}
}

func (c *cloudControllerManagerDeployment) Marshal() ([]byte, error) {
	return MarshalK8SResources(c)
}
