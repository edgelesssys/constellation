package oid

import (
	"encoding/asn1"
)

type Getter interface {
	OID() asn1.ObjectIdentifier
}

// Here we define OIDs for different CSPs. Currently this is used in attested TLS to distinguish the attestation documents.
// OIDs beginning with 1.3.9900 are reserved and can be used without registration.

type Dummy struct{}

func (Dummy) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 1}
}

type AWS struct{}

func (AWS) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 2}
}

type GCP struct{}

func (GCP) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 3}
}

type Azure struct{}

func (Azure) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 4}
}

// TODO: Remove once we no longer use non cvms.
type GCPNonCVM struct{}

func (GCPNonCVM) OID() asn1.ObjectIdentifier {
	return asn1.ObjectIdentifier{1, 3, 9900, 99}
}
