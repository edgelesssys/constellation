/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package validation

import (
	"fmt"
	"reflect"
	"regexp"
)

// Constraint is a constraint on a document or a field of a document.
type Constraint struct {
	// Satisfied returns true if the constraint is satisfied.
	Satisfied func() (valid bool, err error)
}

/*
WithFieldTrace adds a well-formatted error message to the constraint,
which will be displayed if the constraint is not satisfied on the validated document.
*/
func (c *Constraint) WithFieldTrace(doc any, field any) Constraint {
	// we only want to dereference the needle once to dereference the pointer
	// used to pass it to the function without losing reference to it, as the
	// needle could be an arbitrarily long chain of pointers. The same
	// applies to the haystack.
	derefedField := pointerDeref(reflect.ValueOf(field))
	fieldRef := referenceableValue{
		value: derefedField,
		addr:  derefedField.UnsafeAddr(),
		_type: derefedField.Type(),
	}
	derefedDoc := pointerDeref(reflect.ValueOf(doc))
	docRef := referenceableValue{
		value: derefedDoc,
		addr:  derefedDoc.UnsafeAddr(),
		_type: derefedDoc.Type(),
	}
	return c.withErrorMessage(docRef, fieldRef)
}

func (c *Constraint) WithMapFieldTrace(doc any, field any, mapKey string) Constraint {
	// we only want to dereference the needle once to dereference the pointer
	// used to pass it to the function without losing reference to it, as the
	// needle could be an arbitrarily long chain of pointers. The same
	// applies to the haystack.
	derefedField := pointerDeref(reflect.ValueOf(field))
	fieldRef := referenceableValue{
		value:  derefedField,
		addr:   derefedField.UnsafeAddr(),
		_type:  derefedField.Type(),
		mapKey: mapKey,
	}
	derefedDoc := pointerDeref(reflect.ValueOf(doc))
	docRef := referenceableValue{
		value: derefedDoc,
		addr:  derefedDoc.UnsafeAddr(),
		_type: derefedDoc.Type(),
	}
	return c.withErrorMessage(docRef, fieldRef)
}

func (c *Constraint) withErrorMessage(docRef, fieldRef referenceableValue) Constraint {
	return Constraint{
		Satisfied: func() (valid bool, err error) {
			valid, err = c.Satisfied()
			if err != nil {
				return valid, NewValidationError(docRef, fieldRef, err)
			}
			return valid, nil
		},
	}
}

// MatchRegex is a constraint that if s matches regex.
func MatchRegex(s string, regex string) *Constraint {
	return &Constraint{
		Satisfied: func() (valid bool, err error) {
			if !regexp.MustCompile(regex).MatchString(s) {
				return false, fmt.Errorf("%s must match the pattern %s", s, regex)
			}
			return true, nil
		},
	}
}

// Equal is a constraint that if s is equal to t.
func Equal[T comparable](s T, t T) *Constraint {
	return &Constraint{
		Satisfied: func() (valid bool, err error) {
			if s != t {
				return false, fmt.Errorf("%v must be equal to %v", s, t)
			}
			return true, nil
		},
	}
}

// NotEmpty is a constraint that if s is not empty.
func NotEmpty[T comparable](s T) *Constraint {
	return &Constraint{
		Satisfied: func() (valid bool, err error) {
			var zero T
			if s == zero {
				return false, fmt.Errorf("%v must not be empty", s)
			}
			return true, nil
		},
	}
}

// Empty is a constraint that if s is empty.
func Empty[T comparable](s T) *Constraint {
	return &Constraint{
		Satisfied: func() (valid bool, err error) {
			var zero T
			if s != zero {
				return false, fmt.Errorf("%v must be empty", s)
			}
			return true, nil
		},
	}
}
