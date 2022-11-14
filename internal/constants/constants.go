/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package constants contains the constants used by Constellation.
Constants should never be overwritable by command line flags or configuration files.
*/
package constants

import (
	"time"
)

const (
	//
	// Constellation.
	//

	// ConstellationNameLength is the maximum length of a Constellation's name.
	ConstellationNameLength = 37
	// ConstellationMasterSecretStoreName is the name for the Constellation secrets in Kubernetes.
	ConstellationMasterSecretStoreName = "constellation-mastersecret"
	// ConstellationMasterSecretKey is the name of the key for the master secret in the master secret kubernetes secret.
	ConstellationMasterSecretKey = "mastersecret"
	// ConstellationSaltKey is the name of the key for the salt in the master secret kubernetes secret.
	ConstellationSaltKey = "salt"

	//
	// Ports.
	//

	// JoinServicePort is the port for reaching the join service within Kubernetes.
	JoinServicePort = 9090
	// JoinServiceNodePort is the port for reaching the join service outside of Kubernetes.
	JoinServiceNodePort = 30090
	// VerifyServicePortHTTP HTTP port for verification service.
	VerifyServicePortHTTP = 8080
	// VerifyServicePortGRPC GRPC port for verification service.
	VerifyServicePortGRPC = 9090
	// VerifyServiceNodePortHTTP HTTP node port for verification service.
	VerifyServiceNodePortHTTP = 30080
	// VerifyServiceNodePortGRPC GRPC node port for verification service.
	VerifyServiceNodePortGRPC = 30081
	// KMSPort is the port the KMS server listens on.
	KMSPort = 9000
	// BootstrapperPort port of bootstrapper.
	BootstrapperPort = 9000
	// KubernetesPort port for Kubernetes API.
	KubernetesPort = 6443
	// RecoveryPort port for Constellation recovery server.
	RecoveryPort = 9999
	// DebugdPort port for debugd process.
	DebugdPort = 4000
	// KonnectivityPort port for konnectivity k8s service.
	KonnectivityPort = 8132
	// NodePortFrom start of range to use for K8s node port
	// https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	NodePortFrom = 30000
	// NodePortTo end of range to use for K8s node port
	// https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	NodePortTo = 32767

	//
	// Filenames.
	//

	// ClusterIDsFileName filename that contains Constellation clusterID and IP.
	ClusterIDsFileName = "constellation-id.json"
	// ConfigFilename filename of Constellation config file.
	ConfigFilename = "constellation-conf.yaml"
	// LicenseFilename filename of Constellation license file.
	LicenseFilename = "constellation.license"
	// AdminConfFilename filename of KubeConfig for admin access to Constellation.
	AdminConfFilename = "constellation-admin.conf"
	// MasterSecretFilename filename of Constellation mastersecret.
	MasterSecretFilename = "constellation-mastersecret.json"
	// TerraformWorkingDir is the directory name for the TerraformClient workspace.
	TerraformWorkingDir = "constellation-terraform"
	// ControlPlaneAdminConfFilename filepath to control plane kubernetes admin config.
	ControlPlaneAdminConfFilename = "/etc/kubernetes/admin.conf"
	// KubectlPath path to kubectl binary.
	KubectlPath = "/run/state/bin/kubectl"

	//
	// Filenames for Constellation's micro services.
	//

	// ServiceBasePath is the base path for the mounted micro service's files.
	ServiceBasePath = "/var/config"
	// MeasurementsFilename is the filename of CC measurements.
	MeasurementsFilename = "measurements"
	// EnforcedPCRsFilename is the filename for a list PCRs that are required to pass attestation.
	EnforcedPCRsFilename = "enforcedPCRs"
	// MeasurementSaltFilename is the filename of the salt used in creation of the clusterID.
	MeasurementSaltFilename = "measurementSalt"
	// MeasurementSecretFilename is the filename of the secret used in creation of the clusterID.
	MeasurementSecretFilename = "measurementSecret"
	// IDKeyDigestFilename is the name of the file holding the currently enforced idkeydigest.
	IDKeyDigestFilename = "idkeydigest"
	// EnforceIDKeyDigestFilename is the name of the file configuring whether idkeydigest is enforced or not.
	EnforceIDKeyDigestFilename = "enforceIdKeyDigest"
	// AzureCVM is the name of the file indicating whether the cluster is expected to run on CVMs or not.
	AzureCVM = "azureCVM"
	// K8sVersion is the filename of the mapped "k8s-version" configMap file.
	K8sVersion = "k8s-version"

	//
	// CLI.
	//

	// MinControllerCount is the minimum number of control nodes.
	MinControllerCount = 1
	// MinWorkerCount is the minimum number of worker nodes.
	MinWorkerCount = 1
	// EnvVarPrefix is expected prefix for environment variables used to overwrite config parameters.
	EnvVarPrefix = "CONSTELL_"
	// EnvVarAzureClientSecretValue is environment variable to overwrite
	// provider.azure.clientSecretValue .
	EnvVarAzureClientSecretValue = EnvVarPrefix + "AZURE_CLIENT_SECRET_VALUE"

	//
	// Kubernetes.
	//

	// KubernetesJoinTokenTTL time to live for Kubernetes join token.
	KubernetesJoinTokenTTL = 15 * time.Minute
	// ConstellationNamespace namespace to deploy Constellation components into.
	ConstellationNamespace = "kube-system"
	// JoinConfigMap k8s config map with node join config.
	JoinConfigMap = "join-config"
	// InternalConfigMap k8s config map with internal Constellation config.
	InternalConfigMap = "internal-config"

	//
	// Helm.
	//

	// HelmNamespace namespace for helm charts.
	HelmNamespace = "kube-system"

	//
	// Releases.
	//

	// S3PublicBucket contains measurements & releases.
	S3PublicBucket = "https://public-edgeless-constellation.s3.us-east-2.amazonaws.com/"
	// CosignPublicKey signs all our releases.
	CosignPublicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEf8F1hpmwE+YCFXzjGtaQcrL6XZVT
JmEe5iSLvG1SyQSAew7WdMKF6o9t8e2TFuCkzlOhhlws2OHWbiFZnFWCFw==
-----END PUBLIC KEY-----
`
)

// VersionInfo is the version of a binary. Left as a separate variable to allow override during build.
var VersionInfo = "0.0.0"
