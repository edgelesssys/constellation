/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package TDX implements attestation for Intel TDX.
package tdx

type tdxAttestationDocument struct {
	// RawQuote is the raw TDX quote.
	RawQuote []byte
	// UserData is the user data that was passed to the enclave and was included in the quote.
	UserData []byte
}
