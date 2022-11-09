/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package resources

import (
	"fmt"
	"net"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	"google.golang.org/protobuf/proto"
	apps "k8s.io/api/apps/v1"
	k8s "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// VerificationDaemonset groups all k8s resources for the verification service deployment.
type VerificationDaemonset struct {
	DaemonSet    apps.DaemonSet
	Service      k8s.Service
	LoadBalancer k8s.Service
}

// NewVerificationDaemonSet creates a new VerificationDaemonset.
func NewVerificationDaemonSet(csp, loadBalancerIP string) *VerificationDaemonset {
	var err error
	if strings.Contains(loadBalancerIP, ":") {
		loadBalancerIP, _, err = net.SplitHostPort(loadBalancerIP)
		if err != nil {
			panic(err)
		}
	}
	return &VerificationDaemonset{
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
						Containers: []k8s.Container{
							{
								Name:  "verification-service",
								Image: versions.VerificationImage,
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
		LoadBalancer: k8s.Service{
			TypeMeta: meta.TypeMeta{
				APIVersion: "v1",
				Kind:       "Service",
			},
			ObjectMeta: meta.ObjectMeta{
				Name:      "verify",
				Namespace: "kube-system",
			},
			Spec: k8s.ServiceSpec{
				AllocateLoadBalancerNodePorts: proto.Bool(false),
				Type:                          k8s.ServiceTypeLoadBalancer,
				LoadBalancerClass:             proto.String("constellation"),
				ExternalIPs:                   []string{loadBalancerIP},
				Ports: []k8s.ServicePort{
					{
						Name:       "grpc",
						Protocol:   k8s.ProtocolTCP,
						Port:       constants.VerifyServiceNodePortGRPC,
						TargetPort: intstr.FromInt(constants.VerifyServicePortGRPC),
					},
				},
				Selector: map[string]string{
					"k8s-app": "verification-service",
				},
			},
		},
	}
}

// Marshal to Kubernetes YAML.
func (v *VerificationDaemonset) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(v)
}
