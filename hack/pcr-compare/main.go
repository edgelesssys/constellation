/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
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

	// Do the comparison and print output.
	areEqual := compareMeasurements(strippedExpectedMeasurements, strippedActualMeasurements)

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
	// We don't verify metadata here anyway, the client has to do that.
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
func compareMeasurements(expectedMeasurements, actualMeasurements measurements.M) bool {
	// Do the check early.
	areEqual := expectedMeasurements.EqualTo(actualMeasurements)

	// Print values in addition.
	redPrint := color.New(color.FgRed).SprintFunc()
	greenPrint := color.New(color.FgGreen).SprintFunc()

	expectedPCRs := getSortedKeysOfMap(expectedMeasurements)
	for _, pcr := range expectedPCRs {
		if _, ok := actualMeasurements[pcr]; !ok {
			color.Magenta("Expected PCR %d not found in the calculated measurements.\n", pcr)
			continue
		}

		actualValue := actualMeasurements[pcr].Expected
		expectedValue := expectedMeasurements[pcr].Expected

		fmt.Printf("Expected PCR %d: %s\n", pcr, hex.EncodeToString(expectedValue[:]))

		if actualValue == expectedValue {
			fmt.Printf("Measured PCR %d: %s\n", pcr, greenPrint(hex.EncodeToString(actualValue[:])))
			color.Green("PCR %d matches.\n", pcr)
		} else {
			fmt.Printf("Measured PCR %d: %s\n", pcr, redPrint(hex.EncodeToString(actualValue[:])))
			color.Red("PCR %d does not match.\n", pcr)
		}
	}

	return areEqual
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
