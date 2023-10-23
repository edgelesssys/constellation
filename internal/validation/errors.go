package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ValidationError is returned when a document is not valid.
type ValidationError struct {
	Path string
	Err  error
}

/*
newValidationError creates a new ValidationError.

To find the path to the exported field that failed validation, it traverses "doc"
recursively until it finds a field in "doc" that matches the reference to "field".
*/
func newValidationError(doc, field referenceableValue, errMsg error) *ValidationError {
	// traverse the top level struct (i.e. the "haystack") until addr (i.e. the "needle") is found
	path, err := traverse(doc, field, newPathBuilder(doc._type.Name()))
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

/*
traverse reverses "haystack" recursively until it finds a field that matches
the reference saved in "needle", while building a pseudo-JSONPath to the field.

If it traverses a level down, it appends the name of the struct tag
or another entity like array index or map field to path.

When a field matches the reference to the given field, it returns the
path to the field.
*/
func traverse(haystack referenceableValue, needle referenceableValue, path pathBuilder) (string, error) {
	// recursion anchor: doc is the field we are looking for.
	// Join the path and return.
	if foundNeedle(haystack, needle) {
		return path.String(), nil
	}

	kind := haystack._type.Kind()
	switch kind {
	case reflect.Struct:
		// Traverse all visible struct fields.
		for _, field := range reflect.VisibleFields(haystack._type) {
			// skip unexported fields
			if field.IsExported() {
				fieldVal := recPointerDeref(haystack.value.FieldByName(field.Name))
				if isNilPtrOrInvalid(fieldVal) {
					continue
				}
				fieldAddr := haystack.addr + field.Offset
				newHaystack := referenceableValue{
					value: fieldVal,
					addr:  fieldVal.UnsafeAddr(),
					_type: fieldVal.Type(),
				}
				if canTraverse(fieldVal) {
					// When a field is not the needle and cannot be traversed further,
					// a errCannotTraverse is returned. Therefore, we only want to handle
					// the case where the field is the needle.
					if path, err := traverse(newHaystack, needle, path.appendStructField(field)); err == nil {
						return path, nil
					}
				}
				if foundNeedle(referenceableValue{addr: fieldAddr, _type: field.Type}, needle) {
					return path.appendStructField(field).String(), nil
				}
			}
		}
	case reflect.Slice, reflect.Array:
		// Traverse slice / Array elements
		for i := 0; i < haystack.value.Len(); i++ {
			// see struct case
			itemVal := recPointerDeref(haystack.value.Index(i))
			if isNilPtrOrInvalid(itemVal) {
				continue
			}
			newHaystack := referenceableValue{
				value: itemVal,
				addr:  itemVal.UnsafeAddr(),
				_type: itemVal.Type(),
			}
			if canTraverse(itemVal) {
				if path, err := traverse(newHaystack, needle, path.appendArrayIndex(i)); err == nil {
					return path, nil
				}
			}
			if foundNeedle(newHaystack, needle) {
				return path.appendArrayIndex(i).String(), nil
			}
		}
	case reflect.Map:
		// Traverse map elements
		iter := haystack.value.MapRange()
		for iter.Next() {
			// see struct case
			mapKey := iter.Key().String()
			mapVal := recPointerDeref(iter.Value())
			if isNilPtrOrInvalid(mapVal) {
				continue
			}
			if canTraverse(mapVal) {
				newHaystack := referenceableValue{
					value:  mapVal,
					addr:   mapVal.UnsafeAddr(),
					_type:  mapVal.Type(),
					mapKey: mapKey,
				}
				if path, err := traverse(newHaystack, needle, path.appendMapKey(mapKey)); err == nil {
					return path, nil
				}
			}
			// check if reference to map is the needle and the map key matches
			if foundNeedle(referenceableValue{addr: haystack.addr, _type: haystack._type, mapKey: mapKey}, needle) {
				return path.appendMapKey(mapKey).String(), nil
			}
		}
	}

	// Primitive type, but not the value we are looking for.
	return "", errCannotTraverse
}

// referenceableValue is a type that can be passed as any (thus being copied) without losing the reference to the actual value.
type referenceableValue struct {
	value  reflect.Value
	_type  reflect.Type
	mapKey string // special case for map values, which are not addressable
	addr   uintptr
}

// errCannotTraverse is returned when a field cannot be traversed further.
var errCannotTraverse = errors.New("cannot traverse anymore")

// recPointerDeref recursively dereferences pointers and unpacks interfaces until a non-pointer value is found.
func recPointerDeref(val reflect.Value) reflect.Value {
	switch val.Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Interface:
		return recPointerDeref(val.Elem())
	}
	return val
}

// pointerDeref dereferences pointers and unpacks interfaces.
// If the value is not a pointer, it is returned unchanged.
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

// isNilPtrOrInvalid returns true if a value is a nil pointer or if the value is of an invalid kind.
func isNilPtrOrInvalid(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	case reflect.Invalid:
		return true
	}
	return false
}

/*
foundNeedle returns whether the given value is the needle.

It does so by comparing the address and type of the value to the address and type of the needle.
The comparison of types is necessary because the first value of a struct has the same address as the struct itself.
*/
func foundNeedle(haystack, needle referenceableValue) bool {
	return haystack.addr == needle.addr &&
		haystack._type == needle._type &&
		haystack.mapKey == needle.mapKey
}

// pathBuilder is a helper to build a field path.
type pathBuilder struct {
	buf []string // slice can be copied by value when its non-zero, contrary to a strings.Builder
}

// newPathBuilder creates a new pathBuilder from the identifier of a top level document.
func newPathBuilder(topLevelDoc string) pathBuilder {
	return pathBuilder{
		buf: []string{topLevelDoc},
	}
}

// appendStructField appends the JSON / YAML struct tag of a field to the path.
// If no struct tag is present, the field name is used.
func (p pathBuilder) appendStructField(field reflect.StructField) pathBuilder {
	switch {
	case field.Tag.Get("json") != "":
		p.buf = append(p.buf, fmt.Sprintf(".%s", field.Tag.Get("json")))
	case field.Tag.Get("yaml") != "":
		p.buf = append(p.buf, fmt.Sprintf(".%s", field.Tag.Get("yaml")))
	default:
		p.buf = append(p.buf, fmt.Sprintf(".%s", field.Name))
	}
	return p
}

// appendArrayIndex appends the index of an array to the path.
func (p pathBuilder) appendArrayIndex(i int) pathBuilder {
	p.buf = append(p.buf, fmt.Sprintf("[%d]", i))
	return p
}

// appendMapKey appends the key of a map to the path.
func (p pathBuilder) appendMapKey(k string) pathBuilder {
	p.buf = append(p.buf, fmt.Sprintf("[\"%s\"]", k))
	return p
}

// String returns the path.
func (p pathBuilder) String() string {
	// Remove struct tag prefix
	return strings.TrimPrefix(
		strings.Join(p.buf, ""),
		".",
	)
}
