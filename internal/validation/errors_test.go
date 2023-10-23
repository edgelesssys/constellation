package validation

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for primitive / shallow fields

func TestNewValidationErrorSingleField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
	}

	doc, field := references(t, st, &st.OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.otherField: %s", assert.AnError))
}

func TestNewValidationErrorSingleFieldPtr(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		PointerField:  new(int),
	}

	doc, field := references(t, st, &st.PointerField, "")
	err := NewValidationError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.pointerField: %s", assert.AnError))
}

func TestNewValidationErrorSingleFieldDoublePtr(t *testing.T) {
	intp := new(int)
	st := &ErrorTestDoc{
		ExportedField:      "abc",
		OtherField:         42,
		DoublePointerField: &intp,
	}

	doc, field := references(t, st, &st.DoublePointerField, "")
	err := NewValidationError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.doublePointerField: %s", assert.AnError))
}

func TestNewValidationErrorSingleFieldInexistent(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		PointerField:  new(int),
	}

	inexistentField := 123

	doc, field := references(t, st, &inexistentField, "")
	err := NewValidationError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot find path to field: cannot traverse anymore")
}

// Tests for nested structs

func TestNewValidationErrorNestedField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: NestedErrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
		},
	}

	doc, field := references(t, st, &st.NestedField.OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.nestedField.otherField: %s", assert.AnError))
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

	doc, field := references(t, st, &st.NestedField.PointerField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.nestedField.pointerField: %s", assert.AnError))
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

	doc, field := references(t, st, &st.NestedPointerField.OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.nestedPointerField.otherField: %s", assert.AnError))
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

	doc, field := references(t, st, &st.NestedField.NestedField.OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.nestedField.nestedField.otherField: %s", assert.AnError))
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

	doc, field := references(t, st, &st.NestedField.NestedPointerField.OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.nestedField.nestedPointerField.otherField: %s", assert.AnError))
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

	doc, field := references(t, st, &st.NestedPointerField.NestedPointerField.OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.nestedPointerField.nestedPointerField.otherField: %s", assert.AnError))
}

// Tests for slices / arrays

func TestNewValidationErrorPrimitiveSlice(t *testing.T) {
	st := &SliceErrorTestDoc{
		PrimitiveSlice: []string{"abc", "def"},
	}

	doc, field := references(t, st, &st.PrimitiveSlice[1], "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating SliceErrorTestDoc.primitiveSlice[1]: %s", assert.AnError))
}

func TestNewValidationErrorPrimitiveArray(t *testing.T) {
	st := &SliceErrorTestDoc{
		PrimitiveArray: [3]int{1, 2, 3},
	}

	doc, field := references(t, st, &st.PrimitiveArray[1], "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating SliceErrorTestDoc.primitiveArray[1]: %s", assert.AnError))
}

func TestNewValidationErrorStructSlice(t *testing.T) {
	st := &SliceErrorTestDoc{
		StructSlice: []ErrorTestDoc{
			{
				ExportedField: "abc",
				OtherField:    123,
			},
			{
				ExportedField: "def",
				OtherField:    456,
			},
		},
	}

	doc, field := references(t, st, &st.StructSlice[1].OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating SliceErrorTestDoc.structSlice[1].otherField: %s", assert.AnError))
}

func TestNewValidationErrorStructArray(t *testing.T) {
	st := &SliceErrorTestDoc{
		StructArray: [3]ErrorTestDoc{
			{
				ExportedField: "abc",
				OtherField:    123,
			},
			{
				ExportedField: "def",
				OtherField:    456,
			},
		},
	}

	doc, field := references(t, st, &st.StructArray[1].OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating SliceErrorTestDoc.structArray[1].otherField: %s", assert.AnError))
}

func TestNewValidationErrorStructPointerSlice(t *testing.T) {
	st := &SliceErrorTestDoc{
		StructPointerSlice: []*ErrorTestDoc{
			{
				ExportedField: "abc",
				OtherField:    123,
			},
			{
				ExportedField: "def",
				OtherField:    456,
			},
		},
	}

	doc, field := references(t, st, &st.StructPointerSlice[1].OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating SliceErrorTestDoc.structPointerSlice[1].otherField: %s", assert.AnError))
}

func TestNewValidationErrorStructPointerArray(t *testing.T) {
	st := &SliceErrorTestDoc{
		StructPointerArray: [3]*ErrorTestDoc{
			{
				ExportedField: "abc",
				OtherField:    123,
			},
			{
				ExportedField: "def",
				OtherField:    456,
			},
		},
	}

	doc, field := references(t, st, &st.StructPointerArray[1].OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating SliceErrorTestDoc.structPointerArray[1].otherField: %s", assert.AnError))
}

func TestNewValidationErrorPrimitiveSliceSlice(t *testing.T) {
	st := &SliceErrorTestDoc{
		PrimitiveSliceSlice: [][]string{
			{"abc", "def"},
			{"ghi", "jkl"},
		},
	}

	doc, field := references(t, st, &st.PrimitiveSliceSlice[1][1], "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating SliceErrorTestDoc.primitiveSliceSlice[1][1]: %s", assert.AnError))
}

// Tests for maps

func TestNewValidationErrorPrimitiveMap(t *testing.T) {
	st := &MapErrorTestDoc{
		PrimitiveMap: map[string]string{
			"abc": "def",
			"ghi": "jkl",
		},
	}

	doc, field := references(t, st, &st.PrimitiveMap, "ghi")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating MapErrorTestDoc.primitiveMap[\"ghi\"]: %s", assert.AnError))
}

func TestNewValidationErrorStructPointerMap(t *testing.T) {
	st := &MapErrorTestDoc{
		StructPointerMap: map[string]*ErrorTestDoc{
			"abc": {
				ExportedField: "abc",
				OtherField:    123,
			},
			"ghi": {
				ExportedField: "ghi",
				OtherField:    456,
			},
		},
	}

	doc, field := references(t, st, &st.StructPointerMap["ghi"].OtherField, "")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating MapErrorTestDoc.structPointerMap[\"ghi\"].otherField: %s", assert.AnError))
}

func TestNewValidationErrorNestedPrimitiveMap(t *testing.T) {
	st := &MapErrorTestDoc{
		NestedPointerMap: map[string]*map[string]string{
			"abc": {
				"def": "ghi",
			},
			"jkl": {
				"mno": "pqr",
			},
		},
	}

	doc, field := references(t, st, st.NestedPointerMap["jkl"], "mno")
	err := NewValidationError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating MapErrorTestDoc.nestedPointerMap[\"jkl\"][\"mno\"]: %s", assert.AnError))
}

// Special cases

func TestNewValidationErrorTopLevelIsNeedle(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
	}

	doc, field := references(t, st, st, "")
	err := NewValidationError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc: %s", assert.AnError))
}

func TestNewValidationErrorUntaggedField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NoTagField:    123,
	}

	doc, field := references(t, st, &st.NoTagField, "")
	err := NewValidationError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.NoTagField: %s", assert.AnError))
}

