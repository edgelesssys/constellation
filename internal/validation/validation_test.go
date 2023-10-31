/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validDoc = func() *exampleDoc {
	return &exampleDoc{
		StrField: "abc",
		NumField: 42,
		MapField: &map[string]string{
			"empty": "",
		},
		NotEmptyField:   "certainly not.",
		MatchRegexField: "abc",
		OneOfField:      "one",
		OrLeftField:     "left",
		OrRightField:    "right",
		AndLeftField:    "left",
		AndRightField:   "right",
	}
}

func TestValidate(t *testing.T) {
	testCases := map[string]struct {
		doc          func() *exampleDoc
		opts         ValidateOptions
		wantErr      bool
		errAssertion func(*assert.Assertions, error) bool
	}{
		"valid": {
			doc:  validDoc,
			opts: ValidateOptions{},
		},
		"strField is not abc": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.StrField = "def"
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.strField: def must be abc")
			},
			opts: ValidateOptions{},
		},
		"numField is not 42": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.NumField = 43
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.numField: 43 must be equal to 42")
			},
		},
		"multiple errors": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.StrField = "def"
				doc.NumField = 43
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.strField: def must be abc") &&
					assert.Contains(err.Error(), "validating exampleDoc.numField: 43 must be equal to 42")
			},
			opts: ValidateOptions{},
		},
		"multiple errors, fail fast": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.StrField = "def"
				doc.NumField = 43
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.strField: def must be abc")
			},
			opts: ValidateOptions{
				ErrStrategy: FailFast,
			},
		},
		"map field is not empty": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.MapField = &map[string]string{
					"empty": "haha!",
				}
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.mapField[\"empty\"]: haha! must be empty")
			},
			opts: ValidateOptions{
				ErrStrategy: FailFast,
			},
		},
		"not empty field is empty": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.NotEmptyField = ""
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.notEmptyField: must not be empty")
			},
			opts: ValidateOptions{
				ErrStrategy: FailFast,
			},
		},
		"regex doesnt match": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.MatchRegexField = "dontmatch"
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.matchRegexField: dontmatch must match the pattern ^a.c$")
			},
			opts: ValidateOptions{
				ErrStrategy: FailFast,
			},
		},
		"field is not in 'oneof' values": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.OneOfField = "not in oneof"
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.oneOfField: not in oneof must be one of [one two three]")
			},
			opts: ValidateOptions{
				ErrStrategy: FailFast,
			},
		},
		"'or' violated": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.OrLeftField = "not left"
				doc.OrRightField = "not right"
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "at least one of the constraints must be satisfied:") &&
					assert.Contains(err.Error(), "validating exampleDoc.orLeftField: not left must be equal to left") &&
					assert.Contains(err.Error(), "validating exampleDoc.orRightField: not right must be equal to right")
			},
			opts: ValidateOptions{
				ErrStrategy: FailFast,
			},
		},
		"'and' violated": {
			doc: func() *exampleDoc {
				doc := validDoc()
				doc.AndRightField = "not right"
				return doc
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "all of the constraints must be satisfied:") &&
					assert.Contains(err.Error(), "validating exampleDoc.andRightField: not right must be equal to right")
			},
			opts: ValidateOptions{
				ErrStrategy: FailFast,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			err := NewValidator().Validate(tc.doc(), tc.opts)
			if tc.wantErr {
				require.Error(err)
				if !tc.errAssertion(assert, err) {
					t.Fatalf("unexpected error: %v", err)
				}
			} else {
				require.NoError(err)
			}
		})
	}
}

type exampleDoc struct {
	StrField        string             `json:"strField"`
	NumField        int                `json:"numField"`
	MapField        *map[string]string `json:"mapField"`
	NotEmptyField   string             `json:"notEmptyField"`
	MatchRegexField string             `json:"matchRegexField"`
	OneOfField      string             `json:"oneOfField"`
	OrLeftField     string             `json:"orLeftField"`
	OrRightField    string             `json:"orRightField"`
	AndLeftField    string             `json:"andLeftField"`
	AndRightField   string             `json:"andRightField"`
}

// Constraints implements the Validatable interface.
func (d *exampleDoc) Constraints() []*Constraint {
	mapField := *(d.MapField)

	return []*Constraint{
		d.strFieldNeedsToBeAbc().
			WithFieldTrace(d, &d.StrField),
		Equal(d.NumField, 42).
			WithFieldTrace(d, &d.NumField),
		Empty(mapField["empty"]).
			WithMapFieldTrace(d, d.MapField, "empty"),
		NotEmpty(d.NotEmptyField).
			WithFieldTrace(d, &d.NotEmptyField),
		MatchRegex(d.MatchRegexField, "^a.c$").
			WithFieldTrace(d, &d.MatchRegexField),
		OneOf(d.OneOfField, []string{"one", "two", "three"}).
			WithFieldTrace(d, &d.OneOfField),
		Or(
			Equal(d.OrLeftField, "left").
				WithFieldTrace(d, &d.OrLeftField),
			Equal(d.OrRightField, "right").
				WithFieldTrace(d, &d.OrRightField),
		),
		And(
			EvaluateAll,
			Equal(d.AndLeftField, "left").
				WithFieldTrace(d, &d.AndLeftField),
			Equal(d.AndRightField, "right").
				WithFieldTrace(d, &d.AndRightField),
		),
	}
}

// StrFieldNeedsToBeAbc is an example for a custom constraint.
func (d *exampleDoc) strFieldNeedsToBeAbc() *Constraint {
	return &Constraint{
		Satisfied: func() *TreeError {
			if d.StrField != "abc" {
				return NewErrorTree(
					fmt.Errorf("%s must be abc", d.StrField),
				)
			}
			return nil
		},
	}
}

func TestOverrideConstraints(t *testing.T) {
	overrideConstraints := func(t *testing.T, wantCalled bool) func() []*Constraint {
		return func() []*Constraint {
			if !wantCalled {
				t.Fatal("overrideConstraints should not be called")
			}
			return []*Constraint{}
		}
	}

	testCases := map[string]struct {
		doc                exampleDocToOverride
		overrideFunc       func() []*Constraint
		wantOverrideCalled bool
		wantErr            bool
	}{
		"override constraints": {
			doc:                exampleDocToOverride{},
			overrideFunc:       overrideConstraints(t, true),
			wantOverrideCalled: true,
		},
		"do not override constraints": {
			doc: exampleDocToOverride{
				calledDocConstraints: true,
			},
			overrideFunc:       nil,
			wantOverrideCalled: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			validator := NewValidator()
			err := validator.Validate(&tc.doc, ValidateOptions{
				OverrideConstraints: tc.overrideFunc,
			})

			if tc.wantErr {
				require.Error(err)
			} else {
				require.NoError(err)
				if tc.wantOverrideCalled {
					assert.Equal(tc.doc.calledDocConstraints, false)
				}
			}
		})
	}
}

type exampleDocToOverride struct {
	calledDocConstraints bool
}

func (d *exampleDocToOverride) Constraints() []*Constraint {
	d.calledDocConstraints = true
	return []*Constraint{}
}
