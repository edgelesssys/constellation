/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package oid defines OIDs for different CSPs. Currently this is used in attested TLS to distinguish the attestation documents.
OIDs beginning with 1.3.9900 are reserved and can be used without registration.

* The 1.3.9900.1 branch is reserved for placeholder values and testing.

* The 1.3.9900.2 branch is reserved for AWS.

* The 1.3.9900.3 branch is reserved for GCP.

* The 1.3.9900.4 branch is reserved for Azure.

* The 1.3.9900.5 branch is reserved for QEMU.

Deprecated OIDs should never be reused for different purposes.
Instead, new OIDs should be added in the appropriate branch at the next available index.
*/
package oid

import (
	"encoding/asn1"
	"fmt"
)

// Getter returns an ASN.1 Object Identifier.
type Getter interface {
	OID() asn1.ObjectIdentifier
}

// FromString returns the OID for the given string.
func FromString(oid string) (Getter, error) {
	switch oid {
	case dummy:
		return Dummy{}, nil
	case awsNitroTPM:
		return AWSNitroTPM{}, nil
	case gcpSEVES:
		return GCPSEVES{}, nil
	case azureSEVSNP:
		return AzureSEVSNP{}, nil
	case azureTrustedLaunch:
		return AzureTrustedLaunch{}, nil
	case qemuVTPM:
		return QEMUVTPM{}, nil
	case qemuTDX:
		return QEMUTDX{}, nil
	}
	return nil, fmt.Errorf("unknown OID: %q", oid)
}

// Dummy OID for testing.
type Dummy struct{}

// OID returns the struct's object identifier.
func (Dummy) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 1, 1}
}

// String returns the string representation of the OID.
func (Dummy) String() string {
	return dummy
}

// AWSNitroTPM holds the AWS nitro TPM OID.
type AWSNitroTPM struct{}

// OID returns the struct's object identifier.
func (AWSNitroTPM) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 2, 1}
}

// String returns the string representation of the OID.
func (AWSNitroTPM) String() string {
	return awsNitroTPM
}

// GCPSEVES holds the GCP SEV-ES OID.
type GCPSEVES struct{}

// OID returns the struct's object identifier.
func (GCPSEVES) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 3, 1}
}

// String returns the string representation of the OID.
func (GCPSEVES) String() string {
	return gcpSEVES
}

// AzureSEVSNP holds the OID for Azure SNP CVMs.
type AzureSEVSNP struct{}

// OID returns the struct's object identifier.
func (AzureSEVSNP) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 4, 1}
}

// String returns the string representation of the OID.
func (AzureSEVSNP) String() string {
	return azureSEVSNP
}

// AzureTrustedLaunch holds the OID for Azure TrustedLaunch VMs.
type AzureTrustedLaunch struct{}

// OID returns the struct's object identifier.
func (AzureTrustedLaunch) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 4, 2}
}

// String returns the string representation of the OID.
func (AzureTrustedLaunch) String() string {
	return azureTrustedLaunch
}

// QEMUVTPM holds the QEMUVTPM OID.
type QEMUVTPM struct{}

// OID returns the struct's object identifier.
func (QEMUVTPM) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 5, 1}
}

// String returns the string representation of the OID.
func (QEMUVTPM) String() string {
	return qemuVTPM
}

// QEMUTDX holds the QEMU TDX OID.
// Placeholder for dev-cloud integration.
type QEMUTDX struct{}

// OID returns the struct's object identifier.
// Placeholder for dev-cloud integration.
func (QEMUTDX) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 5, 99}
}

// String returns the string representation of the OID.
func (QEMUTDX) String() string {
	return qemuTDX
}

const (
	dummy              = "dummy"
	awsNitroTPM        = "aws-nitro-tpm"
	gcpSEVES           = "gcp-sev-es"
	azureSEVSNP        = "azure-sev-snp"
	azureTrustedLaunch = "azure-trustedlaunch"
	qemuVTPM           = "qemu-vtpm"
	qemuTDX            = "qemu-tdx"
)
