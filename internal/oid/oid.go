/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package oid

import (
	"encoding/asn1"
)

// Getter returns an ASN.1 Object Identifier.
type Getter interface {
	OID() asn1.ObjectIdentifier
}

// Here we define OIDs for different CSPs. Currently this is used in attested TLS to distinguish the attestation documents.
// OIDs beginning with 1.3.9900 are reserved and can be used without registration.

// Dummy OID for testing.
type Dummy struct{}

// OID returns the struct's object identifier.
func (Dummy) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 1}
}

// AWS holds the AWS OID.
type AWS struct{}

// OID returns the struct's object identifier.
func (AWS) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 2}
}

// GCP holds the GCP OID.
type GCP struct{}

// OID returns the struct's object identifier.
func (GCP) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 3}
}

// Azure holds the Azure OID.
type Azure struct{}

// OID returns the struct's object identifier.
func (Azure) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 4}
}

// QEMU holds the QEMU OID.
type QEMU struct{}

// OID returns the struct's object identifier.
func (QEMU) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 5}
}
