/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package kubernetes

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes/scheme"
)

// Marshaler is used by all k8s resources that can be marshaled to YAML.
type Marshaler interface {
	Marshal() ([]byte, error)
}

// MarshalK8SResources marshals every field of a struct into a k8s resource YAML.
func MarshalK8SResources(resources any) ([]byte, error) {
	if resources == nil {
		return nil, errors.New("marshal on nil called")
	}
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	var buf bytes.Buffer

	// reflect over struct containing fields that are k8s resources
	value := reflect.ValueOf(resources)
	if value.Kind() != reflect.Ptr && value.Kind() != reflect.Interface {
		return nil, errors.New("marshal on non-pointer called")
	}
	elem := value.Elem()
	if elem.Kind() == reflect.Struct {
		// iterate over all struct fields
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Field(i)
			var inter any
			// check if value can be converted to interface
			if field.CanInterface() {
				inter = field.Addr().Interface()
			} else {
				continue
			}
			// convert field interface to runtime.Object
			obj, ok := inter.(runtime.Object)
			if !ok {
				continue
			}

			if i > 0 {
				// separate YAML documents
				buf.Write([]byte("---\n"))
			}
			// serialize k8s resource
			if err := serializer.Encode(obj, &buf); err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

// UnmarshalK8SResources takes YAML and converts it into a k8s resources struct.
func UnmarshalK8SResources(data []byte, into any) error {
	if into == nil {
		return errors.New("unmarshal on nil called")
	}
	// reflect over struct containing fields that are k8s resources
	value := reflect.ValueOf(into).Elem()
	if value.Kind() != reflect.Struct {
		return errors.New("can only reflect over struct")
	}

	decoder := serializer.NewCodecFactory(scheme.Scheme).UniversalDecoder()
	documents, err := splitYAML(data)
	if err != nil {
		return fmt.Errorf("splitting deployment YAML into multiple documents: %w", err)
	}
	if len(documents) != value.NumField() {
		return fmt.Errorf("expected %v YAML documents, got %v", value.NumField(), len(documents))
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		var inter any
		// check if value can be converted to interface
		if !field.CanInterface() {
			return fmt.Errorf("cannot use struct field %v as interface", i)
		}
		inter = field.Addr().Interface()
		// convert field interface to runtime.Object
		obj, ok := inter.(runtime.Object)
		if !ok {
			return fmt.Errorf("cannot convert struct field %v as k8s runtime object", i)
		}

		// decode YAML document into struct field
		if err := runtime.DecodeInto(decoder, documents[i], obj); err != nil {
			return err
		}
	}

	return nil
}

// MarshalK8SResourcesList marshals every element of a slice into a k8s resource YAML.
func MarshalK8SResourcesList(resources []runtime.Object) ([]byte, error) {
	serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
	var buf bytes.Buffer

	for i, obj := range resources {
		if i > 0 {
			// separate YAML documents
			buf.Write([]byte("---\n"))
		}
		// serialize k8s resource
		if err := serializer.Encode(obj, &buf); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// splitYAML splits a YAML multidoc into a slice of multiple YAML docs.
func splitYAML(resources []byte) ([][]byte, error) {
	dec := yaml.NewDecoder(bytes.NewReader(resources))
	var res [][]byte
	for {
		var value any
		err := dec.Decode(&value)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		valueBytes, err := yaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		res = append(res, valueBytes)
	}
	return res, nil
}
