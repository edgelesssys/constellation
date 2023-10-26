/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package validation

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorFormatting(t *testing.T) {
	err := &ErrorTree{
		path:     "path",
		err:      fmt.Errorf("error"),
		children: []*ErrorTree{},
	}

	assert.Equal(t, "validating path: error", err.Error())

	err.children = append(err.children, &ErrorTree{
		path:     "child",
		err:      fmt.Errorf("child error"),
		children: []*ErrorTree{},
	})

	assert.Equal(t, "validating path: error\n  validating child: child error", err.Error())

	err.children = append(err.children, &ErrorTree{
		path: "child2",
		err:  fmt.Errorf("child2 error"),
		children: []*ErrorTree{
			{
				path:     "child2child",
				err:      fmt.Errorf("child2child error"),
				children: []*ErrorTree{},
			},
		},
	})
	assert.Equal(t, "validating path: error\n  validating child: child error\n  validating child2: child2 error\n    validating child2child: child2child error", err.Error())
}

// Tests for primitive / shallow fields

func TestNewValidationErrorSingleField(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
	}

	doc, field := references(t, st, &st.OtherField, "")
	err := newTraceError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.otherField: %s", assert.AnError))
}

func TestNewValidationErrorSingleFieldPtr(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		PointerField:  new(int),
	}

	doc, field := references(t, st, &st.PointerField, "")
	err := newTraceError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.pointerField: %s", assert.AnError))
}

func TestNewValidationErrorSingleFieldDoublePtr(t *testing.T) {
	intp := new(int)
	st := &errorTestDoc{
		ExportedField:      "abc",
		OtherField:         42,
		DoublePointerField: &intp,
	}

	doc, field := references(t, st, &st.DoublePointerField, "")
	err := newTraceError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.doublePointerField: %s", assert.AnError))
}

func TestNewValidationErrorSingleFieldInexistent(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		PointerField:  new(int),
	}

	inexistentField := 123

	doc, field := references(t, st, &inexistentField, "")
	err := newTraceError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot find path to field: cannot traverse anymore")
}

// Tests for nested structs

func TestNewValidationErrorNestedField(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: nestederrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
		},
	}

	doc, field := references(t, st, &st.NestedField.OtherField, "")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.nestedField.otherField: %s", assert.AnError))
}

func TestNewValidationErrorPointerInNestedField(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: nestederrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
			PointerField:  new(int),
		},
	}

	doc, field := references(t, st, &st.NestedField.PointerField, "")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.nestedField.pointerField: %s", assert.AnError))
}

func TestNewValidationErrorNestedFieldPtr(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: nestederrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
		},
		NestedPointerField: &nestederrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
		},
	}

	doc, field := references(t, st, &st.NestedPointerField.OtherField, "")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.nestedPointerField.otherField: %s", assert.AnError))
}

func TestNewValidationErrorNestedNestedField(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: nestederrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
			NestedField: nestedNestederrorTestDoc{
				ExportedField: "nested",
				OtherField:    123,
			},
		},
	}

	doc, field := references(t, st, &st.NestedField.NestedField.OtherField, "")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.nestedField.nestedField.otherField: %s", assert.AnError))
}

func TestNewValidationErrorNestedNestedFieldPtr(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedField: nestederrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
			NestedPointerField: &nestedNestederrorTestDoc{
				ExportedField: "nested",
				OtherField:    123,
			},
		},
	}

	doc, field := references(t, st, &st.NestedField.NestedPointerField.OtherField, "")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.nestedField.nestedPointerField.otherField: %s", assert.AnError))
}

func TestNewValidationErrorNestedPtrNestedFieldPtr(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NestedPointerField: &nestederrorTestDoc{
			ExportedField: "nested",
			OtherField:    123,
			NestedPointerField: &nestedNestederrorTestDoc{
				ExportedField: "nested",
				OtherField:    123,
			},
		},
	}

	doc, field := references(t, st, &st.NestedPointerField.NestedPointerField.OtherField, "")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.nestedPointerField.nestedPointerField.otherField: %s", assert.AnError))
}

// Tests for slices / arrays

func TestNewValidationErrorPrimitiveSlice(t *testing.T) {
	st := &sliceErrorTestDoc{
		PrimitiveSlice: []string{"abc", "def"},
	}

	doc, field := references(t, st, &st.PrimitiveSlice[1], "")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating sliceErrorTestDoc.primitiveSlice[1]: %s", assert.AnError))
}

func TestNewValidationErrorPrimitiveArray(t *testing.T) {
	st := &sliceErrorTestDoc{
		PrimitiveArray: [3]int{1, 2, 3},
	}

	doc, field := references(t, st, &st.PrimitiveArray[1], "")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating sliceErrorTestDoc.primitiveArray[1]: %s", assert.AnError))
}

