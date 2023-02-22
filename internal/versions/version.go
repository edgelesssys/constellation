/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versions

import (
	"encoding/json"
	"fmt"

	"golang.org/x/mod/semver"
)

// Version represents a semantic version.
type Version string

// NewVersion returns a Version from a string.
func NewVersion(version string) (Version, error) {
	if !semver.IsValid(version) {
		return Version(""), fmt.Errorf("invalid semver: %s", version)
	}

	return Version(version), nil
}

// String returns the string representation of the Version.
func (v Version) String() string {
	return string(v)
}

// Compare compares two versions. It relies on the semver.Compare function internally.
func (v Version) Compare(other Version) int {
	return semver.Compare(v.String(), other.String())
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
