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

// InvalidUpgradeError present an invalid upgrade. It wraps the source and destination version for improved debuggability.
type InvalidUpgradeError struct {
	from     string
	to       string
	innerErr error
}

// NewInvalidUpgradeError returns a new InvalidUpgradeError.
func NewInvalidUpgradeError(from string, to string, innerErr error) *InvalidUpgradeError {
	return &InvalidUpgradeError{from: from, to: to, innerErr: innerErr}
}

// Unwrap returns the inner error, which is nil in this case.
func (e *InvalidUpgradeError) Unwrap() error {
	return e.innerErr
}

// Error returns the String representation of this error.
func (e *InvalidUpgradeError) Error() string {
	return fmt.Sprintf("upgrading from %s to %s is not a valid upgrade: %s", e.from, e.to, e.innerErr)
}

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
		return NewInvalidUpgradeError(a, b, ErrSemVer)
	}

	if semver.Compare(a, b) >= 0 {
		return NewInvalidUpgradeError(a, b, errors.New("current version newer than or equal to new version"))
	}

	aMajor, aMinor, err := parseCanonicalSemver(a)
	if err != nil {
		return NewInvalidUpgradeError(a, b, err)
	}
	bMajor, bMinor, err := parseCanonicalSemver(b)
	if err != nil {
		return NewInvalidUpgradeError(a, b, err)
	}

	if aMajor != bMajor {
		return NewInvalidUpgradeError(a, b, ErrMajorMismatch)
	}

	if bMinor-aMinor > 1 {
		return NewInvalidUpgradeError(a, b, ErrMinorDrift)
	}

	return nil
}

// BinaryWith tests that this binarie's version is greater or equal than some target version, but not further away than one minor version.
func BinaryWith(binaryVersion, target string) error {
	binaryVersion = EnsurePrefixV(binaryVersion)
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
	// For images we allow newer patch versions, therefore this only checks the minor version.
	if semver.Compare(semver.MajorMinor(binaryVersion), semver.MajorMinor(target)) == -1 {
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

// PriorMinorVersion returns the prior minor version for a given canonical semver.
// The returned format is vMAJOR.MINOR.
func PriorMinorVersion(version string) (string, error) {
	major, minor, err := parseCanonicalSemver(EnsurePrefixV(version))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("v%d.%d", major, minor-1), nil
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
