/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package instancetypes

// AzureCVMInstanceTypes are valid Azure CVM instance types.
var AzureCVMInstanceTypes = []string{
	// CVMs (3rd Generation EPYC 7763v processors)
	// DCasv5-series
	"Standard_DC4as_v5",
	"Standard_DC8as_v5",
	"Standard_DC16as_v5",
	"Standard_DC32as_v5",
	"Standard_DC48as_v5",
	"Standard_DC64as_v5",
	"Standard_DC96as_v5",
	// DCadsv5-series
	"Standard_DC4ads_v5",
	"Standard_DC8ads_v5",
	"Standard_DC16ads_v5",
	"Standard_DC32ads_v5",
	"Standard_DC48ads_v5",
	"Standard_DC64ads_v5",
	"Standard_DC96ads_v5",
	// ECasv5-series
	"Standard_EC4as_v5",
	"Standard_EC8as_v5",
	"Standard_EC16as_v5",
	"Standard_EC20as_v5",
	"Standard_EC32as_v5",
	"Standard_EC48as_v5",
	"Standard_EC64as_v5",
	"Standard_EC96as_v5",
	// ECadsv5-series
	"Standard_EC4ads_v5",
	"Standard_EC8ads_v5",
	"Standard_EC16ads_v5",
	"Standard_EC20ads_v5",
	"Standard_EC32ads_v5",
	"Standard_EC48ads_v5",
	"Standard_EC64ads_v5",
	"Standard_EC96ads_v5",
}

// AzureTrustedLaunchInstanceTypes are valid Azure Trusted Launch instance types.
var AzureTrustedLaunchInstanceTypes = []string{
	// Trusted Launch (2nd Generation AMD EPYC 7452 or 3rd Generation EPYC 7763v processors)
	// Dav4-series
	"Standard_D4a_v4",
	"Standard_D8a_v4",
	"Standard_D16a_v4",
	"Standard_D32a_v4",
	"Standard_D48a_v4",
	"Standard_D64a_v4",
	"Standard_D96a_v4",
	// Dasv4-series
	"Standard_D4as_v4",
	"Standard_D8as_v4",
	"Standard_D16as_v4",
	"Standard_D32as_v4",
	"Standard_D48as_v4",
	"Standard_D64as_v4",
	"Standard_D96as_v4",
	// Eav4-series
	"Standard_E4a_v4",
	"Standard_E8a_v4",
	"Standard_E16a_v4",
	"Standard_E32a_v4",
	"Standard_E48a_v4",
	"Standard_E64a_v4",
	"Standard_E96a_v4",
	// Easv4-series
	"Standard_E4as_v4",
	"Standard_E8as_v4",
	"Standard_E16as_v4",
	"Standard_E20as_v4",
	"Standard_E32as_v4",
	"Standard_E48as_v4",
	"Standard_E64as_v4",
	"Standard_E96as_v4",
}
