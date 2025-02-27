/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package k8sapi

import (
	"path/filepath"

	"github.com/edgelesssys/constellation/v2/bootstrapper/internal/certificate"
	"github.com/edgelesssys/constellation/v2/internal/constants"
	"github.com/edgelesssys/constellation/v2/internal/kubernetes"
	"golang.org/x/mod/semver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletconf "k8s.io/kubelet/config/v1beta1"
	kubeadm "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3"
	kubeconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
)

// Uses types defined here: https://kubernetes.io/docs/reference/config-api/kubeadm-config.v1beta3/
// Slimmed down to the fields we require

const (
	auditLogDir     = "/var/log/kubernetes/audit/"
	auditLogFile    = "audit.log"
	auditPolicyPath = "/etc/kubernetes/audit-policy.yaml"
)

// KubdeadmConfiguration is used to generate kubeadm configurations.
type KubdeadmConfiguration struct{}

// InitConfiguration returns a new init configuration.
func (c *KubdeadmConfiguration) InitConfiguration(externalCloudProvider bool, clusterVersion string) KubeadmInitYAML {
	var cloudProvider string
	if externalCloudProvider {
		cloudProvider = "external"
	}

	initConfig := KubeadmInitYAML{
		InitConfiguration: kubeadm.InitConfiguration{
			TypeMeta: metav1.TypeMeta{
				APIVersion: kubeadm.SchemeGroupVersion.String(),
				Kind:       "InitConfiguration",
			},
			NodeRegistration: kubeadm.NodeRegistrationOptions{
				CRISocket: "unix:///run/containerd/containerd.sock",
				KubeletExtraArgs: map[string]string{
					"cloud-provider": cloudProvider,
				},
			},
			// AdvertiseAddress will be overwritten later
			LocalAPIEndpoint: kubeadm.APIEndpoint{
				BindPort: constants.KubernetesPort,
			},
			Patches: &kubeadm.Patches{Directory: constants.KubeadmPatchDir},
		},
		// https://pkg.go.dev/k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta3#ClusterConfiguration
		ClusterConfiguration: kubeadm.ClusterConfiguration{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ClusterConfiguration",
				APIVersion: kubeadm.SchemeGroupVersion.String(),
			},
			// Target kubernetes version of the control plane.
			KubernetesVersion: clusterVersion,
			// necessary to be able to access the kubeapi server through localhost
			APIServer: kubeadm.APIServer{
				ControlPlaneComponent: kubeadm.ControlPlaneComponent{
					ExtraArgs: map[string]string{
						"audit-policy-file": auditPolicyPath,
						"audit-log-path":    filepath.Join(auditLogDir, auditLogFile), // CIS benchmark
						"audit-log-maxage":  "30",                                     // CIS benchmark - Default value of Rancher
						// log size = 10 files * 100MB + 100 MB (which is currently being written) = 1.1GB
						"audit-log-maxbackup": "10",    // CIS benchmark - Default value of Rancher
						"audit-log-maxsize":   "100",   // CIS benchmark - Default value of Rancher
						"profiling":           "false", // CIS benchmark
						"kubelet-certificate-authority": filepath.Join(
							kubeconstants.KubernetesDir,
							kubeconstants.DefaultCertificateDir,
							kubeconstants.CACertName,
						),
						"tls-cipher-suites": "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256," +
							"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256," +
							"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384," +
							"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256," +
							"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256," +
							"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305," +
							"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,TLS_RSA_WITH_3DES_EDE_CBC_SHA,TLS_RSA_WITH_AES_128_CBC_SHA," +
							"TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_256_GCM_SHA384", // CIS benchmark
					},
					ExtraVolumes: []kubeadm.HostPathMount{
						{
							Name:      "audit-log",
							HostPath:  auditLogDir,
							MountPath: auditLogDir,
							ReadOnly:  false,
							PathType:  corev1.HostPathDirectoryOrCreate,
						},
						{
							Name:      "audit",
							HostPath:  auditPolicyPath,
							MountPath: auditPolicyPath,
							ReadOnly:  true,
							PathType:  corev1.HostPathFile,
						},
					},
				},
				CertSANs: []string{"127.0.0.1"},
			},
			ControllerManager: kubeadm.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"flex-volume-plugin-dir":      "/opt/libexec/kubernetes/kubelet-plugins/volume/exec/",
					"cloud-provider":              cloudProvider,
					"configure-cloud-routes":      "false",
					"profiling":                   "false", // CIS benchmark
					"terminated-pod-gc-threshold": "1000",  // CIS benchmark - Default value of Rancher
				},
			},
			Scheduler: kubeadm.ControlPlaneComponent{
				ExtraArgs: map[string]string{
					"profiling": "false",
				},
			},
		},
		// warning: this config is applied to every node in the cluster!
		KubeletConfiguration: kubeletconf.KubeletConfiguration{
			ProtectKernelDefaults: true, // CIS benchmark
			TLSCipherSuites: []string{
				"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
				"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
				"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
				"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
				"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
				"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
				"TLS_RSA_WITH_AES_256_GCM_SHA384",
				"TLS_RSA_WITH_AES_128_GCM_SHA256",
			}, // CIS benchmark
			StaticPodPath: "/etc/kubernetes/manifests",
			TypeMeta: metav1.TypeMeta{
				APIVersion: kubeletconf.SchemeGroupVersion.String(),
				Kind:       "KubeletConfiguration",
			},
			RegisterWithTaints: []corev1.Taint{
				{
					Key:    "node.cloudprovider.kubernetes.io/uninitialized",
					Value:  "true",
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
				{
					Key:    "node.cilium.io/agent-not-ready",
					Value:  "true",
					Effect: corev1.TaintEffectPreferNoSchedule,
				},
			},
			TLSCertFile:       certificate.CertificateFilename,
			TLSPrivateKeyFile: certificate.KeyFilename,
		},
	}

	if semver.Compare(clusterVersion, "v1.31.0") >= 0 {
		initConfig.ClusterConfiguration.FeatureGates = map[string]bool{"ControlPlaneKubeletLocalMode": true}
	}
	return initConfig
}

