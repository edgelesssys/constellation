/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package testdata contains testing data for an attestation process.
package testdata

import _ "embed"

// ARK is a valid ARK certificate, as returned from the AMD KDS.
//
//go:embed ark.pem
var Ark []byte

// ASK is a valid ASK certificate, as returned from the AMD KDS.
//
//go:embed ask.pem
var Ask []byte
