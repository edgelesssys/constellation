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
	"fmt"
	"time"

	"github.com/edgelesssys/constellation/v2/internal/semver"
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
	// DefaultControlPlaneGroupName is the name of the default control plane node group.
	DefaultControlPlaneGroupName = "control_plane_default"
	// DefaultWorkerGroupName is the name of the default worker node group.
	DefaultWorkerGroupName = "worker_default"
	// CLIDebugLogFile is the name of the debug log file for constellation init/constellation apply.
	CLIDebugLogFile = "constellation-debug.log"
	// SSHCAKeySuffix is the suffix used together with the DEKPrefix to derive an SSH CA key for emergency ssh access.
	SSHCAKeySuffix = "ca_emergency_ssh"
	// SSHCAKeyPath is the path to the emergency SSH CA key on the node.
	SSHCAKeyPath = "/var/run/state/ssh/ssh_ca.pub"
	// SSHHostKeyPath is the path to the SSH host key of the node.
	SSHHostKeyPath = "/var/run/state/ssh/ssh_host_ecdsa_key"
	// SSHHostCertificatePath is the path to the SSH host certificate.
	SSHHostCertificatePath = "/var/run/state/ssh/ssh_host_cert.pub"
	// SSHAdditionalPrincipalsPath stores additional principals (like the public IP of the load balancer) that get added to all host certificates.
	SSHAdditionalPrincipalsPath = "/var/run/state/ssh/principals/additional_principals.txt"

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

	//
	// Filenames.
	//

	// StateFilename filename that contains the entire state of the Constellation cluster.
	StateFilename = "constellation-state.yaml"
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
	// GCPServiceAccountKeyFilename is the file name for the GCP service account key file.
	GCPServiceAccountKeyFilename = "gcpServiceAccountKey.json"
	// ErrorLog file which contains server errors during init.
	ErrorLog = "constellation-cluster.log"
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
	// KubeadmPatchDir directory for kubeadm patches .
	KubeadmPatchDir = "/opt/kubernetes/patches"

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
	// SevSnpCertCacheConfigMapName is the name of the configMap holding the SEV-SNP certificate cache in the join service.
	SevSnpCertCacheConfigMapName = "sev-snp-cert-cache"
	// CertCacheAskKey is the name of the key holding the ASK certificate in the SEV-SNP certificate cache.
	CertCacheAskKey = "ask"
	// CertCacheArkKey is the name of the key holding the ARK certificate in the SEV-SNP certificate cache.
	CertCacheArkKey = "ark"
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
	// MiniConstellationName is a sentinel value for the name of a mini constellation.
	MiniConstellationName = MiniConstellationUID + "-qemu"
	// TerraformLogFile is the file name of the Terraform log file.
	TerraformLogFile = "terraform.log"
	// TerraformUpgradeWorkingDir is the directory name for the Terraform workspace being used in an upgrade.
	TerraformUpgradeWorkingDir = "terraform"
	// TerraformIAMUpgradeWorkingDir is the directory name for the Terraform IAM workspace being used in an upgrade.
	TerraformIAMUpgradeWorkingDir = "terraform-iam"
	// TerraformUpgradeBackupDir is the directory name being used to backup the pre-upgrade state in an upgrade.
	TerraformUpgradeBackupDir = "terraform-backup"
	// TerraformIAMUpgradeBackupDir is the directory name being used to backup the pre-upgrade state of iam in an upgrade.
	TerraformIAMUpgradeBackupDir = "terraform-iam-backup"
	// TerraformEmbeddedDir is the name of the base directory embedded in the CLI binary containing the Terraform files.
	TerraformEmbeddedDir = "infrastructure"
	// UpgradeDir is the name of the directory being used for cluster upgrades.
	UpgradeDir = "constellation-upgrade"
	// ControlPlaneDefault is the name of the default control plane worker group.
	ControlPlaneDefault = "control_plane_default"
	// WorkerDefault is the name of the default worker group.
	WorkerDefault = "worker_default"

	//
	// CSP.
	//

	// MarketplaceImageURIScheme is the scheme used for Constellation marketplace OS images.
	MarketplaceImageURIScheme = "constellation-marketplace-image"

	//
	// Azure.
	//

	// AzureMarketplaceImagePublisherKey is the URI key for the Azure Marketplace image publisher.
	AzureMarketplaceImagePublisherKey = "publisher"
	// AzureMarketplaceImageOfferKey is the URI key for the Azure Marketplace image offer.
	AzureMarketplaceImageOfferKey = "offer"
	// AzureMarketplaceImageSkuKey is the URI key for the Azure Marketplace image SKU.
	AzureMarketplaceImageSkuKey = "sku"
	// AzureMarketplaceImageVersionKey is the URI key for the Azure Marketplace image version.
	AzureMarketplaceImageVersionKey = "version"
	// AzureMarketplaceImagePublisher is the publisher of the Azure Marketplace image.
	AzureMarketplaceImagePublisher = "edgelesssystems"
	// AzureMarketplaceImageOffer is the offer of the Azure Marketplace image.
	AzureMarketplaceImageOffer = "constellation"
	// AzureMarketplaceImagePlan is the plan of the Azure Marketplace image.
	AzureMarketplaceImagePlan = "constellation"

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
	// KubeadmConfigMap k8s config map with kubeadm config
	// (holds ClusterConfiguration).
	KubeadmConfigMap = "kubeadm-config"
	// ClusterConfigurationKey key in kubeadm config map with ClusterConfiguration.
	ClusterConfigurationKey = "ClusterConfiguration"

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
	// CDNAPIBase is the (un-versioned) prefix of the Constellation API.
	CDNAPIBase = "constellation"
	// CDNAPIPrefix is the prefix of the Constellation API (V1).
	CDNAPIPrefix = CDNAPIBase + "/v1"
	// CDNAPIPrefixV2 is the prefix of the Constellation API (v2).
	CDNAPIPrefixV2 = CDNAPIBase + "/v2"
	// CDNAttestationConfigPrefixV1 is the prefix of the Constellation AttestationConfig API (v1).
	CDNAttestationConfigPrefixV1 = CDNAPIPrefix + "/attestation"
	// CDNMeasurementsFile is name of file containing image measurements.
	CDNMeasurementsFile = "measurements.json"
	// CDNMeasurementsSignature is name of file containing signature for CDNMeasurementsFile.
	CDNMeasurementsSignature = "measurements.json.sig"
	// CDNDefaultDistributionID is the default CloudFront distribution ID to use.
	CDNDefaultDistributionID = "E1H77EZTHC3NE4"

	//
	// PKI.
	//

	// CosignPublicKeyReleases signs all our releases.
	CosignPublicKeyReleases = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEf8F1hpmwE+YCFXzjGtaQcrL6XZVT
