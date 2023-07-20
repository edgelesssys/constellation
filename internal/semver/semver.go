/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package semver provides functionality to parse and process semantic versions, as they are used in multiple components of Constellation.

The official semantic versioning specification [1] disallows leading "v" prefixes.
However, the Constellation config uses the "v" prefix for versions to make version strings more recognizable.
This package bridges the gap between Go's semver pkg (doesn't allow "v" prefix) and the Constellation config (requires "v" prefix).

[1] https://semver.org/.
*/
package semver

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"golang.org/x/mod/semver"
)

// Semver represents a semantic version.
type Semver struct {
	major      int
	minor      int
	patch      int
	prerelease string
}

// New returns a Version from a string.
func New(version string) (Semver, error) {
	// ensure that semver has "v" prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	if !semver.IsValid(version) {
		return Semver{}, fmt.Errorf("invalid semver: %s", version)
	}

	version = semver.Canonical(version)

	var major, minor, patch int
	_, pre, _ := strings.Cut(version, "-")
	_, err := fmt.Sscanf(version, "v%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return Semver{}, fmt.Errorf("parsing semver parts: %w", err)
	}

	return Semver{
		major:      major,
		minor:      minor,
		patch:      patch,
		prerelease: pre,
	}, nil
}

// NewFromInt constructs a new Semver from three integers: MAJOR.MINOR.PATCH.
func NewFromInt(major, minor, patch int) Semver {
	return Semver{
		major: major,
		minor: minor,
		patch: patch,
	}
}

// Major returns the major version of the object.
func (v Semver) Major() int {
	return v.major
}

// Minor returns the minor version of the object.
func (v Semver) Minor() int {
	return v.minor
}

// Patch returns the patch version of the object.
func (v Semver) Patch() int {
	return v.patch
}

// Prerelease returns the prerelease section of the object.
func (v Semver) Prerelease() string {
	return v.prerelease
}

// String returns the string representation of the version.
func (v Semver) String() string {
	version := fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
	if v.prerelease != "" {
		return fmt.Sprintf("%s-%s", version, v.prerelease)
	}
	return version
}

// Compare compares two versions. It relies on the semver.Compare function internally.
func (v Semver) Compare(other Semver) int {
	return semver.Compare(v.String(), other.String())
}

// MajorMinorEqual returns if the major and minor version of two versions are equal.
func (v Semver) MajorMinorEqual(other Semver) bool {
	return v.major == other.major && v.minor == other.minor
}

// IsUpgradeTo returns if a version is an upgrade to another version.
// It checks if the version of v is greater than the version of other and allows a drift of at most one minor version.
func (v Semver) IsUpgradeTo(other Semver) bool {
	return v.Compare(other) > 0 && v.major == other.major && v.minor-other.minor <= 1
}

// CompatibleWithBinary returns if a version is compatible version of the current built binary.
// It checks if the version of the binary is equal or greater than the current version and allows a drift of at most one minor version.
func (v Semver) CompatibleWithBinary() bool {
	binaryVersion, err := New(constants.VersionInfo())
	if err != nil {
		return false
	}

	return v.Compare(binaryVersion) == 0 || binaryVersion.IsUpgradeTo(v)
}

// NextMinor returns the next minor version in the format "vm.MINOR".
func (v Semver) NextMinor() string {
	return fmt.Sprintf("v%d.%d", v.major, v.minor+1)
}

// MarshalJSON implements the json.Marshaler interface.
func (v Semver) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, v.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (v *Semver) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	version, err := New(s)
	if err != nil {
		return err
	}

	*v = version
	return nil
}
