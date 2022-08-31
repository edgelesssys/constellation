/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package vmtype

import "strings"

//go:generate stringer -type=VMType

// VMType describes different vm types we support. Introduced for Azure SNP / Trusted Launch attestation.
type VMType uint32

const (
	Unknown VMType = iota
	AzureCVM
	AzureTrustedLaunch
)

// FromString returns a VMType from a string.
func FromString(s string) VMType {
	s = strings.ToLower(s)
	switch s {
	case "azurecvm":
		return AzureCVM
	case "azuretrustedlaunch":
		return AzureTrustedLaunch
	default:
		return Unknown
	}
}
