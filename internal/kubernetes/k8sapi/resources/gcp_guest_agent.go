package resources

import (
	"github.com/edgelesssys/constellation/internal/versions"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type gcpGuestAgentDaemonset struct {
	DaemonSet apps.DaemonSet
}

// NewGCPGuestAgentDaemonset creates a new GCP Guest Agent Daemonset.
// It is used automatically to add loadbalancer IPs to the local routing table of GCP instances.
func NewGCPGuestAgentDaemonset() *gcpGuestAgentDaemonset {
	return &gcpGuestAgentDaemonset{
		DaemonSet: apps.DaemonSet{
			TypeMeta: meta.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "gcp-guest-agent",
				Namespace: "kube-system",
				Labels: map[string]string{
					"k8s-app":                       "gcp-guest-agent",
					"component":                     "gcp-guest-agent",
					"kubernetes.io/cluster-service": "true",
				},
			},
			Spec: apps.DaemonSetSpec{
				Selector: &meta.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "gcp-guest-agent",
					},
				},
				Template: k8s.PodTemplateSpec{
					ObjectMeta: meta.ObjectMeta{
						Labels: map[string]string{
							"k8s-app": "gcp-guest-agent",
						},
					},
					Spec: k8s.PodSpec{
						PriorityClassName: "system-cluster-critical",
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
						Containers: []k8s.Container{
							{
								Name:  "gcp-guest-agent",
								Image: versions.GcpGuestImage,
								SecurityContext: &k8s.SecurityContext{
									Privileged: func(b bool) *bool { return &b }(true),
									Capabilities: &k8s.Capabilities{
										Add: []k8s.Capability{"NET_ADMIN"},
									},
								},
								VolumeMounts: []k8s.VolumeMount{
									{
										Name:      "etcssl",
										ReadOnly:  true,
										MountPath: "/etc/ssl",
									},
									{
										Name:      "etcpki",
										ReadOnly:  true,
										MountPath: "/etc/pki",
									},
									{
										Name:      "bin",
										ReadOnly:  true,
										MountPath: "/bin",
									},
									{
										Name:      "usrbin",
										ReadOnly:  true,
										MountPath: "/usr/bin",
									},
									{
										Name:      "usr",
										ReadOnly:  true,
										MountPath: "/usr",
									},
									{
										Name:      "lib",
										ReadOnly:  true,
										MountPath: "/lib",
									},
									{
										Name:      "lib64",
										ReadOnly:  true,
										MountPath: "/lib64",
									},
								},
							},
						},
						Volumes: []k8s.Volume{
							{
								Name: "etcssl",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/etc/ssl",
									},
								},
							},
							{
								Name: "etcpki",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/etc/pki",
									},
								},
							},
							{
								Name: "bin",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/bin",
									},
								},
							},
							{
								Name: "usrbin",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/usr/bin",
									},
								},
							},
							{
								Name: "usr",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/usr",
									},
								},
							},
							{
								Name: "lib",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/lib",
									},
								},
							},
							{
								Name: "lib64",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/lib64",
									},
								},
							},
						},
						HostNetwork: true,
					},
				},
			},
		},
	}
}

// Marshal marshals the access-manager deployment as YAML documents.
func (c *gcpGuestAgentDaemonset) Marshal() ([]byte, error) {
	return MarshalK8SResources(c)
}
