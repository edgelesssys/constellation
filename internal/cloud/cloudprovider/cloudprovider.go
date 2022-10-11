/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudprovider

import (
	"encoding/json"
	"strings"
)

//go:generate stringer -type=Provider

// Provider is cloud provider used by the CLI.
type Provider uint32

const (
	Unknown Provider = iota
	AWS
	Azure
	GCP
	QEMU
)

// MarshalJSON marshals the Provider to JSON string.
func (p Provider) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON unmarshals the Provider from JSON string.
func (p *Provider) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = FromString(s)
	return nil
}

// FromString returns cloud provider from string.
func FromString(s string) Provider {
	s = strings.ToLower(s)
	switch s {
	case "aws":
		return AWS
	case "azure":
		return Azure
	case "gcp":
		return GCP
	case "qemu":
		return QEMU
	default:
		return Unknown
	}
}
