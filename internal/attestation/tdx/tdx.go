/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package TDX implements attestation for Intel TDX.
package tdx

type tdxAttestationDocument struct {
	RawQuote []byte
	UserData []byte
}
