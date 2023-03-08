/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/edgelesssys/constellation/v2/internal/attestation/measurements"
	"github.com/fatih/color"
)

func main() {
	if len(os.Args) < 3 {
		if len(os.Args) == 0 {
			fmt.Println("Usage:", "pcr-compare", "<expected-measurements> <actual-measurements>")
		} else {
			fmt.Println("Usage:", os.Args[0], "<expected-measurements> <actual-measurements>")
		}
		fmt.Println("<expected-measurements> is supposed to be a JSON file from the 'Build OS image' pipeline.")
		fmt.Println("<actual-measurements> in supposed to be a JSON file with metadata from the PCR reader which is supposed to be verified.")
		os.Exit(1)
	}

	parsedExpectedMeasurements, err := parseMeasurements(os.Args[1])
	if err != nil {
		panic(err)
	}

	parsedActualMeasurements, err := parseMeasurements(os.Args[2])
	if err != nil {
		panic(err)
	}

	// Extract the PCR values we care about.
	strippedActualMeasurements := stripMeasurements(parsedActualMeasurements)
	strippedExpectedMeasurements := stripMeasurements(parsedExpectedMeasurements)

	// Do the check early.
	areEqual := strippedExpectedMeasurements.EqualTo(strippedActualMeasurements)

	// Print values and similarities / differences in addition.
	compareMeasurements(strippedExpectedMeasurements, strippedActualMeasurements)

	if !areEqual {
		os.Exit(1)
	}
}

// parseMeasurements unmarshals a JSON file containing the expected or actual measurements.
func parseMeasurements(filename string) (measurements.M, error) {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return measurements.M{}, err
	}

	// Technically the expected version does not hold metadata, but we can use the same struct as both hold the measurements in `measurements`.
	// This uses the fallback mechanism of the Measurements unmarshaller which accepts strings without the full struct, defaulting to warnOnly = false.
	// warnOnly = false is expected for the expected measurements, so that's fine.
	// We don't verify metadata here, the CLI has to do that.
	var parsedMeasurements measurements.WithMetadata
	if err := json.Unmarshal(fileData, &parsedMeasurements); err != nil {
		return measurements.M{}, err
	}

	return parsedMeasurements.Measurements, nil
}

// stripMeasurements extracts only the measurements we want to compare.
// This excludes PCR 15 since the actual measurements come from an initialized cluster, but the expected measurements are supposed to be from an uninitialized state.
func stripMeasurements(input measurements.M) measurements.M {
	toBeChecked := []uint32{4, 8, 9, 11, 12, 13}

	strippedMeasurements := make(measurements.M, len(toBeChecked))
	for _, pcr := range toBeChecked {
		if _, ok := input[pcr]; ok {
			strippedMeasurements[pcr] = input[pcr]
		}
	}

	return strippedMeasurements
}

// compareMeasurements compares the expected PCRs with the actual PCRs.
func compareMeasurements(expectedMeasurements, actualMeasurements measurements.M) {
	redPrint := color.New(color.FgRed).SprintFunc()
	greenPrint := color.New(color.FgGreen).SprintFunc()

	expectedPCRs := getSortedKeysOfMap(expectedMeasurements)
	for _, pcr := range expectedPCRs {
		if _, ok := actualMeasurements[pcr]; !ok {
			color.Magenta("Expected PCR %d not found in the calculated measurements.\n", pcr)
			continue
		}

		actualValue := actualMeasurements[pcr]
		expectedValue := expectedMeasurements[pcr]

		fmt.Printf("Expected PCR %02d: %s (warnOnly: %t)\n", pcr, hex.EncodeToString(expectedValue.Expected[:]), expectedValue.WarnOnly)

		var foundMismatch bool
		var coloredValue string
		var coloredWarnOnly string
		if bytes.Equal(actualValue.Expected, expectedValue.Expected) {
			coloredValue = greenPrint(hex.EncodeToString(actualValue.Expected[:]))
		} else {
			coloredValue = redPrint(hex.EncodeToString(actualValue.Expected[:]))
			foundMismatch = true
		}

		if actualValue.WarnOnly == expectedValue.WarnOnly {
			coloredWarnOnly = greenPrint(fmt.Sprintf("%t", actualValue.WarnOnly))
		} else {
			coloredWarnOnly = redPrint(fmt.Sprintf("%t", actualValue.WarnOnly))
			foundMismatch = true
		}

		fmt.Printf("Measured PCR %02d: %s (warnOnly: %s)\n", pcr, coloredValue, coloredWarnOnly)
		if !foundMismatch {
			color.Green("PCR %02d matches.\n", pcr)
		} else {
			color.Red("PCR %02d does not match.\n", pcr)
		}
	}
}

// getSortedKeysOfMap returns the sorted keys of a map to allow the PCR output to be ordered in the output.
func getSortedKeysOfMap(inputMap measurements.M) []uint32 {
	keys := make([]uint32, 0, len(inputMap))
	for singleKey := range inputMap {
		keys = append(keys, singleKey)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	return keys
}
