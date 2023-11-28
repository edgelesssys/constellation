/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import (
	"context"
)

// singleUUIDVerifier constructs a RekorVerifier that returns a single UUID and no errors,
// and should work for most tests on the happy path.
func singleUUIDVerifier() *stubRekorVerifier {
	return &stubRekorVerifier{
		SearchByHashUUIDs: []string{"11111111111111111111111111111111111111111111111111111111111111111111111111111111"},
		SearchByHashError: nil,
		VerifyEntryError:  nil,
	}
}

// SubRekorVerifier is a stub for RekorVerifier.
type stubRekorVerifier struct {
	SearchByHashUUIDs []string
	SearchByHashError error
	VerifyEntryError  error
}

// SearchByHash returns the exported fields SearchByHashUUIDs, SearchByHashError.
func (v *stubRekorVerifier) SearchByHash(context.Context, string) ([]string, error) {
	return v.SearchByHashUUIDs, v.SearchByHashError
}

// VerifyEntry returns the exported field VerifyEntryError.
func (v *stubRekorVerifier) VerifyEntry(context.Context, string, string) error {
	return v.VerifyEntryError
}

type stubCosignVerifier struct {
	verifyError error
}

func (v *stubCosignVerifier) VerifySignature(_, _ []byte) error {
	return v.verifyError
}
