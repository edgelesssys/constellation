/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package cloudprovider

import "strings"

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

func ToString(p Provider) string {
	switch p {
	case AWS:
		return "aws"
	case Azure:
		return "azure"
	case GCP:
		return "gcp"
	case QEMU:
		return "qemu"
	default:
		return "unknown"
	}
}
