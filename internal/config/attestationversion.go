/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package config

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const placeholderVersionValue = 0

// NewLatestPlaceholderVersion returns the latest version with a placeholder version value.
func NewLatestPlaceholderVersion() AttestationVersion {
	return AttestationVersion{
		Value:      placeholderVersionValue,
		WantLatest: true,
	}
}

// AttestationVersion is a type that represents a version of a SNP.
type AttestationVersion struct {
	Value      uint8
	WantLatest bool
}

// MarshalYAML implements a custom marshaller to resolve "latest" values.
func (v AttestationVersion) MarshalYAML() (any, error) {
	if v.WantLatest {
		return "latest", nil
	}
	return v.Value, nil
}

// UnmarshalYAML implements a custom unmarshaller to resolve "atest" values.
func (v *AttestationVersion) UnmarshalYAML(unmarshal func(any) error) error {
	var rawUnmarshal string
	if err := unmarshal(&rawUnmarshal); err != nil {
		return fmt.Errorf("raw unmarshal: %w", err)
	}

	return v.parseRawUnmarshal(rawUnmarshal)
}

// MarshalJSON implements a custom marshaller to resolve "latest" values.
func (v AttestationVersion) MarshalJSON() ([]byte, error) {
	if v.WantLatest {
		return json.Marshal("latest")
	}
	return json.Marshal(v.Value)
}

// UnmarshalJSON implements a custom unmarshaller to resolve "latest" values.
func (v *AttestationVersion) UnmarshalJSON(data []byte) (err error) {
	// JSON has two distinct ways to represent numbers and strings.
	// This means we cannot simply unmarshal to string, like with YAML.
	// Unmarshalling to `any` causes Go to unmarshal numbers to float64.
	// Therefore, try to unmarshal to string, and then to int, instead of using type assertions.
	var unmarshalString string
	if err := json.Unmarshal(data, &unmarshalString); err != nil {
		var unmarshalInt int64
		if err := json.Unmarshal(data, &unmarshalInt); err != nil {
			return fmt.Errorf("unable to unmarshal to string or int: %w", err)
		}
		unmarshalString = strconv.FormatInt(unmarshalInt, 10)
	}

	return v.parseRawUnmarshal(unmarshalString)
}

func (v *AttestationVersion) parseRawUnmarshal(str string) error {
	if strings.HasPrefix(str, "0") && len(str) != 1 {
		return fmt.Errorf("no format with prefixed 0 (octal, hexadecimal) allowed: %s", str)
	}
	if strings.ToLower(str) == "latest" {
		v.WantLatest = true
		v.Value = placeholderVersionValue
	} else {
		ui, err := strconv.ParseUint(str, 10, 8)
		if err != nil {
			return fmt.Errorf("invalid version value: %s", str)
		}
		if ui > math.MaxUint8 {
			return fmt.Errorf("integer value is out ouf uint8 range: %d", ui)
		}
		v.Value = uint8(ui)
	}
	return nil
}
