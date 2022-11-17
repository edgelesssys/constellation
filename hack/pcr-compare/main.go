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

// Based on image/measured-boot/precalculate_pcr_<4,9,12>.sh & pcr-stable.json
// This holds hex-encoded PCR values.
type expectedMeasurementsFile struct {
	Measurements expectedMeasurements `json:"measurements"`
}

type expectedMeasurements struct {
	PCR4  string `json:"4"`
	PCR8  string `json:"8"`
	PCR9  string `json:"9"`
	PCR11 string `json:"11"`
	PCR12 string `json:"12"`
	PCR13 string `json:"13"`
}

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

	parsedExpectedMeasurements, err := parseExpectedMeasurements(os.Args[1])
	if err != nil {
		panic(err)
	}

	parsedActualMeasurements, err := parseActualMeasurements(os.Args[2])
	if err != nil {
		panic(err)
	}

	mapExpectedMeasurements := createMapFromStrings(parsedExpectedMeasurements)
	mapActualMeasurements := convertBytesMeasurementsToHexMeasurements(parsedActualMeasurements)

	foundMismatch := compareMeasurements(mapExpectedMeasurements, mapActualMeasurements)

	if foundMismatch {
		os.Exit(1)
	}
}

// parseExpectedMeasurements unmarshals a JSON file containing the expected measurements from the "Build OS image" pipeline.
func parseExpectedMeasurements(filename string) (expectedMeasurements, error) {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return expectedMeasurements{}, err
	}
	var parsedExpectedMeasurementsFile expectedMeasurementsFile
	if err := json.Unmarshal(fileData, &parsedExpectedMeasurementsFile); err != nil {
		return expectedMeasurements{}, err
	}
	return parsedExpectedMeasurementsFile.Measurements, nil
}

// parseActualMeasurements unmarshals a JSON file containing the actual measurements extracted from a running cluster via pcr-reader.
func parseActualMeasurements(filename string) (measurements.M, error) {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return measurements.M{}, err
	}
	var parsedActualMeasurements measurements.WithMetadata
	if err := json.Unmarshal(fileData, &parsedActualMeasurements); err != nil {
		return measurements.M{}, err
	}
	return parsedActualMeasurements.Measurements, nil
}

// createMapFromStrings converts the "Build OS Image" pipeline's PCR values from single string entries to a map.
func createMapFromStrings(measurements expectedMeasurements) map[uint32]string {
	mapMeasurements := make(map[uint32]string)

	// Note: We do not check for PCR 15 here, as the expected measurements will have 0x0 as a value since its uninitialized,
	// but the running cluster will have a value here since it is measured from a running, initialized cluster.
	mapMeasurements[4] = measurements.PCR4
	mapMeasurements[8] = measurements.PCR8
	mapMeasurements[9] = measurements.PCR9
	mapMeasurements[11] = measurements.PCR11
	mapMeasurements[12] = measurements.PCR12
	mapMeasurements[13] = measurements.PCR13

	return mapMeasurements
}

// convertBytesMeasurementsToHexMeasurements converts the base64 SHA-256 digests to hex SHA-256 digests.
func convertBytesMeasurementsToHexMeasurements(measurements measurements.M) map[uint32]string {
	hexMeasurements := make(map[uint32]string)
	for key, value := range measurements {
		hexMeasurements[key] = hex.EncodeToString(value.Expected[:])
	}

	return hexMeasurements
}

// compareMeasurements compares the expected PCRs with the actual PCRs.
func compareMeasurements(mapExpectedMeasurements, mapActualMeasurements map[uint32]string) bool {
	redPrint := color.New(color.FgRed).SprintFunc()
	greenPrint := color.New(color.FgGreen).SprintFunc()

	expectedPCRs := getSortedKeysOfMap(mapExpectedMeasurements)

	var foundMismatch bool
	for _, pcr := range expectedPCRs {
		if _, ok := mapActualMeasurements[pcr]; !ok {
			color.Magenta("Expected PCR %d not found in the calculated measurements.\n", pcr)
			continue
		}

		actualValue := mapActualMeasurements[pcr]
		expectedValue := mapExpectedMeasurements[pcr]

		fmt.Printf("Expected PCR %d: %s\n", pcr, expectedValue)

		if actualValue == expectedValue {
			fmt.Printf("Measured PCR %d: %s\n", pcr, greenPrint(actualValue))
			color.Green("PCR %d matches.\n", pcr)
		} else {
			fmt.Printf("Measured PCR %d: %s\n", pcr, redPrint(actualValue))
			color.Red("PCR %d does not match.\n", pcr)
			foundMismatch = true
		}
	}

	return foundMismatch
}

// getSortedKeysOfMap returns the sorted keys of a map to allow the PCR output to be ordered in the output.
func getSortedKeysOfMap(inputMap map[uint32]string) []uint32 {
	keys := make([]uint32, 0, len(inputMap))
	for singleKey := range inputMap {
		keys = append(keys, singleKey)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	return keys
}
