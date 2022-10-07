/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cmd

import "context"

// SingleUUIDVerifier constructs a RekorVerifier that returns a single UUID and no errors,
// and should work for most tests on the happy path.
func SingleUUIDVerifier() *StubRekorVerifier {
	return &StubRekorVerifier{
		SearchByHashUUIDs: []string{"11111111111111111111111111111111111111111111111111111111111111111111111111111111"},
		SearchByHashError: nil,
		VerifyEntryError:  nil,
	}
}

// SubRekorVerifier is a stub for RekorVerifier.
type StubRekorVerifier struct {
	SearchByHashUUIDs []string
	SearchByHashError error
	VerifyEntryError  error
}

// SearchByHash returns the exported fields SearchByHashUUIDs, SearchByHashError.
func (v *StubRekorVerifier) SearchByHash(context.Context, string) ([]string, error) {
	return v.SearchByHashUUIDs, v.SearchByHashError
}

// VerifyEntry returns the exported field VerifyEntryError.
func (v *StubRekorVerifier) VerifyEntry(context.Context, string, string) error {
	return v.VerifyEntryError
}
