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

// Latest is the latest version of a kind of resource.
type Latest struct {
	// Ref is the branch name this latest version belongs to.
	Ref string `json:"ref,omitempty"`
	// Stream is stream name this latest version belongs to.
	Stream string `json:"stream,omitempty"`
	// Kind is the kind of resource this latest version is for.
	Kind VersionKind `json:"kind,omitempty"`
	// Version is the latest version for this ref, stream and kind.
	Version string `json:"version,omitempty"`
}

// JSONPath returns the S3 JSON path for this object.
func (l Latest) JSONPath() string {
	return path.Join(
		constants.CDNAPIPrefix,
		"ref", l.Ref,
		"stream", l.Stream,
		"versions",
		"latest",
		l.Kind.String()+".json",
	)
}

// URL returns the URL for this object.
func (l Latest) URL() (string, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = l.JSONPath()
	return url.String(), nil
}

// Validate checks if this latest version is valid.
func (l Latest) Validate() error {
	var retErr error
	if err := ValidateRef(l.Ref); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if err := ValidateStream(l.Ref, l.Stream); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if l.Kind == VersionKindUnknown {
		retErr = errors.Join(retErr, fmt.Errorf("version of kind %q is not supported", l.Kind))
	}
	if !semver.IsValid(l.Version) {
		retErr = errors.Join(retErr, fmt.Errorf("version %q is not a valid semver", l.Version))
	}

	return retErr
}

// ValidateRequest checks if this latest version beside values that are fetched.
func (l Latest) ValidateRequest() error {
	var retErr error
	if err := ValidateRef(l.Ref); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if err := ValidateStream(l.Ref, l.Stream); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if l.Kind == VersionKindUnknown {
		retErr = errors.Join(retErr, fmt.Errorf("version of kind %q is not supported", l.Kind))
	}
	if l.Version != "" {
		retErr = errors.Join(retErr, fmt.Errorf("version %q must be empty for request", l.Version))
	}
	return retErr
}

// ShortPath returns the short path of the latest version.
func (l Latest) ShortPath() string {
	return shortPath(l.Ref, l.Stream, l.Version)
}
