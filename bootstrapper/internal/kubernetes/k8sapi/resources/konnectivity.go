/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package resources

import (
	"crypto/x509"
	"crypto/x509/pkix"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/certificate"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"github.com/edgelesssys/constellation/v2/internal/versions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apiserver/pkg/apis/apiserver"
)

const (
	// KonnectivityCertificateFilename is the path to the kubelets certificate.
	KonnectivityCertificateFilename = "/etc/kubernetes/konnectivity.crt"
	// KonnectivityKeyFilename is the path to the kubelets private key.
	KonnectivityKeyFilename = "/etc/kubernetes/konnectivity.key"
)

type KonnectivityAgents struct {
	DaemonSet          appsv1.DaemonSet
	ClusterRoleBinding rbacv1.ClusterRoleBinding
	ServiceAccount     corev1.ServiceAccount
}

type KonnectivityServerStaticPod struct {
	StaticPod corev1.Pod
}

type EgressSelectorConfiguration struct {
	EgressSelectorConfiguration apiserver.EgressSelectorConfiguration
}

func NewKonnectivityAgents(konnectivityServerAddress string) *KonnectivityAgents {
	return &KonnectivityAgents{
		DaemonSet: appsv1.DaemonSet{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "konnectivity-agent",
				Namespace: "kube-system",
				Labels: map[string]string{
					"k8s-app":                         "konnectivity-agent",
					"addonmanager.kubernetes.io/mode": "Reconcile",
				},
			},
			Spec: appsv1.DaemonSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "konnectivity-agent",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"k8s-app": "konnectivity-agent",
						},
					},
					Spec: corev1.PodSpec{
						PriorityClassName: "system-cluster-critical",
						Tolerations: []corev1.Toleration{
							{
								Key:      "node-role.kubernetes.io/master",
								Operator: corev1.TolerationOpExists,
								Effect:   corev1.TaintEffectNoSchedule,
							},
							{
								Key:      "node-role.kubernetes.io/control-plane",
								Operator: corev1.TolerationOpExists,
								Effect:   corev1.TaintEffectNoSchedule,
							},
							{
								Key:      "CriticalAddonsOnly",
								Operator: corev1.TolerationOpExists,
							},
							{
								Key:      "node.kubernetes.io/not-ready",
								Operator: corev1.TolerationOpExists,
								Effect:   corev1.TaintEffectNoExecute,
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "konnectivity-agent",
								Image: versions.KonnectivityAgentImage,
								Command: []string{
									"/proxy-agent",
								},
								Args: []string{
									"--logtostderr=true",
									"--proxy-server-host=" + konnectivityServerAddress,
									"--ca-cert=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
									"--proxy-server-port=8132",
									"--admin-server-port=8133",
									"--health-server-port=8134",
									"--service-account-token-path=/var/run/secrets/tokens/konnectivity-agent-token",
									"--agent-identifiers=host=$(HOST_IP)",
									// we will be able to avoid constant polling when either one is done:
									// https://github.com/kubernetes-sigs/apiserver-network-proxy/issues/358
									// https://github.com/kubernetes-sigs/apiserver-network-proxy/issues/273
									"--sync-forever=true",
									// Ensure stable connection to the konnectivity server.
									"--keepalive-time=60m",
									"--sync-interval=5s",
									"--sync-interval-cap=30s",
									"--probe-interval=5s",
									"--v=3",
								},
								Env: []corev1.EnvVar{
									{
										Name: "HOST_IP",
										ValueFrom: &corev1.EnvVarSource{
											FieldRef: &corev1.ObjectFieldSelector{
												APIVersion: "v1",
												FieldPath:  "status.hostIP",
											},
										},
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "konnectivity-agent-token",
										MountPath: "/var/run/secrets/tokens",
										ReadOnly:  true,
									},
								},
								LivenessProbe: &corev1.Probe{
									ProbeHandler: corev1.ProbeHandler{
										HTTPGet: &corev1.HTTPGetAction{
											Path: "/healthz",
											Port: intstr.FromInt(8134),
										},
									},
									InitialDelaySeconds: 15,
									TimeoutSeconds:      15,
								},
							},
						},
						ServiceAccountName: "konnectivity-agent",
						Volumes: []corev1.Volume{
							{
								Name: "konnectivity-agent-token",
								VolumeSource: corev1.VolumeSource{
									Projected: &corev1.ProjectedVolumeSource{
										Sources: []corev1.VolumeProjection{
											{
												ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
													Audience: "system:konnectivity-server",
													Path:     "konnectivity-agent-token",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		ClusterRoleBinding: rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:konnectivity-server",
				Labels: map[string]string{
					"kubernetes.io/cluster-service":   "true",
					"addonmanager.kubernetes.io/mode": "Reconcile",
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "system:auth-delegator",
			},
			Subjects: []rbacv1.Subject{
				{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "User",
					Name:     "system:konnectivity-server",
				},
			},
		},
		ServiceAccount: corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "konnectivity-agent",
				Namespace: "kube-system",
				Labels: map[string]string{
					"kubernetes.io/cluster-service":   "true",
					"addonmanager.kubernetes.io/mode": "Reconcile",
				},
			},
		},
	}
}

func NewKonnectivityServerStaticPod() *KonnectivityServerStaticPod {
	udsHostPathType := corev1.HostPathDirectoryOrCreate
	return &KonnectivityServerStaticPod{
		StaticPod: corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "konnectivity-server",
				Namespace: "kube-system",
			},
			Spec: corev1.PodSpec{
				PriorityClassName: "system-cluster-critical",
				HostNetwork:       true,
				Containers: []corev1.Container{
					{
						Name:    "konnectivity-server-container",
						Image:   versions.KonnectivityServerImage,
						Command: []string{"/proxy-server"},
						Args: []string{
							"--logtostderr=true",
							// This needs to be consistent with the value set in egressSelectorConfiguration.
							"--uds-name=/etc/kubernetes/konnectivity-server/konnectivity-server.socket",
							// The following two lines assume the Konnectivity server is
							// deployed on the same machine as the apiserver, and the certs and
							// key of the API Server are at the specified location.
							"--cluster-cert=/etc/kubernetes/pki/apiserver.crt",
							"--cluster-key=/etc/kubernetes/pki/apiserver.key",
							// This needs to be consistent with the value set in egressSelectorConfiguration.
							"--mode=grpc",
							"--server-port=0",
							"--agent-port=8132",
							"--admin-port=8133",
							"--health-port=8134",
							"--v=5",
							"--agent-namespace=kube-system",
							"--agent-service-account=konnectivity-agent",
							"--kubeconfig=/etc/kubernetes/konnectivity-server.conf",
							"--authentication-audience=system:konnectivity-server",
							"--proxy-strategies=default",
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz",
									Port: intstr.FromInt(8134),
								},
							},
							InitialDelaySeconds: 30,
							TimeoutSeconds:      60,
						},
						Ports: []corev1.ContainerPort{
							{
								Name:          "agent-port",
								ContainerPort: 8132,
								HostPort:      8132,
							},
							{
								Name:          "admin-port",
								ContainerPort: 8133,
								HostPort:      8133,
							},
							{
								Name:          "health-port",
								ContainerPort: 8134,
								HostPort:      8134,
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "k8s-certs",
								MountPath: "/etc/kubernetes/pki",
								ReadOnly:  true,
							},
							{
								Name:      "kubeconfig",
								MountPath: "/etc/kubernetes/konnectivity-server.conf",
								ReadOnly:  true,
							},
							{
								Name:      "konnectivity-uds",
								MountPath: "/etc/kubernetes/konnectivity-server",
								ReadOnly:  false,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "k8s-certs",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/etc/kubernetes/pki",
							},
						},
					},
					{
						Name: "kubeconfig",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/etc/kubernetes/konnectivity-server.conf",
							},
						},
					},
					{
						Name: "konnectivity-uds",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/etc/kubernetes/konnectivity-server",
								Type: &udsHostPathType,
							},
						},
					},
				},
			},
		},
	}
}

func NewEgressSelectorConfiguration() *EgressSelectorConfiguration {
	return &EgressSelectorConfiguration{
		EgressSelectorConfiguration: apiserver.EgressSelectorConfiguration{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "apiserver.k8s.io/v1beta1",
				Kind:       "EgressSelectorConfiguration",
			},
			EgressSelections: []apiserver.EgressSelection{
				{
					Name: "cluster",
					Connection: apiserver.Connection{
						ProxyProtocol: "GRPC",
						Transport: &apiserver.Transport{
							UDS: &apiserver.UDSTransport{
								UDSName: "/etc/kubernetes/konnectivity-server/konnectivity-server.socket",
							},
						},
					},
				},
			},
		},
	}
}

func (v *KonnectivityAgents) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(v)
}

func (v *KonnectivityServerStaticPod) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(v)
}

func (v *EgressSelectorConfiguration) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(v)
}

// GetCertificateRequest returns a certificate request and matching private key for the konnectivity server.
func GetKonnectivityCertificateRequest() (certificateRequest []byte, privateKey []byte, err error) {
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "system:konnectivity-server",
		},
	}
	return certificate.GetCertificateRequest(csrTemplate)
}
