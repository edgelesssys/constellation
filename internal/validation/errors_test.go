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

func TestNewValidationErrorSingleFieldDoublePtr(t *testing.T) {
	intp := new(int)
	st := &ErrorTestDoc{
		ExportedField:      "abc",
		OtherField:         42,
		DoublePointerField: &intp,
	}

	err := NewValidationError(st, &st.DoublePointerField, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating doublePointerField: %s", assert.AnError))
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
		NestedField: NestedErrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
		},
	}

	err := NewValidationError(st, &st.NestedField.OtherField, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating nestedField.otherField: %s", assert.AnError))
}

func TestNewValidationErrorPointerInNestedField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: NestedErrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
			PointerField:  new(int),
		},
	}

	err := NewValidationError(st, &st.NestedField.PointerField, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating nestedField.pointerField: %s", assert.AnError))
}

func TestNewValidationErrorNestedFieldPtr(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: NestedErrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
		},
		NestedPointerField: &NestedErrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
		},
	}

	err := NewValidationError(st, &st.NestedPointerField.OtherField, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating nestedPointerField.otherField: %s", assert.AnError))
}

func TestNewValidationErrorNestedNestedField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: NestedErrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
			NestedField: NestedNestedErrorTestDoc{
				ExportedField: "nested",
				OtherField:    123,
			},
		},
	}

	err := NewValidationError(st, &st.NestedField.NestedField.OtherField, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating nestedField.nestedField.otherField: %s", assert.AnError))
}

func TestNewValidationErrorNestedNestedFieldPtr(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: NestedErrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
			NestedPointerField: &NestedNestedErrorTestDoc{
				ExportedField: "nested",
				OtherField:    123,
			},
		},
	}

	err := NewValidationError(st, &st.NestedField.NestedField.OtherField, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating nestedField.nestedField.otherField: %s", assert.AnError))
}

func TestNewValidationErrorNestedPtrNestedFieldPtr(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedPointerField: &NestedErrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
			NestedPointerField: &NestedNestedErrorTestDoc{
				ExportedField: "nested",
				OtherField:    123,
			},
		},
	}

	err := NewValidationError(st, &st.NestedField.NestedField.OtherField, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating nestedField.nestedField.otherField: %s", assert.AnError))
}

type ErrorTestDoc struct {
	ExportedField      string              `json:"exportedField" yaml:"exportedField"`
	OtherField         int                 `json:"otherField" yaml:"otherField"`
	PointerField       *int                `json:"pointerField" yaml:"pointerField"`
	DoublePointerField **int               `json:"doublePointerField" yaml:"doublePointerField"`
	NestedField        NestedErrorTestDoc  `json:"nestedField" yaml:"nestedField"`
	NestedPointerField *NestedErrorTestDoc `json:"nestedPointerField" yaml:"nestedPointerField"`
}

type NestedErrorTestDoc struct {
	ExportedField      string                    `json:"exportedField" yaml:"exportedField"`
	OtherField         int                       `json:"otherField" yaml:"otherField"`
	PointerField       *int                      `json:"pointerField" yaml:"pointerField"`
	NestedField        NestedNestedErrorTestDoc  `json:"nestedField" yaml:"nestedField"`
	NestedPointerField *NestedNestedErrorTestDoc `json:"nestedPointerField" yaml:"nestedPointerField"`
}

type NestedNestedErrorTestDoc struct {
	ExportedField string `json:"exportedField" yaml:"exportedField"`
	OtherField    int    `json:"otherField" yaml:"otherField"`
	PointerField  *int   `json:"pointerField" yaml:"pointerField"`
}
