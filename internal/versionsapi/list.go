/*
Copyright (c) Edgeless Systems GmbH

SPDX-License-Identifier: AGPL-3.0-only
*/

package versionsapi

import (
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/edgelesssys/constellation/v2/internal/constants"
	"golang.org/x/mod/semver"
)

// List represents a list of versions for a kind of resource.
// It has a granularity of either "major" or "minor".
//
// For example, a List with granularity "major" could contain
// the base version "v1" and a list of minor versions "v1.0", "v1.1", "v1.2" etc.
// A List with granularity "minor" could contain the base version
// "v1.0" and a list of patch versions "v1.0.0", "v1.0.1", "v1.0.2" etc.
type List struct {
	// Ref is the branch name the list belongs to.
	Ref string `json:"ref,omitempty"`
	// Stream is the update stream of the list.
	Stream string `json:"stream,omitempty"`
	// Granularity is the granularity of the base version of this list.
	// It can be either "major" or "minor".
	Granularity Granularity `json:"granularity,omitempty"`
	// Base is the base version of the list.
	// Every version in the list is a finer-grained version of this base version.
	Base string `json:"base,omitempty"`
	// Kind is the kind of resource this list is for.
	Kind VersionKind `json:"kind,omitempty"`
	// Versions is a list of all versions in this list.
	Versions []string `json:"versions,omitempty"`
}

// JSONPath returns the S3 JSON path for this object.
func (l List) JSONPath() string {
	return path.Join(
		constants.CDNAPIPrefix,
		"ref", l.Ref,
		"stream", l.Stream,
		"versions", l.Granularity.String(), l.Base,
		l.Kind.String()+".json",
	)
}

// URL returns the URL for this object.
func (l List) URL() (string, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = l.JSONPath()
	return url.String(), nil
}

// ValidateRequest validates the request parameters of the list.
// The versions field must be empty.
func (l List) ValidateRequest() error {
	var retErr error
	if err := ValidateRef(l.Ref); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if err := ValidateStream(l.Ref, l.Stream); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if l.Granularity != GranularityMajor && l.Granularity != GranularityMinor {
		retErr = errors.Join(retErr, fmt.Errorf("granularity %q is not supported", l.Granularity))
	}
	if l.Kind != VersionKindImage {
		retErr = errors.Join(retErr, fmt.Errorf("kind %q is not supported", l.Kind))
	}
	if !semver.IsValid(l.Base) {
		retErr = errors.Join(retErr, fmt.Errorf("base version %q is not a valid semantic version", l.Base))
	}
	var normalizeFunc func(string) string
	switch l.Granularity {
	case GranularityMajor:
		normalizeFunc = semver.Major
	case GranularityMinor:
		normalizeFunc = semver.MajorMinor
	default:
		normalizeFunc = func(s string) string { return s }
	}
	if normalizeFunc(l.Base) != l.Base {
		retErr = errors.Join(retErr, fmt.Errorf("base version %q does not match granularity %q", l.Base, l.Granularity))
	}
	if len(l.Versions) != 0 {
		retErr = errors.Join(retErr, fmt.Errorf("versions must be empty for request"))
	}
	return retErr
}

// Validate checks if the list is valid.
// This performs the following checks:
// - The ref is set.
// - The stream is supported.
// - The granularity is "major" or "minor".
// - The kind is supported.
// - The base version is a valid semantic version that matches the granularity.
// - All versions in the list are valid semantic versions that are finer-grained than the base version.
func (l List) Validate() error {
	var retErr error
	if err := ValidateRef(l.Ref); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if err := ValidateStream(l.Ref, l.Stream); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if l.Granularity != GranularityMajor && l.Granularity != GranularityMinor {
		retErr = errors.Join(retErr, fmt.Errorf("granularity %q is not supported", l.Granularity))
	}
	if l.Kind != VersionKindImage {
		retErr = errors.Join(retErr, fmt.Errorf("kind %q is not supported", l.Kind))
	}
	if !semver.IsValid(l.Base) {
		retErr = errors.Join(retErr, fmt.Errorf("base version %q is not a valid semantic version", l.Base))
	}
	var normalizeFunc func(string) string
	switch l.Granularity {
	case GranularityMajor:
		normalizeFunc = semver.Major
	case GranularityMinor:
		normalizeFunc = semver.MajorMinor
	default:
		normalizeFunc = func(s string) string { return s }
	}
	if normalizeFunc(l.Base) != l.Base {
		retErr = errors.Join(retErr, fmt.Errorf("base version %q does not match granularity %q", l.Base, l.Granularity))
	}
	for _, ver := range l.Versions {
		if !semver.IsValid(ver) {
			retErr = errors.Join(retErr, fmt.Errorf("version %q is not a valid semantic version", ver))
		}
		if normalizeFunc(ver) != l.Base || normalizeFunc(ver) == ver {
			retErr = errors.Join(retErr, fmt.Errorf("version %q is not finer-grained than base version %q", ver, l.Base))
		}
	}

	return retErr
}

// Contains returns true if the list contains the given version.
func (l List) Contains(version string) bool {
	for _, v := range l.Versions {
		if v == version {
			return true
		}
	}
	return false
}

// StructuredVersions returns the versions of the list as slice of
// Version structs.
func (l List) StructuredVersions() []Version {
	versions := make([]Version, len(l.Versions))
	for i, v := range l.Versions {
		versions[i] = Version{
			Ref:     l.Ref,
			Stream:  l.Stream,
			Version: v,
			Kind:    l.Kind,
		}
	}
	return versions
}
