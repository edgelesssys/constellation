/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package variant defines Attestation variants for different CSPs.

Each variant defines an OID, a string representation, and a function to compare it to other OIDs.

The OID is used in attested TLS to distinguish the attestation documents.
OIDs beginning with 1.3.9900 are reserved and can be used without registration.

* The 1.3.9900.1 branch is reserved for placeholder values and testing.

* The 1.3.9900.2 branch is reserved for AWS.

* The 1.3.9900.3 branch is reserved for GCP.

* The 1.3.9900.4 branch is reserved for Azure.

* The 1.3.9900.5 branch is reserved for QEMU.

Deprecated OIDs should never be reused for different purposes.
Instead, new OIDs should be added in the appropriate branch at the next available index.

String representation should be lowercase and contain only letters, numbers, and hyphens.
They should be prefixed with the branch name, e.g. all variants in the 1.3.9900.2 (AWS) branch should start with "aws-".
Each variant should have a unique string representation.
*/
package variant

import (
	"encoding/asn1"
	"fmt"
)

const (
	dummy              = "dummy"
	awsNitroTPM        = "aws-nitro-tpm"
	gcpSEVES           = "gcp-sev-es"
	azureSEVSNP        = "azure-sev-snp"
	azureTrustedLaunch = "azure-trustedlaunch"
	qemuVTPM           = "qemu-vtpm"
	qemuTDX            = "qemu-tdx"
)

// Getter returns an ASN.1 Object Identifier.
type Getter interface {
	OID() asn1.ObjectIdentifier
}

// Variant describes an attestation variant.
type Variant interface {
	Getter
	String() string
	Equal(other Getter) bool
}

// FromString returns the OID for the given string.
func FromString(oid string) (Variant, error) {
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

// Equal returns true if the other variant is also a Dummy.
func (Dummy) Equal(other Getter) bool {
	return other.OID().Equal(Dummy{}.OID())
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

// Equal returns true if the other variant is also AWSNitroTPM.
func (AWSNitroTPM) Equal(other Getter) bool {
	return other.OID().Equal(AWSNitroTPM{}.OID())
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

// Equal returns true if the other variant is also GCPSEVES.
func (GCPSEVES) Equal(other Getter) bool {
	return other.OID().Equal(GCPSEVES{}.OID())
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

// Equal returns true if the other variant is also AzureSEVSNP.
func (AzureSEVSNP) Equal(other Getter) bool {
	return other.OID().Equal(AzureSEVSNP{}.OID())
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

// Equal returns true if the other variant is also AzureTrustedLaunch.
func (AzureTrustedLaunch) Equal(other Getter) bool {
	return other.OID().Equal(AzureTrustedLaunch{}.OID())
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

// Equal returns true if the other variant is also QEMUVTPM.
func (QEMUVTPM) Equal(other Getter) bool {
	return other.OID().Equal(QEMUVTPM{}.OID())
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

// Equal returns true if the other variant is also QEMUTDX.
func (QEMUTDX) Equal(other Getter) bool {
	return other.OID().Equal(QEMUTDX{}.OID())
}