JmEe5iSLvG1SyQSAew7WdMKF6o9t8e2TFuCkzlOhhlws2OHWbiFZnFWCFw==
-----END PUBLIC KEY-----
`
	// CosignPublicKeyDev signs all our development builds.
	CosignPublicKeyDev = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAELcPl4Ik+qZuH4K049wksoXK/Os3Z
b92PDCpM7FZAINQF88s1TZS/HmRXYk62UJ4eqPduvUnJmXhNikhLbMi6fw==
-----END PUBLIC KEY-----
`

	//
	// Terraform Provider.
	//

	// ConstellationClusterURIScheme is the scheme used in Terraform Constellation cluster import URIs.
	ConstellationClusterURIScheme = "constellation-cluster"
	// KubeConfigURIKey is the key used for the KubeConfig in Terraform Constellation cluster import URIs.
	KubeConfigURIKey = "kubeConfig"
	// ClusterEndpointURIKey is the key used for the cluster endpoint in Terraform Constellation cluster import URIs.
	ClusterEndpointURIKey = "clusterEndpoint"
	// MasterSecretURIKey is the key used for the master secret in Terraform Constellation cluster import URIs.
	MasterSecretURIKey = "masterSecret"
	// MasterSecretSaltURIKey is the key used for the master secret salt in Terraform Constellation cluster import URIs.
	MasterSecretSaltURIKey = "masterSecretSalt"
)

// BinaryVersion returns the version of this Binary.
func BinaryVersion() semver.Semver {
	version, err := semver.New(versionInfo)
	if err != nil {
		// This is not user input, unrecoverable, should never happen.
		panic(fmt.Sprintf("parsing embedded version information: %s", err))
	}

	return version
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
