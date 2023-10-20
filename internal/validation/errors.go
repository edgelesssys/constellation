package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type ValidationError struct {
	Path string
	Err  error
}

// NewValidationError creates a new ValidationError.
//
// To find the path to the exported field that failed validation, it traverses the
// top level struct recursively until it finds a field that matches the
// reference to the field that failed validation.
func NewValidationError(topLevelStruct any, field any, errMsg error) *ValidationError {
	path, err := getDocumentPath(topLevelStruct, field)
	if err != nil {
		return &ValidationError{
			Path: "unknown",
			Err:  fmt.Errorf("cannot find path to field: %v. original error: %w", err, errMsg),
		}
	}

	return &ValidationError{
		Path: path,
		Err:  errMsg,
	}
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validating %s: %s", e.Path, e.Err)
}

// Unwrap implements the error interface.
func (e *ValidationError) Unwrap() error {
	return e.Err
}

// getDocumentPath finds the JSON / YAML path of field in doc.
func getDocumentPath(doc any, field any) (string, error) {
	// we only want to dereference the needle once to dereference the pointer
	// used to pass it to the function without losing reference to it, as the
	// needle could be an arbitrarily long chain of pointers. The same
	// applies to the haystack.
	derefedNeedle := pointerDeref(reflect.ValueOf(field))
	needleRef := referenceableValue{
		value: derefedNeedle,
		addr:  derefedNeedle.UnsafeAddr(),
		_type: derefedNeedle.Type(),
	}
	derefedHaystack := pointerDeref(reflect.ValueOf(doc))
	haystackRef := referenceableValue{
		value: derefedHaystack,
		addr:  derefedHaystack.UnsafeAddr(),
		_type: derefedHaystack.Type(),
	}

	// traverse the top level struct (i.e. the "haystack") until addr (i.e. the "needle") is found
	return traverse(haystackRef, needleRef, []string{})
}

// traverse reverses haystack recursively until it finds a field that matches
// the reference in needle.
//
// If it traverses a level down, it
// appends the name of the struct tag of the field to path.
//
// When a field matches the reference to the given field, it returns the
// path to the field, joined with ".".
func traverse(haystack referenceableValue, needle referenceableValue, path []string) (string, error) {
	// recursion anchor: doc is the field we are looking for.
	// Join the path and return.
	if foundNeedle(haystack.addr, haystack._type, needle.addr, needle._type) {
		return strings.Join(path, "."), nil
	}

	kind := haystack._type.Kind()
	switch kind {
	case reflect.Struct:
		// Traverse all visible struct fields.
		for _, field := range reflect.VisibleFields(haystack._type) {
			// skip unexported fields
			if field.IsExported() {
				fieldVal := haystack.value.FieldByName(field.Name)
				fieldAddr := haystack.addr + field.Offset
				if canTraverse(fieldVal) {
					// When a field is not the needle and cannot be traversed further,
					// a errCannotTraverse is returned. Therefore, we only want to handle
					// the case where the field is the needle.
					newHaystack := referenceableValue{
						value: fieldVal,
						addr:  fieldAddr,
						_type: field.Type,
					}
					if path, err := traverse(newHaystack, needle, appendByStructTag(path, field)); err == nil {
						return path, nil
					}
				}
				if foundNeedle(fieldAddr, field.Type, needle.addr, needle._type) {
					return strings.Join(appendByStructTag(path, field), "."), nil
				}
			}
		}
		// case reflect.Slice, reflect.Array:
		// 	// Traverse slice / Array elements
		// 	for i := 0; i < derefedHaystack.Len(); i++ {
		// 		// see struct case
		// 		if path, err := traverse(derefedHaystack.Index(i), needleAddr, needleType, append(path, fmt.Sprintf("%d", i))); err == nil {
		// 			return path, nil
		// 		}
		// 	}
		// case reflect.Map:
		// 	// Traverse map elements
		// 	for _, key := range derefedHaystack.MapKeys() {
		// 		// see struct case
		// 		if path, err := traverse(derefedHaystack.MapIndex(key), needleAddr, needleType, append(path, key.String())); err == nil {
		// 			return path, nil
		// 		}
		// 	}
	}
	// Primitive type, but not the value we are looking for.
	// Return a
	return "", errCannotTraverse
}

// referenceableValue is a type that can be passed as any (thus being copied) without losing the reference to the actual value.
type referenceableValue struct {
	value reflect.Value
	_type reflect.Type
	addr  uintptr
}

// errCannotTraverse is returned when a field cannot be traversed further.
var errCannotTraverse = errors.New("cannot traverse anymore")

// appendByStructTag appends the name of the JSON / YAML struct tag of field to path.
// If no struct tag is present, path is returned unchanged.
func appendByStructTag(path []string, field reflect.StructField) []string {
	switch {
	case field.Tag.Get("json") != "":
		return append(path, field.Tag.Get("json"))
	case field.Tag.Get("yaml") != "":
		return append(path, field.Tag.Get("yaml"))
	}
	return path
}

// recPointerDeref recursively dereferences pointers and unpacks interfaces until a non-pointer value is found.
func recPointerDeref(val reflect.Value) reflect.Value {
	switch val.Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Interface:
		return recPointerDeref(val.Elem())
	}
	return val
}

// pointerDeref dereferences pointers and unpacks interfaces.
// If the value is not a pointer, it is returned unchanged
func pointerDeref(val reflect.Value) reflect.Value {
	switch val.Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Interface:
		return val.Elem()
	}
	return val
}

/*
canTraverse whether a value can be further traversed.

For pointer types, false is returned.
*/
func canTraverse(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
		return true
	}
	return false
}

/*
foundNeedle returns whether the given value is the needle.

It does so by comparing the address and type of the value to the address and type of the needle.
The comparison of types is necessary because the first value of a struct has the same address as the struct itself.
*/
func foundNeedle(addr uintptr, _type reflect.Type, needleAddr uintptr, needleType reflect.Type) bool {
	if addr == needleAddr {
		if _type == needleType {
			return true
		}
	}
	return false
}
