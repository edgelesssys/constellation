/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

/*
Package semver provides functionality to parse and process semantic versions, as they are used in multiple components of Constellation.

The official [semantic versioning specification] disallows leading "v" prefixes.
However, the Constellation config uses the "v" prefix for versions to make version strings more recognizable.
This package bridges the gap between Go's semver pkg (doesn't allow "v" prefix) and the Constellation config (requires "v" prefix).

[semantic versioning specification]: https://semver.org/
*/
package semver

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/edgelesssys/constellation/v2/internal/compatibility"
	"golang.org/x/mod/semver"
)

// Sort sorts a list of semantic version strings using [ByVersion].
func Sort(list []Semver) {
	sort.Sort(byVersion(list))
}

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

// NewFromInt constructs a new Semver from three integers and prerelease string: MAJOR.MINOR.PATCH-PRERELEASE.
func NewFromInt(major, minor, patch int, prerelease string) Semver {
	return Semver{
		major:      major,
		minor:      minor,
		patch:      patch,
		prerelease: prerelease,
	}
}

// NewSlice returns a slice of Semver from a slice of strings.
func NewSlice(in []string) ([]Semver, error) {
	var out []Semver
	for _, version := range in {
		semVersion, err := New(version)
		if err != nil {
			return nil, fmt.Errorf("parsing version %s: %w", version, err)
		}
		out = append(out, semVersion)
	}

	return out, nil
}

// ToStrings converts a slice of Semver to a slice of strings.
func ToStrings(in []Semver) []string {
	var out []string
	for _, v := range in {
		out = append(out, v.String())
	}

	return out
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
// The result will be 0 if v == w, -1 if v < w, or +1 if v > w.
func (v Semver) Compare(other Semver) int {
	return semver.Compare(v.String(), other.String())
}

// MajorMinorEqual returns if the major and minor version of two versions are equal.
func (v Semver) MajorMinorEqual(other Semver) bool {
	return v.major == other.major && v.minor == other.minor
}

// IsUpgradeTo returns if a version is an upgrade to another version.
// It checks if the version of v is greater than the version of other and allows a drift of at most one minor version.
func (v Semver) IsUpgradeTo(other Semver) error {
	if v.Compare(other) <= 0 {
		return compatibility.NewInvalidUpgradeError(other.String(), v.String(), errors.New("current version newer than or equal to new version"))
	}
	if v.major != other.major {
		return compatibility.NewInvalidUpgradeError(other.String(), v.String(), compatibility.ErrMajorMismatch)
	}

	if v.minor-other.minor > 1 {
		return compatibility.NewInvalidUpgradeError(other.String(), v.String(), compatibility.ErrMinorDrift)
	}

	return nil
}

// NextMinor returns the next minor version in the format "vMAJOR.MINOR+1".
func (v Semver) NextMinor() string {
	return fmt.Sprintf("v%d.%d", v.major, v.minor+1)
}

// MarshalYAML implements the yaml.Marshaller interface.
func (v Semver) MarshalYAML() (any, error) {
	return v.String(), nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (v *Semver) UnmarshalYAML(unmarshal func(any) error) error {
	var raw string
	if err := unmarshal(&raw); err != nil {
		return fmt.Errorf("unmarshalling to string: %w", err)
	}

	version, err := New(raw)
	if err != nil {
		return fmt.Errorf("parsing semantic version: %w", err)
	}

	*v = version

	return nil
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

// byVersion implements [sort.Interface] for sorting semantic version strings.
// Copied from Go's semver pkg with minimal modification.
// https://cs.opensource.google/go/x/mod/+/master:semver/semver.go
type byVersion []Semver

func (vs byVersion) Len() int      { return len(vs) }
func (vs byVersion) Swap(i, j int) { vs[i], vs[j] = vs[j], vs[i] }
func (vs byVersion) Less(i, j int) bool {
	cmp := vs[i].Compare(vs[j])
	if cmp != 0 {
		return cmp < 0
	}

	// if versions are equal, sort by lexicographic order
	return vs[i].String() < vs[j].String()
}
