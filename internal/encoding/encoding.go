/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

// Package encoding provides data types and functions for JSON or YAML encoding/decoding.
package encoding

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// HexBytes is a byte slice that is marshalled to and from a hex string.
type HexBytes []byte

// String returns the hex encoded string representation of the byte slice.
func (h HexBytes) String() string {
	return hex.EncodeToString(h)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (h *HexBytes) UnmarshalJSON(data []byte) error {
	var hexString string
	if err := json.Unmarshal(data, &hexString); err != nil {
		return err
	}
	// special case to stay consistent with yaml unmarshaler:
	// on empty string, unmarshal to nil
	if hexString == "" {
		*h = nil
		return nil
	}
	return h.unmarshal(hexString)
}

// MarshalJSON implements the json.Marshaler interface.
func (h HexBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (h *HexBytes) UnmarshalYAML(unmarshal func(any) error) error {
	var hexString string
	if err := unmarshal(&hexString); err != nil {
		// compatibility mode for old state file format:
		// fall back to unmarshalling as a byte slice for backwards compatibility
		var oldHexBytes []byte
		if err := unmarshal(&oldHexBytes); err != nil {
			return fmt.Errorf("unmarshalling hex bytes: %w", err)
		}
		hexString = hex.EncodeToString(oldHexBytes)
	}
	return h.unmarshal(hexString)
}

// MarshalYAML implements the yaml.Marshaler interface.
func (h HexBytes) MarshalYAML() (any, error) {
	return h.String(), nil
}

func (h *HexBytes) unmarshal(hexString string) error {
	bytes, err := hex.DecodeString(hexString)
	if err != nil {
		return fmt.Errorf("decoding hex bytes: %w", err)
	}
	*h = bytes
	return nil
}
