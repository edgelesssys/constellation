/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package validation provides a unified document validation interface for use within the Constellation CLI.

It validates documents that specify a set of constraints on their content.
*/
package validation

import (
	"fmt"
)

// ErrStrategy is the strategy to use when encountering an error during validation.
type ErrStrategy int

const (
	// EvaluateAll continues evaluating all constraints even if one is not satisfied.
	EvaluateAll ErrStrategy = iota
	// FailFast stops validation on the first error.
	FailFast
)

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validator validates documents.
type Validator struct{}

// Validatable is implemented by documents that can be validated.
// It returns a list of constraints that must be satisfied for the document to be valid.
type Validatable interface {
	Constraints() []*Constraint
}

// ValidateOptions are the options to use when validating a document.
type ValidateOptions struct {
	// ErrStrategy is the strategy to use when encountering an error during validation.
	ErrStrategy ErrStrategy
}

// Validate validates a document using the given options.
func (v *Validator) Validate(doc Validatable, opts ValidateOptions) error {
	retErr := newListError(fmt.Errorf("validating document: "))
	for _, c := range doc.Constraints() {
		if err := c.Satisfied(); err != nil {
			if opts.ErrStrategy == FailFast {
				return err
			}
			if lErr, ok := err.(*listError); ok {
				retErr.addChild(lErr)
			} else {
				retErr.addChild(newListError(err))
			}
		}
	}
	if len(retErr.childs) > 0 {
		return retErr
	}
	return nil
}