func TestNewValidationErrorOnlyYamlTaggedField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NoTagField:    123,
		OnlyYamlKey:   "abc",
	}

	doc, field := references(t, st, &st.OnlyYamlKey, "")
	err := NewValidationError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.onlyYamlKey: %s", assert.AnError))
}

type ErrorTestDoc struct {
	ExportedField      string              `json:"exportedField" yaml:"exportedField"`
	OtherField         int                 `json:"otherField" yaml:"otherField"`
	PointerField       *int                `json:"pointerField" yaml:"pointerField"`
	DoublePointerField **int               `json:"doublePointerField" yaml:"doublePointerField"`
	NestedField        NestedErrorTestDoc  `json:"nestedField" yaml:"nestedField"`
	NestedPointerField *NestedErrorTestDoc `json:"nestedPointerField" yaml:"nestedPointerField"`
	NoTagField         int
	OnlyYamlKey        string `yaml:"onlyYamlKey"`
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

type SliceErrorTestDoc struct {
	PrimitiveSlice      []string         `json:"primitiveSlice" yaml:"primitiveSlice"`
	PrimitiveArray      [3]int           `json:"primitiveArray" yaml:"primitiveArray"`
	StructSlice         []ErrorTestDoc   `json:"structSlice" yaml:"structSlice"`
	StructArray         [3]ErrorTestDoc  `json:"structArray" yaml:"structArray"`
	StructPointerSlice  []*ErrorTestDoc  `json:"structPointerSlice" yaml:"structPointerSlice"`
	StructPointerArray  [3]*ErrorTestDoc `json:"structPointerArray" yaml:"structPointerArray"`
	PrimitiveSliceSlice [][]string       `json:"primitiveSliceSlice" yaml:"primitiveSliceSlice"`
}

type MapErrorTestDoc struct {
	PrimitiveMap     map[string]string             `json:"primitiveMap" yaml:"primitiveMap"`
	StructPointerMap map[string]*ErrorTestDoc      `json:"structPointerMap" yaml:"structPointerMap"`
	NestedPointerMap map[string]*map[string]string `json:"nestedPointerMap" yaml:"nestedPointerMap"`
}

// references returns referenceableValues for the given doc and field for testing purposes.
func references(t *testing.T, doc, field any, mapKey string) (haystack, needle referenceableValue) {
	t.Helper()
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
	return docRef, fieldRef
}
