/*
Package constants contains the constants used by Constellation.
Constants should never be overwritable by command line flags or configuration files.
*/
package constants

import "time"

const (
	//
	// Ports.
	//

	CoordinatorPort = 9000
	EnclaveSSHPort  = 2222
	SSHPort         = 22
	WireguardPort   = 51820
	NVMEOverTCPPort = 8009
	// Default NodePort Range
	// https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	NodePortFrom = 30000
	NodePortTo   = 32767

	//
	// Filenames.
	//

	StateFilename         = "constellation-state.json"
	ConfigFilename        = "constellation-conf.yaml"
	AdminConfFilename     = "constellation-admin.conf"
	MasterSecretFilename  = "constellation-mastersecret.base64"
	WGQuickConfigFilename = "wg0.conf"

	//
	// Cryptographic constants.
	//
	StateDiskKeyLength      = 32
	DerivedKeyLengthDefault = 32

	//
	// CLI.
	//

	MinControllerCount = 1
	MinWorkerCount     = 1

	//
	// Kubernetes.
	//

	// KubernetesVersion installed by kubeadm.
	KubernetesVersion      = "stable-1.23"
	KubernetesJoinTokenTTL = 15 * time.Minute
)

// CliVersion is the version of the CLI. Left as a separate variable to allow override during build.
var CliVersion = "0.0.0"
