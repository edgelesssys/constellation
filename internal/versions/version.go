/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"golang.org/x/mod/semver"
)

// Version represents a semantic version.
type Version struct {
	Major int
	Minor int
	Patch int
}

// NewVersion returns a Version from a string.
func NewVersion(version string) (Version, error) {
	// ensure that semver has "v" prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	if !semver.IsValid(version) {
		return Version{}, fmt.Errorf("invalid semver: %s", version)
	}

	version = semver.Canonical(version)

	var major, minor, patch int
	_, err := fmt.Sscanf(version, "v%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return Version{}, fmt.Errorf("parsing semver parts: %w", err)
	}

	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// String returns the string representation of the version.
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares two versions. It relies on the semver.Compare function internally.
func (v Version) Compare(other Version) int {
	return semver.Compare(v.String(), other.String())
}

// CanUpgradeTo returns if a version can be upgraded to another version.
// It checks if the version of other is greater than the current version and allows a drift of at most one minor version.
func (v Version) CanUpgradeTo(other Version) bool {
	return v.Compare(other) < 0 && other.Minor-v.Minor <= 1
}

// CanUpgradeToBinary returns if a version can upgrade to the version of the current built binary.
// It checks if the version of the binary is greater than the current version and allows a drift of at most one minor version.
func (v Version) CanUpgradeToBinary() bool {
	binaryVersion, err := NewVersion(constants.VersionInfo)
	if err != nil {
		return false
	}

	return v.CanUpgradeTo(binaryVersion)
}

// NextMinor returns the next minor version in the format "vMAJOR.MINOR".
func (v Version) NextMinor() string {
	return fmt.Sprintf("v%d.%d", v.Major, v.Minor+1)
}

// MarshalJSON implements the json.Marshaler interface.
func (v Version) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, v.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (v *Version) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	version, err := NewVersion(s)
	if err != nil {
		return err
	}

	*v = version
	return nil
}
