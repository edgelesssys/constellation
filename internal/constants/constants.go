/*
Package constants contains the constants used by Constellation.
Constants should never be overwritable by command line flags or configuration files.
*/
package constants

const (
	//
	// Ports.
	//

	CoordinatorPort = 9000
	EnclaveSSHPort  = 2222
	SSHPort         = 22
	WireguardPort   = 51820
	NVMEOverTCPPort = 8009

	//
	// Filenames.
	//

	StateFilename         = "constellation-state.json"
	AdminConfFilename     = "constellation-admin.conf"
	MasterSecretFilename  = "constellation-mastersecret.base64"
	WGQuickConfigFilename = "wg0.conf"
)

// CliVersion is the version of the CLI. Left as a separate variable to allow override during build.
var CliVersion = "0.0.0"
