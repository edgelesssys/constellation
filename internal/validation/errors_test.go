package validation

import (
	"fmt"
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

	err := NewValidationError(st, &st.OtherField, "", assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.otherField: %s", assert.AnError))
}

func TestNewValidationErrorSingleFieldPtr(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		PointerField:  new(int),
	}

	err := NewValidationError(st, &st.PointerField, "", assert.AnError)
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

	err := NewValidationError(st, &st.DoublePointerField, "", assert.AnError)
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

	err := NewValidationError(st, &inexistentField, "", assert.AnError)
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

	err := NewValidationError(st, &st.NestedField.OtherField, "", assert.AnError)
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

	err := NewValidationError(st, &st.NestedField.PointerField, "", assert.AnError)
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

	err := NewValidationError(st, &st.NestedPointerField.OtherField, "", assert.AnError)
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

	err := NewValidationError(st, &st.NestedField.NestedField.OtherField, "", assert.AnError)
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

	err := NewValidationError(st, &st.NestedField.NestedField.OtherField, "", assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.nestedField.nestedField.otherField: %s", assert.AnError))
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

	err := NewValidationError(st, &st.NestedField.NestedField.OtherField, "", assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc.nestedField.nestedField.otherField: %s", assert.AnError))
}

// Tests for slices / arrays

func TestNewValidationErrorPrimitiveSlice(t *testing.T) {
	st := &SliceErrorTestDoc{
		PrimitiveSlice: []string{"abc", "def"},
	}

	err := NewValidationError(st, &st.PrimitiveSlice[1], "", assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating SliceErrorTestDoc.primitiveSlice[1]: %s", assert.AnError))
}

func TestNewValidationErrorPrimitiveArray(t *testing.T) {
	st := &SliceErrorTestDoc{
		PrimitiveArray: [3]int{1, 2, 3},
	}

	err := NewValidationError(st, &st.PrimitiveArray[1], "", assert.AnError)
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

	err := NewValidationError(st, &st.StructSlice[1].OtherField, "", assert.AnError)
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

	err := NewValidationError(st, &st.StructArray[1].OtherField, "", assert.AnError)
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

	err := NewValidationError(st, &st.StructPointerSlice[1].OtherField, "", assert.AnError)
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

	err := NewValidationError(st, &st.StructPointerArray[1].OtherField, "", assert.AnError)
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

	err := NewValidationError(st, &st.PrimitiveSliceSlice[1][1], "", assert.AnError)
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

	err := NewValidationError(st, &st.PrimitiveMap, "ghi", assert.AnError)
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

	err := NewValidationError(st, &st.StructPointerMap["ghi"].OtherField, "", assert.AnError)
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

	err := NewValidationError(st, st.NestedPointerMap["jkl"], "mno", assert.AnError)
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

	err := NewValidationError(st, st, "", assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating ErrorTestDoc: %s", assert.AnError))
}

func TestNewValidationErrorUntaggedField(t *testing.T) {
	st := &ErrorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NoTagField:    123,
	}

	err := NewValidationError(st, &st.NoTagField, "", assert.AnError)
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

	err := NewValidationError(st, &st.OnlyYamlKey, "", assert.AnError)
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
