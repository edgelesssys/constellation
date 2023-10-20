package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationErrorSingleField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
	}

	err := NewValidationError(st, &st.OtherField, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating otherField: %s", assert.AnError))
}

func TestNewValidationErrorSingleFieldPtr(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		PointerField:  new(int),
	}

	err := NewValidationError(st, &st.PointerField, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating pointerField: %s", assert.AnError))
}

type ErrorTestDoc struct {
	ExportedField string `json:"exportedField" yaml:"exportedField"`
	OtherField    int    `json:"otherField" yaml:"otherField"`
	PointerField  *int   `json:"pointerField" yaml:"pointerField"`
}
