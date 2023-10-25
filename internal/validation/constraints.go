/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package validation

import (
	"fmt"
	"net"
	"reflect"
	"regexp"
)

// Constraint is a constraint on a document or a field of a document.
type Constraint struct {
	// Satisfied returns no error if the constraint is satisfied.
	// Otherwise, it returns the reason why the constraint is not satisfied.
	Satisfied func() error
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
func (c *Constraint) WithFieldTrace(doc any, field any) *Constraint {
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

	Equal(d.IntField, 42).WithMapFieldTrace(d, &d.MapField, mapKey)

For non-map fields, WithFieldTrace should be used instead of WithMapFieldTrace.
*/
func (c *Constraint) WithMapFieldTrace(doc any, field any, mapKey string) *Constraint {
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
func (c *Constraint) withTrace(docRef, fieldRef referenceableValue) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			if err := c.Satisfied(); err != nil {
				return newError(docRef, fieldRef, err)
			}
			return nil
		},
	}
}

// MatchRegex is a constraint that if s matches regex.
func MatchRegex(s string, regex string) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			if !regexp.MustCompile(regex).MatchString(s) {
				return fmt.Errorf("%s must match the pattern %s", s, regex)
			}
			return nil
		},
	}
}

// Equal is a constraint that checks if s is equal to t.
func Equal[T comparable](s T, t T) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			if s != t {
				return fmt.Errorf("%v must be equal to %v", s, t)
			}
			return nil
		},
	}
}

// Empty is a constraint that checks if s is empty.
func Empty[T comparable](s T) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			var zero T
			if s != zero {
				return fmt.Errorf("%v must be empty", s)
			}
			return nil
		},
	}
}

// NotEmpty is a constraint that checks if s is not empty.
func NotEmpty[T comparable](s T) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			if Empty(s).Satisfied() == nil {
				return fmt.Errorf("%v must not be empty", s)
			}
			return nil
		},
	}
}

// OneOf is a constraint that s is in the set of values p.
func OneOf[T comparable](s T, p []T) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			for _, v := range p {
				if s == v {
					return nil
				}
			}
			return fmt.Errorf("%v must be one of %v", s, p)
		},
	}
}

// IPAddress is a constraint that checks if s is a valid IP address.
func IPAddress(s string) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			if net.ParseIP(s) == nil {
				return fmt.Errorf("%s must be a valid IP address", s)
			}
			return nil
		},
	}
}

// CIDR is a constraint that checks if s is a valid CIDR.
func CIDR(s string) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			if _, _, err := net.ParseCIDR(s); err != nil {
				return fmt.Errorf("%s must be a valid CIDR", s)
			}
			return nil
		},
	}
}

// DNSName is a constraint that checks if s is a valid DNS name.
func DNSName(s string) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			if _, err := net.LookupHost(s); err != nil {
				return fmt.Errorf("%s must be a valid DNS name", s)
			}
			return nil
		},
	}
}

// EmptySlice is a constraint that checks if s is an empty slice.
func EmptySlice[T comparable](s []T) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			if len(s) != 0 {
				return fmt.Errorf("%v must be empty", s)
			}
			return nil
		},
	}
}

// All is a constraint that checks if all elements of s satisfy the constraint c.
// The constraint should be parametric in regards to the index of the element in s,
// as well as the element itself.
func All[T comparable](s []T, c func(i int, v T) *Constraint) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			retErr := newListError(fmt.Errorf("all of the constraints must be satisfied: "))
			for i, v := range s {
				if err := c(i, v).Satisfied(); err != nil {
					retErr.addChild(newListError(err))
				}
			}
			if len(retErr.childs) > 0 {
				return retErr
			}
			return nil
		},
	}
}

// And groups multiple constraints in an "and" relation and fails according to the given strategy.
func And(errStrat ErrStrategy, constraints ...*Constraint) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			retErr := newListError(fmt.Errorf("all of the constraints must be satisfied: "))
			for _, constraint := range constraints {
				if err := constraint.Satisfied(); err != nil {
					if errStrat == FailFast {
						return err
					}
					retErr.addChild(newListError(err))
				}
			}
			if len(retErr.childs) > 0 {
				return retErr
			}
			return nil
		},
	}
}

// Or groups multiple constraints in an "or" relation.
func Or(constraints ...*Constraint) *Constraint {
	return &Constraint{
		Satisfied: func() error {
			retErr := newListError(fmt.Errorf("at least one of the constraints must be satisfied: "))
			for _, constraint := range constraints {
				err := constraint.Satisfied()
				if err == nil {
					return nil
				}
				retErr.addChild(newListError(err))
			}
			if len(retErr.childs) > 0 {
				return retErr
			}
			return nil
		},
	}
}
