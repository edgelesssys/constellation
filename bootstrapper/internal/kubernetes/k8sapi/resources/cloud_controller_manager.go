package resources

import (
	"fmt"

	"github.com/edgelesssys/constellation/internal/kubernetes"
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

// NewDefaultCloudControllerManagerDeployment creates a new *cloudControllerManagerDeployment, customized for the CSP.
func NewDefaultCloudControllerManagerDeployment(cloudProvider, image, path, podCIDR string, extraArgs []string, extraVolumes []k8s.Volume, extraVolumeMounts []k8s.VolumeMount, env []k8s.EnvVar) *cloudControllerManagerDeployment {
	command := []string{
		path,
		fmt.Sprintf("--cloud-provider=%s", cloudProvider),
		"--leader-elect=true",
		fmt.Sprintf("--cluster-cidr=%s", podCIDR),
		"-v=2",
	}
	command = append(command, extraArgs...)
	volumes := []k8s.Volume{
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
	}
	volumes = append(volumes, extraVolumes...)
	volumeMounts := []k8s.VolumeMount{
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
	}
	volumeMounts = append(volumeMounts, extraVolumeMounts...)

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
								Name:         "cloud-controller-manager",
								Image:        image,
								Command:      command,
								VolumeMounts: volumeMounts,
								Env:          env,
							},
						},
						Volumes: volumes,
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
							{
								Key:      "node-role.kubernetes.io/control-plane",
								Operator: k8s.TolerationOpExists,
								Effect:   k8s.TaintEffectNoSchedule,
							},
							{
								Key:    "node.kubernetes.io/not-ready",
								Effect: k8s.TaintEffectNoSchedule,
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

func (c *cloudControllerManagerDeployment) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(c)
}
