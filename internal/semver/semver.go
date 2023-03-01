/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package semver provides functionality to parse and process semantic versions, as they are used in multiple components of Constellation.
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
	Major int
	Minor int
	Patch int
}

// NewSemver returns a Version from a string.
func NewSemver(version string) (Semver, error) {
	// ensure that semver has "v" prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	if !semver.IsValid(version) {
		return Semver{}, fmt.Errorf("invalid semver: %s", version)
	}

	version = semver.Canonical(version)

	var major, minor, patch int
	_, err := fmt.Sscanf(version, "v%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return Semver{}, fmt.Errorf("parsing semver parts: %w", err)
	}

	return Semver{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// String returns the string representation of the version.
func (v Semver) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares two versions. It relies on the semver.Compare function internally.
func (v Semver) Compare(other Semver) int {
	return semver.Compare(v.String(), other.String())
}

// IsUpgradeTo returns if a version is an upgrade to another version.
// It checks if the version of v is greater than the version of other and allows a drift of at most one minor version.
func (v Semver) IsUpgradeTo(other Semver) bool {
	return v.Compare(other) > 0 && v.Major == other.Major && v.Minor-other.Minor <= 1
}

// CompatibleWithBinary returns if a version is compatible version of the current built binary.
// It checks if the version of the binary is equal or greater than the current version and allows a drift of at most one minor version.
func (v Semver) CompatibleWithBinary() bool {
	binaryVersion, err := NewSemver(constants.VersionInfo())
	if err != nil {
		return false
	}

	return v.Compare(binaryVersion) == 0 || binaryVersion.IsUpgradeTo(v)
}

// NextMinor returns the next minor version in the format "vMAJOR.MINOR".
func (v Semver) NextMinor() string {
	return fmt.Sprintf("v%d.%d", v.Major, v.Minor+1)
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

	version, err := NewSemver(s)
	if err != nil {
		return err
	}

	*v = version
	return nil
}
