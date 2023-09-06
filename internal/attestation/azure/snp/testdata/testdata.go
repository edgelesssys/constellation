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

// AzureThimVCEK is an example VCEK certificate (PEM, as returned from Azure THIM) for the AttestationReport.
//
//go:embed vcek.pem
var AzureThimVCEK []byte

// AmdKdsVCEK is an example VCEK certificate (DER, as returned from AMD KDS) for the AttestationReport.
//
//go:embed vcek.cert
var AmdKdsVCEK []byte

// RuntimeData is an example runtime data from the TPM for the AttestationReport.
//
//go:embed runtimedata.bin
var RuntimeData []byte

// CertChain is a valid certificate chain (PEM, as returned from Azure THIM) for the VCEK certificate.
//
//go:embed certchain.pem
var CertChain []byte
