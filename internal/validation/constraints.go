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
WithFieldTrace adds a well-formatted trace to the field to the error message
shown when the constraint is not satisfied. Both "doc" and "field" must be pointers:
  - "doc" must be a pointer to the top level document
  - "field" must be a pointer to the field to be validated

Example for a non-pointer field:

	Equal(d.IntField, 42).WithFieldTrace(d, &d.IntField)

Example for a pointer field:

	NotEmpty(d.StrPtrField).WithFieldTrace(d, d.StrPtrField)

Due to Go's addressability limititations regarding maps, if a map field is
to be validated, WithMapFieldTrace must be used instead of WithFieldTrace.
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
	return c.withTrace(docRef, fieldRef)
}

/*
WithMapFieldTrace adds a well-formatted trace to the map field to the error message
shown when the constraint is not satisfied. Both "doc" and "field" must be pointers:
  - "doc" must be a pointer to the top level document
  - "field" must be a pointer to the map containing the field to be validated
  - "mapKey" must be the key of the field to be validated in the map pointed to by "field"

Example:

	Equal(d.IntField, 42).WithMapFieldTrace(d, &d.IntField)

For non-map fields, WithFieldTrace should be used instead of WithMapFieldTrace.
*/
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
	return c.withTrace(docRef, fieldRef)
}

// withTrace wraps the constraint's error message with a well-formatted trace.
func (c *Constraint) withTrace(docRef, fieldRef referenceableValue) Constraint {
	return Constraint{
		Satisfied: func() (valid bool, err error) {
			valid, err = c.Satisfied()
			if err != nil {
				return valid, newError(docRef, fieldRef, err)
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
