/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: BUSL-1.1
*/

// Package testdata contains testing data for an attestation process.
package testdata

import _ "embed"

// Ark is a valid ARK certificate, as returned from the AMD KDS.
//
//go:embed ark.pem
var Ark []byte

// Ask is a valid ASK certificate, as returned from the AMD KDS.
//
//go:embed ask.pem
var Ask []byte
