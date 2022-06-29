package resources

import (
	"fmt"

	"github.com/edgelesssys/constellation/internal/constants"
	"github.com/edgelesssys/constellation/internal/secrets"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type verificationDaemonset struct {
	DaemonSet apps.DaemonSet
	Service   k8s.Service
}

func NewVerificationDaemonSet(csp string) *verificationDaemonset {
	return &verificationDaemonset{
		DaemonSet: apps.DaemonSet{
			TypeMeta: meta.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "verification-service",
				Namespace: "kube-system",
				Labels: map[string]string{
					"k8s-app":   "verification-service",
					"component": "verification-service",
				},
			},
			Spec: apps.DaemonSetSpec{
				Selector: &meta.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "verification-service",
					},
				},
				Template: k8s.PodTemplateSpec{
					ObjectMeta: meta.ObjectMeta{
						Labels: map[string]string{
							"k8s-app": "verification-service",
						},
					},
					Spec: k8s.PodSpec{
						Tolerations: []k8s.Toleration{
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
						ImagePullSecrets: []k8s.LocalObjectReference{
							{
								Name: secrets.PullSecretName,
							},
						},
						Containers: []k8s.Container{
							{
								Name:  "verification-service",
								Image: verificationImage,
								Ports: []k8s.ContainerPort{
									{
										Name:          "http",
										ContainerPort: constants.VerifyServicePortHTTP,
									},
									{
										Name:          "grpc",
										ContainerPort: constants.VerifyServicePortGRPC,
									},
								},
								SecurityContext: &k8s.SecurityContext{
									Privileged: func(b bool) *bool { return &b }(true),
								},
								Args: []string{
									fmt.Sprintf("--cloud-provider=%s", csp),
								},
								VolumeMounts: []k8s.VolumeMount{
									{
										Name:      "event-log",
										ReadOnly:  true,
										MountPath: "/sys/kernel/security/",
									},
								},
							},
						},
						Volumes: []k8s.Volume{
							{
								Name: "event-log",
								VolumeSource: k8s.VolumeSource{
									HostPath: &k8s.HostPathVolumeSource{
										Path: "/sys/kernel/security/",
									},
								},
							},
						},
					},
				},
			},
		},
		Service: k8s.Service{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "verification-service",
				Namespace: "kube-system",
			},
			Spec: k8s.ServiceSpec{
				Type: k8s.ServiceTypeNodePort,
				Ports: []k8s.ServicePort{
					{
						Name:       "http",
						Protocol:   k8s.ProtocolTCP,
						Port:       constants.VerifyServicePortHTTP,
						TargetPort: intstr.FromInt(constants.VerifyServicePortHTTP),
						NodePort:   constants.VerifyServiceNodePortHTTP,
					},
					{
						Name:       "grpc",
						Protocol:   k8s.ProtocolTCP,
						Port:       constants.VerifyServicePortGRPC,
						TargetPort: intstr.FromInt(constants.VerifyServicePortGRPC),
						NodePort:   constants.VerifyServiceNodePortGRPC,
					},
				},
				Selector: map[string]string{
					"k8s-app": "verification-service",
				},
			},
		},
	}
}

func (v *verificationDaemonset) Marshal() ([]byte, error) {
	return MarshalK8SResources(v)
}
