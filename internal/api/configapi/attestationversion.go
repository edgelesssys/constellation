/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package configapi

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	Bootloader AzureSEVSNPVersionType = "bootloader" // Bootloader is the version of the Azure SEVSNP bootloader.
	TEE        AzureSEVSNPVersionType = "tee"        // TEE is the version of the Azure SEVSNP TEE.
	SNP        AzureSEVSNPVersionType = "snp"        // SNP is the version of the Azure SEVSNP SNP.
	Microcode  AzureSEVSNPVersionType = "microcode"  // Microcode is the version of the Azure SEVSNP microcode.
)

const dummyVersionValue = 0

// NewLatestDummyVersion returns the latest version with a dummy version value.
func NewLatestDummyVersion() AttestationVersion {
	return AttestationVersion{
		Value:    dummyVersionValue,
		IsLatest: true,
	}
}

// AttestationVersion is a type that represents a version of a SNP.
type AttestationVersion struct {
	Value    uint8
	IsLatest bool
}

// MarshalYAML implements a custom marshaller to resolve "latest" values.
func (v AttestationVersion) MarshalYAML() (any, error) {
	if v.IsLatest {
		return "latest", nil
	}
	return v.Value, nil
}

// UnmarshalYAML implements a custom unmarshaller to resolve "atest" values.
func (v *AttestationVersion) UnmarshalYAML(unmarshal func(any) error) error {
	var rawUnmarshal any
	if err := unmarshal(&rawUnmarshal); err != nil {
		return fmt.Errorf("raw unmarshal: %w", err)
	}

	return v.parseRawUnmarshal(rawUnmarshal)
}

// MarshalJSON implements a custom marshaller to resolve "latest" values.
func (v AttestationVersion) MarshalJSON() ([]byte, error) {
	if v.IsLatest {
		return json.Marshal("latest")
	}
	return json.Marshal(v.Value)
}

// UnmarshalJSON implements a custom unmarshaller to resolve "latest" values.
func (v *AttestationVersion) UnmarshalJSON(data []byte) (err error) {
	var rawUnmarshal any
	if err := json.Unmarshal(data, &rawUnmarshal); err != nil {
		return fmt.Errorf("raw unmarshal: %w", err)
	}
	return v.parseRawUnmarshal(rawUnmarshal)
}

func (v *AttestationVersion) parseRawUnmarshal(rawUnmarshal any) error {
	switch s := rawUnmarshal.(type) {
	case string:
		if strings.ToLower(s) == "latest" {
			v.IsLatest = true
		} else {
			return fmt.Errorf("invalid version value: %s", s)
		}
	case float64:
		v.Value = uint8(s)
	case int:
		v.Value = uint8(s)
	default:
		return fmt.Errorf("invalid version value type: %s", s)
	}
	return nil
}
