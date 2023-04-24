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
	// AWSConstellationNameLength is the maximum length of a Constellation's name on AWS.
	AWSConstellationNameLength = 10
	// ConstellationMasterSecretStoreName is the name for the Constellation secrets in Kubernetes.
	ConstellationMasterSecretStoreName = "constellation-mastersecret"
	// ConstellationMasterSecretKey is the name of the key for the master secret in the master secret kubernetes secret.
	ConstellationMasterSecretKey = "mastersecret"
	// ConstellationSaltKey is the name of the key for the salt in the master secret kubernetes secret.
	ConstellationSaltKey = "salt"
	// ConstellationVerifyServiceUserData is the user data that the verification service includes in the attestation.
	ConstellationVerifyServiceUserData = "VerifyService"
	// AttestationVariant is the name of the environment variable that contains the attestation variant.
	AttestationVariant = "CONSTEL_ATTESTATION_VARIANT"

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
	// KeyServicePort is the port the KMS server listens on.
	KeyServicePort = 9000
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
	// TerraformIAMWorkingDir is the directory name for the Terraform IAM Client workspace.
	TerraformIAMWorkingDir = "constellation-iam-terraform"
	// GCPServiceAccountKeyFile is the file name for the GCP service account key file.
	GCPServiceAccountKeyFile = "gcpServiceAccountKey.json"
	// ControlPlaneAdminConfFilename filepath to control plane kubernetes admin config.
	ControlPlaneAdminConfFilename = "/etc/kubernetes/admin.conf"
	// KubectlPath path to kubectl binary.
	KubectlPath = "/run/state/bin/kubectl"
	// UpgradeAgentSocketPath is the path to the UDS that is used for the gRPC connection to the upgrade agent.
	UpgradeAgentSocketPath = "/run/constellation-upgrade-agent.sock"
	// UpgradeAgentMountPath is the path inside the operator container where the UDS is mounted.
	UpgradeAgentMountPath = "/etc/constellation-upgrade-agent.sock"
	// CniPluginsDir path directory for CNI plugins.
	CniPluginsDir = "/opt/cni/bin"
	// BinDir install path for CNI config.
	BinDir = "/run/state/bin"
	// KubeadmPath install path for kubeadm.
	KubeadmPath = "/run/state/bin/kubeadm"
	// KubeletPath install path for kubelet.
	KubeletPath = "/run/state/bin/kubelet"

	//
	// Filenames for Constellation's micro services.
	//

	// ServiceBasePath is the base path for the mounted micro service's files.
	ServiceBasePath = "/var/config"
	// AttestationConfigFilename is the filename of the config used for CC validation.
	AttestationConfigFilename = "attestationConfig"
	// MeasurementSaltFilename is the filename of the salt used in creation of the clusterID.
	MeasurementSaltFilename = "measurementSalt"
	// MeasurementSecretFilename is the filename of the secret used in creation of the clusterID.
	MeasurementSecretFilename = "measurementSecret"

	// K8sVersionFieldName is the name of the of the key holding the wanted Kubernetes version.
	K8sVersionFieldName = "cluster-version"
	// ComponentsListKey is the name of the key holding the list of components in the components configMap.
	ComponentsListKey = "components"
	// NodeVersionResourceName resource name used for NodeVersion in constellation-operator and CLI.
	NodeVersionResourceName = "constellation-version"
	// NodeKubernetesComponentsAnnotationKey is the name of the annotation holding the reference to the ConfigMap listing all K8s components.
	NodeKubernetesComponentsAnnotationKey = "constellation.edgeless.systems/kubernetes-components"
	// JoiningNodesConfigMapName is the name of the configMap holding the joining nodes with the components hashes the node-operator should annotate the nodes with.
	JoiningNodesConfigMapName = "joining-nodes"

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
	// EnvVarOpenStackPassword is environment variable to overwrite
	// provider.openstack.password .
	EnvVarOpenStackPassword = EnvVarPrefix + "OS_PASSWORD"
	// EnvVarNoSpinner is environment variable used to disable the loading indicator (spinner)
	// displayed in Constellation CLI. Any non-empty value, e.g., CONSTELL_NO_SPINNER=1,
	// can be used to disable the spinner.
	EnvVarNoSpinner = EnvVarPrefix + "NO_SPINNER"
	// MiniConstellationUID is a sentinel value for the UID of a mini constellation.
	MiniConstellationUID = "mini"
	// TerraformLogFile is the file name of the Terraform log file.
	TerraformLogFile = "terraform.log"
	// UpgradePlanFile is the file name of the Terraform plan file for Constellation upgrades.
	UpgradePlanFile = "upgrade-plan.json"

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

	// CDNRepositoryURL is the base URL of the Constellation CDN artifact repository.
	CDNRepositoryURL = "https://cdn.confidential.cloud"
	// CDNAPIPrefix is the prefix of the Constellation API.
	CDNAPIPrefix = "constellation/v1"
	// CDNMeasurementsFile is name of file containing image measurements.
	CDNMeasurementsFile = "measurements.json"
	// CDNMeasurementsSignature is name of file containing signature for CDNMeasurementsFile.
	CDNMeasurementsSignature = "measurements.json.sig"
)

// VersionInfo returns the version of a binary.
func VersionInfo() string {
	return versionInfo
}

// Timestamp returns the commit timestamp of a binary.
func Timestamp() string {
	return timestamp
}

// Commit returns the commit hash of a binary.
func Commit() string {
	return commit
}

// State returns the git state of the working directory.
func State() string {
	return state
}

var (
	// versionInfo is the version of a binary. Left as a separate variable to allow override during build.
	versionInfo = "0.0.0"
	// timestamp is the commit timestamp of a binary. Left as a separate variable to allow override during build.
	timestamp = "1970-01-01T00:00:00Z"
	// commit is the commit hash of a binary. Left as a separate variable to allow override during build.
	commit = "0000000000000000000000000000000000000000"
	// state is the git state of the working directory. Left as a separate variable to allow override during build.
	state = "unknown"
)
