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
				strField: "abc",
			},
			opts: ValidateOptions{},
		},
		"invalid": {
			doc: &exampleDoc{
				strField: "def",
			},
			wantErr: true,
			errAssertion: func(assert *assert.Assertions, err error) bool {
				return assert.Contains(err.Error(), "strField must be abc")
			},
			opts: ValidateOptions{},
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
	strField  string
	numField  int
	nested    nestedExampleDoc
	nestedPtr *nestedExampleDoc
}

type nestedExampleDoc struct {
	strField string
	numField int
}

func (d *exampleDoc) Constraints() []Constraint {
	return []Constraint{
		d.strFieldNeedsToBeAbc,
		MatchRegex(d.strField, "^[a-z]+$"),
		Equal(d.numField, 42),
	}
}

// StrFieldNeedsToBeAbc is an example for a custom constraint.
func (d *exampleDoc) strFieldNeedsToBeAbc() (bool, error) {
	if d.strField != "abc" {
		return false, fmt.Errorf("%s must be abc", d.strField)
	}
	return true, nil
}
