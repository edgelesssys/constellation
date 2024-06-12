/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/encoding"
)

type versionValue interface {
	encoding.HexBytes | uint8 | uint16
}

func placeholderVersionValue[T versionValue]() T {
	var placeholder T
	return placeholder
}

// NewLatestPlaceholderVersion returns the latest version with a placeholder version value.
func NewLatestPlaceholderVersion[T versionValue]() AttestationVersion[T] {
	return AttestationVersion[T]{
		Value:      placeholderVersionValue[T](),
		WantLatest: true,
	}
}

// AttestationVersion holds version information.
type AttestationVersion[T versionValue] struct {
	Value      T
	WantLatest bool
}

// MarshalYAML implements a custom marshaller to write "latest" as the type's value, if set.
func (v AttestationVersion[T]) MarshalYAML() (any, error) {
	if v.WantLatest {
		return "latest", nil
	}
	return v.Value, nil
}

// UnmarshalYAML implements a custom unmarshaller to resolve "latest" values.
func (v *AttestationVersion[T]) UnmarshalYAML(unmarshal func(any) error) error {
	return v.unmarshal(unmarshal)
}

// MarshalJSON implements a custom marshaller to write "latest" as the type's value, if set.
func (v AttestationVersion[T]) MarshalJSON() ([]byte, error) {
	if v.WantLatest {
		return json.Marshal("latest")
	}
	return json.Marshal(v.Value)
}

// UnmarshalJSON implements a custom unmarshaller to resolve "latest" values.
func (v *AttestationVersion[T]) UnmarshalJSON(data []byte) (err error) {
	return v.unmarshal(func(a any) error {
		return json.Unmarshal(data, a)
	})
}

// unmarshal takes care of unmarshalling the value from YAML or JSON.
func (v *AttestationVersion[T]) unmarshal(unmarshal func(any) error) error {
	// Start by trying to unmarshal to the distinct type
	var distinctType T
	if err := unmarshal(&distinctType); err == nil {
		v.Value = distinctType
		return nil
	}

	var unmarshalString string
	if err := unmarshal(&unmarshalString); err != nil {
		return fmt.Errorf("failed unmarshalling to %T or string: %w", distinctType, err)
	}

	if strings.ToLower(unmarshalString) == "latest" {
		v.WantLatest = true
		v.Value = placeholderVersionValue[T]()
		return nil
	}

	return fmt.Errorf("failed unmarshalling to %T or string: invalid value: %s", distinctType, unmarshalString)
}
