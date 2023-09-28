/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package verify provides the the types for the types of the verify report in JSON format.

At the moment the package is concerned with providing an interface for constellation verify and the attestationconfigapi upload tool through JSON serialization.
*/
package verify

import (
	"crypto/x509"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// Report contains the entire data reported by constellation verify.
type Report struct {
	SNPReport SNPReport      `json:"snp_report"`
	VCEK      []Certificate  `json:"vcek"`
	CertChain []Certificate  `json:"cert_chain"`
	MAAToken  MaaTokenClaims `json:"maa_token"`
}

// Certificate contains the certificate data and additional information.
type Certificate struct {
	*x509.Certificate `json:"certificate"`
	CertTypeName      string     `json:"cert_type_name"`
	StructVersion     uint8      `json:"struct_version"`
	ProductName       string     `json:"product_name"`
	HardwareID        []byte     `json:"hardware_id"`
	TCBVersion        TCBVersion `json:"tcb_version"`
}

// TCBVersion contains the TCB version data.
type TCBVersion struct {
	Bootloader uint8
	TEE        uint8
	SNP        uint8
	Microcode  uint8
	Spl4       uint8
	Spl5       uint8
	Spl6       uint8
	Spl7       uint8
	UcodeSpl   uint8 // UcodeSpl is the microcode security patch level.
}

// PlatformInfo contains the platform information.
type PlatformInfo struct {
	SMT  bool `json:"smt"`
	TSME bool `json:"tsme"`
}

// SignerInfo contains the signer information.
type SignerInfo struct {
	AuthorKeyEn bool         `json:"author_key_en"`
	MaskChipKey bool         `json:"mask_chip_key"`
	SigningKey  fmt.Stringer `json:"signing_key"`
}

// SNPReport contains the SNP report data.
type SNPReport struct {
	Version              uint32       `json:"version"`
	GuestSvn             uint32       `json:"guest_svn"`
	PolicyABIMinor       uint8        `json:"policy_abi_minor"`
	PolicyABIMajor       uint8        `json:"policy_abi_major"`
	PolicySMT            bool         `json:"policy_symmetric_multi_threading"`
	PolicyMigrationAgent bool         `json:"policy_migration_agent"`
	PolicyDebug          bool         `json:"policy_debug"`
	PolicySingleSocket   bool         `json:"policy_single_socket"`
	FamilyID             []byte       `json:"family_id"`
	ImageID              []byte       `json:"image_id"`
	Vmpl                 uint32       `json:"vmpl"`
	SignatureAlgo        uint32       `json:"signature_algo"`
	CurrentTCB           TCBVersion   `json:"current_tcb"`
	PlatformInfo         PlatformInfo `json:"platform_info"`
	SignerInfo           SignerInfo   `json:"signer_info"`
	ReportData           []byte       `json:"report_data"`
	Measurement          []byte       `json:"measurement"`
	HostData             []byte       `json:"host_data"`
	IDKeyDigest          []byte       `json:"id_key_digest"`
	AuthorKeyDigest      []byte       `json:"author_key_digest"`
	ReportID             []byte       `json:"report_id"`
	ReportIDMa           []byte       `json:"report_id_ma"`
	ReportedTCB          TCBVersion   `json:"reported_tcb"`
	ChipID               []byte       `json:"chip_id"`
	CommittedTCB         TCBVersion   `json:"committed_tcb"`
	CurrentBuild         uint32       `json:"current_build"`
	CurrentMinor         uint32       `json:"current_minor"`
	CurrentMajor         uint32       `json:"current_major"`
	CommittedBuild       uint32       `json:"committed_build"`
	CommittedMinor       uint32       `json:"committed_minor"`
	CommittedMajor       uint32       `json:"committed_major"`
	LaunchTCB            TCBVersion   `json:"launch_tcb"`
	Signature            []byte       `json:"signature"`
}

// MaaTokenClaims contains the MAA token claims.
type MaaTokenClaims struct {
	jwt.RegisteredClaims
	Secureboot                               bool   `json:"secureboot,omitempty"`
	XMsAttestationType                       string `json:"x-ms-attestation-type,omitempty"`
	XMsAzurevmAttestationProtocolVer         string `json:"x-ms-azurevm-attestation-protocol-ver,omitempty"`
	XMsAzurevmAttestedPcrs                   []int  `json:"x-ms-azurevm-attested-pcrs,omitempty"`
	XMsAzurevmBootdebugEnabled               bool   `json:"x-ms-azurevm-bootdebug-enabled,omitempty"`
	XMsAzurevmDbvalidated                    bool   `json:"x-ms-azurevm-dbvalidated,omitempty"`
	XMsAzurevmDbxvalidated                   bool   `json:"x-ms-azurevm-dbxvalidated,omitempty"`
	XMsAzurevmDebuggersdisabled              bool   `json:"x-ms-azurevm-debuggersdisabled,omitempty"`
	XMsAzurevmDefaultSecurebootkeysvalidated bool   `json:"x-ms-azurevm-default-securebootkeysvalidated,omitempty"`
	XMsAzurevmElamEnabled                    bool   `json:"x-ms-azurevm-elam-enabled,omitempty"`
	XMsAzurevmFlightsigningEnabled           bool   `json:"x-ms-azurevm-flightsigning-enabled,omitempty"`
	XMsAzurevmHvciPolicy                     int    `json:"x-ms-azurevm-hvci-policy,omitempty"`
	XMsAzurevmHypervisordebugEnabled         bool   `json:"x-ms-azurevm-hypervisordebug-enabled,omitempty"`
	XMsAzurevmIsWindows                      bool   `json:"x-ms-azurevm-is-windows,omitempty"`
	XMsAzurevmKerneldebugEnabled             bool   `json:"x-ms-azurevm-kerneldebug-enabled,omitempty"`
	XMsAzurevmOsbuild                        string `json:"x-ms-azurevm-osbuild,omitempty"`
	XMsAzurevmOsdistro                       string `json:"x-ms-azurevm-osdistro,omitempty"`
	XMsAzurevmOstype                         string `json:"x-ms-azurevm-ostype,omitempty"`
	XMsAzurevmOsversionMajor                 int    `json:"x-ms-azurevm-osversion-major,omitempty"`
	XMsAzurevmOsversionMinor                 int    `json:"x-ms-azurevm-osversion-minor,omitempty"`
	XMsAzurevmSigningdisabled                bool   `json:"x-ms-azurevm-signingdisabled,omitempty"`
	XMsAzurevmTestsigningEnabled             bool   `json:"x-ms-azurevm-testsigning-enabled,omitempty"`
	XMsAzurevmVmid                           string `json:"x-ms-azurevm-vmid,omitempty"`
	XMsIsolationTee                          struct {
		XMsAttestationType  string `json:"x-ms-attestation-type,omitempty"`
		XMsComplianceStatus string `json:"x-ms-compliance-status,omitempty"`
		XMsRuntime          struct {
			Keys []struct {
				E      string   `json:"e,omitempty"`
				KeyOps []string `json:"key_ops,omitempty"`
				Kid    string   `json:"kid,omitempty"`
				Kty    string   `json:"kty,omitempty"`
				N      string   `json:"n,omitempty"`
			} `json:"keys,omitempty"`
			VMConfiguration struct {
				ConsoleEnabled bool   `json:"console-enabled,omitempty"`
				CurrentTime    int    `json:"current-time,omitempty"`
				SecureBoot     bool   `json:"secure-boot,omitempty"`
				TpmEnabled     bool   `json:"tpm-enabled,omitempty"`
				VMUniqueID     string `json:"vmUniqueId,omitempty"`
			} `json:"vm-configuration,omitempty"`
		} `json:"x-ms-runtime,omitempty"`
		XMsSevsnpvmAuthorkeydigest   string `json:"x-ms-sevsnpvm-authorkeydigest,omitempty"`
		XMsSevsnpvmBootloaderSvn     int    `json:"x-ms-sevsnpvm-bootloader-svn,omitempty"`
		XMsSevsnpvmFamilyID          string `json:"x-ms-sevsnpvm-familyId,omitempty"`
		XMsSevsnpvmGuestsvn          int    `json:"x-ms-sevsnpvm-guestsvn,omitempty"`
		XMsSevsnpvmHostdata          string `json:"x-ms-sevsnpvm-hostdata,omitempty"`
		XMsSevsnpvmIdkeydigest       string `json:"x-ms-sevsnpvm-idkeydigest,omitempty"`
		XMsSevsnpvmImageID           string `json:"x-ms-sevsnpvm-imageId,omitempty"`
		XMsSevsnpvmIsDebuggable      bool   `json:"x-ms-sevsnpvm-is-debuggable,omitempty"`
		XMsSevsnpvmLaunchmeasurement string `json:"x-ms-sevsnpvm-launchmeasurement,omitempty"`
		XMsSevsnpvmMicrocodeSvn      int    `json:"x-ms-sevsnpvm-microcode-svn,omitempty"`
		XMsSevsnpvmMigrationAllowed  bool   `json:"x-ms-sevsnpvm-migration-allowed,omitempty"`
		XMsSevsnpvmReportdata        string `json:"x-ms-sevsnpvm-reportdata,omitempty"`
		XMsSevsnpvmReportid          string `json:"x-ms-sevsnpvm-reportid,omitempty"`
		XMsSevsnpvmSmtAllowed        bool   `json:"x-ms-sevsnpvm-smt-allowed,omitempty"`
		XMsSevsnpvmSnpfwSvn          int    `json:"x-ms-sevsnpvm-snpfw-svn,omitempty"`
		XMsSevsnpvmTeeSvn            int    `json:"x-ms-sevsnpvm-tee-svn,omitempty"`
		XMsSevsnpvmVmpl              int    `json:"x-ms-sevsnpvm-vmpl,omitempty"`
	} `json:"x-ms-isolation-tee,omitempty"`
	XMsPolicyHash string `json:"x-ms-policy-hash,omitempty"`
	XMsRuntime    struct {
		ClientPayload struct {
			Nonce string `json:"nonce,omitempty"`
		} `json:"client-payload,omitempty"`
		Keys []struct {
			E      string   `json:"e,omitempty"`
			KeyOps []string `json:"key_ops,omitempty"`
			Kid    string   `json:"kid,omitempty"`
			Kty    string   `json:"kty,omitempty"`
			N      string   `json:"n,omitempty"`
		} `json:"keys,omitempty"`
	} `json:"x-ms-runtime,omitempty"`
	XMsVer string `json:"x-ms-ver,omitempty"`
}
