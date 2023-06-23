/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package instancetypes

// AWSSupportedInstanceFamilies is derived from:
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enable-nitrotpm-prerequisites.html (Last updated: October 20th, 2022).
var AWSSupportedInstanceFamilies = []string{
	"C5",
	"C5a",
	"C5ad",
	"C5d",
	"C5n",
	"C6i",
	"D3",
	"D3en",
	"G4dn",
	"G5",
	"Hpc6a",
	"I3en",
	"I4i",
	"Inf1",
	"M5",
	"M5a",
	"M5ad",
	"M5d",
	"M5dn",
	"M5n",
	"M5zn",
	"M6a",
	"M6i",
	"R5",
	"R5a",
	"R5ad",
	"R5b",
	"R5d",
	"R5dn",
	"R5n",
	"R6i",
	"U-3tb1",
	"U-6tb1",
	"U-9tb1",
	"U-12tb1",
	"X2idn",
	"X2iedn",
	"X2iezn",
	"z1d",
}

// AWSSNPSupportedInstanceFamilies is derived from:
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/snp-requirements.html
var AWSSNPSupportedInstanceFamilies = []string{
	"C6a",
	"M6a",
	"R6a",
}
