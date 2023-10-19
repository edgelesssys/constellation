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
// To find the path to the field that failed validation, it traverses the
// top level struct recursively until it finds a field that matches the
// reference to the field that failed validation.
func NewValidationError(topLevelStruct any, field any, errMsg error) *ValidationError {
	path, err := getDocumentPath(topLevelStruct, field)
	if err != nil {
		panic(fmt.Sprintf("cannot find path to field: %v", err))
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
	needleAddr := reflect.ValueOf(field).Elem().UnsafeAddr()
	needleType := reflect.TypeOf(field)

	var haystackAddr uintptr
	switch reflect.TypeOf(doc).Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Interface:
		haystackAddr = reflect.ValueOf(doc).Elem().UnsafeAddr()
	default:
		haystackAddr = reflect.ValueOf(doc).UnsafeAddr()
	}
	haystackType := reflect.TypeOf(doc)

	// traverse the top level struct (i.e. the "haystack") until addr (i.e. the "needle") is found
	return traverse(haystackAddr, haystackType, needleAddr, needleType, []string{})
}

// traverse reverses haystack recursively until it finds a field that matches
// the reference in needle.
//
// If it traverses a level down, it
// appends the name of the struct tag of the field to path.
//
// When a field matches the reference to the given field, it returns the
// path to the field, joined with ".".
func traverse(haystackAddr uintptr, haystackType reflect.Type, needleAddr uintptr, needleType reflect.Type, path []string) (string, error) {
	// recursion anchor: doc is the field we are looking for.
	// Join the path and return. Since the first value of a struct has
	// the same address as the struct itself, we need to check the type as well.
	if haystackAddr == needleAddr && haystackType == needleType {
		return strings.Join(path, "."), nil
	}

	kind := haystackType.Kind()
	switch kind {
	case reflect.Pointer, reflect.UnsafePointer:
		// Dereference pointer and continue.
		return traverse(haystackVal.Elem(), needleAddr, needleType, path)
	case reflect.Struct:
		// Traverse all visible struct fields.
		for _, field := range reflect.VisibleFields(reflect.TypeOf(haystack)) {
			// skip unexported fields
			if field.IsExported() {
				// When a field is not the needle and cannot be traversed further,
				// a errCannotTraverse is returned. Therefore, we only want to handle
				// the case where the field is the needle.
				if path, err := traverse(field, needleAddr, needleType, appendByStructTag(path, field)); err == nil {
					return path, nil
				}
			}
		}
	case reflect.Slice, reflect.Array:
		// Traverse slice / Array elements
		for i := 0; i < haystackVal.Len(); i++ {
			// see struct case
			if path, err := traverse(haystackVal.Index(i), needleAddr, needleType, append(path, fmt.Sprintf("%d", i))); err == nil {
				return path, nil
			}
		}
	case reflect.Map:
		// Traverse map elements
		for _, key := range haystackVal.MapKeys() {
			// see struct case
			if path, err := traverse(haystackVal.MapIndex(key), needleAddr, needleType, append(path, key.String())); err == nil {
				return path, nil
			}
		}
	}
	// Primitive type, but not the value we are looking for.
	// Return a
	return "", errCannotTraverse
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
