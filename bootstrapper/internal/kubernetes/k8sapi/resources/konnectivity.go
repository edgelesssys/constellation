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
	corev1 "k8s.io/api/core/v1"
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

// KonnectivityServerStaticPod deployment.
type KonnectivityServerStaticPod struct {
	StaticPod corev1.Pod
}

// EgressSelectorConfiguration deployment.
type EgressSelectorConfiguration struct {
	EgressSelectorConfiguration apiserver.EgressSelectorConfiguration
}

// NewKonnectivityServerStaticPod create a new KonnectivityServerStaticPod.
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

// NewEgressSelectorConfiguration creates a new EgressSelectorConfiguration.
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

// Marshal to Kubernetes YAML.
func (v *KonnectivityServerStaticPod) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(v)
}

// Marshal to Kubernetes YAML.
func (v *EgressSelectorConfiguration) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(v)
}

// GetKonnectivityCertificateRequest returns a certificate request and matching private key for the konnectivity server.
func GetKonnectivityCertificateRequest() (certificateRequest []byte, privateKey []byte, err error) {
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "system:konnectivity-server",
		},
	}
	return certificate.GetCertificateRequest(csrTemplate)
}
