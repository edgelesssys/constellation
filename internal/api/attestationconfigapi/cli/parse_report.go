/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/api/attestationconfigapi"
)

// ParseSNPReport parses the SNP report and returns the version information.
func ParseSNPReport(reader io.Reader) (attestationconfigapi.AzureSEVSNPVersion, error) {
	mp := parseLaunchTCBSection(reader)
	return parseVersion(mp)
}

func parseLaunchTCBSection(reader io.Reader) map[string]string {
	scanner := bufio.NewScanner(reader)
	parsedValues := make(map[string]string)
	inLaunchTCBSection := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Launch TCB:") {
			inLaunchTCBSection = true
			continue // skip the current line as it is the section header
		}

		// Stop scanning if we have reached the end of the Launch TCB section
		if inLaunchTCBSection && strings.Contains(line, "Signature (DER)") {
			break
		}

		if inLaunchTCBSection {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				parsedValues[key] = value
			}
		}
	}

	return parsedValues
}

func parseVersion(versionMap map[string]string) (res attestationconfigapi.AzureSEVSNPVersion, err error) {
	version := attestationconfigapi.AzureSEVSNPVersion{}
	for key, value := range versionMap {
		val, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return res, fmt.Errorf("could not parse value for key %s: %w", key, err)
		}

		switch key {
		case "Secure Processor bootloader SVN":
			version.Bootloader = uint8(val)
		case "Secure Processor operating system SVN":
			version.TEE = uint8(val)
		case "SEV-SNP firmware SVN":
			version.SNP = uint8(val)
		case "Microcode SVN":
			version.Microcode = uint8(val)
		}
	}
	return version, nil
}
