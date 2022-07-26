/*
Package constants contains the constants used by Constellation.
Constants should never be overwritable by command line flags or configuration files.
*/
package constants

import "time"

const (
	//
	// Constellation.
	//

	// ConstellationNameLength is the maximum length of a Constellation's name.
	ConstellationNameLength = 37
	// ConstellationMasterSecretStoreName is the name for the Constellation secrets in Kubernetes.
	ConstellationMasterSecretStoreName = "constellation-mastersecret"
	// ConstellationMasterSecretKey is the name of the key for master secret in the master secret store secret.
	ConstellationMasterSecretKey = "mastersecret"

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
	EnclaveSSHPort   = 2222
	SSHPort          = 22
	NVMEOverTCPPort  = 8009
	// Default NodePort Range
	// https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	NodePortFrom   = 30000
	NodePortTo     = 32767
	KubernetesPort = 6443

	//
	// Filenames.
	//
	StateFilename           = "constellation-state.json"
	ClusterIDsFileName      = "constellation-id.json"
	ConfigFilename          = "constellation-conf.yaml"
	DebugdConfigFilename    = "cdbg-conf.yaml"
	AdminConfFilename       = "constellation-admin.conf"
	MasterSecretFilename    = "constellation-mastersecret.base64"
	WGQuickConfigFilename   = "wg0.conf"
	CoreOSAdminConfFilename = "/etc/kubernetes/admin.conf"
	KubeadmCertificateDir   = "/etc/kubernetes/pki"

	//
	// Filenames for Constellation's micro services.
	//

	// ServiceBasePath is the base path for the mounted micro service's files.
	ServiceBasePath = "/var/config"
	// MeasurementsFilename is the filename of CC measurements.
	MeasurementsFilename = "measurements"
	// MeasurementSaltFilename is the filename of the salt used in creation of the clusterID.
	MeasurementSaltFilename = "measurementSalt"
	// MeasurementSecretFilename is the filename of the secret used in creation of the clusterID.
	MeasurementSecretFilename = "measurementSecret"
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
)

// VersionInfo is the version of a binary. Left as a separate variable to allow override during build.
var VersionInfo = "0.0.0"
