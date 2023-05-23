package snpversion

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

const defaultVersionValue = math.MaxUint8

// NewLatestDummyVersion returns the latest version with a dummy value (should only be used for testing).
func NewLatestDummyVersion() Version {
	return Version{
		Value:    defaultVersionValue,
		IsLatest: true,
	}
}

// Version is a type that represents a version of a SNP.
type Version struct {
	Value    uint8
	IsLatest bool
}

// UnmarshalYAML implements a custom unmarshaller to resolve "atest" values.
func (v *Version) UnmarshalYAML(unmarshal func(any) error) error {
	var rawUnmarshal any
	if err := unmarshal(&rawUnmarshal); err != nil {
		return fmt.Errorf("raw unmarshal: %w", err)
	}

	return v.parseRawUnmarshal(rawUnmarshal)
}

func (v *Version) parseRawUnmarshal(rawUnmarshal any) error {
	switch s := rawUnmarshal.(type) {
	case string:
		if strings.ToLower(s) == "latest" {
			v.IsLatest = true
		} else {
			return fmt.Errorf("invalid version  value: %s", string(s))
		}
	case float64:
		v.Value = uint8(s)
	default:
		return fmt.Errorf("invalid version value input: %s", s)
	}
	return nil
}

// MarshalYAML implements a custom marshaller to resolve "latest" values.
func (v Version) MarshalYAML() (any, error) {
	if v.IsLatest {
		return "latest", nil
	}
	return v.Value, nil
}

// UnmarshalJSON implements a custom unmarshaller to resolve "latest" values.
func (v *Version) UnmarshalJSON(data []byte) (err error) {
	var rawUnmarshal any
	if err := json.Unmarshal(data, &rawUnmarshal); err != nil {
		return fmt.Errorf("raw unmarshal: %w", err)
	}
	return v.parseRawUnmarshal(rawUnmarshal)
}
