/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

// Package testdata contains testing data for an attestation process.
package testdata

import _ "embed"

// CertChain is a valid certificate chain (PEM, as returned from Azure THIM) for the VCEK certificate.
//
//go:embed certchain.pem
var CertChain []byte
