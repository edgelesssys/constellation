/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package compatibility offers helper functions for comparing and filtering versions.
*/
package compatibility

import (
	"errors"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"golang.org/x/mod/semver"
)

var (
	// ErrMajorMismatch signals that the major version of two compared versions don't match.
	ErrMajorMismatch = errors.New("missmatching major version")
	// ErrMinorDrift signals that the minor version of two compared versions are further apart than one.
	ErrMinorDrift = errors.New("version difference larger than one minor version")
	// ErrSemVer signals that a given version does not adhere to the Semver syntax.
	ErrSemVer = errors.New("invalid semantic version")
	// ErrOutdatedCLI signals that the configured version is newer than the CLI. This is not allowed.
	ErrOutdatedCLI = errors.New("target version newer than cli version")
)

// EnsurePrefixV returns the input string prefixed with the letter "v", if the string doesn't already start with that letter.
func EnsurePrefixV(str string) string {
	if strings.HasPrefix(str, "v") {
		return str
	}
	return "v" + str
}

// IsValidUpgrade checks that a and b adhere to a version drift of 1 and b is greater than a.
func IsValidUpgrade(a, b string) error {
	a = EnsurePrefixV(a)
	b = EnsurePrefixV(b)

	if !semver.IsValid(a) || !semver.IsValid(b) {
		return ErrSemVer
	}

	if semver.Compare(a, b) >= 0 {
		return errors.New("current version newer than or equal to new version")
	}

	aMajor, aMinor, err := parseCanonicalSemver(a)
	if err != nil {
		return err
	}
	bMajor, bMinor, err := parseCanonicalSemver(b)
	if err != nil {
		return err
	}

	if aMajor != bMajor {
		return ErrMajorMismatch
	}

	if bMinor-aMinor > 1 {
		return ErrMinorDrift
	}

	return nil
}

// BinaryWith tests that this binarie's version is greater or equal than some target version, but not further away than one minor version.
func BinaryWith(target string) error {
	binaryVersion := EnsurePrefixV(constants.VersionInfo)
	target = EnsurePrefixV(target)
	if !semver.IsValid(binaryVersion) || !semver.IsValid(target) {
		return ErrSemVer
	}
	cliMajor, cliMinor, err := parseCanonicalSemver(binaryVersion)
	if err != nil {
		return err
	}
	targetMajor, targetMinor, err := parseCanonicalSemver(target)
	if err != nil {
		return err
	}

	// Major versions always have to match.
	if cliMajor != targetMajor {
		return ErrMajorMismatch
	}
	if semver.Compare(binaryVersion, target) == -1 {
		return ErrOutdatedCLI
	}
	// Abort if minor version drift between CLI and versionA value is greater than 1.
	if cliMinor-targetMinor > 1 {
		return ErrMinorDrift
	}

	return nil
}

// FilterNewerVersion filters the list of versions to only include versions newer than currentVersion.
func FilterNewerVersion(currentVersion string, newVersions []string) []string {
	currentVersion = EnsurePrefixV(currentVersion)
	var result []string

	for _, image := range newVersions {
		image = EnsurePrefixV(image)
		// check if image is newer than current version
		if semver.Compare(image, currentVersion) <= 0 {
			continue
		}
		result = append(result, image)
	}
	return result
}

// NextMinorVersion returns the next minor version for a given canonical semver.
// The returned format is vMAJOR.MINOR.
func NextMinorVersion(version string) (string, error) {
	major, minor, err := parseCanonicalSemver(EnsurePrefixV(version))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("v%d.%d", major, minor+1), nil
}

func parseCanonicalSemver(version string) (major int, minor int, err error) {
	version = semver.MajorMinor(version) // ensure version is in canonical form (vX.Y.Z)
	if version == "" {
		return 0, 0, fmt.Errorf("invalid semver: '%s'", version)
	}
	_, err = fmt.Sscanf(version, "v%d.%d", &major, &minor)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing version: %w", err)
	}

	return major, minor, nil
}
