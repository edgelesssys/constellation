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

func TestValidate(t *testing.T) {
	testCases := map[string]struct {
		doc          Validatable
		opts         ValidateOptions
		wantErr      bool
		errAssertion func(*assert.Assertions, error) bool
	}{
		"valid": {
			doc: &exampleDoc{
				StrField: "abc",
				NumField: 42,
				MapField: &map[string]string{
					"empty": "",
				},
				NotEmptyField:   "certainly not.",
				MatchRegexField: "abc",
			},
			opts: ValidateOptions{},
		},
		"strField is not abc": {
			doc: &exampleDoc{
				StrField: "def",
				NumField: 42,
				MapField: &map[string]string{
					"empty": "",
				},
				NotEmptyField:   "certainly not.",
				MatchRegexField: "abc",
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.strField: def must be abc")
			},
			opts: ValidateOptions{},
		},
		"numField is not 42": {
			doc: &exampleDoc{
				StrField: "abc",
				NumField: 43,
				MapField: &map[string]string{
					"empty": "",
				},
				NotEmptyField:   "certainly not.",
				MatchRegexField: "abc",
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.numField: 43 must be equal to 42")
			},
		},
		"multiple errors": {
			doc: &exampleDoc{
				StrField: "def",
				NumField: 43,
				MapField: &map[string]string{
					"empty": "",
				},
				NotEmptyField:   "certainly not.",
				MatchRegexField: "abc",
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.strField: def must be abc") &&
					assert.Contains(err.Error(), "validating exampleDoc.numField: 43 must be equal to 42")
			},
			opts: ValidateOptions{},
		},
		"multiple errors, fail fast": {
			doc: &exampleDoc{
				StrField: "def",
				NumField: 43,
				MapField: &map[string]string{
					"empty": "",
				},
				NotEmptyField:   "certainly not.",
				MatchRegexField: "abc",
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.strField: def must be abc")
			},
			opts: ValidateOptions{
				FailFast: true,
			},
		},
		"map field is not empty": {
			doc: &exampleDoc{
				StrField: "abc",
				NumField: 42,
				MapField: &map[string]string{
					"empty": "haha!",
				},
				NotEmptyField:   "certainly not.",
				MatchRegexField: "abc",
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.mapField[\"empty\"]: haha! must be empty")
			},
			opts: ValidateOptions{
				FailFast: true,
			},
		},
		"empty field is not empty": {
			doc: &exampleDoc{
				StrField: "abc",
				NumField: 42,
				MapField: &map[string]string{
					"empty": "",
				},
				NotEmptyField:   "",
				MatchRegexField: "abc",
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.notEmptyField:  must not be empty")
			},
			opts: ValidateOptions{
				FailFast: true,
			},
		},
		"regex doesnt match": {
			doc: &exampleDoc{
				StrField: "abc",
				NumField: 42,
				MapField: &map[string]string{
					"empty": "",
				},
				NotEmptyField:   "certainly not!",
				MatchRegexField: "dontmatch",
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "validating exampleDoc.matchRegexField: dontmatch must match the pattern ^a.c$")
			},
			opts: ValidateOptions{
				FailFast: true,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			err := NewValidator().Validate(tc.doc, tc.opts)
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
}

// Constraints implements the Validatable interface.
func (d *exampleDoc) Constraints() []Constraint {
	mapField := *(d.MapField)

	return []Constraint{
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
	}
}

// StrFieldNeedsToBeAbc is an example for a custom constraint.
func (d *exampleDoc) strFieldNeedsToBeAbc() *Constraint {
	return &Constraint{
		Satisfied: func() error {
			if d.StrField != "abc" {
				return fmt.Errorf("%s must be abc", d.StrField)
			}
			return nil
		},
	}
}
