/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

// Package testdata contains testing data for an attestation process.
package testdata

import _ "embed"

// SNPReport holds a valid VLEK-signed SNP report.
//
//go:embed report.txt
var SNPReport string

// AKDigest holds the AK digest embedded in SNPReport.REPORT_DATA.
const AKDigest = "87ab7caf510e1b3520dc3cceb64ee44128e10976fb0d3fc5e274699d8aaf506154af4c1de0a026b49fdf861e9ac75551551b3534d1c61369a3b08f5baed0db2f"

// VLEK for SNPReport.
//
//go:embed vlek.pem
var VLEK []byte

// CertChain is a valid certificate chain for the VLEK certificate. Queried from AMD KDS.
//
//go:embed certchain.pem
var CertChain []byte
