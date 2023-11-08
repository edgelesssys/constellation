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

// Used to validate DNS names.
var domainRegex = regexp.MustCompile(`^(?i)[a-z0-9-]+(\.[a-z0-9-]+)+\.?$`)

// Constraint is a constraint on a document or a field of a document.
type Constraint struct {
	// Satisfied returns no error if the constraint is satisfied.
	// Otherwise, it returns the reason why the constraint is not satisfied,
	// possibly including its child errors, i.e., errors returned by constraints
	// that are embedded in this constraint.
	Satisfied func() *TreeError
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
		Satisfied: func() *TreeError {
			if err := c.Satisfied(); err != nil {
				return newTraceError(docRef, fieldRef, err)
			}
			return nil
		},
	}
}

// MatchRegex is a constraint that if s matches regex.
func MatchRegex(s string, regex string) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if !regexp.MustCompile(regex).MatchString(s) {
				return NewErrorTree(fmt.Errorf("%s must match the pattern %s", s, regex))
			}
			return nil
		},
	}
}

// Equal is a constraint that checks if s is equal to t.
func Equal[T comparable](s T, t T) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if s != t {
				return NewErrorTree(fmt.Errorf("%v must be equal to %v", s, t))
			}
			return nil
		},
	}
}

// NotEqual is a constraint that checks if s is not equal to t.
func NotEqual[T comparable](s T, t T) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if Equal(s, t).Satisfied() == nil {
				return NewErrorTree(fmt.Errorf("%v must not be equal to %v", s, t))
			}
			return nil
		},
	}
}

// Empty is a constraint that checks if s is empty.
func Empty[T comparable](s T) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			var zero T
			if s != zero {
				return NewErrorTree(fmt.Errorf("%v must be empty", s))
			}
			return nil
		},
	}
}

// NotEmpty is a constraint that checks if s is not empty.
func NotEmpty[T comparable](s T) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if Empty(s).Satisfied() == nil {
				return NewErrorTree(fmt.Errorf("must not be empty"))
			}
			return nil
		},
	}
}

// OneOf is a constraint that s is in the set of values p.
func OneOf[T comparable](s T, p []T) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			for _, v := range p {
				if s == v {
					return nil
				}
			}
			return NewErrorTree(fmt.Errorf("%v must be one of %v", s, p))
		},
	}
}

// IPAddress is a constraint that checks if s is a valid IP address.
func IPAddress(s string) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if net.ParseIP(s) == nil {
				return NewErrorTree(fmt.Errorf("%s must be a valid IP address", s))
			}
			return nil
		},
	}
}

// CIDR is a constraint that checks if s is a valid CIDR.
func CIDR(s string) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if _, _, err := net.ParseCIDR(s); err != nil {
				return NewErrorTree(fmt.Errorf("%s must be a valid CIDR", s))
			}
			return nil
		},
	}
}

// DNSName is a constraint that checks if s is a valid DNS name.
func DNSName(s string) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if !domainRegex.MatchString(s) {
				return NewErrorTree(fmt.Errorf("%s must be a valid DNS name", s))
			}
			return nil
		},
	}
}

// EmptySlice is a constraint that checks if s is an empty slice.
func EmptySlice[T comparable](s []T) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if len(s) != 0 {
				return NewErrorTree(fmt.Errorf("%v must be empty", s))
			}
			return nil
		},
	}
}

// NotEmptySlice is a constraint that checks if slice s is not empty.
func NotEmptySlice[T comparable](s []T) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if EmptySlice(s).Satisfied() == nil {
				return NewErrorTree(fmt.Errorf("must not be empty"))
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
		Satisfied: func() *TreeError {
			retErr := NewErrorTree(fmt.Errorf("all of the constraints must be satisfied: "))
			for i, v := range s {
				if err := c(i, v).Satisfied(); err != nil {
					retErr.appendChild(err)
				}
			}
			if len(retErr.children) == 0 {
				return nil
			}
			return retErr
		},
	}
}

// And groups multiple constraints in an "and" relation and fails according to the given strategy.
func And(errStrat ErrStrategy, constraints ...*Constraint) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			retErr := NewErrorTree(fmt.Errorf("all of the constraints must be satisfied: "))
			for _, constraint := range constraints {
				if err := constraint.Satisfied(); err != nil {
					if errStrat == FailFast {
						return err
					}
					retErr.appendChild(err)
				}
			}
			if len(retErr.children) == 0 {
				return nil
			}
			return retErr
		},
	}
}

// Or groups multiple constraints in an "or" relation.
func Or(constraints ...*Constraint) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			retErr := NewErrorTree(fmt.Errorf("at least one of the constraints must be satisfied: "))
			for _, constraint := range constraints {
				err := constraint.Satisfied()
				if err == nil {
					return nil
				}
				retErr.appendChild(err)
			}
			if len(retErr.children) == 0 {
				return nil
			}
			return retErr
		},
	}
}

// IfNotNil evaluates a constraint if and only if s is not nil.
func IfNotNil[T comparable](s *T, c func() *Constraint) *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if s == nil {
				return nil
			}
			return c().Satisfied()
		},
	}
}
