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

// CLIInfo contains information about a specific CLI version (i.e. it's compatibility with Kubernetes versions).
type CLIInfo struct {
	// Ref is the reference name of the CLI.
	Ref string `json:"ref,omitempty"`
	// Stream is the stream name of the CLI.
	Stream string `json:"stream,omitempty"`
	// Version is the version of the CLI.
	Version string `json:"version,omitempty"`
	// Kubernetes contains all compatible Kubernetes versions.
	Kubernetes []string `json:"kubernetes,omitempty"`
}

// JSONPath returns the S3 JSON path for this object.
func (c CLIInfo) JSONPath() string {
	return path.Join(
		constants.CDNAPIPrefix,
		"ref", c.Ref,
		"stream", c.Stream,
		c.Version,
		"cli",
		"info.json",
	)
}

// URL returns the URL to the JSON file for this object.
func (c CLIInfo) URL() (string, error) {
	url, err := url.Parse(constants.CDNRepositoryURL)
	if err != nil {
		return "", fmt.Errorf("parsing CDN URL: %w", err)
	}
	url.Path = c.JSONPath()
	return url.String(), nil
}

// ValidateRequest validates the request parameters of the list.
// The Kubernetes slice must be empty.
func (c CLIInfo) ValidateRequest() error {
	var retErr error
	if err := ValidateRef(c.Ref); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if err := ValidateStream(c.Ref, c.Stream); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if !semver.IsValid(c.Version) {
		retErr = errors.Join(retErr, fmt.Errorf("version %q is not a valid semver", c.Version))
	}
	if len(c.Kubernetes) != 0 {
		retErr = errors.Join(retErr, errors.New("Kubernetes slice must be empty for request"))
	}

	return retErr
}

// Validate checks if the CLI info is valid.
func (c CLIInfo) Validate() error {
	var retErr error
	if err := ValidateRef(c.Ref); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if err := ValidateStream(c.Ref, c.Stream); err != nil {
		retErr = errors.Join(retErr, err)
	}
	if !semver.IsValid(c.Version) {
		retErr = errors.Join(retErr, fmt.Errorf("version %q is not a valid semver", c.Version))
	}
	if len(c.Kubernetes) == 0 {
		retErr = errors.Join(retErr, errors.New("Kubernetes slice must not be empty"))
	}
	for _, k := range c.Kubernetes {
		if !semver.IsValid(k) {
			retErr = errors.Join(retErr, fmt.Errorf("Kubernetes version %q is not a valid semver", k))
		}
	}

	return retErr
}
