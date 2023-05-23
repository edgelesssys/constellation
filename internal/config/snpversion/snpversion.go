/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package snpversion

const (
	Bootloader Type = "bootloader" // Bootloader is the version of the Azure SEVSNP bootloader.
	TEE        Type = "tee"        // TEE is the version of the Azure SEVSNP TEE.
	SNP        Type = "snp"        // SNP is the version of the Azure SEVSNP SNP.
	Microcode  Type = "microcode"  // Microcode is the version of the Azure SEVSNP microcode.
)

// Type is the type of the version to be requested.
type Type string

// GetLatest returns the version of the given type.
func GetLatest(t Type) (res Version) {
	res.IsLatest = true
	switch t {
	case Bootloader:
		res.Value = 2
	case TEE:
		res.Value = 0
	case SNP:
		res.Value = 6
	case Microcode:
		res.Value = 93
	default:
		panic("invalid version type")
	}
	return
}
