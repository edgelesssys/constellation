/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"fmt"
	"os"
	"strings"
)

/*
	Generates internal/config/instancetypes/aws.go from a list of supported AWS Nitro families or instance types.
	Derived from: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-types.html
	Date: September 29, 2022
*/

/*
	Supported Instance Families
*/

// AWSInstanceTypesGeneralPurpose contains the list of valid AWS Nitro "General purpose" families.
var AWSInstanceTypesGeneralPurpose = []string{"M5", "M5a", "M5ad", "M5d", "M5dn", "M5n", "M5zn", "M6a", "M6g", "M6gd", "M6i", "M6id", "T3", "T3a", "T4g"}

// AWSInstanceTypesComputeOptimized contains the list of valid AWS Nitro "Compute optimized" families.
var AWSInstanceTypesComputeOptimized = []string{"C5", "C5a", "C5ad", "C5d", "C5n", "C6a", "C6g", "C6gd", "C6gn", "C6i", "C6id", "Hpc6a"}

// AWSInstanceTypesMemoryOptimized contains the list of valid AWS Nitro "Memory optimized" families.
var AWSInstanceTypesMemoryOptimized = []string{"R5", "R5a", "R5ad", "R5b", "R5d", "R5dn", "R5n", "R6a", "R6g", "R6gd", "R6i", "R6id", "X2gd", "X2idn", "X2iedn", "X2iezn", "z1d"}

// AWSIntanceTypesStorageOptimized contains the list of valid AWS Nitro "Storage optimized" families.
var AWSInstanceTypesStorageOptimized = []string{"D3", "D3en", "I3en", "I4i", "Im4gn", "Is4gen"}

// AWSInstanceTypesAcceleratedComputing contains the list of valid AWS Nitro "Accelerated computing" families.
var AWSInstanceTypesAcceleratedComputing = []string{"DL1", "G4", "G4ad", "G5", "G5g", "Inf1", "P4", "VT1"}

/*
	Single Instance Types
	Special supported instances from certain families + supported metal instances
*/

// AWSSingleInstanceTypesMemoryOptimized contains the "single instance type" special cases for valid AWS Nitro "Memory optimized" families.
var AWSSingleInstanceTypesMemoryOptimized = []string{"u-3tb1.56xlarge", "u-6tb1.56xlarge", "u-6tb1.112xlarge", "u-9tb1.112xlarge", "u-12tb1.112xlarge"}

// AWSSingleInstanceTypesAcceleratedComputing contains the "single instance type" special cases for valid AWS Nitro "Accelerated computing" families.
var AWSSingleInstanceTypesAcceleratedComputing = []string{"p3dn.24xlarge"}

// AWSSingleInstanceTypesGeneralPurposeMetal contains the list of valid AWS Nitro "General purpose" metal instances.
var AWSSingleInstanceTypesGeneralPurposeMetal = []string{"m5.metal", "m5d.metal", "m5dn.metal", "m5n.metal", "m5zn.metal", "m6a.metal", "m6g.metal", "m6gd.metal", "m6i.metal", "m6id.metal"}

// AWSSingleInstanceTypesComputeOptimizedMetal contains the list of valid AWS Nitro "Compute optimized" metal instances.
var AWSSingleInstanceTypesComputeOptimizedMetal = []string{"c5.metal", "c5d.metal", "c5n.metal", "c6a.metal", "c6g.metal", "c6gd.metal", "c6i.metal", "c6id.metal"}

// AWSSingleInstanceTypesMemoryOptimizedMetal contains the list of valid AWS Nitro "Memory optimized" metal instances.
var AWSSingleInstanceTypesMemoryOptimizedMetal = []string{"r5.metal", "r5b.metal", "r5d.metal", "r5dn.metal", "r5n.metal", "r6a.metal", "r6g.metal", "r6gd.metal", "r6i.metal", "r6id.metal", "u-6tb1.metal", "u-9tb1.metal", "u-12tb1.metal", "u-18tb1.metal", "u-24tb1.metal", "x2gd.metal", "x2idn.metal", "x2iedn.metal", "x2iezn.metal", "z1d.metal"}

// AWSSingleInstanceTypesStorageOptimizedMetal contains the list of valid AWS Nitro "Storage optimized" metal instances.
var AWSSingleInstanceTypesStorageOptimizedMetal = []string{"i3.metal", "i3en.metal", "i4i.metal"}

// AWSSingleInstanceTypesAcceleratedComputingMetal contains the list of valid AWS Nitro "Accelerated computing" metal instances.
var AWSSingleInstanceTypesAcceleratedComputingMetal = []string{"g4dn.metal", "g5g.metal"}

/*
	Composite groups
	Used for easier iteration during compatibility check
*/

var (
	AWSInstanceFamilyGroupsSupported = [][]string{AWSInstanceTypesGeneralPurpose, AWSInstanceTypesComputeOptimized, AWSInstanceTypesMemoryOptimized, AWSInstanceTypesStorageOptimized, AWSInstanceTypesAcceleratedComputing}
	AWSSingleInstancesSupported      = [][]string{AWSSingleInstanceTypesMemoryOptimized, AWSSingleInstanceTypesAcceleratedComputing, AWSSingleInstanceTypesGeneralPurposeMetal, AWSSingleInstanceTypesComputeOptimizedMetal, AWSSingleInstanceTypesMemoryOptimizedMetal, AWSSingleInstanceTypesStorageOptimizedMetal, AWSSingleInstanceTypesAcceleratedComputingMetal}
	AWSAllInstances                  = append(AWSInstanceFamilyGroupsSupported, AWSSingleInstancesSupported...)
)

func main() {
	allInstances := make([]string, 0)
	for _, family := range AWSAllInstances {
		for _, instance := range family {
			// Filter Graviton based on whether lower-case "g" is present in the instance name but not directly at the beginning (e.g. G5 is not Graviton, but G5g is)
			// Downstream code still has to verify that the instance entered by the user is not Graviton.
			// This can be removed if/once Graviton support is added (same as with any downstream checks).
			family := strings.Split(instance, ".")[0]
			if strings.LastIndex(family, "g") <= 0 {
				allInstances = append(allInstances, instance)
			}
		}
	}

	// If called via "go generate", overwrites itself.
	file, err := os.OpenFile("aws.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Fprintf(file, `/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

//go:generate go run ../../../hack/generate-aws-instancetypes/main.go > aws.go
// Code generated by hack/generate-aws-instancetypes tool. DO NOT EDIT.

package instancetypes

var AWSSupportedInstanceTypesOrFamilies = []string{
	"%s",
}
`, strings.Join(allInstances, "\",\n\t\""))
}
