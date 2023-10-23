/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package validation provides a unified document validation interface for use within the Constellation CLI.

It validates documents that specify a set of constraints on their content.
*/
package validation

import "errors"

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validator validates documents.
type Validator struct{}

// Validatable is implemented by documents that can be validated.
// It returns a list of constraints that must be satisfied for the document to be valid.
type Validatable interface {
	Constraints() []Constraint
}

// ValidateOptions are the options to use when validating a document.
type ValidateOptions struct {
	// FailFast stops validation on the first error.
	FailFast bool
}

// Validate validates a document using the given options.
func (v *Validator) Validate(doc Validatable, opts ValidateOptions) error {
	var retErr error
	for _, c := range doc.Constraints() {
		if valid, err := c.Satisfied(); !valid {
			if opts.FailFast {
				return err
			}
			retErr = errors.Join(retErr, err)
		}
	}
	return retErr
}
