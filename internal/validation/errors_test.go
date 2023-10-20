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

func TestNewValidationErrorSingleFieldInexistent(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		PointerField:  new(int),
	}

	inexistentField := 123

	err := NewValidationError(st, &inexistentField, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot find path to field: cannot traverse anymore")
}

func TestNewValidationErrorNestedField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedErrorTestDoc: NestedErrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
		},
	}

	err := NewValidationError(st, &st.NestedErrorTestDoc.OtherField, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating nestedErrorTestDoc.otherField: %s", assert.AnError))
}

type ErrorTestDoc struct {
	ExportedField      string             `json:"exportedField" yaml:"exportedField"`
	OtherField         int                `json:"otherField" yaml:"otherField"`
	PointerField       *int               `json:"pointerField" yaml:"pointerField"`
	NestedErrorTestDoc NestedErrorTestDoc `json:"nestedErrorTestDoc" yaml:"nestedErrorTestDoc"`
}

type NestedErrorTestDoc struct {
	ExportedField string `json:"exportedField" yaml:"exportedField"`
	OtherField    int    `json:"otherField" yaml:"otherField"`
	PointerField  *int   `json:"pointerField" yaml:"pointerField"`
}
