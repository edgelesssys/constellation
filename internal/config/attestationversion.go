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
	var rawUnmarshal string
	if err := json.Unmarshal(data, &rawUnmarshal); err != nil {
		return fmt.Errorf("raw unmarshal: %w", err)
	}
	return v.parseRawUnmarshal(rawUnmarshal)
}

func (v *AttestationVersion) parseRawUnmarshal(str string) error {
	if strings.HasPrefix(str, "0") {
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