// JoinConfiguration returns a new kubeadm join configuration.
func (c *KubdeadmConfiguration) JoinConfiguration(externalCloudProvider bool) KubeadmJoinYAML {
	var cloudProvider string
	if externalCloudProvider {
		cloudProvider = "external"
	}
	return KubeadmJoinYAML{
		JoinConfiguration: kubeadm.JoinConfiguration{
			TypeMeta: metav1.TypeMeta{
				APIVersion: kubeadm.SchemeGroupVersion.String(),
				Kind:       "JoinConfiguration",
			},
			NodeRegistration: kubeadm.NodeRegistrationOptions{
				CRISocket: "unix:///run/containerd/containerd.sock",
				KubeletExtraArgs: map[string]string{
					"cloud-provider": cloudProvider,
				},
			},
			Discovery: kubeadm.Discovery{
				BootstrapToken: &kubeadm.BootstrapTokenDiscovery{},
			},
			Patches: &kubeadm.Patches{Directory: constants.KubeadmPatchDir},
		},
		KubeletConfiguration: kubeletconf.KubeletConfiguration{
			TypeMeta: metav1.TypeMeta{
				APIVersion: kubeletconf.SchemeGroupVersion.String(),
				Kind:       "KubeletConfiguration",
			},
		},
	}
}

// KubeadmJoinYAML holds configuration for kubeadm join workflow.
type KubeadmJoinYAML struct {
	JoinConfiguration    kubeadm.JoinConfiguration
	KubeletConfiguration kubeletconf.KubeletConfiguration
}

// SetNodeName sets the node name.
func (k *KubeadmJoinYAML) SetNodeName(nodeName string) {
	k.JoinConfiguration.NodeRegistration.Name = nodeName
}

// SetAPIServerEndpoint sets the api server endpoint.
func (k *KubeadmJoinYAML) SetAPIServerEndpoint(apiServerEndpoint string) {
	k.JoinConfiguration.Discovery.BootstrapToken.APIServerEndpoint = apiServerEndpoint
}

// SetToken sets the boostrap token.
func (k *KubeadmJoinYAML) SetToken(token string) {
	k.JoinConfiguration.Discovery.BootstrapToken.Token = token
}

