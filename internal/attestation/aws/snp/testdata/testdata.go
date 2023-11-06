/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package testdata contains testing data for an attestation process.
package testdata

import _ "embed"

// SNPReport holds a valid VLEK-signed SNP report.
//
//go:embed report.txt
var SNPReport string

// AKDigest holds the AK digest embedded in SNPReport.REPORT_DATA.
const AKDigest = "032635613c8e331fa29e096371910fe6a1f69383dda02c9461400a70b66d87a3da5dd863002522be43afc34f2c233989bd6e401e351d10d7cc800d6f5dfcf019"

// VLEK for SNPReport.
//
//go:embed vlek.pem
var VLEK []byte

// CertChain is a valid certificate chain for the VLEK certificate. Queried from AMD KDS.
//
//go:embed certchain.pem
var CertChain []byte