func TestNewValidationErrorStructSlice(t *testing.T) {
	st := &sliceErrorTestDoc{
		StructSlice: []errorTestDoc{
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
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating sliceErrorTestDoc.structSlice[1].otherField: %s", assert.AnError))
}

func TestNewValidationErrorStructArray(t *testing.T) {
	st := &sliceErrorTestDoc{
		StructArray: [3]errorTestDoc{
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
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating sliceErrorTestDoc.structArray[1].otherField: %s", assert.AnError))
}

func TestNewValidationErrorStructPointerSlice(t *testing.T) {
	st := &sliceErrorTestDoc{
		StructPointerSlice: []*errorTestDoc{
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
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating sliceErrorTestDoc.structPointerSlice[1].otherField: %s", assert.AnError))
}

func TestNewValidationErrorStructPointerArray(t *testing.T) {
	st := &sliceErrorTestDoc{
		StructPointerArray: [3]*errorTestDoc{
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
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating sliceErrorTestDoc.structPointerArray[1].otherField: %s", assert.AnError))
}

func TestNewValidationErrorPrimitiveSliceSlice(t *testing.T) {
	st := &sliceErrorTestDoc{
		PrimitiveSliceSlice: [][]string{
			{"abc", "def"},
			{"ghi", "jkl"},
		},
	}

	doc, field := references(t, st, &st.PrimitiveSliceSlice[1][1], "")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating sliceErrorTestDoc.primitiveSliceSlice[1][1]: %s", assert.AnError))
}

// Tests for maps

func TestNewValidationErrorPrimitiveMap(t *testing.T) {
	st := &mapErrorTestDoc{
		PrimitiveMap: map[string]string{
			"abc": "def",
			"ghi": "jkl",
		},
	}

	doc, field := references(t, st, &st.PrimitiveMap, "ghi")
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating mapErrorTestDoc.primitiveMap[\"ghi\"]: %s", assert.AnError))
}

func TestNewValidationErrorStructPointerMap(t *testing.T) {
	st := &mapErrorTestDoc{
		StructPointerMap: map[string]*errorTestDoc{
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
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating mapErrorTestDoc.structPointerMap[\"ghi\"].otherField: %s", assert.AnError))
}

func TestNewValidationErrorNestedPrimitiveMap(t *testing.T) {
	st := &mapErrorTestDoc{
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
	err := newTraceError(doc, field, assert.AnError)
	t.Log(err)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating mapErrorTestDoc.nestedPointerMap[\"jkl\"][\"mno\"]: %s", assert.AnError))
}

// Special cases

func TestNewValidationErrorTopLevelIsNeedle(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
	}

	doc, field := references(t, st, st, "")
	err := newTraceError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc: %s", assert.AnError))
}

func TestNewValidationErrorUntaggedField(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NoTagField:    123,
	}

	doc, field := references(t, st, &st.NoTagField, "")
	err := newTraceError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.NoTagField: %s", assert.AnError))
}

func TestNewValidationErrorOnlyYamlTaggedField(t *testing.T) {
	st := &errorTestDoc{
		ExportedField: "abc",
		OtherField:    42,
		NoTagField:    123,
		OnlyYamlKey:   "abc",
	}

	doc, field := references(t, st, &st.OnlyYamlKey, "")
	err := newTraceError(doc, field, assert.AnError)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("validating errorTestDoc.onlyYamlKey: %s", assert.AnError))
}

type errorTestDoc struct {
	ExportedField      string              `json:"exportedField" yaml:"exportedField"`
	OtherField         int                 `json:"otherField" yaml:"otherField"`
	PointerField       *int                `json:"pointerField" yaml:"pointerField"`
	DoublePointerField **int               `json:"doublePointerField" yaml:"doublePointerField"`
	NestedField        nestederrorTestDoc  `json:"nestedField" yaml:"nestedField"`
	NestedPointerField *nestederrorTestDoc `json:"nestedPointerField" yaml:"nestedPointerField"`
	NoTagField         int
	OnlyYamlKey        string `yaml:"onlyYamlKey"`
}

type nestederrorTestDoc struct {
	ExportedField      string                    `json:"exportedField" yaml:"exportedField"`
	OtherField         int                       `json:"otherField" yaml:"otherField"`
	PointerField       *int                      `json:"pointerField" yaml:"pointerField"`
	NestedField        nestedNestederrorTestDoc  `json:"nestedField" yaml:"nestedField"`
	NestedPointerField *nestedNestederrorTestDoc `json:"nestedPointerField" yaml:"nestedPointerField"`
}

type nestedNestederrorTestDoc struct {
	ExportedField string `json:"exportedField" yaml:"exportedField"`
	OtherField    int    `json:"otherField" yaml:"otherField"`
	PointerField  *int   `json:"pointerField" yaml:"pointerField"`
}

type sliceErrorTestDoc struct {
	PrimitiveSlice      []string         `json:"primitiveSlice" yaml:"primitiveSlice"`
	PrimitiveArray      [3]int           `json:"primitiveArray" yaml:"primitiveArray"`
	StructSlice         []errorTestDoc   `json:"structSlice" yaml:"structSlice"`
	StructArray         [3]errorTestDoc  `json:"structArray" yaml:"structArray"`
	StructPointerSlice  []*errorTestDoc  `json:"structPointerSlice" yaml:"structPointerSlice"`
	StructPointerArray  [3]*errorTestDoc `json:"structPointerArray" yaml:"structPointerArray"`
	PrimitiveSliceSlice [][]string       `json:"primitiveSliceSlice" yaml:"primitiveSliceSlice"`
}

type mapErrorTestDoc struct {
	PrimitiveMap     map[string]string             `json:"primitiveMap" yaml:"primitiveMap"`
	StructPointerMap map[string]*errorTestDoc      `json:"structPointerMap" yaml:"structPointerMap"`
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
