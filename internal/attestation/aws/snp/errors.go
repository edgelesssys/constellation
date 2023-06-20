/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snp

import "fmt"

// DecodeError is used to signal an error during decoding of a public key.
// It only wrapps an error.
type DecodeError struct {
	inner error
}

// NewDecodeError an error in a DecodeError.
func NewDecodeError(err error) *DecodeError {
	return &DecodeError{inner: err}
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf("error decoding public key: %v", e.inner)
}

func (e *DecodeError) Unwrap() error {
	return e.inner
}

// ValidationError is used to signal an invalid SNP report.
// It only wrapps an error.
// Used during testing to error conditions more precisely.
type ValidationError struct {
	inner error
}

// NewValidationError wraps an error in a ValidationError.
func NewValidationError(err error) *ValidationError {
	return &ValidationError{inner: err}
}

func (e *ValidationError) Error() string {
	return e.inner.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.inner
}
