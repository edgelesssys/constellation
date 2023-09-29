/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package testdata contains testing data for an attestation process.
package testdata

import _ "embed"

// CertChain is a valid certificate chain (PEM, as returned from Azure THIM) for the VCEK certificate.
//
//go:embed certchain.pem
var CertChain []byte
