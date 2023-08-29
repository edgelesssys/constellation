/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package testdata contains testing data for an attestation process.
package testdata

import _ "embed"

// AttestationBytes is an example attestation report from a Constellation VM.
//
//go:embed attestation.bin
var AttestationReport []byte

// VCEK is an example VCEK certificate for the AttestationReport.
//
//go:embed vcek.cert
var VCEK []byte

// CertChain is a valid certificate chain for the VCEK certificate.
//
//go:embed chain.cert
var CertChain []byte
