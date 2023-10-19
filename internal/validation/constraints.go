/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package validation

import (
	"fmt"
	"regexp"
)

// Constraint is a constraint on a document or a field of a document.
type Constraint func() (valid bool, err error)

// MatchRegex is a constraint that if s matches regex.
func MatchRegex(s string, regex string) Constraint {
	return func() (valid bool, err error) {
		if !regexp.MustCompile(regex).MatchString(s) {
			return false, fmt.Errorf("%s must match the pattern %s", s, regex)
		}
		return true, nil
	}
}

// Equal is a constraint that if s is equal to t.
func Equal[T comparable](s T, t T) Constraint {
	return func() (valid bool, err error) {
		if s != t {
			return false, fmt.Errorf("%v must be equal to %v", s, t)
		}
		return true, nil
	}
}

// NotEmpty is a constraint that if s is not empty.
func NotEmpty[T comparable](s T) Constraint {
	return func() (valid bool, err error) {
		var zero T
		if s == zero {
			return false, fmt.Errorf("%v must not be empty", s)
		}
		return true, nil
	}
}

// Empty is a constraint that if s is empty.
func Empty[T comparable](s T) Constraint {
	return func() (valid bool, err error) {
		var zero T
		if s != zero {
			return false, fmt.Errorf("%v must be empty", s)
		}
		return true, nil
	}
}
