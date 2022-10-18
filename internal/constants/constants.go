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
	JoinServiceNodePort       = 30090
	VerifyServicePortHTTP     = 8080
	VerifyServicePortGRPC     = 9090
	VerifyServiceNodePortHTTP = 30080
	VerifyServiceNodePortGRPC = 30081
	// KMSPort is the port the KMS server listens on.
	KMSPort          = 9000
	BootstrapperPort = 9000
	KubernetesPort   = 6443
	RecoveryPort     = 9999
	EnclaveSSHPort   = 2222
	SSHPort          = 22
	NVMEOverTCPPort  = 8009
	DebugdPort       = 4000
	KonnectivityPort = 8132
	// Default NodePort Range
	// https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	NodePortFrom = 30000
	NodePortTo   = 32767

	//
	// Filenames.
	//
	ClusterIDsFileName      = "constellation-id.json"
	ConfigFilename          = "constellation-conf.yaml"
	LicenseFilename         = "constellation.license"
	DebugdConfigFilename    = "cdbg-conf.yaml"
	AdminConfFilename       = "constellation-admin.conf"
	MasterSecretFilename    = "constellation-mastersecret.json"
	WGQuickConfigFilename   = "wg0.conf"
	CoreOSAdminConfFilename = "/etc/kubernetes/admin.conf"
	KubeadmCertificateDir   = "/etc/kubernetes/pki"
	KubectlPath             = "/run/state/bin/kubectl"

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

	MinControllerCount = 1
	MinWorkerCount     = 1

	//
	// Kubernetes.
	//

	KubernetesJoinTokenTTL = 15 * time.Minute
	ConstellationNamespace = "kube-system"
	JoinConfigMap          = "join-config"
	InternalConfigMap      = "internal-config"

	//
	// Helm.
	//

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
