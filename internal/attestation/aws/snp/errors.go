/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import "fmt"

// decodeError is used to signal an error during decoding of a public key.
// It only wrapps an error.
type decodeError struct {
	inner error
}

// newDecodeError an error in a DecodeError.
func newDecodeError(err error) *decodeError {
	return &decodeError{inner: err}
}

func (e *decodeError) Error() string {
	return fmt.Sprintf("decoding public key: %v", e.inner)
}

func (e *decodeError) Unwrap() error {
	return e.inner
}

// validationError is used to signal an invalid SNP report.
// It only wrapps an error.
// Used during testing to error conditions more precisely.
type validationError struct {
	inner error
}

// newValidationError wraps an error in a ValidationError.
func newValidationError(err error) *validationError {
	return &validationError{inner: err}
}

func (e *validationError) Error() string {
	return e.inner.Error()
}

func (e *validationError) Unwrap() error {
	return e.inner
}
