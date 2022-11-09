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
	// Unknown is the default value for VMType and should not be used.
	Unknown VMType = iota
	// AzureCVM is an Azure Confidential Virtual Machine (CVM).
	AzureCVM
	// AzureTrustedLaunch is an Azure Trusted Launch VM.
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