// AppendDiscoveryTokenCaCertHash appends another trusted discovery token CA hash.
func (k *KubeadmJoinYAML) AppendDiscoveryTokenCaCertHash(discoveryTokenCaCertHash string) {
	k.JoinConfiguration.Discovery.BootstrapToken.CACertHashes = append(k.JoinConfiguration.Discovery.BootstrapToken.CACertHashes, discoveryTokenCaCertHash)
}

// SetNodeIP sets the node IP.
func (k *KubeadmJoinYAML) SetNodeIP(nodeIP string) {
	if k.JoinConfiguration.NodeRegistration.KubeletExtraArgs == nil {
		k.JoinConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{"node-ip": nodeIP}
	} else {
		k.JoinConfiguration.NodeRegistration.KubeletExtraArgs["node-ip"] = nodeIP
	}
}

// SetProviderID sets the provider ID.
func (k *KubeadmJoinYAML) SetProviderID(providerID string) {
	k.KubeletConfiguration.ProviderID = providerID
}

// SetControlPlane sets the control plane with the advertised address.
func (k *KubeadmJoinYAML) SetControlPlane(advertiseAddress string) {
	k.JoinConfiguration.ControlPlane = &kubeadm.JoinControlPlane{
		LocalAPIEndpoint: kubeadm.APIEndpoint{
			AdvertiseAddress: advertiseAddress,
			BindPort:         constants.KubernetesPort,
		},
	}
	k.JoinConfiguration.SkipPhases = []string{"control-plane-prepare/download-certs"}
}

// Marshal into a k8s resource YAML.
func (k *KubeadmJoinYAML) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(k)
}

// KubeadmInitYAML holds configuration for kubeadm init workflow.
type KubeadmInitYAML struct {
	InitConfiguration    kubeadm.InitConfiguration
	ClusterConfiguration kubeadm.ClusterConfiguration
	KubeletConfiguration kubeletconf.KubeletConfiguration
}

// SetNodeName sets name of node.
func (k *KubeadmInitYAML) SetNodeName(nodeName string) {
	k.InitConfiguration.NodeRegistration.Name = nodeName
}

// SetClusterName sets the name of the Kubernetes cluster.
// This name is reflected in the kubeconfig file and in the name of the default admin user.
func (k *KubeadmInitYAML) SetClusterName(clusterName string) {
	k.ClusterConfiguration.ClusterName = clusterName
}

// SetCertSANs sets the SANs for the certificate.
func (k *KubeadmInitYAML) SetCertSANs(certSANs []string) {
	for _, certSAN := range certSANs {
		if certSAN == "" {
			continue
		}
		k.ClusterConfiguration.APIServer.CertSANs = append(k.ClusterConfiguration.APIServer.CertSANs, certSAN)
	}
}

// SetControlPlaneEndpoint sets the control plane endpoint if controlPlaneEndpoint is not empty.
func (k *KubeadmInitYAML) SetControlPlaneEndpoint(controlPlaneEndpoint string) {
	if controlPlaneEndpoint != "" {
		k.ClusterConfiguration.ControlPlaneEndpoint = controlPlaneEndpoint
	}
}

// SetNodeIP sets the node IP.
func (k *KubeadmInitYAML) SetNodeIP(nodeIP string) {
	if k.InitConfiguration.NodeRegistration.KubeletExtraArgs == nil {
		k.InitConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{"node-ip": nodeIP}
	} else {
		k.InitConfiguration.NodeRegistration.KubeletExtraArgs["node-ip"] = nodeIP
	}
}

// SetProviderID sets the provider ID.
func (k *KubeadmInitYAML) SetProviderID(providerID string) {
	if k.InitConfiguration.NodeRegistration.KubeletExtraArgs == nil {
		k.InitConfiguration.NodeRegistration.KubeletExtraArgs = map[string]string{"provider-id": providerID}
	} else {
		k.InitConfiguration.NodeRegistration.KubeletExtraArgs["provider-id"] = providerID
	}
}

// SetServiceSubnet sets the service subnet.
func (k *KubeadmInitYAML) SetServiceSubnet(subnet string) {
	if subnet != "" {
		k.ClusterConfiguration.Networking.ServiceSubnet = subnet
	}
}

// Marshal into a k8s resource YAML.
func (k *KubeadmInitYAML) Marshal() ([]byte, error) {
	return kubernetes.MarshalK8SResources(k)
}
