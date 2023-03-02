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
	"errors"
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
		return AWS{}, nil
	case gcpSEVES:
		return GCP{}, nil
	case azureSEVSNP:
		return AzureSNP{}, nil
	case azureTL:
		return AzureTrustedLaunch{}, nil
	case qemuVTPM:
		return QEMU{}, nil
	}
	return nil, errors.New("unknown OID")
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

// AWS holds the AWS OID.
type AWS struct{}

// OID returns the struct's object identifier.
func (AWS) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 2, 1}
}

// String returns the string representation of the OID.
func (AWS) String() string {
	return awsNitroTPM
}

// GCP holds the GCP OID.
type GCP struct{}

// OID returns the struct's object identifier.
func (GCP) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 3, 1}
}

// String returns the string representation of the OID.
func (GCP) String() string {
	return gcpSEVES
}

// AzureSNP holds the OID for Azure SNP CVMs.
type AzureSNP struct{}

// OID returns the struct's object identifier.
func (AzureSNP) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 4, 1}
}

// String returns the string representation of the OID.
func (AzureSNP) String() string {
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
	return azureTL
}

// QEMU holds the QEMU OID.
type QEMU struct{}

// OID returns the struct's object identifier.
func (QEMU) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 5, 1}
}

// String returns the string representation of the OID.
func (QEMU) String() string {
	return qemuVTPM
}

const (
	dummy       = "dummy"
	awsNitroTPM = "aws-nitrotpm"
	gcpSEVES    = "gcp-seves"
	azureSEVSNP = "azure-sevsnp"
	azureTL     = "azure-trustedlaunch"
	qemuVTPM    = "qemu-vtpm"
)
